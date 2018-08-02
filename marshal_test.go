package cc111x

import (
	"bytes"
	"fmt"
	"math"
	"testing"
)

func TestMarshalUint16(t *testing.T) {
	cases := []struct {
		val uint16
		rep []byte
	}{
		{0x1234, []byte{0x12, 0x34}},
		{0, []byte{0, 0}},
		{math.MaxUint16, []byte{0xFF, 0xFF}},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("marshal16_%d", c.val), func(t *testing.T) {
			rep := marshalUint16(c.val)
			if !bytes.Equal(rep, c.rep) {
				t.Errorf("marshalUint16(%04X) == % X, want % X", c.val, rep, c.rep)
			}
		})
	}
}

func TestMarshalUint32(t *testing.T) {
	cases := []struct {
		val uint32
		rep []byte
	}{
		{0x12345678, []byte{0x12, 0x34, 0x56, 0x78}},
		{0, []byte{0, 0, 0, 0}},
		{math.MaxUint32, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("marshal32_%d", c.val), func(t *testing.T) {
			rep := marshalUint32(c.val)
			if !bytes.Equal(rep, c.rep) {
				t.Errorf("marshalUint32(%08X) == % X, want % X", c.val, rep, c.rep)
			}
		})
	}
}
