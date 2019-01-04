package adb

import "testing"

func TestVersion(t *testing.T) {
	version, err := NewClient("127.0.0.1:5037").Version()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(version)
}
