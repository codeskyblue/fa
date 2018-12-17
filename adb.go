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

// SendPacket data is like "000chost:version"
func (conn *AdbConnection) SendPacket(data string) error {
	pktData := fmt.Sprintf("%04x%s", len(data), data)
	_, err := conn.Write([]byte(pktData))
	return err
}

func (conn *AdbConnection) readN(n int) (v string, err error) {
	buf := make([]byte, n)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return
	}
	return string(buf), nil
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

// RecvPacket receive data like "OKAY00040028"
func (conn *AdbConnection) RecvPacket() (data string, err error) {
	stat, err := conn.readN(4)
	if err != nil {
		return "", err
	}
	switch stat {
	case _OKAY:
		return conn.readString()
	case _FAIL:
		data, err = conn.readString()
		if err != nil {
			return
		}
		err = errors.New(data)
		return
	default:
		return "", fmt.Errorf("Unknown stat: %s", strconv.Quote(stat))
	}
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
	if err := conn.SendPacket("host:version"); err != nil {
		return "", err
	}
	return conn.RecvPacket()
}
