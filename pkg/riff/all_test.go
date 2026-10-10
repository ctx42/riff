// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"encoding/binary"
	"io"

	"github.com/ctx42/testing/pkg/tester"
)

// IDUNKN is a test chunk ID "ABCD" which is unlikely to be in the wild.
const IDUNKN uint32 = 0x41424344

// writeUint32LE writes v to dst in little endian. Fails the test on error.
func writeUint32LE(t tester.T, dst io.Writer, v uint32) {
	t.Helper()
	if err := binary.Write(dst, binary.LittleEndian, v); err != nil {
		t.Fatal(err)
	}
}

// writeUint32BE writes v to dst in big endian, the byte order of chunk IDs.
// Fails the test on error.
func writeUint32BE(t tester.T, dst io.Writer, v uint32) {
	t.Helper()
	if err := binary.Write(dst, binary.BigEndian, v); err != nil {
		t.Fatal(err)
	}
}

// writeUint16LE writes v to dst in little endian. Fails the test on error.
func writeUint16LE(t tester.T, dst io.Writer, v uint16) {
	t.Helper()
	if err := binary.Write(dst, binary.LittleEndian, v); err != nil {
		t.Fatal(err)
	}
}

// writeBytes writes p to dst. Fails the test on error.
func writeBytes(t tester.T, dst io.Writer, p []byte) {
	t.Helper()
	if _, err := dst.Write(p); err != nil {
		t.Fatal(err)
	}
}

// skip4B reads and discards exactly 4 bytes from r. Fails the test on error.
func skip4B(t tester.T, r io.Reader) {
	t.Helper()
	if _, err := io.ReadFull(r, make([]byte, 4)); err != nil {
		t.Fatal(err)
	}
}

// isAllRead returns true if r has no bytes left to read.
func isAllRead(r io.Reader) bool {
	_, err := r.Read(make([]byte, 1))
	return err == io.EOF
}
