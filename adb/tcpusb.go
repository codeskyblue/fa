package adb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
)

const (
	SYNC = "SYNC"
	CXNN = "CXNN"
	OPEN = "OPEN"
	OKAY = "OKAY"
	CLSE = "CLSE"
	WRTE = "WRTE"
	AUTH = "AUTH"
)

func checksum(data []byte) byte {
	sum := byte(0)
	for _, c := range data {
		sum += c
	}
	return sum
}

func xorBytes(a, b []byte) []byte {
	if len(a) != len(b) {
		panic(fmt.Sprintf("xorBytes a:%x b:%x have different size", a, b))
	}
	dst := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		dst[i] = a[i] ^ b[i]
	}
	return dst
}

type AdbServer struct {
	version       int
	mayPayload    int
	authorized    bool
	syncToken     int
	remoteID      int
	services      map[string]string
	remoteAddress string
	token         string
	signature     string
}

type PacketReader struct {
	reader io.Reader
	err    error
}

type errReader struct{}

func (e errReader) Read(p []byte) (int, error) {
	return 0, errors.New("package already read error")
}

func (p *PacketReader) r() io.Reader {
	if p.err != nil { // use p.err to short error checks
		return errReader{}
	}
	return p.reader
}

// Receive packet example
// 00000000  43 4e 58 4e 01 00 00 01  00 00 10 00 23 00 00 00  |CNXN........#...|
// 00000010  3c 0d 00 00 bc b1 a7 b1  68 6f 73 74 3a 3a 66 65  |<.......host::fe|
// 00000020  61 74 75 72 65 73 3d 63  6d 64 2c 73 74 61 74 5f  |atures=cmd,stat_|
// 00000030  76 32 2c 73 68 65 6c 6c  5f 76 32                 |v2,shell_v2|
func (p *PacketReader) readPacket() {
	var (
		command = p.readStringN(4)
		arg0    = p.readN(4)
		arg1    = p.readN(4)
		length  = p.readInt32()
		check   = p.readInt32()
		magic   = p.readN(4)
	)
	if p.err != nil {
		return
		// log.Println("ERR:", p.err)
	}
	if !bytes.Equal(xorBytes([]byte(command), magic), []byte{0xff, 0xff, 0xff, 0xff}) {
		p.err = errors.New("verify magic failed")
		return
	}
	log.Printf("cmd:%s, arg0:%x, arg1:%x, len:%d, check:%d, magic:%x",
		command, arg0, arg1, length, check, magic)
	log.Printf("cmd:%x", []byte(command))
	body := p.readStringN(int(length))
	log.Println(body)
}

func (p *PacketReader) readN(n int) []byte {
	buf := make([]byte, n)
	_, p.err = io.ReadFull(p.r(), buf)
	return buf
}

func (p *PacketReader) readStringN(n int) string {
	return string(p.readN(n))
}

func (p *PacketReader) readInt32() int32 {
	var i int32
	p.err = binary.Read(p.r(), binary.LittleEndian, &i)
	return i
}
