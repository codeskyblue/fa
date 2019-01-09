package adb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var client = NewClient("127.0.0.1:5037")

func TestServerVersion(t *testing.T) {
	version, err := client.ServerVersion()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(version)
}

func TestDevices(t *testing.T) {
	devs, err := client.ListDevices()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(devs)
}

func TestKillServer(t *testing.T) {
	err := client.KillServer()
	assert.NoError(t, err)
	time.Sleep(2 * time.Second)
}

func TestDeviceStat(t *testing.T) {
	device := client.Device(AnyUsbDevice())
	info, err := device.Stat("/data/local/tmp/minicap")
	if !assert.NoError(t, err) {
		return
	}
	t.Log(info.Name(), info.Mode().String(), info.Size(), info.ModTime())
}

func TestDeviceRunCommand(t *testing.T) {
	device := client.Device(AnyUsbDevice())
	output, err := device.RunCommand("pwd")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "/\n", output)
}
