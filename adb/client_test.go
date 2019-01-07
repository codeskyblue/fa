package adb

import "testing"

var client = NewClient("127.0.0.1:5037")

func TestVersion(t *testing.T) {
	version, err := client.Version()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(version)
}

func TestDevices(t *testing.T) {
	devs, err := client.Devices()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(devs)
}
