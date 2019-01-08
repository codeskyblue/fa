package adb

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
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

type adbFileInfo struct {
	name  string
	mode  os.FileMode
	size  uint32
	mtime time.Time
}

func (f *adbFileInfo) Name() string {
	return f.name
}

func (f *adbFileInfo) Size() int64 {
	return int64(f.size)
}
func (f *adbFileInfo) Mode() os.FileMode {
	return f.mode
}

func (f *adbFileInfo) ModTime() time.Time {
	return f.mtime
}

func (f *adbFileInfo) IsDir() bool {
	return f.mode.IsDir()
}

func (f *adbFileInfo) Sys() interface{} {
	return nil
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
	rw.EncodeString("sync:")
	rw.respCheck()
	rw.WriteObjects("STAT", uint32(len(path)), path)

	id, err := rw.ReadNString(4)
	if err != nil {
		return
	}
	if id != "STAT" {
		return nil, fmt.Errorf("Invalid status: %q", id)
	}
	adbMode, _ := rw.ReadUint32()
	size, _ := rw.ReadUint32()
	seconds, err := rw.ReadUint32()
	if err != nil {
		return nil, err
	}
	return &adbFileInfo{
		name:  path,
		size:  size,
		mtime: time.Unix(int64(seconds), 0).Local(),
		mode:  fileModeFromAdb(adbMode),
	}, nil
}

type PropValue string

func (p PropValue) Bool() bool {
	return p == "true"
}

func (ad *Device) Properties() (props map[string]PropValue, err error) {
	return
}
