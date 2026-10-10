// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"io"
	"math"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"
	"github.com/ctx42/testkit/pkg/iokit"
)

// fmtChunkWithoutExtraBytes constructs fmt chunk without extra bytes.
func fmtChunkWithoutExtraBytes(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDfmt)   // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 16)      // ( 4) 4 - Chunk size
	writeUint16LE(t, src, CompPCM) // ( 6) 2 - CompCode
	writeUint16LE(t, src, 1)       // ( 8) 2 - ChannelCnt
	writeUint32LE(t, src, 44100)   // (10) 4 - SampleRate
	writeUint32LE(t, src, 88200)   // (14) 4 - AvgByteRate
	writeUint16LE(t, src, 2)       // (18) 2 - BlockAlign
	writeUint16LE(t, src, 16)      // (20) 2 - BitsPerSample
	// Total length: 8+16=24
	return src
}

// fmtChunkWithExtraBytesEven constructs fmt chunk with extra bytes of
// even length.
func fmtChunkWithExtraBytesEven(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDfmt)     // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 16+4)      // ( 4) 4 - Chunk size
	writeUint16LE(t, src, CompPCM)   // ( 6) 2 - CompCode
	writeUint16LE(t, src, 1)         // ( 8) 2 - ChannelCnt
	writeUint32LE(t, src, 44100)     // (10) 4 - SampleRate
	writeUint32LE(t, src, 88200)     // (14) 4 - AvgByteRate
	writeUint16LE(t, src, 2)         // (18) 2 - BlockAlign
	writeUint16LE(t, src, 16)        // (20) 2 - BitsPerSample
	writeUint16LE(t, src, 2)         // (22) 2 - ExtraBytes
	writeBytes(t, src, []byte{0, 1}) // (24) 2 - Extra bytes
	// Total length: 8+16+2+2=28
	return src
}

// fmtChunkWithExtraBytesOdd constructs fmt chunk with extra bytes of
// odd length.
func fmtChunkWithExtraBytesOdd(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDfmt)        // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 16+2+4)       // ( 4) 4 - Chunk size
	writeUint16LE(t, src, CompPCM)      // ( 6) 2 - CompCode
	writeUint16LE(t, src, 1)            // ( 8) 2 - ChannelCnt
	writeUint32LE(t, src, 44100)        // (10) 4 - SampleRate
	writeUint32LE(t, src, 88200)        // (14) 4 - AvgByteRate
	writeUint16LE(t, src, 2)            // (18) 2 - BlockAlign
	writeUint16LE(t, src, 16)           // (20) 2 - BitsPerSample
	writeUint16LE(t, src, 3)            // (22) 2 - ExtraBytes
	writeBytes(t, src, []byte{0, 1, 2}) // (24) 3 - Extra bytes
	writeBytes(t, src, []byte{0})       // (27) 1 - Padding byte
	// Total length: 8+16+2+4=30
	return src
}

// fmtChunkWithExtraOfZeroLen constructs fmt chunk with extra bytes of
// zero length.
func fmtChunkWithExtraOfZeroLen(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDfmt)   // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 16+2)    // ( 4) 4 - Chunk size
	writeUint16LE(t, src, CompPCM) // ( 6) 2 - CompCode
	writeUint16LE(t, src, 1)       // ( 8) 2 - ChannelCnt
	writeUint32LE(t, src, 44100)   // (10) 4 - SampleRate
	writeUint32LE(t, src, 88200)   // (14) 4 - AvgByteRate
	writeUint16LE(t, src, 2)       // (18) 2 - BlockAlign
	writeUint16LE(t, src, 16)      // (20) 2 - BitsPerSample
	writeUint16LE(t, src, 0)       // (22) 2 - ExtraBytes
	// Total length: 8+16+2+0=26
	return src
}

// fmtChunkWithExtraInvalidSize constructs fmt chunk with extra bytes declaring
// invalid even size.
func fmtChunkWithExtraInvalidSizeEven(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDfmt)        // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 16+2)         // ( 4) 4 - Chunk size
	writeUint16LE(t, src, CompPCM)      // ( 6) 2 - CompCode
	writeUint16LE(t, src, 1)            // ( 8) 2 - ChannelCnt
	writeUint32LE(t, src, 44100)        // (10) 4 - SampleRate
	writeUint32LE(t, src, 88200)        // (14) 4 - AvgByteRate
	writeUint16LE(t, src, 2)            // (18) 2 - BlockAlign
	writeUint16LE(t, src, 16)           // (20) 2 - BitsPerSample
	writeUint16LE(t, src, 7000)         // (22) 2 - ExtraBytes
	writeBytes(t, src, []byte{0, 1, 2}) // (24) 3 - Extra bytes
	// Total length: 8+16+2+3=29
	return src
}

// fmtChunkWithExtraInvalidSizeOdd constructs fmt chunk with extra bytes
// declaring invalid odd size.
func fmtChunkWithExtraInvalidSizeOdd(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDfmt)        // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 16+2)         // ( 4) 4 - Chunk size
	writeUint16LE(t, src, CompPCM)      // ( 6) 2 - CompCode
	writeUint16LE(t, src, 1)            // ( 8) 2 - ChannelCnt
	writeUint32LE(t, src, 44100)        // (10) 4 - SampleRate
	writeUint32LE(t, src, 88200)        // (14) 4 - AvgByteRate
	writeUint16LE(t, src, 2)            // (18) 2 - BlockAlign
	writeUint16LE(t, src, 16)           // (20) 2 - BitsPerSample
	writeUint16LE(t, src, 7001)         // (22) 2 - ExtraBytes
	writeBytes(t, src, []byte{0, 1, 2}) // (24) 3 - Extra bytes
	// Total length: 8+16+2+3=29
	return src
}

func Test_ChunkFMT_Size(t *testing.T) {
	t.Run("new chunk", func(t *testing.T) {
		// --- Given ---
		ch := FMT()
		ch.CompCode = CompPCM

		// --- When ---
		have := ch.Size()

		// --- Then ---
		assert.Equal(t, FMTChunkSize, have)
	})

	t.Run("write zero extra set after SetExtra", func(t *testing.T) {
		// --- Given ---
		ch := FMT()
		ch.SetExtra(nil)
		ch.WriteZeroExtra = true

		// --- When ---
		have := ch.Size()

		// --- Then ---
		assert.Equal(t, uint32(18), have)
	})

	t.Run("RIFF size matches written bytes", func(t *testing.T) {
		// --- Given ---
		ch := FMT()
		ch.CompCode = CompPCM
		rif := Compose(Chunks{ch})
		dst := &bytes.Buffer{}

		// --- When ---
		n, err := rif.WriteTo(dst)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(36), n)
		assert.Equal(t, uint32(n-8), le.Uint32(dst.Bytes()[4:8]))
	})
}

func Test_FMT(t *testing.T) {
	// --- When ---
	have := FMT()

	// --- Then ---
	assert.Equal(t, IDfmt, have.ID())
	assert.Equal(t, FMTChunkSize, have.Size())
	assert.Equal(t, uint32(0), have.Type())
	assert.False(t, have.Multi())
	assert.Nil(t, have.Chunks())
	assert.False(t, have.Raw())
	assert.False(t, have.WriteZeroExtra)
}

func Test_ChunkFMT_SetExtra_tabular(t *testing.T) {
	tt := []struct {
		testN string

		ch    func(tester.T) io.Reader
		extra []byte
		wSize uint32
	}{
		{"fmtWithoutExtraBytes1", fmtChunkWithoutExtraBytes, []byte{}, 16},
		{"fmtWithoutExtraBytes2", fmtChunkWithoutExtraBytes, []byte{5, 6}, 20},
		{
			"fmtWithoutExtraBytes3",
			fmtChunkWithoutExtraBytes,
			[]byte{5, 6, 7},
			22,
		},

		{"fmtWithExtraBytesEven1", fmtChunkWithExtraBytesEven, []byte{}, 16},
		{
			"fmtWithExtraBytesEven2",
			fmtChunkWithExtraBytesEven,
			[]byte{5, 6},
			20,
		},
		{
			"fmtWithExtraBytesEven3",
			fmtChunkWithExtraBytesEven,
			[]byte{5, 6, 7},
			22,
		},

		{"fmtWithExtraBytesOdd1", fmtChunkWithExtraBytesOdd, []byte{}, 16},
		{"fmtWithExtraBytesOdd2", fmtChunkWithExtraBytesOdd, []byte{5, 6}, 20},
		{
			"fmtWithExtraBytesOdd3",
			fmtChunkWithExtraBytesOdd,
			[]byte{5, 6, 7},
			22,
		},

		{"fmtWithExtraOfZeroLen1", fmtChunkWithExtraOfZeroLen, []byte{}, 18},
		{
			"fmtWithExtraOfZeroLen2",
			fmtChunkWithExtraOfZeroLen,
			[]byte{5, 6},
			20,
		},
		{
			"fmtWithExtraOfZeroLen3",
			fmtChunkWithExtraOfZeroLen,
			[]byte{5, 6, 7},
			22,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			src := tc.ch(t)
			skip4B(t, src)

			ch := FMT()
			must.Value(ch.ReadFrom(src))

			// --- When ---
			ch.SetExtra(tc.extra)

			// --- Then ---
			assert.Equal(t, tc.extra, must.Value(io.ReadAll(ch.Extra())))
			assert.Equal(t, tc.wSize, ch.Size())
		})
	}
}

func Test_ChunkFMT_Duration(t *testing.T) {
	t.Run("byte rate", func(t *testing.T) {
		// --- Given ---
		src := fmtChunkWithoutExtraBytes(t)
		skip4B(t, src)

		ch := FMT()
		must.Value(ch.ReadFrom(src))

		// --- When ---
		have := ch.Duration(88200)

		// --- Then ---
		assert.Equal(t, time.Second, have)
	})

	t.Run("zero byte rate", func(t *testing.T) {
		// --- Given ---
		ch := FMT()

		// --- When ---
		have := ch.Duration(88200)

		// --- Then ---
		assert.Equal(t, time.Duration(0), have)
	})
}

func Test_ChunkFMT_ReadFrom(t *testing.T) {
	t.Run("without extra bytes", func(t *testing.T) {
		// --- Given ---
		src := fmtChunkWithoutExtraBytes(t)
		skip4B(t, src) // Skip chunk ID.

		ch := FMT()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(20), n)
		assert.Equal(t, uint32(16), ch.Size())
		assert.Equal(t, CompPCM, ch.CompCode)
		assert.Equal(t, uint16(1), ch.ChannelCnt)
		assert.Equal(t, uint32(44100), ch.SampleRate)
		assert.Equal(t, uint32(88200), ch.AvgByteRate)
		assert.Equal(t, uint16(2), ch.BlockAlign)
		assert.Equal(t, uint16(16), ch.BitsPerSample)
		assert.Equal(t, []byte(nil), ch.extra)
		assert.True(t, isAllRead(src))
	})

	t.Run("even extra bytes", func(t *testing.T) {
		// --- Given ---
		src := fmtChunkWithExtraBytesEven(t)
		skip4B(t, src) // Skip chunk ID.

		ch := FMT()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(24), n)
		assert.Equal(t, uint32(20), ch.Size())
		assert.Equal(t, CompPCM, ch.CompCode)
		assert.Equal(t, uint16(1), ch.ChannelCnt)
		assert.Equal(t, uint32(44100), ch.SampleRate)
		assert.Equal(t, uint32(88200), ch.AvgByteRate)
		assert.Equal(t, uint16(2), ch.BlockAlign)
		assert.Equal(t, uint16(16), ch.BitsPerSample)
		assert.Equal(t, []byte{0, 1}, ch.extra)
		assert.True(t, isAllRead(src))
	})

	t.Run("odd extra bytes", func(t *testing.T) {
		// --- Given ---
		src := fmtChunkWithExtraBytesOdd(t)
		skip4B(t, src) // Skip chunk ID.

		ch := FMT()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(26), n)
		assert.Equal(t, uint32(22), ch.Size())
		assert.Equal(t, CompPCM, ch.CompCode)
		assert.Equal(t, uint16(1), ch.ChannelCnt)
		assert.Equal(t, uint32(44100), ch.SampleRate)
		assert.Equal(t, uint32(88200), ch.AvgByteRate)
		assert.Equal(t, uint16(2), ch.BlockAlign)
		assert.Equal(t, uint16(16), ch.BitsPerSample)
		assert.Equal(t, []byte{0, 1, 2}, ch.extra)
		assert.True(t, isAllRead(src))
	})

	t.Run("zero length extra bytes", func(t *testing.T) {
		// --- Given ---
		src := fmtChunkWithExtraOfZeroLen(t)
		skip4B(t, src) // Skip chunk ID.

		ch := FMT()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(22), n)
		assert.Equal(t, uint32(18), ch.Size())
		assert.Equal(t, CompPCM, ch.CompCode)
		assert.Equal(t, uint16(1), ch.ChannelCnt)
		assert.Equal(t, uint32(44100), ch.SampleRate)
		assert.Equal(t, uint32(88200), ch.AvgByteRate)
		assert.Equal(t, uint16(2), ch.BlockAlign)
		assert.Equal(t, uint16(16), ch.BitsPerSample)
		assert.Equal(t, []byte(nil), ch.extra)
		assert.True(t, isAllRead(src))
	})

	t.Run("error - size limit", func(t *testing.T) {
		// --- Given ---
		src := fmtChunkWithExtraBytesOdd(t)
		skip4B(t, src) // Skip chunk ID.

		ch := FMT(WithSizeLimit(21))

		// --- When ---
		_, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrTooLarge, err)
	})

	t.Run("error - truncated input", func(t *testing.T) {
		// Reading less than 26 bytes should always result in an error.
		for i := 1; i < 26; i++ {
			// --- Given ---
			src := fmtChunkWithExtraBytesOdd(t)
			skip4B(t, src) // Skip chunk ID.

			ch := FMT()

			// --- When ---
			_, err := ch.ReadFrom(io.LimitReader(src, int64(i)))

			// --- Then ---
			if !assert.Error(t, err) {
				t.Logf("errro i=%d", i)
			}
		}
	})

	t.Run("error - too short", func(t *testing.T) {
		// --- Given ---
		src := &bytes.Buffer{}
		writeUint32LE(t, src, 15)

		ch := FMT()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrTooShort, err)
		assert.ErrorEqual(t, "error decoding fmt  chunk: length too short", err)
		assert.Equal(t, int64(4), n)
	})

	t.Run("error - read", func(t *testing.T) {
		// --- Given ---
		src := fmtChunkWithExtraBytesOdd(t)
		skip4B(t, src) // Skip chunk ID.
		src = iokit.ErrReader(src, 24)
		ch := FMT()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, iokit.ErrRead, err)
		assert.Equal(t, int64(24), n)
	})

	t.Run("error - invalid even extra size", func(t *testing.T) {
		// --- Given ---
		src := fmtChunkWithExtraInvalidSizeEven(t)
		skip4B(t, src) // Skip chunk ID.

		ch := FMT()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrChunkSizeMismatch, err)
		want := "error decoding fmt  chunk: chunk size mismatch"
		assert.ErrorEqual(t, want, err)
		assert.Equal(t, int64(22), n)
	})

	t.Run("error - invalid odd extra size", func(t *testing.T) {
		// --- Given ---
		src := fmtChunkWithExtraInvalidSizeOdd(t)
		skip4B(t, src) // Skip chunk ID.

		ch := FMT()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrChunkSizeMismatch, err)
		want := "error decoding fmt  chunk: chunk size mismatch"
		assert.ErrorEqual(t, want, err)
		assert.Equal(t, int64(22), n)
	})
}

func Test_ChunkFMT_WriteTo(t *testing.T) {
	t.Run("error - extra bytes over uint16", func(t *testing.T) {
		// --- Given ---
		ch := FMT()
		ch.SetExtra(make([]byte, math.MaxUint16+1))
		dst := &bytes.Buffer{}

		// --- When ---
		n, err := ch.WriteTo(dst)

		// --- Then ---
		assert.ErrorIs(t, ErrTooLarge, err)
		assert.Equal(t, int64(0), n)
		assert.Equal(t, 0, dst.Len())
	})

	t.Run("error - write", func(t *testing.T) {
		// Writing less then 30 bytes should always result in an error.
		for i := 30; i > 0; i-- {
			// --- Given ---
			src := fmtChunkWithExtraBytesOdd(t)
			skip4B(t, src) // Skip chunk ID.

			ch := FMT()
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

func Test_ChunkFMT_WriteTo_tabular(t *testing.T) {
	// decoded returns a builder of a chunk decoded from the fixture.
	decoded := func(fix func(tester.T) io.Reader) func(*testing.T) *ChunkFMT {
		return func(t *testing.T) *ChunkFMT {
			t.Helper()
			src := fix(t)
			skip4B(t, src) // Skip chunk ID.

			ch := FMT()
			must.Value(ch.ReadFrom(src))
			return ch
		}
	}

	// created returns a builder of a new chunk with the extra bytes set.
	created := func(extra []byte) func(*testing.T) *ChunkFMT {
		return func(t *testing.T) *ChunkFMT {
			t.Helper()
			ch := FMT()
			ch.CompCode = CompPCM
			ch.ChannelCnt = 1
			ch.SampleRate = 44100
			ch.AvgByteRate = 88200
			ch.BlockAlign = 2
			ch.BitsPerSample = 16
			ch.SetExtra(extra)
			return ch
		}
	}

	tt := []struct {
		testN string

		ch   func(*testing.T) *ChunkFMT
		want func(tester.T) io.Reader
		n    int64
	}{
		{
			"decoded without extra bytes",
			decoded(fmtChunkWithoutExtraBytes),
			fmtChunkWithoutExtraBytes,
			24,
		},
		{
			"decoded with even extra bytes",
			decoded(fmtChunkWithExtraBytesEven),
			fmtChunkWithExtraBytesEven,
			28,
		},
		{
			"decoded with odd extra bytes",
			decoded(fmtChunkWithExtraBytesOdd),
			fmtChunkWithExtraBytesOdd,
			30,
		},
		{
			"decoded with zero length extra bytes",
			decoded(fmtChunkWithExtraOfZeroLen),
			fmtChunkWithExtraOfZeroLen,
			26,
		},
		{
			"created without extra bytes",
			created([]byte{}),
			fmtChunkWithoutExtraBytes,
			24,
		},
		{
			"created with even extra bytes",
			created([]byte{0, 1}),
			fmtChunkWithExtraBytesEven,
			28,
		},
		{
			"created with odd extra bytes",
			created([]byte{0, 1, 2}),
			fmtChunkWithExtraBytesOdd,
			30,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			ch := tc.ch(t)
			dst := &bytes.Buffer{}

			// --- When ---
			n, err := ch.WriteTo(dst)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.n, n)
			assert.Equal(t, uint32(n-8), ch.Size())
			assert.Equal(t, must.Value(io.ReadAll(tc.want(t))), dst.Bytes())
		})
	}
}

func Test_ChunkFMT_Reset(t *testing.T) {
	// --- Given ---
	src := fmtChunkWithExtraBytesOdd(t)
	skip4B(t, src) // Skip chunk ID.

	ch := FMT()
	must.Value(ch.ReadFrom(src))

	// --- When ---
	ch.Reset()

	// --- Then ---
	assert.Equal(t, FMTChunkSize, ch.Size())
	assert.Equal(t, CompNone, ch.CompCode)
	assert.Equal(t, uint16(0), ch.ChannelCnt)
	assert.Equal(t, uint32(0), ch.SampleRate)
	assert.Equal(t, uint32(0), ch.AvgByteRate)
	assert.Equal(t, uint16(0), ch.BlockAlign)
	assert.Equal(t, uint16(0), ch.BitsPerSample)
	assert.False(t, ch.WriteZeroExtra)
	assert.Len(t, 0, ch.extra)
	assert.True(t, isAllRead(src))
}
