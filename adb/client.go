package adb

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	shellquote "github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
)

type Client struct {
	Addr string
}

func NewClient(addr string) *Client {
	if addr == "" {
		addr = "127.0.0.1:5037"
	}
	return &Client{
		Addr: addr,
	}
}

func (c *Client) dial() (conn *ADBConn, err error) {
	nc, err := net.DialTimeout("tcp", c.Addr, 2*time.Second)
	if err != nil {
		if err = c.StartServer(); err != nil {
			err = errors.Wrap(err, "adb start-server")
			return
		}
		nc, err = net.DialTimeout("tcp", c.Addr, 2*time.Second)
	}
	return NewADBConn(nc), err
}

func (c *Client) roundTrip(data string) (conn *ADBConn, err error) {
	conn, err = c.dial()
	if err != nil {
		return
	}
	if len(data) > 0 {
		err = conn.Encode([]byte(data))
	}
	return
}

func (c *Client) roundTripSingleResponse(data string) (string, error) {
	conn, err := c.roundTrip(data)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	if err := conn.CheckOKAY(); err != nil {
		return "", err
	}
	return conn.DecodeString()
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

func (c *Client) StartServer() (err error) {
	cmd := exec.Command("adb", "start-server")
	return cmd.Run()
}

// KillServer tells the server to quit immediately
func (c *Client) KillServer() error {
	conn, err := c.roundTrip("host:kill")
	if err != nil {
		if _, ok := err.(net.Error); ok { // adb is already stopped if connection refused
			return nil
		}
		return err
	}
	defer conn.Close()
	return conn.CheckOKAY()
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
	descriptor DeviceDescriptor
	client     *Client
}

func (d *Device) String() string {
	return d.descriptor.String()
	// return fmt.Sprintf("%s:%v", ad.serial, ad.State)
}

func (d *Device) Serial() (serial string, err error) {
	return
}

// OpenTransport is a low level function
// Connect to adbd.exe and send <host-prefix>:transport and check OKAY
// conn should be Close after using
func (d *Device) OpenTransport() (conn *ADBConn, err error) {
	req := "host:" + d.descriptor.getTransportDescriptor()
	conn, err = d.client.roundTrip(req)
	if err != nil {
		return
	}
	conn.CheckOKAY()
	if conn.Err() != nil {
		conn.Close()
	}
	return conn, conn.Err()
}

func (d *Device) OpenShell(cmd string) (rwc io.ReadWriteCloser, err error) {
	req := "host:" + d.descriptor.getTransportDescriptor()
	conn, err := d.client.roundTrip(req)
	if err != nil {
		return
	}
	conn.CheckOKAY()
	conn.EncodeString("shell:" + cmd)
	conn.CheckOKAY()
	if conn.Err() != nil {
		conn.Close()
	}
	return conn, conn.Err()
}

func (d *Device) RunCommand(args ...string) (output string, err error) {
	cmd := shellquote.Join(args...)
	rwc, err := d.OpenShell(cmd)
	if err != nil {
		return
	}
	data, err := ioutil.ReadAll(rwc)
	if err != nil {
		return
	}
	return string(data), err
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
	conn, err := d.client.roundTrip(req)
	if err != nil {
		return
	}
	defer conn.Close()
	if err = conn.CheckOKAY(); err != nil {
		return
	}
	conn.EncodeString("sync:")
	conn.CheckOKAY()
	conn.WriteObjects("STAT", uint32(len(path)), path)

	id, err := conn.ReadNString(4)
	if err != nil {
		return
	}
	if id != "STAT" {
		return nil, fmt.Errorf("Invalid status: %q", id)
	}
	adbMode, _ := conn.ReadUint32()
	size, _ := conn.ReadUint32()
	seconds, err := conn.ReadUint32()
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
