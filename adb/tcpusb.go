package adb

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"log"
	"net"
	"os"
)

func handleConnect(conn net.Conn) {
	defer conn.Close()
	log.Printf("remote: %s", conn.RemoteAddr())
	prd := NewPacketReader(conn)

HANDLE_COMMAND:
	for pkt := range prd.C {
		switch pkt.Command {
		case CNXN:
			version := pkt.swapu32(pkt.Arg0)
			log.Println("Version", version)
			maxPayload := pkt.Arg1
			log.Println("MaxPayload:", maxPayload)
			if maxPayload > 0xFFFF { // UINT16_MAX
				maxPayload = 0xFFFF
			}
			log.Println("MaxPayload:", maxPayload)

			// Ref link
			// https://github.com/openstf/adbkit/blob/master/src/adb/tcpusb/socket.coffee
			token := make([]byte, TOKEN_LENGTH)
			rand.Read(token)
			log.Println("Create challenge", base64.StdEncoding.EncodeToString(token))
			sendData := Packet{
				Command: AUTH,
				Arg0:    AUTH_TOKEN,
				Body:    token,
			}.EncodeToBytes()
			stdoutDumper := hex.Dumper(os.Stdout)
			stdoutDumper.Write(sendData)
			stdoutDumper.Close()
			_, err := conn.Write(sendData)
			if err != nil {
				log.Println("Err:", err)
				return
			}
		case AUTH:
			log.Println("Handle conn")
			break HANDLE_COMMAND
		default:
			log.Printf("Unknown command: %#v", pkt)
			break HANDLE_COMMAND
		}
	}
}

// RunAdbServer listen for a address for command `adb connect`
func RunAdbServer(serial string) error {
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		return err
	}
	defer lis.Close()
	conn, err := lis.Accept()
	if err != nil {
		return err
	}
	handleConnect(conn)
	return nil
}
