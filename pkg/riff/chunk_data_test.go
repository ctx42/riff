// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"
	"github.com/ctx42/testkit/pkg/iokit"
)

// dataChunkEven constructs data chunk with even number of data bytes.
func dataChunkEven(t tester.T) io.Reader {
	t.Helper()
	data := []byte{
		0x00, 0x01, 0x02, 0x03,
		0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f,
	}

	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDdata) // (0)  4 - Chunk ID
	writeUint32LE(t, src, 16)     // (4)  4 - Chunk size
	writeBytes(t, src, data)      // (8) 16 - Data
	// Total length: 8+16=24
	return src
}

// dataChunkOdd constructs data chunk with odd number of data bytes.
func dataChunkOdd(t tester.T) io.Reader {
	t.Helper()
	data := []byte{
		0x00, 0x01, 0x02, 0x03,
		0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e,
	}

	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDdata) // ( 0)  4 - Chunk ID
	writeUint32LE(t, src, 15)     // ( 4)  4 - Chunk size
	writeBytes(t, src, data)      // ( 8) 15 - Data
	writeBytes(t, src, []byte{0}) // (24)  1 - Padding byte
	// Total length: 8+16=24
	return src
}

func Test_DATA(t *testing.T) {
	t.Run("skip data mode", func(t *testing.T) {
		// --- When ---
		have := DATA()

		// --- Then ---
		assert.Equal(t, IDdata, have.ID())
		assert.Equal(t, uint32(0), have.Size())
		assert.Equal(t, uint32(0), have.Type())
		assert.False(t, have.Multi())
		assert.Nil(t, have.Chunks())
		assert.False(t, have.Raw())
		assert.Nil(t, have.data)
		assert.Equal(t, []byte{}, must.Value(io.ReadAll(have.Data())))
	})

	t.Run("load data mode", func(t *testing.T) {
		// --- When ---
		have := DATA(WithLoadData())

		// --- Then ---
		assert.Equal(t, IDdata, have.ID())
		assert.Equal(t, uint32(0), have.Size())
		assert.Equal(t, uint32(0), have.Type())
		assert.False(t, have.Multi())
		assert.Nil(t, have.Chunks())
		assert.False(t, have.Raw())
		assert.NotNil(t, have.data)
		assert.Equal(t, []byte{}, must.Value(io.ReadAll(have.Data())))
	})
}

func Test_ChunkDATA_SetData(t *testing.T) {
	t.Run("even size", func(t *testing.T) {
		// --- Given ---
		ch := DATA(WithLoadData())

		// --- When ---
		err := ch.SetData([]byte{0, 1, 2, 3})

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, uint32(4), ch.Size())
		assert.Equal(t, []byte{0, 1, 2, 3}, must.Value(io.ReadAll(ch.Data())))
	})

	t.Run("odd size", func(t *testing.T) {
		// --- Given ---
		ch := DATA(WithLoadData())

		// --- When ---
		err := ch.SetData([]byte{0, 1, 2})

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, uint32(3), ch.Size())
		assert.Equal(t, []byte{0, 1, 2}, must.Value(io.ReadAll(ch.Data())))
	})

	t.Run("error - skip data mode", func(t *testing.T) {
		// --- Given ---
		ch := DATA()

		// --- When ---
		err := ch.SetData([]byte{0, 1, 2})

		// --- Then ---
		assert.ErrorIs(t, err, ErrSkipDataMode)
		assert.Equal(t, uint32(0), ch.Size())
		assert.Equal(t, []byte{}, must.Value(io.ReadAll(ch.Data())))
	})
}

func Test_ChunkDATA_Duration(t *testing.T) {
	t.Run("byte rate", func(t *testing.T) {
		// --- Given ---
		ch := DATA()
		ch.size = 88200
		ch.data = make([]byte, 88200)

		// --- When ---
		have := ch.Duration(88200)

		// --- Then ---
		assert.Equal(t, time.Second, have)
	})

	t.Run("zero byte rate", func(t *testing.T) {
		// --- Given ---
		ch := DATA()
		ch.size = 88200

		// --- When ---
		have := ch.Duration(0)

		// --- Then ---
		assert.Equal(t, time.Duration(0), have)
	})
}

func Test_ChunkDATA_ReadFrom(t *testing.T) {
	t.Run("even size", func(t *testing.T) {
		// --- Given ---
		src := dataChunkEven(t)
		skip4B(t, src) // Skip chunk ID.

		ch := DATA(WithLoadData())

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(20), n)
		assert.Equal(t, uint32(16), ch.Size())
		want := []byte{
			0x00, 0x01, 0x02, 0x03,
			0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0a, 0x0b,
			0x0c, 0x0d, 0x0e, 0x0f,
		}
		assert.Equal(t, want, must.Value(io.ReadAll(ch.Data())))
		assert.True(t, isAllRead(src))
	})

	t.Run("odd size", func(t *testing.T) {
		// --- Given ---
		src := dataChunkOdd(t)
		skip4B(t, src) // Skip chunk ID.

		ch := DATA(WithLoadData())

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(20), n)
		assert.Equal(t, uint32(15), ch.Size())
		want := []byte{
			0x00, 0x01, 0x02, 0x03,
			0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0a, 0x0b,
			0x0c, 0x0d, 0x0e,
		}
		assert.Equal(t, want, must.Value(io.ReadAll(ch.Data())))
		assert.True(t, isAllRead(src))
	})

	t.Run("error - size limit", func(t *testing.T) {
		// --- Given ---
		src := dataChunkOdd(t)
		skip4B(t, src) // Skip chunk ID.

		ch := DATA(WithSizeLimit(15))

		// --- When ---
		_, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrTooLarge, err)
	})

	t.Run("error - truncated input", func(t *testing.T) {
		// Reading less than 20 bytes should always result in an error.
		for i := 1; i < 20; i++ {
			// --- Given ---
			src := dataChunkOdd(t)
			skip4B(t, src) // Skip chunk ID.

			ch := DATA(WithLoadData())

			// --- When ---
			_, err := ch.ReadFrom(io.LimitReader(src, int64(i)))

			// --- Then ---
			if !assert.Error(t, err) {
				t.Logf("error i=%d", i)
			}
		}
	})
}

func Test_ChunkDATA_WriteTo(t *testing.T) {
	t.Run("error - skip data mode", func(t *testing.T) {
		// --- Given ---
		src := dataChunkOdd(t)
		skip4B(t, src) // Skip chunk ID.

		ch := DATA()
		_, err := ch.ReadFrom(src)
		assert.NoError(t, err)

		// --- When ---
		n, err := ch.WriteTo(&bytes.Buffer{})

		// --- Then ---
		assert.ErrorIs(t, err, ErrSkipDataMode)
		assert.Equal(t, int64(0), n)
	})

	t.Run("error - write", func(t *testing.T) {
		// Writing less than 24 bytes should always result in an error.
		for i := 24; i > 0; i-- {
			// --- Given ---
			src := dataChunkOdd(t)
			skip4B(t, src) // Skip chunk ID.

			ch := DATA(WithLoadData())
			_, err := ch.ReadFrom(src)
			if assert.NoError(t, err) {
				t.Logf("error i=%d", i)
			}

			dst := &bytes.Buffer{}

			// --- When ---
			_, err = ch.WriteTo(iokit.ErrWriter(dst, i))

			// --- Then ---
			if !assert.Error(t, err) {
				t.Logf("error i=%d", i)
			}
		}
	})
}

func Test_ChunkDATA_WriteTo_tabular(t *testing.T) {
	tt := []struct {
		testN string

		n  int64
		ch func(tester.T) io.Reader
	}{
		{"dataChunkEven", 24, dataChunkEven},
		{"dataChunkOdd", 24, dataChunkOdd},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			src := tc.ch(t)
			skip4B(t, src) // Skip chunk ID.

			ch := DATA(WithLoadData())
			must.Value(ch.ReadFrom(src))

			dst := &bytes.Buffer{}

			// --- When ---
			n, err := ch.WriteTo(dst)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.n, n)
			want := must.Value(io.ReadAll(tc.ch(t)))
			assert.Equal(t, want, dst.Bytes())
		})
	}
}

func Test_ChunkDATA_Reset(t *testing.T) {
	// --- Given ---
	ch := DATA(WithLoadData())
	ch.size = 88200
	err := ch.SetData(make([]byte, 88200))
	assert.NoError(t, err)

	// --- When ---
	ch.Reset()

	// --- Then ---
	assert.Equal(t, uint32(0), ch.Size())
	assert.Len(t, 0, ch.data)
}
