// Ref link
// https://github.com/openstf/adbkit/blob/master/src/adb/tcpusb/socket.coffee
package adb

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

type Session struct {
	conn          net.Conn
	signature     []byte
	err           error
	token         []byte
	version       uint32
	maxPayload    uint32
	remoteAddress string
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
	}
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
		default:
			log.Printf("unknown cmd: %s", pkt.Command)
			return
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
	pkt.DumpToStdout()
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
