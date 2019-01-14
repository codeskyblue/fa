package main

import (
	"fmt"
	"log"
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

// Plugged in: StateDisconnected->StateOffline->StateOnline
// Unplugged:  StateOnline->StateDisconnected
type DeviceState string

const (
	StateInvalid      = ""
	StateDisconnected = "disconnected"
	StateOffline      = "offline"
	StateOnline       = "device"
	StateUnauthorized = "unauthorized"
)

func newDeviceState(s string) DeviceState {
	switch s {
	case "device":
		return StateOnline
	case "offline":
		return StateOffline
	case "disconnected":
		return StateDisconnected
	case "unauthorized":
		return StateUnauthorized
	default:
		return StateInvalid
	}
}

type DeviceStateChangedEvent struct {
	Serial   string
	OldState DeviceState
	NewState DeviceState
}

func (s DeviceStateChangedEvent) String() string {
	return fmt.Sprintf("%s: %s->%s", s.Serial, s.OldState, s.NewState)
}

// CameOnline returns true if this event represents a device coming online.
func (s DeviceStateChangedEvent) CameOnline() bool {
	return s.OldState != StateOnline && s.NewState == StateOnline
}

// WentOffline returns true if this event represents a device going offline.
func (s DeviceStateChangedEvent) WentOffline() bool {
	return s.OldState == StateOnline && s.NewState != StateOnline
}

func (c *AdbClient) Watch() (C chan DeviceStateChangedEvent, err error) {
	C = make(chan DeviceStateChangedEvent, 0)
	conn, err := c.newConnection()
	if err != nil {
		return
	}
	conn.WritePacket("host:track-devices")
	go func() {
		defer close(C)
		var lastKnownStates = make(map[string]DeviceState)
		for {
			line, err := conn.readString()
			if err != nil {
				break
			}
			line = strings.TrimSpace(line)
			log.Println("TRACK", strconv.Quote(line))
			if line == "" {
				continue
			}
			parts := strings.Split(line, "\t")
			if len(parts) != 2 {
				continue
			}
			serial, state := parts[0], newDeviceState(parts[1])

			C <- DeviceStateChangedEvent{
				Serial:   serial,
				OldState: lastKnownStates[serial],
				NewState: state,
			}
			lastKnownStates[serial] = state
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
