package adb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeviceProperties(t *testing.T) {
	device := NewClient("").Device(AnyDevice())
	props, err := device.Properties()
	if assert.NoError(t, err) {
		t.Log(props["ro.product.name"])
	}
}
