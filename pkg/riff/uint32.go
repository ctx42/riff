// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Compile time checks.
var (
	_ io.Reader    = Uint32(0)
	_ fmt.Stringer = Uint32(0)
)

// Uint32 is a chunk ID or type with methods to print and encode it.
type Uint32 uint32

// String returns the ASCII representation of the uint32 integer.
func (u32 Uint32) String() string {
	return string(byte(u32>>24&0x000000FF)) +
		string(byte(u32>>16&0x000000FF)) +
		string(byte(u32>>8&0x000000FF)) +
		string(byte(u32&0x000000FF))
}

// Read writes the value to p in big endian and returns [io.EOF] with it, so a
// single read consumes the whole value. It returns [ErrTooShort] when p is
// shorter than four bytes.
func (u32 Uint32) Read(p []byte) (int, error) {
	if len(p) < 4 {
		return 0, fmt.Errorf("buffer too small for uint32: %w", ErrTooShort)
	}
	binary.BigEndian.PutUint32(p, uint32(u32))
	return 4, io.EOF
}
