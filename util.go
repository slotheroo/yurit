// Copyright 2015, David Howden
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yurit

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"
)

//checkLen checks to see if a specifix index is within range of a slice of bytes
func checkLen(bytes []byte, check int) error {
	if len(bytes) < check {
		return fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", check, len(bytes))
	}
	return nil
}

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

func get8Dot8FixedPointAsFloat(b []byte) float64 {
	return float64(getUint16AsInt(b)) / 256.0
}

func get16Dot16FixedPointAsFloat(b []byte) float64 {
	return float64(getUint32AsInt64(b)) / 65536.0
}

func getFloat64(b []byte) float64 {
	i := binary.BigEndian.Uint64(b)
	return math.Float64frombits(i)
}

func getInt16AsInt(b []byte) int {
	return int(int16(binary.BigEndian.Uint16(b)))
}

func getInt32AsInt(b []byte) int {
	return int(int32(binary.BigEndian.Uint32(b)))
}

func getInt64(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

func getInt32LittleAsInt(b []byte) int {
	return int(binary.LittleEndian.Uint32(b))
}

func getUint16AsInt(b []byte) int {
	return int(binary.BigEndian.Uint16(b))
}

func getUint24AsInt(b []byte) int {
	b2 := []byte{0}
	b2 = append(b2, b...)
	return int(binary.BigEndian.Uint32(b2))
}

func getUint32AsInt64(b []byte) int64 {
	return int64(binary.BigEndian.Uint32(b))
}

func getUint32Little(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b)
}

func getUint32LittleAsInt64(b []byte) int64 {
	return int64(getUint32Little(b))
}

func getUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
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

func readInt64Little(r io.Reader) (int64, error) {
	u, err := readUint64LittleEndian(r)
	return int64(u), err
}

func readUint64LittleEndian(r io.Reader) (uint64, error) {
	b, err := readBytes(r, 8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b), nil
}

func trimString(x string) string {
	return strings.TrimSpace(strings.Trim(x, "\x00"))
}
