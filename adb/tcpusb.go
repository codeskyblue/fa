// Ref link
// https://github.com/openstf/adbkit/blob/master/src/adb/tcpusb/socket.coffee
package adb

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/qiniu/log"
)

type TransportService struct {
	opened            bool
	sess              *Session
	device            *Device
	localId, remoteId uint32
	transport         *ADBConn
}

func (t *TransportService) handle(pkt Packet) {
	switch pkt.Command {
	// case _OPEN:
	// 	t.handleOpenPacket(pkt)
	case _OKAY:
		t.handleOkayPacket(pkt)
	}
}

// func (t *TransportService) handleOpenPacket(pkt Packet){

// }

func (t *TransportService) handleOkayPacket(pkt Packet) {
	buf := make([]byte, 1024)
	n, err := t.transport.Read(buf)
	if err != nil {
		t.sess.err = errors.Wrap(err, "copy data")
	}
	t.sess.writePacket(_WRTE, t.localId, t.remoteId, buf[0:n])
}

type Session struct {
	conn          net.Conn
	signature     []byte
	err           error
	token         []byte
	version       uint32
	maxPayload    uint32
	remoteAddress string
	services      map[uint32]*TransportService

	mu             sync.Mutex
	tmpLocalIdLock sync.Mutex
	tmpLocalId     uint32
}

func NewSession(conn net.Conn) *Session {
	// generate challenge
	token := make([]byte, TOKEN_LENGTH)
	rand.Read(token)
	log.Println("Create challenge", base64.StdEncoding.EncodeToString(token))

	return &Session{
		conn:          conn,
		token:         token,
		version:       1,
		remoteAddress: conn.RemoteAddr().String(),
		services:      make(map[uint32]*TransportService),
	}
}

func (s *Session) nextLocalId() uint32 {
	s.tmpLocalIdLock.Lock()
	defer s.tmpLocalIdLock.Unlock()
	s.tmpLocalId += 1
	return s.tmpLocalId
}

func (s *Session) writePacket(cmd string, arg0, arg1 uint32, body []byte) error {
	_, err := Packet{
		Command: cmd,
		Arg0:    arg0,
		Arg1:    arg1,
		Body:    body,
	}.WriteTo(s.conn)
	return err
}

func (s *Session) Serve() {
	defer s.conn.Close()
	pr := NewPacketReader(s.conn)

	for pkt := range pr.C {
		switch pkt.Command {
		case _CNXN:
			s.onConnection(pkt)
		case _AUTH:
			s.onAuth(pkt)
		case _OPEN:
			s.onOpen(pkt)
		case _OKAY, _WRTE, _CLSE:
			s.forwardServicePacket(pkt)
		default:
			s.err = errors.New("unknown cmd: " + pkt.Command)
		}
		if s.err != nil {
			log.Printf("unexpect err: %v", s.err)
			break
		}
	}
	log.Println("Session")
}

func (sess *Session) onConnection(pkt Packet) {
	// version := pkt.swapu32(pkt.Arg0)
	sess.version = pkt.Arg0
	log.Printf("Version: %x", pkt.Arg0)
	maxPayload := pkt.Arg1
	log.Println("MaxPayload:", maxPayload)
	if maxPayload > 0xFFFF { // UINT16_MAX
		maxPayload = 0xFFFF
	}
	sess.maxPayload = maxPayload
	log.Println("MaxPayload:", maxPayload)
	sess.err = sess.writePacket(_AUTH, AUTH_TOKEN, 0, sess.token)
	pkt.DumpToStdout()
}

func (sess *Session) authVerified() {
	version := swapUint32(1)
	log.Printf("send version: %x", version)
	connProps := []string{
		"ro.product.name=2014011",
		"ro.product.model=2014011",
		"ro.product.device=HM2014011",
	}
	// connProps = append(connProps, "features=cmd,stat_v2,shell_v2")
	deviceBanner := "device"
	payload := fmt.Sprintf("%s::%s", deviceBanner, strings.Join(connProps, ";"))
	// id := "device::;;\x00"
	sess.err = sess.writePacket(_CNXN, version, sess.maxPayload, []byte(payload))
	Packet{_CNXN, sess.version, sess.maxPayload, []byte(payload)}.DumpToStdout()
}

func (sess *Session) onAuth(pkt Packet) {
	log.Println("Handle AUTH")
	switch pkt.Arg0 {
	case AUTH_SIGNATURE:
		sess.signature = pkt.Body
		// The real logic is
		// If already have rsa_publickey, then verify signature, send CNXN if passed
		// If no rsa pubkey, then send AUTH to request it
		// Check signature again and send CNXN if passed
		log.Printf("Receive signature: %s", base64.StdEncoding.EncodeToString(pkt.Body))
		// sess.err = sess.writePacket(_AUTH, AUTH_TOKEN, 0, sess.token)
		sess.authVerified()
	case AUTH_RSAPUBLICKEY:
		if sess.signature == nil {
			sess.err = errors.New("Public key sent before signature")
			return
		}
		log.Printf("Receive public key: %s", pkt.Body)
		// TODO(ssx): parse public key from body and verify signature
		// pkt.DumpToStdout()
		log.Println("receive RSA PublicKey")
		// pkt.DumpToStdout()
		// send deviceId
		// time.Sleep(10 * time.Second)
		// sess.err = errors.New("retry")
		// adb 1.0.40 will show "failed to authenticate to x.x.x.x:5555"
		// but actually connected.
		// sess.authVerified()
	default:
		sess.err = fmt.Errorf("unknown authentication method: %d", pkt.Arg0)
	}
}

func (sess *Session) onOpen(pkt Packet) {
	remoteId := pkt.Arg0
	localId := sess.nextLocalId()
	if len(pkt.Body) < 2 {
		sess.err = errors.New("empty service name")
		return // Not throw error ?
	}
	name := string(pkt.BodySkipNull())
	log.Infof("Calling #%s, remoteId: %d, localId: %d", name, remoteId, localId)

	// Session service
	device := NewClient("").Device(AnyUsbDevice())
	conn, err := device.OpenTransport()
	if err != nil {
		sess.err = err
		return
	}
	conn.Encode(pkt.BodySkipNull())
	if err := conn.CheckOKAY(); err != nil {
		sess.err = errors.Wrap(err, "session.OPEN")
		return
	}

	sess.writePacket(_OKAY, localId, remoteId, nil)

	service := &TransportService{
		localId:   localId,
		remoteId:  remoteId,
		sess:      sess,
		opened:    false,
		transport: conn,
	}

	sess.mu.Lock()
	sess.services[localId] = service
	sess.mu.Unlock()

	go func() {
		buf := make([]byte, 1024)
		for sess.err == nil {
			n, err := conn.Read(buf)
			if err != nil {
				sess.err = errors.Wrap(err, "copy data")
			}
			sess.writePacket(_WRTE, localId, remoteId, buf[0:n])
		}
	}()
	// sess.writePacket(_OKAY)
	// Packet{}

	pkt.DumpToStdout()
}

func (sess *Session) forwardServicePacket(pkt Packet) {
	sess.mu.Lock()
	service, ok := sess.services[pkt.Arg1] // localId
	sess.mu.Unlock()
	if !ok {
		sess.err = errors.New("transport service already closed")
	}
	service.handle(pkt)
}

// RunAdbServer listen for a address for command `adb connect`
func RunAdbServer(serial string) error {
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		return err
	}
	defer lis.Close()
	return Serve(lis)
}

func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		sess := NewSession(conn)
		go sess.Serve()
	}
}
