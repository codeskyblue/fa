package adb

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type Client struct {
	Addr string
}

func NewClient(addr string) *Client {
	return &Client{
		Addr: addr,
	}
}

func (c *Client) roundTrip(data string) (conn net.Conn, rw *ADBConn, err error) {
	conn, err = net.Dial("tcp", c.Addr)
	if err != nil {
		return
	}
	rw = NewADBConn(conn)
	err = rw.Encode([]byte(data))
	return
}

func (c *Client) roundTripSingleResponse(data string) (string, error) {
	conn, rw, err := c.roundTrip(data)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	if err := rw.respCheck(); err != nil {
		return "", err
	}
	return rw.DecodeString()
}

// ServerVersion returns int. 39 means 1.0.39
func (c *Client) ServerVersion() (v int, err error) {
	verstr, err := c.roundTripSingleResponse("host:version")
	if err != nil {
		return
	}
	_, err = fmt.Sscanf(verstr, "%x", &v)
	return
}

type DeviceState string

const (
	StateUnauthorized = DeviceState("unauthorized")
	StateDisconnected = DeviceState("disconnected")
	StateOffline      = DeviceState("offline")
	StateOnline       = DeviceState("device")
)

// ListDevices returns the list of connected devices
func (c *Client) ListDevices() (devs []*Device, err error) {
	lines, err := c.roundTripSingleResponse("host:devices")
	if err != nil {
		return nil, err
	}

	devs = make([]*Device, 0)
	for _, line := range strings.Split(lines, "\n") {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		devs = append(devs, c.Device(DeviceWithSerial(parts[0])))
	}
	return
}

// KillServer tells the server to quit immediately
func (c *Client) KillServer() error {
	conn, rw, err := c.roundTrip("host:kill")
	if err != nil {
		if _, ok := err.(net.Error); ok { // adb is already stopped if connection refused
			return nil
		}
		return err
	}
	defer conn.Close()
	return rw.respCheck()
}

func (c *Client) Device(descriptor DeviceDescriptor) *Device {
	return &Device{
		client:     c,
		descriptor: descriptor,
	}
}

func (c *Client) DeviceWithSerial(serial string) *Device {
	return c.Device(DeviceWithSerial(serial))
}

// Device
type Device struct {
	// serial string // always set
	// State DeviceState

	descriptor DeviceDescriptor
	client     *Client
}

func (d *Device) String() string {
	return d.descriptor.String()
	// return fmt.Sprintf("%s:%v", ad.serial, ad.State)
}

func (d *Device) OpenShell(cmd string) (rwc io.ReadWriteCloser, err error) {
	return
}

func (d *Device) Stat(path string) (info os.FileInfo, err error) {
	req := "host:" + d.descriptor.getTransportDescriptor()
	conn, rw, err := d.client.roundTrip(req)
	if err != nil {
		return
	}
	defer conn.Close()
	if err = rw.respCheck(); err != nil {
		return
	}
	rw.Encode([]byte("sync:"))
	rw.respCheck()
	rw.Write([]byte("LIST"))
	rw.Encode([]byte("/data/local/tmp"))

	// rw.Write([]byte("abcd"))
	rw.DecodeString()
	// rw.Encode([]byte("abcd"))
	return
}

type PropValue string

func (p PropValue) Bool() bool {
	return p == "true"
}

func (ad *Device) Properties() (props map[string]PropValue, err error) {
	return
}
