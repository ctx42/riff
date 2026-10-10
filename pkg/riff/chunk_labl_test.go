// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"io"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testkit/pkg/iokit"
)

func lablChunkTextLenEven(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDlabl)                // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 8)                     // ( 4) 4 - Chunk size
	writeUint32LE(t, src, 123)                   // ( 8) 4 - Cue ID
	writeBytes(t, src, []byte{'a', 'b', 'c', 0}) // (12) 4 - Text
	// Total length: 8+4+4=16
	return src
}

func lablChunkTextLenOdd(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDlabl)           // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 7)                // ( 4) 4 - Chunk size
	writeUint32LE(t, src, 123)              // ( 8) 4 - Cue ID
	writeBytes(t, src, []byte{'a', 'b', 0}) // (12) 3 - Text
	writeBytes(t, src, []byte{0})           // (15) 1 - Padding byte
	// Total length: 8+4+3+1=16
	return src
}

func Test_LABL(t *testing.T) {
	// --- When ---
	have := LABL()

	// --- Then ---
	assert.Equal(t, IDlabl, have.ID())
	assert.Equal(t, uint32(0), have.Size())
	assert.Equal(t, uint32(0), have.Type())
	assert.True(t, have.Multi())
	assert.Nil(t, have.Chunks())
	assert.False(t, have.Raw())
}

func Test_ChunkLABL_ReadFrom(t *testing.T) {
	t.Run("even text length", func(t *testing.T) {
		// --- Given ---
		src := lablChunkTextLenEven(t)
		skip4B(t, src) // Skip chunk ID.

		ch := LABL()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)

		assert.Equal(t, int64(12), n)
		assert.Equal(t, IDlabl, ch.ID())
		assert.Equal(t, uint32(8), ch.Size())
		assert.Equal(t, uint32(123), ch.CuePointID)
		assert.Equal(t, []byte{'a', 'b', 'c', 0}, ch.label)
		assert.Equal(t, []byte{'a', 'b', 'c'}, must.Value(io.ReadAll(ch.Label())))
		assert.True(t, isAllRead(src))
	})

	t.Run("odd text length", func(t *testing.T) {
		// --- Given ---
		src := lablChunkTextLenOdd(t)
		skip4B(t, src) // Skip chunk ID.

		// --- When ---
		ch := LABL()
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)

		assert.Equal(t, int64(12), n)
		assert.Equal(t, IDlabl, ch.ID())
		assert.Equal(t, uint32(7), ch.Size())
		assert.Equal(t, uint32(123), ch.CuePointID)
		assert.Equal(t, []byte{'a', 'b', 0}, ch.label)
		assert.Equal(t, []byte{'a', 'b'}, must.Value(io.ReadAll(ch.Label())))
		assert.True(t, isAllRead(src))
	})

	t.Run("error - truncated input", func(t *testing.T) {
		// Reading less than 12 bytes should always result in an error.
		for i := 1; i < 12; i++ {
			// --- Given ---
			src := lablChunkTextLenEven(t)
			skip4B(t, src) // Skip chunk ID.

			// --- When ---
			_, err := LABL().ReadFrom(io.LimitReader(src, int64(i)))

			// --- Then ---
			if !assert.Error(t, err) {
				t.Logf("errro i=%d", i)
			}
		}
	})

	t.Run("error - too short", func(t *testing.T) {
		// --- Given ---
		src := &bytes.Buffer{}
		writeUint32LE(t, src, 3)

		ch := LABL()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrTooShort, err)
		assert.ErrorContain(t, "adtl:labl chunk", err)
		assert.Equal(t, int64(4), n)
	})

	t.Run("error - size limit", func(t *testing.T) {
		// --- Given ---
		src := lablChunkTextLenEven(t)
		skip4B(t, src) // Skip chunk ID.

		ch := LABL(WithSizeLimit(7))

		// --- When ---
		_, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrTooLarge, err)
	})
}

func Test_ChunkLABL_WriteTo(t *testing.T) {
	t.Run("error - write", func(t *testing.T) {
		// Writing less then 16 bytes should always result in an error.
		for i := 16; i > 0; i-- {
			// --- Given ---
			src := lablChunkTextLenOdd(t)
			skip4B(t, src) // Skip chunk ID.

			ch := LABL()
			_, err := ch.ReadFrom(src)
			if !assert.NoError(t, err) {
				t.Logf("errro i=%d", i)
			}

			// --- When ---
			dst := &bytes.Buffer{}
			_, err = ch.WriteTo(iokit.ErrWriter(dst, i))

			// --- Then ---
			if !assert.Error(t, err) {
				t.Logf("errro i=%d", i)
			}
		}
	})
}

func Test_ChunkLABL_WriteTo_tabular(t *testing.T) {
	tt := []struct {
		testN string

		n  int64
		ch func(*testing.T) io.Reader
	}{
		{"lablChunkTextLenEven", 16, lablChunkTextLenEven},
		{"lablChunkTextLenOdd", 16, lablChunkTextLenOdd},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			src := tc.ch(t)
			skip4B(t, src) // Skip chunk ID.

			ch := LABL()
			_, err := ch.ReadFrom(src)
			assert.NoError(t, err)

			// --- When ---
			dst := &bytes.Buffer{}
			n, err := ch.WriteTo(dst)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.n, n)

			want := must.Value(io.ReadAll(tc.ch(t)))
			assert.Equal(t, want, dst.Bytes())
		})
	}
}

func Test_ChunkLABL_Reset(t *testing.T) {
	// --- Given ---
	ch := LABL()
	ch.size = 3
	ch.CuePointID = 123
	ch.label = []byte{'a', 'b', 0}

	// --- When ---
	ch.Reset()

	// --- Then ---
	assert.Equal(t, uint32(IDlabl), ch.ID())
	assert.Equal(t, uint32(0), ch.Size())
	assert.Equal(t, uint32(0), ch.CuePointID)
	assert.Equal(t, []byte{}, ch.label)
	assert.Equal(t, []byte{}, must.Value(io.ReadAll(ch.Label())))
}
