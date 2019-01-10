package adb

import "testing"

func TestTcpUsb(t *testing.T) {
	t.Log("adb connect localhost:9000")
	err := RunAdbServer("12345678")
	if err != nil {
		t.Fatal(err)
	}
}

// func TestSliceBytes(t *testing.T) {
// buf := make([]byte, 0)
// t.Log(buf[0 : math.Max len(buf)-1])
// }
