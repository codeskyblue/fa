package adb

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type ADBConn struct {
	io.ReadWriter
}

func NewADBConn(rw io.ReadWriter) *ADBConn {
	prw := DebugProxyConn{
		R:     bufio.NewReader(rw),
		W:     rw,
		Debug: true}

	return &ADBConn{
		ReadWriter: prw,
	}
}

func (conn *ADBConn) Encode(v []byte) error {
	val := string(v)
	data := fmt.Sprintf("%04x%s", len(val), val)
	_, err := conn.Write([]byte(data))
	return err
}

func (conn *ADBConn) ReadN(n int) (data []byte, err error) {
	buf := make([]byte, n)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return
	}
	return buf, nil
}

func (conn *ADBConn) ReadNString(n int) (data string, err error) {
	bdata, err := conn.ReadN(n)
	return string(bdata), err
}

func (conn *ADBConn) DecodeString() (string, error) {
	hexlen, err := conn.ReadNString(4)
	if err != nil {
		return "", err
	}
	var length int
	_, err = fmt.Sscanf(hexlen, "%04x", &length)
	if err != nil {
		return "", err
	}
	return conn.ReadNString(length)
}

// respCheck check OKAY, or FAIL
func (conn *ADBConn) respCheck() error {
	status, _ := conn.ReadNString(4)
	switch status {
	case _OKAY:
		return nil
	case _FAIL:
		data, err := conn.DecodeString()
		if err != nil {
			return err
		}
		return errors.New(data)
	default:
		return fmt.Errorf("Unexpected response: %s, should be OKAY or FAIL", strconv.Quote(status))
	}
}

type DebugProxyConn struct {
	R     io.Reader
	W     io.Writer
	Debug bool
}

func (px DebugProxyConn) Write(data []byte) (int, error) {
	if px.Debug {
		fmt.Printf("-> %s\n", string(data))
	}
	return px.W.Write(data)
}

func (px DebugProxyConn) Read(data []byte) (int, error) {
	n, err := px.R.Read(data)
	if px.Debug {
		fmt.Printf("<- %s\n", string(data[0:n]))
	}
	return n, err
}
