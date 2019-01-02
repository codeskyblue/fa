package adb

import (
	"net"
	"testing"
)

func TestTcpUsb(t *testing.T) {
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()
	conn, err := lis.Accept()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("remote: %s", conn.RemoteAddr())
	prd := PacketReader{reader: conn}
	prd.readPacket()
}
