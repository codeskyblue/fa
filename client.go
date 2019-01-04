package main

import (
	"net"
	"os/exec"
	"strconv"
	"strings"
)

type AdbClient struct {
	Addr string
}

func NewAdbClient() *AdbClient {
	return &AdbClient{
		Addr: defaultHost + ":" + strconv.Itoa(defaultPort),
	}
}

var DefaultAdbClient = &AdbClient{
	Addr: "127.0.0.1:5037",
}

func (c *AdbClient) newConnection() (conn *AdbConnection, err error) {
	netConn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return nil, err
	}
	return &AdbConnection{netConn}, nil
}

// Version return 4 size string
func (c *AdbClient) Version() (string, error) {
	ver, err := c.rawVersion()
	if err == nil {
		return ver, nil
	}
	exec.Command(adbPath(), "start-server").Run()
	return c.rawVersion()
}

func (c *AdbClient) Watch() (C chan string, err error) {
	C = make(chan string, 0)
	// c.Version()
	conn, err := c.newConnection()
	if err != nil {
		return
	}
	conn.WritePacket("host:track-devices")
	go func() {
		defer close(C)
		for {
			line, err := conn.readString()
			if err != nil {
				break
			}
			line = strings.TrimSpace(line)
			if line != "" {
				C <- line
			}
		}
	}()
	return
}

// Version returns adb server version
func (c *AdbClient) rawVersion() (string, error) {
	conn, err := c.newConnection()
	if err != nil {
		return "", err
	}
	defer conn.Close()
	if err := conn.WritePacket("host:version"); err != nil {
		return "", err
	}
	return conn.readString()
}

func (c *AdbClient) DeviceWithSerial(serial string) *AdbDevice {
	return &AdbDevice{
		AdbClient: c,
		Serial:    serial,
	}
}
