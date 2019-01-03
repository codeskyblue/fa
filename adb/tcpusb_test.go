package adb

import (
	"testing"
)

func TestTcpUsb(t *testing.T) {
	err := RunAdbServer("12345678")
	if err != nil {
		t.Fatal(err)
	}
}
