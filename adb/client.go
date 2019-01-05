package adb

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
)

const (
	// _OKAY = "OKAY"
	_FAIL = "FAIL"
)

type ADBEncoder struct {
	wr io.Writer
}

func (e *ADBEncoder) Encode(v []byte) error {
	val := string(v)
	data := fmt.Sprintf("%04x%s", len(val), val)
	_, err := e.wr.Write([]byte(data))
	return err
}

type ADBDecoder struct {
	rd io.Reader
}

func (d *ADBDecoder) ReadN(n int) (data []byte, err error) {
	buf := make([]byte, n)
	_, err = io.ReadFull(d.rd, buf)
	if err != nil {
		return
	}
	return buf, nil
}

func (d *ADBDecoder) ReadNString(n int) (data string, err error) {
	bdata, err := d.ReadN(n)
	return string(bdata), err
}

func (d *ADBDecoder) DecodeString() (string, error) {
	hexlen, err := d.ReadNString(4)
	if err != nil {
		return "", err
	}
	var length int
	_, err = fmt.Sscanf(hexlen, "%04x", &length)
	if err != nil {
		return "", err
	}
	return d.ReadNString(length)
}

// respCheck check OKAY, or FAIL
func (d *ADBDecoder) respCheck() error {
	status, err := d.ReadNString(4)
	if err != nil {
		return err
	}
	switch status {
	case _OKAY:
		return nil
	case _FAIL:
		data, err := d.DecodeString()
		if err != nil {
			return err
		}
		return errors.New(data)
	default:
		return fmt.Errorf("Unexpected response: %s, should be OKAY or FAIL", strconv.Quote(status))
	}
}

func assembleString(data string) []byte {
	pktData := fmt.Sprintf("%04x%s", len(data), data)
	return []byte(pktData)
}

// func assembleBytes(data []byte) []byte {
// }

type Client struct {
	Addr string
}

func NewClient(addr string) *Client {
	return &Client{
		Addr: addr,
	}
}

type DebugProxyConn struct {
	R io.Reader
	W io.Writer
}

func (px DebugProxyConn) Write(data []byte) (int, error) {
	fmt.Printf("-> %s\n", string(data))
	return px.W.Write(data)
}

func (px DebugProxyConn) Read(data []byte) (int, error) {
	n, err := px.R.Read(data)
	fmt.Printf("<- %s\n", string(data[0:n]))
	return n, err
}

// TODO(ssx): test not passed yet.
func (c *Client) Version() (string, error) {
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	dconn := DebugProxyConn{
		R: bufio.NewReader(conn),
		W: conn}

	writer := ADBEncoder{dconn}
	writer.Encode([]byte("host:version"))
	reader := ADBDecoder{dconn}
	if err := reader.respCheck(); err != nil {
		return "", err
	}
	return reader.DecodeString()
}

func (c *Client) DeviceWithSerial(serial string) *ADevice {
	return &ADevice{
		client: c,
		serial: serial,
	}
}

// Device
type ADevice struct {
	client *Client
	serial string
}

func (ad *ADevice) OpenShell(cmd string) (rwc io.ReadWriteCloser, err error) {
	return
}

func (ad *ADevice) Stat(path string) {

}

type PropValue string

func (p PropValue) Bool() bool {
	return p == "true"
}

func (ad *ADevice) Properties() (props map[string]PropValue, err error) {
	return
}
