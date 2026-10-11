// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"slices"
)

// StrToID converts a four-byte ASCII string to uint32.
// If the string is shorter than 4 bytes, it will be padded with spaces.
// If the string is longer than four bytes, it will be trimmed to 4 bytes.
func StrToID(s string) uint32 {
	b := []byte(s)
	if len(b) > 4 {
		b = b[:4]
	}
	if len(b) < 4 {
		b = append(b, bytes.Repeat([]byte{' '}, 4-len(b))...)
	}
	var id uint32
	for _, b := range b {
		id = (id << 8) | uint32(b)
	}
	return id
}

// ReadChunkID reads a four-byte chunk ID into id. It returns [io.EOF] when r
// has no bytes left and [io.ErrUnexpectedEOF] when it has fewer than four.
func ReadChunkID(r io.Reader, id *uint32) error {
	return binary.Read(r, be, id)
}

// RealSize returns size rounded up to an even number: the size of a chunk body
// with its padding byte.
func RealSize(size uint32) uint32 {
	if size%2 != 0 {
		return size + 1
	}
	return size
}

// ReadPaddingIf reads the padding byte from r when size is odd. It returns the
// number of bytes read, [io.ErrUnexpectedEOF] when the padding byte is
// missing, and [ErrPadByte] when it is not zero. When r is a [ChunkReader]
// whose source input ends where the padding byte should be, the missing byte
// is tolerated: no bytes are read and no error is returned.
func ReadPaddingIf(r io.Reader, size uint32) (int64, error) {
	if size%2 == 0 {
		return 0, nil
	}

	var pad [1]byte
	if _, err := io.ReadFull(r, pad[:]); err != nil {
		if cr, ok := r.(*ChunkReader); ok && err == io.EOF && cr.srcEOF {
			return 0, nil
		}
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
		return 0, err
	}
	if pad[0] != 0 {
		return 1, ErrPadByte
	}
	return 1, nil
}

// discardPaddingIf reads and discards a padding byte of any value from r when
// size is odd. It returns [io.ErrUnexpectedEOF] when the byte is missing.
func discardPaddingIf(r io.Reader, size uint32) error {
	if size%2 == 0 {
		return nil
	}

	if _, err := io.ReadFull(r, make([]byte, 1)); err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
		return err
	}
	return nil
}

// shortRead reports whether read bytes of chunk content fall short of the
// declared size. Chunks take an even number of bytes, so an odd read one byte
// short of size means the input ended without the final padding byte, which
// is tolerated.
func shortRead(size, read int64) bool {
	return read < size && (read%2 == 0 || read+1 != size)
}

// WriteIDAndSize writes the chunk id (big endian) and size (little endian)
// to w.
func WriteIDAndSize(w io.Writer, id, size uint32) (int64, error) {
	var sum int64
	if err := binary.Write(w, be, id); err != nil {
		return 0, err
	}
	sum += 4
	if err := binary.Write(w, le, size); err != nil {
		return sum, err
	}
	sum += 4
	return sum, nil
}

// WritePaddingIf writes a zero padding byte to w when size is odd.
func WritePaddingIf(w io.Writer, size uint32) (int64, error) {
	if size%2 == 1 {
		n, err := w.Write([]byte{0})
		return int64(n), err
	}
	return 0, nil
}

// TrimZeroRight removes all zero bytes from the end of the slice.
func TrimZeroRight(b []byte) []byte {
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] > 0 {
			return b[:i+1]
		}
	}
	return b[:0]
}

// linkids returns the ASCII representations of two chunk IDs joined with a
// colon.
func linkids(id1, id2 uint32) string {
	return Uint32(id1).String() + ":" + Uint32(id2).String()
}

// grow returns b resized to n bytes, reallocating when its capacity is short.
func grow(b []byte, n int) []byte {
	if cap(b) >= n {
		return b[:n]
	}

	tmp := make([]byte, n)
	copy(tmp, b)
	return tmp
}

// minReadStep is the smallest number of bytes readBytes grows a buffer by.
const minReadStep = 512

// readBytes reads exactly n bytes from r into b, reusing its capacity, and
// returns the resized slice. A buffer too small for n bytes grows in step with
// the bytes actually read, so a declared size larger than the input cannot
// force a large allocation up front. It returns [io.EOF] when no bytes were
// read and [io.ErrUnexpectedEOF] when fewer than n bytes were read.
func readBytes(r io.Reader, b []byte, n int) ([]byte, error) {
	if n <= cap(b) {
		b = b[:n]
		_, err := io.ReadFull(r, b)
		return b, err
	}

	b = b[:0]
	for len(b) < n {
		if len(b) == cap(b) {
			b = slices.Grow(b, min(n-len(b), max(cap(b), minReadStep)))
		}
		end := min(cap(b), n)
		m, err := io.ReadFull(r, b[len(b):end])
		b = b[:len(b)+m]
		if err != nil {
			if errors.Is(err, io.EOF) && len(b) > 0 {
				err = io.ErrUnexpectedEOF
			}
			return b, err
		}
	}
	return b, nil
}
