package adb

import (
	"bytes"
	"encoding/binary"
)

// Packet is a meta for adb connect
type Packet struct {
	Command string
	Arg0    []byte
	Arg1    []byte
	Body    []byte
}

func (pkt *Packet) magic() []byte {
	return xorBytes([]byte(pkt.Command), []byte{0xff, 0xff, 0xff, 0xff})
}

func (pkt *Packet) EncodeToBytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 24+len(pkt.Body)))
	binary.Write(buf, binary.LittleEndian, []byte(pkt.Command))
	binary.Write(buf, binary.LittleEndian, pkt.Arg0)
	binary.Write(buf, binary.LittleEndian, pkt.Arg1)
	binary.Write(buf, binary.LittleEndian, uint32(len(pkt.Body)))
	binary.Write(buf, binary.LittleEndian, pkt.magic())
	buf.Write(pkt.Body)
	return buf.Bytes()
}
