// +build darwin

package main

import (
	"io"
	"os"
	"testing"
)

func TestAdbVersion(t *testing.T) {
	version, err := DefaultAdbClient.Version()
	if err != nil {
		panic(err)
	}
	t.Logf("version: %s", version)
}

func TestAdbShell(t *testing.T) {
	t.Log("Shell Test")
	d := DefaultAdbClient.DeviceWithSerial("0123456789ABCDEF")
	rd, err := d.OpenShell("pwd")
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(os.Stdout, rd)
	// output, exitCode, err := DefaultAdbClient.Shell("pwd")
	// if err != nil {
	// 	t.Log(err)
	// }
	// t.Log(output)
	// t.Log(exitCode)
}
