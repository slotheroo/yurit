// Copyright 2015, David Howden
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yurit

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"
)

//Checks to see if all bytes in a slice are 0
func areZero(b []byte) bool {
	for _, byte := range b {
		if byte != 0 {
			return false
		}
	}
	return true
}

//bufferPeek looks at the next n bytes in buffer without advancing the position.
//If there are fewer than n bytes in the buffer, bufferPeek returns the entire
//buffer.
func bufferPeek(buf *bytes.Buffer, n int) []byte {
	if buf.Len() < n {
		n = buf.Len()
	}
	return buf.Bytes()[:n]
}

//checkLen checks to see if a specific index is within range of a slice of bytes
func checkLen(bytes []byte, check uint) error {
	if uint(len(bytes)) < check {
		return fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", check, len(bytes))
	}
	return nil
}

//Does strings contain str
func containsString(strings []string, str string) bool {
	for _, s := range strings {
		if s == str {
			return true
		}
	}
	return false
}

func getBit(b byte, n uint) bool {
	if n > 7 {
		return false
	}
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

//b must have length of 2 bytes or function will panic
func get8Dot8FixedPointAsFloat(b []byte) float64 {
	return float64(getInt16AsInt(b)) / 256.0
}

//b must have length of 4 bytes or function will panic
func get16Dot16FixedPointAsFloat(b []byte) float64 {
	return float64(getInt32AsInt(b)) / 65536.0
}

//b must have length of 8 bytes or function will panic
func getFloat64(b []byte) float64 {
	i := binary.BigEndian.Uint64(b)
	return math.Float64frombits(i)
}

//treats an unknown number of bytes as a uint and returns as an int
func getInt(b []byte) int {
	var n int
	for _, x := range b {
		n = n << 8
		n |= int(x)
	}
	return n
}

//b must have length of 2 bytes or function will panic
func getInt16AsInt(b []byte) int {
	return int(int16(binary.BigEndian.Uint16(b)))
}

//b must have length of 4 bytes or function will panic
func getInt32AsInt(b []byte) int {
	return int(int32(binary.BigEndian.Uint32(b)))
}

//b must have length of 4 bytes or function will panic
func getInt32LittleAsInt(b []byte) int {
	return int(int32(binary.LittleEndian.Uint32(b)))
}

//b must have length of 8 bytes or function will panic
func getInt64(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

func getString(b []byte) string {
	return trimString(string(b))
}

//treats an unknown number of bytes as a uint
func getUint(b []byte) uint {
	var n uint
	for _, x := range b {
		n = n << 8
		n |= uint(x)
	}
	return n
}

//b must have length of 2 bytes or function will panic
func getUint16AsInt(b []byte) int {
	return int(binary.BigEndian.Uint16(b))
}

//b must have length of 3 bytes or function will panic
func getUint24AsInt(b []byte) int {
	b2 := []byte{0}
	b2 = append(b2, b...)
	return int(binary.BigEndian.Uint32(b2))
}

//b must have length of 4 bytes or function will panic
func getUint32AsInt64(b []byte) int64 {
	return int64(binary.BigEndian.Uint32(b))
}

//b must have length of 4 bytes or function will panic
func getUint32LittleAsInt64(b []byte) int64 {
	return int64(binary.LittleEndian.Uint32(b))
}

//b must have length of 8 bytes or function will panic
func getUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

/*func getUintLittleEndian(b []byte) uint {
	var n uint
	for i, x := range b {
		n += uint(x) << uint(8*i)
	}
	return n
}*/

////////

func readBytes(r io.Reader, n uint) ([]byte, error) {
	b := make([]byte, n)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

//TODO: Replace with readBytes plus cast to string
func readString(r io.Reader, n uint) (string, error) {
	b, err := readBytes(r, n)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

//TODO: Replace with readBytes plus appropriate Get function
func readUint(r io.Reader, n uint) (uint, error) {
	x, err := readInt(r, n)
	if err != nil {
		return 0, err
	}
	return uint(x), nil
}

//TODO: Replace with readBytes plus appropriate Get function
func readInt(r io.Reader, n uint) (int, error) {
	b, err := readBytes(r, n)
	if err != nil {
		return 0, err
	}
	return getInt(b), nil
}

//TODO: Replace with readBytes plus appropriate Get function
func read7BitChunkedUint(r io.Reader, n uint) (uint, error) {
	b, err := readBytes(r, n)
	if err != nil {
		return 0, err
	}
	return uint(get7BitChunkedInt(b)), nil
}

//TODO: Replace with readBytes plus appropriate Get function
func readUint32Little(r io.Reader) (uint32, error) {
	b, err := readBytes(r, 4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

//TODO: Replace with readBytes plus appropriate Get function
func readInt64Little(r io.Reader) (int64, error) {
	u, err := readUint64Little(r)
	return int64(u), err
}

//TODO: Replace with readBytes plus appropriate Get function
func readUint64Little(r io.Reader) (uint64, error) {
	b, err := readBytes(r, 8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b), nil
}

func trimString(x string) string {
	return strings.TrimSpace(strings.Trim(x, "\x00"))
}
