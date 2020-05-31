// Copyright 2015, David Howden
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yurit

import (
	"encoding/binary"
	"io"
	"math"
	"strings"
)

//Does strings contain aString
func containsString(strings []string, aString string) bool {
	for _, s := range strings {
		if s == aString {
			return true
		}
	}
	return false
}

func getBit(b byte, n uint) bool {
	x := byte(1 << n)
	return (b & x) == x
}

func get7BitChunkedInt(b []byte) int {
	var n int
	for _, x := range b {
		n = n << 7
		n |= int(x)
	}
	return n
}

func getInt(b []byte) int {
	var n int
	for _, x := range b {
		n = n << 8
		n |= int(x)
	}
	return n
}

func getString(b []byte) string {
	return trimString(string(b))
}

func getUintLittleEndian(b []byte) uint {
	var n uint
	for i, x := range b {
		n += uint(x) << uint(8*i)
	}
	return n
}

func getSignedInt32LittleEndian(b []byte) int32 {
	return int32(binary.LittleEndian.Uint32(b))
}

func get16Dot16FixedPointAsFloat(b []byte) float64 {
	return float64(getUint32(b)) / 65536.0
}

func getFloat64(b []byte) float64 {
	i := binary.BigEndian.Uint64(b)
	return math.Float64frombits(i)
}

func getInt16(b []byte) int64 {
	return int64(int16(binary.BigEndian.Uint16(b)))
}

func getInt32(b []byte) int64 {
	return int64(int32(binary.BigEndian.Uint32(b)))
}

func getUint16(b []byte) int64 {
	return int64(binary.BigEndian.Uint16(b))
}

func getUint24(b []byte) int64 {
	b2 := []byte{0}
	b2 = append(b2, b...)
	return int64(binary.BigEndian.Uint32(b2))
}

func getUint32(b []byte) int64 {
	return int64(binary.BigEndian.Uint32(b))
}

func readUint64LittleEndian(r io.Reader) (uint64, error) {
	b, err := readBytes(r, 8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b), nil
}

func readBytes(r io.Reader, n uint) ([]byte, error) {
	b := make([]byte, n)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func readString(r io.Reader, n uint) (string, error) {
	b, err := readBytes(r, n)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func readUint(r io.Reader, n uint) (uint, error) {
	x, err := readInt(r, n)
	if err != nil {
		return 0, err
	}
	return uint(x), nil
}

func readInt(r io.Reader, n uint) (int, error) {
	b, err := readBytes(r, n)
	if err != nil {
		return 0, err
	}
	return getInt(b), nil
}

func read7BitChunkedUint(r io.Reader, n uint) (uint, error) {
	b, err := readBytes(r, n)
	if err != nil {
		return 0, err
	}
	return uint(get7BitChunkedInt(b)), nil
}

func readUint32LittleEndian(r io.Reader) (uint32, error) {
	b, err := readBytes(r, 4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

//For reading signed int
func readSignedInt32LittleEndian(r io.Reader) (int32, error) {
	b, err := readBytes(r, 4)
	if err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(b)), nil
}

func trimString(x string) string {
	return strings.TrimSpace(strings.Trim(x, "\x00"))
}
