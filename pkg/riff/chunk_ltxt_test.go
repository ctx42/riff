// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"io"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"
	"github.com/ctx42/testkit/pkg/iokit"
)

func ltxtChunkTextLenEven(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDltxt)                // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 24)                    // ( 4) 4 - Chunk size
	writeUint32LE(t, src, 1)                     // ( 8) 4 - PID
	writeUint32LE(t, src, 2)                     // (12) 4 - SamLen
	writeUint32LE(t, src, 3)                     // (16) 4 - PurID
	writeUint16LE(t, src, 4)                     // (20) 2 - Country
	writeUint16LE(t, src, 5)                     // (22) 2 - Language
	writeUint16LE(t, src, 6)                     // (24) 2 - Dialect
	writeUint16LE(t, src, 7)                     // (26) 2 - CodePage
	writeBytes(t, src, []byte{'a', 'b', 'c', 0}) // (28) 4 - Text
	// Total length: 8+4*3+2*4+4=32
	return src
}

func ltxtChunkTextLenOdd(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDltxt)           // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 23)               // ( 4) 4 - Chunk size
	writeUint32LE(t, src, 1)                // ( 8) 4 - PID
	writeUint32LE(t, src, 2)                // (12) 4 - SamLen
	writeUint32LE(t, src, 3)                // (16) 4 - PurID
	writeUint16LE(t, src, 4)                // (20) 2 - Country
	writeUint16LE(t, src, 5)                // (22) 2 - Language
	writeUint16LE(t, src, 6)                // (24) 2 - Dialect
	writeUint16LE(t, src, 7)                // (26) 2 - CodePage
	writeBytes(t, src, []byte{'a', 'b', 0}) // (28) 3 - Text
	writeBytes(t, src, []byte{0})           // (31) 1 - Padding byte
	// Total length: 8+4*3+2*4+3+1=32
	return src
}

func Test_ltxtStatic_decode(t *testing.T) {
	// --- Given ---
	b := make([]byte, 20)
	le.PutUint32(b[0:], 1)
	le.PutUint32(b[4:], 2)
	le.PutUint32(b[8:], 3)
	le.PutUint16(b[12:], 4)
	le.PutUint16(b[14:], 5)
	le.PutUint16(b[16:], 6)
	le.PutUint16(b[18:], 7)
	lts := &ltxtStatic{}

	// --- When ---
	lts.decode(b)

	// --- Then ---
	want := &ltxtStatic{
		CuePointID: 1,
		SamLen:     2,
		PurID:      3,
		Country:    4,
		Language:   5,
		Dialect:    6,
		CodePage:   7,
	}
	assert.Equal(t, want, lts)
}

func Test_ltxtStatic_encode(t *testing.T) {
	// --- Given ---
	want := make([]byte, 20)
	for i := range want {
		want[i] = byte(i + 1)
	}
	lts := &ltxtStatic{}
	lts.decode(want)
	have := make([]byte, 20)

	// --- When ---
	lts.encode(have)

	// --- Then ---
	assert.Equal(t, want, have)
}

func Test_LTXTMake(t *testing.T) {
	// --- Given ---
	mkr := LTXTMake(WithLoadData())

	// --- When ---
	have := mkr()

	// --- Then ---
	assert.SameType(t, &ChunkLTXT{}, have)
	assert.True(t, have.(*ChunkLTXT).opts.LoadData())
}

func Test_LTXT(t *testing.T) {
	// --- When ---
	have := LTXT()

	// --- Then ---
	assert.Equal(t, IDltxt, have.ID())
	assert.Equal(t, uint32(0), have.Size())
	assert.Equal(t, uint32(0), have.Type())
	assert.True(t, have.Multi())
	assert.Nil(t, have.Chunks())
	assert.False(t, have.Raw())
}

func Test_ChunkLTXT_ReadFrom(t *testing.T) {
	t.Run("even text length", func(t *testing.T) {
		// --- Given ---
		src := ltxtChunkTextLenEven(t)
		skip4B(t, src) // Skip chunk ID.

		ch := LTXT()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(28), n)
		assert.Equal(t, IDltxt, ch.ID())
		assert.Equal(t, uint32(24), ch.Size())
		assert.Equal(t, uint32(1), ch.CuePointID)
		assert.Equal(t, uint32(2), ch.SamLen)
		assert.Equal(t, uint32(3), ch.PurID)
		assert.Equal(t, uint16(4), ch.Country)
		assert.Equal(t, uint16(5), ch.Language)
		assert.Equal(t, uint16(6), ch.Dialect)
		assert.Equal(t, uint16(7), ch.CodePage)
		assert.Equal(t, []byte{'a', 'b', 'c', 0}, ch.text)
		text := must.Value(io.ReadAll(ch.Text()))
		assert.Equal(t, []byte{'a', 'b', 'c'}, text)
		assert.True(t, isAllRead(src))
	})

	t.Run("odd text length", func(t *testing.T) {
		// --- Given ---
		src := ltxtChunkTextLenOdd(t)
		skip4B(t, src) // Skip chunk ID.

		ch := LTXT()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(28), n)
		assert.Equal(t, IDltxt, ch.ID())
		assert.Equal(t, uint32(23), ch.Size())
		assert.Equal(t, uint32(1), ch.CuePointID)
		assert.Equal(t, uint32(2), ch.SamLen)
		assert.Equal(t, uint32(3), ch.PurID)
		assert.Equal(t, uint16(4), ch.Country)
		assert.Equal(t, uint16(5), ch.Language)
		assert.Equal(t, uint16(6), ch.Dialect)
		assert.Equal(t, uint16(7), ch.CodePage)
		assert.Equal(t, []byte{'a', 'b', 0}, ch.text)
		assert.Equal(t, []byte{'a', 'b'}, must.Value(io.ReadAll(ch.Text())))
		assert.True(t, isAllRead(src))
	})

	t.Run("error - truncated input", func(t *testing.T) {
		// Reading less than 12 bytes should always result in an error.
		for i := 1; i < 12; i++ {
			// --- Given ---
			src := ltxtChunkTextLenEven(t)
			skip4B(t, src) // Skip chunk ID.

			ch := LTXT()

			// --- When ---
			_, err := ch.ReadFrom(io.LimitReader(src, int64(i)))

			// --- Then ---
			assert.Error(t, err)
		}
	})

	t.Run("error - too short", func(t *testing.T) {
		// --- Given ---
		src := &bytes.Buffer{}
		writeUint32LE(t, src, 19)

		ch := LTXT()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrTooShort, err)
		want := "error decoding adtl:ltxt chunk: length too short"
		assert.ErrorEqual(t, want, err)
		assert.Equal(t, int64(4), n)
	})

	t.Run("error - size limit", func(t *testing.T) {
		// --- Given ---
		src := ltxtChunkTextLenEven(t)
		skip4B(t, src) // Skip chunk ID.

		ch := LTXT(WithSizeLimit(23))

		// --- When ---
		_, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrTooLarge, err)
	})

	t.Run("error - truncated text", func(t *testing.T) {
		// --- Given ---
		src := ltxtChunkTextLenOdd(t)
		skip4B(t, src) // Skip chunk ID.

		ch := LTXT()

		// --- When ---
		_, err := ch.ReadFrom(io.LimitReader(src, 25))

		// --- Then ---
		want := "error decoding adtl:ltxt chunk: unexpected EOF"
		assert.ErrorEqual(t, want, err)
		assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
	})

	t.Run("error - non-zero padding", func(t *testing.T) {
		// --- Given ---
		data := must.Value(io.ReadAll(ltxtChunkTextLenOdd(t)))
		data[len(data)-1] = 1

		ch := LTXT()

		// --- When ---
		_, err := ch.ReadFrom(bytes.NewReader(data[4:]))

		// --- Then ---
		want := "error decoding adtl:ltxt chunk: non-zero pad byte"
		assert.ErrorEqual(t, want, err)
		assert.ErrorIs(t, ErrPadByte, err)
	})

	t.Run("missing padding at end of input", func(t *testing.T) {
		// --- Given ---
		src := ltxtChunkTextLenOdd(t)
		skip4B(t, src) // Skip chunk ID.

		ch := LTXT()

		// --- When ---
		n, err := ch.ReadFrom(io.LimitReader(src, 27))

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(27), n)
	})
}

func Test_ChunkLTXT_errDecode(t *testing.T) {
	// --- Given ---
	ch := LTXT()

	// --- When ---
	err := ch.errDecode(io.EOF)

	// --- Then ---
	assert.ErrorEqual(t, "error decoding adtl:ltxt chunk: EOF", err)
	assert.ErrorIs(t, io.EOF, err)
}

func Test_ChunkLTXT_WriteTo(t *testing.T) {
	t.Run("error - write", func(t *testing.T) {
		// Writing less then 32 bytes should always result in an error.
		for i := 32; i > 0; i-- {
			// --- Given ---
			src := ltxtChunkTextLenOdd(t)
			skip4B(t, src) // Skip chunk ID.

			ch := LTXT()
			must.Value(ch.ReadFrom(src))
			dst := &bytes.Buffer{}

			// --- When ---
			_, err := ch.WriteTo(iokit.ErrWriter(dst, i))

			// --- Then ---
			if !assert.Error(t, err) {
				t.Logf("errro i=%d", i)
			}
		}
	})
}

func Test_ChunkLTXT_WriteTo_tabular(t *testing.T) {
	tt := []struct {
		testN string

		n  int64
		ch func(tester.T) io.Reader
	}{
		{"ltxtChunkTextLenEven", 32, ltxtChunkTextLenEven},
		{"ltxtChunkTextLenOdd", 32, ltxtChunkTextLenOdd},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			src := tc.ch(t)
			skip4B(t, src) // Skip chunk ID.

			ch := LTXT()
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

func Test_ChunkLTXT_Reset(t *testing.T) {
	// --- Given ---
	ch := LTXT()
	ch.size = 3
	ch.CuePointID = 1
	ch.SamLen = 2
	ch.PurID = 3
	ch.Country = 4
	ch.Language = 5
	ch.Dialect = 6
	ch.CodePage = 7
	ch.text = []byte{'a', 'b', 0}

	// --- When ---
	ch.Reset()

	// --- Then ---
	assert.Equal(t, IDltxt, ch.ID())
	assert.Equal(t, uint32(0), ch.Size())
	assert.Equal(t, uint32(0), ch.CuePointID)
	assert.Equal(t, uint32(0), ch.SamLen)
	assert.Equal(t, uint32(0), ch.PurID)
	assert.Equal(t, uint16(0), ch.Country)
	assert.Equal(t, uint16(0), ch.Language)
	assert.Equal(t, uint16(0), ch.Dialect)
	assert.Equal(t, uint16(0), ch.CodePage)
	assert.Equal(t, []byte{}, ch.text)
}
