package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	shellquote "github.com/kballard/go-shellquote"
)

const (
	_OKAY = "OKAY"
	_FAIL = "FAIL"
)

func adbCommand(serial string, args ...string) *exec.Cmd {
	if debug {
		fmt.Println("+ adb", "-s", serial, strings.Join(args, " "))
	}
	c := exec.Command(adbPath(), args...)
	c.Env = append(os.Environ(), "ANDROID_SERIAL="+serial)
	return c
}

func runCommand(name string, args ...string) (err error) {
	if filepath.Base(name) == name {
		name, err = exec.LookPath(name)
		if err != nil {
			return err
		}
	}
	procAttr := new(os.ProcAttr)
	procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	proc, err := os.StartProcess(name, append([]string{name}, args...), procAttr)
	if err != nil {
		return err
	}
	procState, err := proc.Wait()
	if err != nil {
		return err
	}
	ws, ok := procState.Sys().(syscall.WaitStatus)
	if !ok {
		return errors.New("exit code unknown")
	}
	exitCode := ws.ExitStatus()
	if exitCode == 0 {
		return nil
	}
	return errors.New("exit code " + strconv.Itoa(exitCode))
}

func panicError(e error) {
	if e != nil {
		panic(e)
	}
}

type AdbConnection struct {
	net.Conn
}

func (conn *AdbConnection) WritePacket(data string) error {
	pktData := fmt.Sprintf("%04x%s", len(data), data)
	_, err := conn.Write([]byte(pktData))
	if err != nil {
		return err
	}
	return conn.respCheck()
}

func (conn *AdbConnection) readN(n int) (v string, err error) {
	buf := make([]byte, n)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return
	}
	return string(buf), nil
}

// respCheck check OKAY, or FAIL
func (conn *AdbConnection) respCheck() error {
	status, err := conn.readN(4)
	if err != nil {
		return err
	}
	switch status {
	case _OKAY:
		return nil
	case _FAIL:
		data, err := conn.readString()
		if err != nil {
			return err
		}
		return errors.New(data)
	default:
		return fmt.Errorf("Unexpected response: %s, should be OKAY or FAIL", strconv.Quote(status))
	}
}

func (conn *AdbConnection) readString() (string, error) {
	hexlen, err := conn.readN(4)
	if err != nil {
		return "", err
	}
	var length int
	_, err = fmt.Sscanf(hexlen, "%04x", &length)
	if err != nil {
		return "", err
	}
	return conn.readN(length)
}

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

type AdbDevice struct {
	*AdbClient
	Serial string
}

func (c *AdbClient) DeviceWithSerial(serial string) *AdbDevice {
	return &AdbDevice{
		AdbClient: c,
		Serial:    serial,
	}
}

func (c *AdbDevice) OpenShell(cmd string) (rw io.ReadWriteCloser, err error) {
	conn, err := c.newConnection()
	if err != nil {
		return
	}
	err = conn.WritePacket("host:transport:" + c.Serial)
	if err != nil {
		return
	}
	err = conn.WritePacket("shell:" + cmd) //shellquote.Join(args...)) // + " ; echo :$?")
	if err != nil {
		return
	}
	return conn, nil
}

// OpenCommand accept list of args return combined output reader
func (c *AdbDevice) OpenCommand(args ...string) (reader io.ReadWriteCloser, err error) {
	return c.OpenShell(shellquote.Join(args...))
}

func (c *AdbDevice) RunCommand(args ...string) (exitCode int, err error) {
	// TODO
	reader, err := c.OpenCommand(args...)
	if err != nil {
		return
	}
	defer reader.Close()
	_, err = io.Copy(os.Stdout, reader)
	return
}
