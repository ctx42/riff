// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testkit/pkg/iokit"
)

func Test_RAWC(t *testing.T) {
	// --- When ---
	have := RAWC(IDUNKN, WithLoadData())

	// --- Then ---
	assert.Equal(t, uint32(IDUNKN), have.ID())
	assert.Equal(t, uint32(0), have.Size())
	assert.Equal(t, uint32(0), have.Type())
	assert.True(t, have.Multi())
	assert.Len(t, 0, have.Chunks())
	assert.True(t, have.Raw())
}

func Test_ChunkRAWC_ReadFrom(t *testing.T) {
	t.Run("load data", func(t *testing.T) {
		// --- Given ---
		src := &bytes.Buffer{}
		writeUint32LE(t, src, 3)         // Chunk size (4).
		src.Write([]byte{'A', 'B', 'C'}) // Chunk data (*).
		src.WriteByte(0)                 // Padding byte (1).

		ch := RAWC(IDUNKN, WithLoadData())

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(8), n)
		assert.Equal(t, uint32(3), ch.Size())
		assert.Equal(t, uint32(0), ch.Type())
		assert.Equal(t, []byte{'A', 'B', 'C'}, ch.data)
		assert.True(t, isAllRead(src))
	})

	t.Run("skip data with seeker", func(t *testing.T) {
		// --- Given ---
		buf := &bytes.Buffer{}
		writeUint32LE(t, buf, 3)                  // Chunk size (4).
		writeBytes(t, buf, []byte{'A', 'B', 'C'}) // Chunk data (*).
		writeBytes(t, buf, []byte{0})             // Padding byte (1).

		// The bytes.Reader implements io.Seeker.
		src := bytes.NewReader(buf.Bytes())
		ch := RAWC(IDUNKN)

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(8), n)
		assert.Equal(t, uint32(3), ch.Size())
		assert.Nil(t, ch.data)
		assert.True(t, isAllRead(src))
	})

	t.Run("skip data without seeker", func(t *testing.T) {
		// --- Given ---
		// The bytes.Buffer doesn't implement io.Seeker.
		src := &bytes.Buffer{}
		writeUint32LE(t, src, 3)                  // Chunk size (4).
		writeBytes(t, src, []byte{'A', 'B', 'C'}) // Chunk data (*).
		writeBytes(t, src, []byte{0})             // Padding byte (1).
		ch := RAWC(IDUNKN)

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(8), n)
		assert.Equal(t, uint32(3), ch.Size())
		assert.Nil(t, ch.data)
		assert.True(t, isAllRead(src))
	})

	t.Run("error - short size", func(t *testing.T) {
		// --- Given ---
		src := &bytes.Buffer{}
		src.Write([]byte{0, 1, 2}) // Too short chunk size.

		ch := RAWC(IDUNKN, WithLoadData())

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
		assert.ErrorContain(t, "error decoding RAWC:ABCD chunk: ", err)
		assert.Equal(t, int64(0), n)
	})

	t.Run("error - reading data", func(t *testing.T) {
		// --- Given ---
		src := &bytes.Buffer{}
		writeUint32LE(t, src, 10)        // Chunk size (4).
		src.Write([]byte{'A', 'B', 'C'}) // Chunk data (*).
		src.WriteByte(0)                 // Padding byte (1).
		ch := RAWC(IDUNKN, WithLoadData())

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
		assert.ErrorContain(t, "error decoding RAWC:ABCD chunk: ", err)
		assert.Equal(t, int64(8), n)
	})

	t.Run("error - reading padding", func(t *testing.T) {
		// --- Given ---
		src := &bytes.Buffer{}
		writeUint32LE(t, src, 3)         // Chunk size (4).
		src.Write([]byte{'A', 'B', 'C'}) // Chunk data (*).
		ch := RAWC(IDUNKN, WithLoadData())

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
		assert.ErrorContain(t, "error decoding RAWC:ABCD chunk: ", err)
		assert.Equal(t, int64(7), n)
	})

	t.Run("error - size limit", func(t *testing.T) {
		// --- Given ---
		src := &bytes.Buffer{}
		writeUint32LE(t, src, 4)              // Chunk size (4).
		src.Write([]byte{'A', 'B', 'C', 'D'}) // Chunk data (*).

		ch := RAWC(IDUNKN, WithLoadData(), WithSizeLimit(3))

		// --- When ---
		_, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrTooLarge, err)
	})
}

func Test_ChunkRAWC_WriteTo(t *testing.T) {
	t.Run("without padding", func(t *testing.T) {
		// --- Given ---
		ch := RAWC(IDUNKN, WithLoadData())
		ch.size = 2
		ch.data = []byte{0, 1}
		dst := &bytes.Buffer{}

		// --- When ---
		n, err := ch.WriteTo(dst)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(10), n)

		want := []byte{
			0x41, 0x42, 0x43, 0x44, // ID.
			0x2, 0x0, 0x0, 0x0, // Size.
			0x0, 0x1, // Data.
		}
		assert.Equal(t, want, dst.Bytes())
	})

	t.Run("with padding", func(t *testing.T) {
		// --- Given ---
		ch := RAWC(IDUNKN, WithLoadData())
		ch.size = 3
		ch.data = []byte{0, 1, 2}
		dst := &bytes.Buffer{}

		// --- When ---
		n, err := ch.WriteTo(dst)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(12), n)

		want := []byte{
			0x41, 0x42, 0x43, 0x44, // ID.
			0x3, 0x0, 0x0, 0x0, // Size.
			0x0, 0x1, 0x2, // Data.
			0x0, // Padding.
		}
		assert.Equal(t, want, dst.Bytes())
	})

	t.Run("error - writing ID", func(t *testing.T) {
		// --- Given ---
		ch := RAWC(IDUNKN, WithLoadData())
		ch.size = 3
		ch.data = []byte{0, 1, 2}

		buf := &bytes.Buffer{}
		dst := iokit.ErrWriter(buf, 3)

		// --- When ---
		n, err := ch.WriteTo(dst)

		// --- Then ---
		assert.ErrorIs(t, iokit.ErrWrite, err)
		assert.ErrorContain(t, "error encoding RAWC:ABCD chunk: ", err)
		assert.Equal(t, int64(0), n)

		want := []byte{0x41, 0x42, 0x43}
		assert.Equal(t, want, buf.Bytes())
	})

	t.Run("error - writing data", func(t *testing.T) {
		// --- Given ---
		ch := RAWC(IDUNKN, WithLoadData())
		ch.size = 3
		ch.data = []byte{0, 1, 2}

		buf := &bytes.Buffer{}
		dst := iokit.ErrWriter(buf, 10)

		// --- When ---
		n, err := ch.WriteTo(dst)

		// --- Then ---
		assert.ErrorIs(t, iokit.ErrWrite, err)
		assert.ErrorContain(t, "error encoding RAWC:ABCD chunk: ", err)
		assert.Equal(t, int64(10), n)

		want := []byte{
			0x41, 0x42, 0x43, 0x44, // ID.
			0x3, 0x0, 0x0, 0x0, // Size.
			0x0, 0x1, // Data.
		}
		assert.Equal(t, want, buf.Bytes())
	})

	t.Run("error - writing padding", func(t *testing.T) {
		// --- Given ---
		ch := RAWC(IDUNKN, WithLoadData())
		ch.size = 3
		ch.data = []byte{0, 1, 2}

		buf := &bytes.Buffer{}
		dst := iokit.ErrWriter(buf, 12)

		// --- When ---
		n, err := ch.WriteTo(dst)

		// --- Then ---
		assert.ErrorIs(t, iokit.ErrWrite, err)
		assert.ErrorContain(t, "error encoding RAWC:ABCD chunk: ", err)
		assert.Equal(t, int64(12), n)

		want := []byte{
			0x41, 0x42, 0x43, 0x44, // ID.
			0x3, 0x0, 0x0, 0x0, // Size.
			0x0, 0x1, 0x2, // Data.
			0x0, // Padding
		}
		assert.Equal(t, want, buf.Bytes())
	})

	t.Run("error - skip data mode", func(t *testing.T) {
		// --- Given ---
		ch := RAWC(IDUNKN)
		ch.size = 3

		dst := &bytes.Buffer{}

		// --- When ---
		n, err := ch.WriteTo(dst)

		// --- Then ---
		assert.True(t, errors.Is(err, ErrSkipDataMode))
		assert.Equal(t, int64(0), n)
	})
}

func Test_ChunkRAWC_Reset(t *testing.T) {
	// --- Given ---
	ch := RAWC(IDUNKN, WithLoadData())
	ch.size = 3
	ch.data = []byte{0, 1, 2}

	// --- When ---
	ch.Reset()

	// --- Then ---
	assert.Equal(t, uint32(IDUNKN), ch.ID())
	assert.Equal(t, uint32(0), ch.Size())
	assert.Len(t, 0, ch.data)
}
