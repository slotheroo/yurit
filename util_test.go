// Copyright 2015, David Howden
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yurit

import (
	"bytes"
	"testing"
)

func TestCheckLen(t *testing.T) {
	tests := []struct {
		bytes      []byte
		check      uint
		makesError bool
	}{
		{nil, 0, false},
		{[]byte{}, 0, false},
		{[]byte{}, 1, true},
		{[]byte{0x01, 0x02, 0x03}, 2, false},
		{[]byte{0x01, 0x02, 0x03}, 3, false},
		{[]byte{0x01, 0x02, 0x03}, 4, true},
	}

	for ii, tt := range tests {
		err := checkLen(tt.bytes, tt.check)
		if (err != nil) != tt.makesError {
			t.Errorf("[%d] checkLen(%v, %v) expected error: %v - Error was: %v", ii, tt.bytes, tt.check, tt.makesError, err)
		}
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		strings []string
		str     string
		output  bool
	}{
		{nil, "test", false},
		{[]string{}, "test", false},
		{[]string{"no"}, "test", false},
		{[]string{"test"}, "test", true},
		{[]string{"no", "no", "not"}, "test", false},
		{[]string{"test", "no", "not"}, "test", true},
		{[]string{"no", "test", "not"}, "test", true},
		{[]string{"no", "no", "test"}, "test", true},
	}

	for ii, tt := range tests {
		got := containsString(tt.strings, tt.str)
		if got != tt.output {
			t.Errorf("[%d] containsString(%v, %v) = %v, expected %v", ii, tt.strings, tt.str, got, tt.output)
		}
	}
}

func TestGetBit(t *testing.T) {
	all := byte(0xFF)
	none := byte(0x00)
	for i := uint(0); i < 9; i++ {
		got := getBit(all, i)
		if !got && i < 8 {
			t.Errorf("getBit(%v, %v) = %v, expected %v", all, i, got, true)
		} else if got && i == 8 {
			t.Errorf("getBit(%v, %v) = %v, expected %v", all, i, got, false)
		}
		got = getBit(none, i)
		if got {
			t.Errorf("getBit(%v, %v) = %v, expected %v", none, i, got, false)
		}
	}
}

func TestGet7BitChunkedInt(t *testing.T) {
	tests := []struct {
		input  []byte
		output int
	}{
		{nil, 0},
		{[]byte{}, 0},
		{[]byte{0x01}, 1},
		{[]byte{0x7F, 0x7F}, 0x3FFF},
	}

	for ii, tt := range tests {
		got := get7BitChunkedInt(tt.input)
		if got != tt.output {
			t.Errorf("[%d] get7BitChunkedInt(%v) = %v, expected %v", ii, tt.input, got, tt.output)
		}
	}
}

func TestGet8Dot8FixedPointAsFloat(t *testing.T) {
	tests := []struct {
		input  []byte
		output float64
	}{
		{[]byte{0x00, 0x00}, 0.0},
		{[]byte{0x00, 0x80}, 0.5},
		{[]byte{0x01, 0x80}, 1.5},
		{[]byte{0x7F, 0xC0}, 127.75},
		{[]byte{0xFE, 0x80}, -1.5},
	}

	for ii, tt := range tests {
		got := get8Dot8FixedPointAsFloat(tt.input)
		if got != tt.output {
			t.Errorf("[%d] get8Dot8FixedPointAsFloat(%v) = %v, expected %v", ii, tt.input, got, tt.output)
		}
	}
}

func TestGet16Dot16FixedPointAsFloat(t *testing.T) {
	tests := []struct {
		input  []byte
		output float64
	}{
		{[]byte{0x00, 0x00, 0x00, 0x00}, 0.0},
		{[]byte{0x00, 0x01, 0x80, 0x00}, 1.5},
		{[]byte{0xFF, 0xFE, 0xFF, 0xF0}, -1.000244140625},
	}

	for ii, tt := range tests {
		got := get16Dot16FixedPointAsFloat(tt.input)
		if got != tt.output {
			t.Errorf("[%d] get16Dot16FixedPointAsFloat(%v) = %v, expected %v", ii, tt.input, got, tt.output)
		}
	}
}

func TestGetFloat64(t *testing.T) {
	tests := []struct {
		input  []byte
		output float64
	}{
		{[]byte{0x3F, 0xF0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 1.0},
		{[]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, -2.0},
	}

	for ii, tt := range tests {
		got := getFloat64(tt.input)
		if got != tt.output {
			t.Errorf("[%d] getFloat64(%v) = %v, expected %v", ii, tt.input, got, tt.output)
		}
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		input  []byte
		output int
	}{
		{[]byte{}, 0},
		{[]byte{0x01}, 1},
		{[]byte{0xF1, 0xF2}, 0xF1F2},
		{[]byte{0xF1, 0xF2, 0xF3}, 0xF1F2F3},
		{[]byte{0xF1, 0xF2, 0xF3, 0xF4}, 0xF1F2F3F4},
	}

	for ii, tt := range tests {
		got := getInt(tt.input)
		if got != tt.output {
			t.Errorf("[%d] getInt(%v) = %v, expected %v", ii, tt.input, got, tt.output)
		}
	}
}

func TestGetInt16AsInt(t *testing.T) {
	tests := []struct {
		input  []byte
		output int
	}{
		{[]byte{0x00, 0x00}, 0},
		{[]byte{0x00, 0x01}, 1},
		{[]byte{0xFF, 0xFF}, -1},
	}

	for ii, tt := range tests {
		got := getInt16AsInt(tt.input)
		if got != tt.output {
			t.Errorf("[%d] getInt16AsInt(%v) = %v, expected %v", ii, tt.input, got, tt.output)
		}
	}
}

func TestGetInt32AsInt(t *testing.T) {
	tests := []struct {
		input  []byte
		output int
	}{
		{[]byte{0x00, 0x00, 0x00, 0x00}, 0},
		{[]byte{0x00, 0x00, 0x00, 0x01}, 1},
		{[]byte{0xFF, 0xFF, 0xFF, 0xFF}, -1},
	}

	for ii, tt := range tests {
		got := getInt32AsInt(tt.input)
		if got != tt.output {
			t.Errorf("[%d] getInt32AsInt(%v) = %v, expected %v", ii, tt.input, got, tt.output)
		}
	}
}

func TestGetInt32LittleAsInt(t *testing.T) {
	tests := []struct {
		input  []byte
		output int
	}{
		{[]byte{0x00, 0x00, 0x00, 0x00}, 0},
		{[]byte{0x01, 0x00, 0x00, 0x00}, 1},
		{[]byte{0xFF, 0xFF, 0xFF, 0xFF}, -1},
	}

	for ii, tt := range tests {
		got := getInt32LittleAsInt(tt.input)
		if got != tt.output {
			t.Errorf("[%d] getInt32LittleAsInt(%v) = %v, expected %v", ii, tt.input, got, tt.output)
		}
	}
}

func TestGetInt64(t *testing.T) {
	tests := []struct {
		input  []byte
		output int64
	}{
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0},
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, 1},
		{[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, -1},
	}

	for ii, tt := range tests {
		got := getInt64(tt.input)
		if got != tt.output {
			t.Errorf("[%d] getInt64(%v) = %v, expected %v", ii, tt.input, got, tt.output)
		}
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		input  []byte
		output string
	}{
		{nil, ""},
		{[]byte(""), ""},
		{[]byte("test"), "test"},
		{[]byte("	test "), "test"},
	}

	for ii, tt := range tests {
		got := getString(tt.input)
		if got != tt.output {
			t.Errorf("[%d] getString(%v) = %v, expected %v", ii, tt.input, got, tt.output)
		}
	}
}

func TestGetUint16AsInt(t *testing.T) {
	tests := []struct {
		input  []byte
		output int
	}{
		{[]byte{0x00, 0x00}, 0},
		{[]byte{0x00, 0x01}, 1},
		{[]byte{0xFF, 0xFF}, 65535},
	}

	for ii, tt := range tests {
		got := getUint16AsInt(tt.input)
		if got != tt.output {
			t.Errorf("[%d] getUint16AsInt(%v) = %v, expected %v", ii, tt.input, got, tt.output)
		}
	}
}

func TestGetUint24AsInt(t *testing.T) {
	tests := []struct {
		input  []byte
		output int
	}{
		{[]byte{0x00, 0x00, 0x00}, 0},
		{[]byte{0x00, 0x00, 0x01}, 1},
		{[]byte{0xFF, 0xFF, 0xFF}, 16777215},
	}

	for ii, tt := range tests {
		got := getUint24AsInt(tt.input)
		if got != tt.output {
			t.Errorf("[%d] getUint24AsInt(%v) = %v, expected %v", ii, tt.input, got, tt.output)
		}
	}
}

func TestGetUint32AsInt64(t *testing.T) {
	tests := []struct {
		input  []byte
		output int64
	}{
		{[]byte{0x00, 0x00, 0x00, 0x00}, 0},
		{[]byte{0x00, 0x00, 0x00, 0x01}, 1},
		{[]byte{0xFF, 0xFF, 0xFF, 0xFF}, 4294967295},
	}

	for ii, tt := range tests {
		got := getUint32AsInt64(tt.input)
		if got != tt.output {
			t.Errorf("[%d] getUint32AsInt64(%v) = %v, expected %v", ii, tt.input, got, tt.output)
		}
	}
}

func TestGetUint32LittleAsInt64(t *testing.T) {
	tests := []struct {
		input  []byte
		output int64
	}{
		{[]byte{0x00, 0x00, 0x00, 0x00}, 0},
		{[]byte{0x01, 0x00, 0x00, 0x00}, 1},
		{[]byte{0xFF, 0xFF, 0xFF, 0xFF}, 4294967295},
	}

	for ii, tt := range tests {
		got := getUint32LittleAsInt64(tt.input)
		if got != tt.output {
			t.Errorf("[%d] getUint32LittleAsInt64(%v) = %v, expected %v", ii, tt.input, got, tt.output)
		}
	}
}

func TestGetUint64(t *testing.T) {
	tests := []struct {
		input  []byte
		output uint64
	}{
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0},
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, 1},
		{[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, 18446744073709551615},
	}

	for ii, tt := range tests {
		got := getUint64(tt.input)
		if got != tt.output {
			t.Errorf("[%d] getUint64(%v) = %v, expected %v", ii, tt.input, got, tt.output)
		}
	}
}

func TestReadUint64Little(t *testing.T) {
	tests := []struct {
		input  []byte
		output uint64
	}{
		{[]byte{0, 0, 0, 0, 0, 0, 0, 0}, 0},
		{[]byte{0x01, 0, 0, 0, 0, 0, 0, 0}, 1},
		{[]byte{0xF1, 0xF2, 0, 0, 0, 0, 0, 0}, 0xF2F1},
		{[]byte{0xF1, 0xF2, 0xF3, 0, 0, 0, 0, 0}, 0xF3F2F1},
		{[]byte{0xF1, 0xF2, 0xF3, 0xF4, 0, 0, 0, 0}, 0xF4F3F2F1},
		{[]byte{0xF1, 0xF2, 0xF3, 0xF4, 0xF5, 0, 0, 0}, 0xF5F4F3F2F1},
	}

	for ii, tt := range tests {
		got, err := readUint64Little(bytes.NewReader(tt.input))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if got != tt.output {
			t.Errorf("[%d] readUint64Little(%v) = %v, expected %v", ii, tt.input, got, tt.output)
		}
	}
}
