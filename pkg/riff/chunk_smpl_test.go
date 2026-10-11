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

func smplWithoutLoopsNoData(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDsmpl) // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 36)     // ( 4) 4 - Chunk size
	writeUint32LE(t, src, 1)      // ( 8) 4 - Manufacturer
	writeUint32LE(t, src, 2)      // (12) 4 - Product
	writeUint32LE(t, src, 3)      // (16) 4 - SamplePeriod
	writeUint32LE(t, src, 4)      // (20) 4 - MIDIUnityNote
	writeUint32LE(t, src, 5)      // (24) 4 - MIDIPitchFraction
	writeUint32LE(t, src, 6)      // (28) 4 - SMPTEFormat
	writeUint32LE(t, src, 7)      // (32) 4 - SMPTEOffset
	writeUint32LE(t, src, 0)      // (36) 4 - SampleLoopCnt
	writeUint32LE(t, src, 0)      // (40) 4 - SamplerData
	// Total length: 8+36=44
	return src
}

func smplWithoutLoopsWithData(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDsmpl)       // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 39)           // ( 4) 4 - Chunk size
	writeUint32LE(t, src, 1)            // ( 8) 4 - Manufacturer
	writeUint32LE(t, src, 2)            // (12) 4 - Product
	writeUint32LE(t, src, 3)            // (16) 4 - SamplePeriod
	writeUint32LE(t, src, 4)            // (20) 4 - MIDIUnityNote
	writeUint32LE(t, src, 5)            // (24) 4 - MIDIPitchFraction
	writeUint32LE(t, src, 6)            // (28) 4 - SMPTEFormat
	writeUint32LE(t, src, 7)            // (32) 4 - SMPTEOffset
	writeUint32LE(t, src, 0)            // (36) 4 - SampleLoopCnt
	writeUint32LE(t, src, 3)            // (40) 4 - SamplerData
	writeBytes(t, src, []byte{0, 1, 2}) // (44) 3 - Data
	writeBytes(t, src, []byte{0})       // (47) 1 - Padding byte
	// Total length: 8+36+3+1=48
	return src
}

func smplWithLoopsNoData(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDsmpl) // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 60)     // ( 4) 4 - Chunk size
	writeUint32LE(t, src, 1)      // ( 8) 4 - Manufacturer
	writeUint32LE(t, src, 2)      // (12) 4 - Product
	writeUint32LE(t, src, 3)      // (16) 4 - SamplePeriod
	writeUint32LE(t, src, 4)      // (20) 4 - MIDIUnityNote
	writeUint32LE(t, src, 5)      // (24) 4 - MIDIPitchFraction
	writeUint32LE(t, src, 6)      // (28) 4 - SMPTEFormat
	writeUint32LE(t, src, 7)      // (32) 4 - SMPTEOffset
	writeUint32LE(t, src, 1)      // (36) 4 - SampleLoopCnt
	writeUint32LE(t, src, 0)      // (40) 4 - SamplerData
	writeUint32LE(t, src, 9)      // (44) 4 - CuePointID
	writeUint32LE(t, src, 10)     // (48) 4 - Type
	writeUint32LE(t, src, 11)     // (52) 4 - Start
	writeUint32LE(t, src, 12)     // (56) 4 - End
	writeUint32LE(t, src, 13)     // (60) 4 - Fraction
	writeUint32LE(t, src, 14)     // (64) 4 - PlayCnt
	// Total length: 8+36+24=68
	return src
}

func smplWithLoopsWithData(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDsmpl)       // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 63)           // ( 4) 4 - Chunk size
	writeUint32LE(t, src, 1)            // ( 8) 4 - Manufacturer
	writeUint32LE(t, src, 2)            // (12) 4 - Product
	writeUint32LE(t, src, 3)            // (16) 4 - SamplePeriod
	writeUint32LE(t, src, 4)            // (20) 4 - MIDIUnityNote
	writeUint32LE(t, src, 5)            // (24) 4 - MIDIPitchFraction
	writeUint32LE(t, src, 6)            // (28) 4 - SMPTEFormat
	writeUint32LE(t, src, 7)            // (32) 4 - SMPTEOffset
	writeUint32LE(t, src, 1)            // (36) 4 - SampleLoopCnt
	writeUint32LE(t, src, 3)            // (40) 4 - SamplerData
	writeUint32LE(t, src, 9)            // (44) 4 - CuePointID
	writeUint32LE(t, src, 10)           // (48) 4 - Type
	writeUint32LE(t, src, 11)           // (52) 4 - Start
	writeUint32LE(t, src, 12)           // (56) 4 - End
	writeUint32LE(t, src, 13)           // (60) 4 - Fraction
	writeUint32LE(t, src, 14)           // (64) 4 - PlayCnt
	writeBytes(t, src, []byte{0, 1, 2}) // (68) 3 - Data
	writeBytes(t, src, []byte{0})       // (71) 1 - Padding byte
	// Total length: 8+36+24+3+1=72
	return src
}

// smplInvalidSize declares one SampleLoop but the chunk size declares only
// 36 bytes.
func smplInvalidSize(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDsmpl) // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 36)     // ( 4) 4 - Chunk size
	writeUint32LE(t, src, 1)      // ( 8) 4 - Manufacturer
	writeUint32LE(t, src, 2)      // (12) 4 - Product
	writeUint32LE(t, src, 3)      // (16) 4 - SamplePeriod
	writeUint32LE(t, src, 4)      // (20) 4 - MIDIUnityNote
	writeUint32LE(t, src, 5)      // (24) 4 - MIDIPitchFraction
	writeUint32LE(t, src, 6)      // (28) 4 - SMPTEFormat
	writeUint32LE(t, src, 7)      // (32) 4 - SMPTEOffset
	writeUint32LE(t, src, 1)      // (36) 4 - SampleLoopCnt
	writeUint32LE(t, src, 0)      // (40) 4 - SamplerData
	// Total length: 8+36=44
	return src
}

func Test_SMPL(t *testing.T) {
	// --- When ---
	have := SMPL()

	// --- Then ---
	assert.Equal(t, IDsmpl, have.ID())
	assert.Equal(t, SMPLChunkSize, have.Size())
	assert.Equal(t, uint32(0), have.Type())
	assert.False(t, have.Multi())
	assert.Nil(t, have.Chunks())
	assert.False(t, have.Raw())
	assert.Equal(t, uint32(0), have.Manufacturer)
	assert.Equal(t, uint32(0), have.Product)
	assert.Equal(t, uint32(0), have.SamplePeriod)
	assert.Equal(t, uint32(0), have.MIDIUnityNote)
	assert.Equal(t, uint32(0), have.MIDIPitchFraction)
	assert.Equal(t, uint32(0), have.SMPTEFormat)
	assert.Equal(t, uint32(0), have.SMPTEOffset)
	assert.Equal(t, uint32(0), have.SampleLoopCnt)
	assert.Equal(t, uint32(0), have.SamplerDataCnt)
	assert.Equal(t, []byte{}, must.Value(io.ReadAll(have.SamplerData())))
	assert.Len(t, 0, have.SampleLoops)
}

func Test_ChunkSMPL_Size(t *testing.T) {
	t.Run("new chunk", func(t *testing.T) {
		// --- Given ---
		ch := SMPL()

		// --- When ---
		have := ch.Size()

		// --- Then ---
		assert.Equal(t, SMPLChunkSize, have)
	})

	t.Run("loop appended", func(t *testing.T) {
		// --- Given ---
		ch := SMPL()
		ch.SampleLoops = append(ch.SampleLoops, &SampleLoop{})
		dst := &bytes.Buffer{}

		// --- When ---
		have := ch.Size()

		// --- Then ---
		assert.Equal(t, uint32(60), have)
		n, err := ch.WriteTo(dst)
		assert.NoError(t, err)
		assert.Equal(t, int64(have)+8, n)
	})
}

func Test_ChunkSMPL_ReadFrom(t *testing.T) {
	t.Run("without loops", func(t *testing.T) {
		// --- Given ---
		src := smplWithoutLoopsNoData(t)
		skip4B(t, src) // Skip chunk ID.

		ch := SMPL()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(40), n)
		assert.Equal(t, uint32(36), ch.Size())
		assert.Equal(t, uint32(1), ch.Manufacturer)
		assert.Equal(t, uint32(2), ch.Product)
		assert.Equal(t, uint32(3), ch.SamplePeriod)
		assert.Equal(t, uint32(4), ch.MIDIUnityNote)
		assert.Equal(t, uint32(5), ch.MIDIPitchFraction)
		assert.Equal(t, uint32(6), ch.SMPTEFormat)
		assert.Equal(t, uint32(7), ch.SMPTEOffset)
		assert.Equal(t, uint32(0), ch.SampleLoopCnt)
		assert.Equal(t, uint32(0), ch.SamplerDataCnt)
		assert.Len(t, 0, ch.SampleLoops)
		assert.Equal(t, []byte{}, must.Value(io.ReadAll(ch.SamplerData())))
		assert.True(t, isAllRead(src))
	})

	t.Run("without loops with data", func(t *testing.T) {
		// --- Given ---
		src := smplWithoutLoopsWithData(t)
		skip4B(t, src) // Skip chunk ID.

		ch := SMPL()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(44), n)
		assert.Equal(t, uint32(39), ch.Size())
		assert.Equal(t, uint32(1), ch.Manufacturer)
		assert.Equal(t, uint32(2), ch.Product)
		assert.Equal(t, uint32(3), ch.SamplePeriod)
		assert.Equal(t, uint32(4), ch.MIDIUnityNote)
		assert.Equal(t, uint32(5), ch.MIDIPitchFraction)
		assert.Equal(t, uint32(6), ch.SMPTEFormat)
		assert.Equal(t, uint32(7), ch.SMPTEOffset)
		assert.Equal(t, uint32(0), ch.SampleLoopCnt)
		assert.Equal(t, uint32(3), ch.SamplerDataCnt)
		assert.Len(t, 0, ch.SampleLoops)
		data := must.Value(io.ReadAll(ch.SamplerData()))
		assert.Equal(t, []byte{0, 1, 2}, data)
		assert.True(t, isAllRead(src))
	})

	t.Run("with loops", func(t *testing.T) {
		// --- Given ---
		src := smplWithLoopsNoData(t)
		skip4B(t, src) // Skip chunk ID.

		ch := SMPL()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(64), n)
		assert.Equal(t, uint32(60), ch.Size())
		assert.Equal(t, uint32(1), ch.Manufacturer)
		assert.Equal(t, uint32(2), ch.Product)
		assert.Equal(t, uint32(3), ch.SamplePeriod)
		assert.Equal(t, uint32(4), ch.MIDIUnityNote)
		assert.Equal(t, uint32(5), ch.MIDIPitchFraction)
		assert.Equal(t, uint32(6), ch.SMPTEFormat)
		assert.Equal(t, uint32(7), ch.SMPTEOffset)
		assert.Equal(t, uint32(1), ch.SampleLoopCnt)
		assert.Equal(t, uint32(0), ch.SamplerDataCnt)
		assert.Len(t, 1, ch.SampleLoops)
		assert.Equal(t, uint32(9), ch.SampleLoops[0].CuePointID)
		assert.Equal(t, uint32(10), ch.SampleLoops[0].Type)
		assert.Equal(t, uint32(11), ch.SampleLoops[0].Start)
		assert.Equal(t, uint32(12), ch.SampleLoops[0].End)
		assert.Equal(t, uint32(13), ch.SampleLoops[0].Fraction)
		assert.Equal(t, uint32(14), ch.SampleLoops[0].PlayCnt)
		assert.Equal(t, []byte{}, must.Value(io.ReadAll(ch.SamplerData())))
		assert.True(t, isAllRead(src))
	})

	t.Run("with loops with data", func(t *testing.T) {
		// --- Given ---
		src := smplWithLoopsWithData(t)
		skip4B(t, src) // Skip chunk ID.

		ch := SMPL()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(68), n)
		assert.Equal(t, uint32(63), ch.Size())
		assert.Equal(t, uint32(1), ch.Manufacturer)
		assert.Equal(t, uint32(2), ch.Product)
		assert.Equal(t, uint32(3), ch.SamplePeriod)
		assert.Equal(t, uint32(4), ch.MIDIUnityNote)
		assert.Equal(t, uint32(5), ch.MIDIPitchFraction)
		assert.Equal(t, uint32(6), ch.SMPTEFormat)
		assert.Equal(t, uint32(7), ch.SMPTEOffset)
		assert.Equal(t, uint32(1), ch.SampleLoopCnt)
		assert.Equal(t, uint32(3), ch.SamplerDataCnt)
		assert.Len(t, 1, ch.SampleLoops)
		assert.Equal(t, uint32(9), ch.SampleLoops[0].CuePointID)
		assert.Equal(t, uint32(10), ch.SampleLoops[0].Type)
		assert.Equal(t, uint32(11), ch.SampleLoops[0].Start)
		assert.Equal(t, uint32(12), ch.SampleLoops[0].End)
		assert.Equal(t, uint32(13), ch.SampleLoops[0].Fraction)
		assert.Equal(t, uint32(14), ch.SampleLoops[0].PlayCnt)
		data := must.Value(io.ReadAll(ch.SamplerData()))
		assert.Equal(t, []byte{0, 1, 2}, data)
		assert.True(t, isAllRead(src))
	})

	t.Run("held loop not reused", func(t *testing.T) {
		// --- Given ---
		src := smplWithLoopsNoData(t)
		skip4B(t, src) // Skip chunk ID.

		ch := SMPL()
		must.Value(ch.ReadFrom(src))
		held := ch.SampleLoops[0]
		ch.Reset()

		src = smplWithLoopsNoData(t)
		skip4B(t, src) // Skip chunk ID.

		other := SMPL()

		// --- When ---
		_, err := other.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.NotSame(t, held, other.SampleLoops[0])
	})

	t.Run("error - truncated input", func(t *testing.T) {
		// Reading less than 64 bytes should always result in an error.
		for i := 1; i < 64; i++ {
			// --- Given ---
			src := smplWithLoopsNoData(t)
			skip4B(t, src) // Skip chunk ID.

			ch := SMPL()

			// --- When ---
			_, err := ch.ReadFrom(io.LimitReader(src, int64(i)))

			// --- Then ---
			assert.Error(t, err)
		}
	})

	t.Run("error - too short", func(t *testing.T) {
		// --- Given ---
		src := &bytes.Buffer{}
		writeUint32LE(t, src, 35)

		ch := SMPL()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrTooShort, err)
		assert.ErrorEqual(t, "error decoding smpl chunk: length too short", err)
		assert.Equal(t, int64(4), n)
	})

	t.Run("error - invalid size", func(t *testing.T) {
		// --- Given ---
		src := smplInvalidSize(t)
		skip4B(t, src) // Skip chunk ID.

		ch := SMPL()

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrChunkSizeMismatch, err)
		want := "error decoding smpl chunk: chunk size mismatch"
		assert.ErrorEqual(t, want, err)
		assert.Equal(t, int64(40), n)
	})

	t.Run("error - truncated sampler data", func(t *testing.T) {
		// --- Given ---
		src := smplWithoutLoopsWithData(t)
		skip4B(t, src) // Skip chunk ID.

		// --- When ---
		n, err := SMPL().ReadFrom(io.LimitReader(src, 42))

		// --- Then ---
		assert.ErrorEqual(t, "error decoding smpl chunk: unexpected EOF", err)
		assert.Equal(t, int64(42), n)
	})

	t.Run("error - truncated loop", func(t *testing.T) {
		// --- Given ---
		src := smplWithLoopsNoData(t)
		skip4B(t, src) // Skip chunk ID.

		// --- When ---
		n, err := SMPL().ReadFrom(io.LimitReader(src, 50))

		// --- Then ---
		assert.ErrorEqual(t, "error decoding smpl chunk: unexpected EOF", err)
		assert.Equal(t, int64(50), n)
	})

	t.Run("error - loop count overflow", func(t *testing.T) {
		// --- Given ---
		src := &bytes.Buffer{}
		writeUint32LE(t, src, 44) // ( 0) 4 - Chunk size
		for range 7 {
			writeUint32LE(t, src, 0) // ( 4) 28 - Manufacturer to SMPTEOffset
		}
		writeUint32LE(t, src, 0x0AAAAAAB) // (32) 4 - SampleLoopCnt
		writeUint32LE(t, src, 0)          // (36) 4 - SamplerData
		writeBytes(t, src, make([]byte, 8))

		// --- When ---
		_, err := SMPL().ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrChunkSizeMismatch, err)
	})

	t.Run("error - size limit", func(t *testing.T) {
		// --- Given ---
		src := smplWithoutLoopsWithData(t)
		skip4B(t, src) // Skip chunk ID.

		ch := SMPL(WithSizeLimit(39))

		// --- When ---
		_, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrTooLarge, err)
	})

	t.Run("missing padding at end of input", func(t *testing.T) {
		// --- Given ---
		src := smplWithoutLoopsWithData(t)
		skip4B(t, src) // Skip chunk ID.

		ch := SMPL()

		// --- When ---
		n, err := ch.ReadFrom(io.LimitReader(src, 43))

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(43), n)
	})
}

func Test_ChunkSMPL_WriteTo(t *testing.T) {
	t.Run("error - write", func(t *testing.T) {
		// Writing less then 64 bytes should always result in an error.
		for i := 64; i > 0; i-- {
			// --- Given ---
			src := smplWithLoopsNoData(t)
			skip4B(t, src) // Skip chunk ID.

			ch := SMPL()
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

	t.Run("error - writing sampler data", func(t *testing.T) {
		// --- Given ---
		src := smplWithoutLoopsWithData(t)
		skip4B(t, src) // Skip chunk ID.

		ch := SMPL()
		must.Value(ch.ReadFrom(src))

		// --- When ---
		_, err := ch.WriteTo(iokit.ErrWriter(&bytes.Buffer{}, 47))

		// --- Then ---
		assert.ErrorEqual(t, "error encoding smpl chunk: write error", err)
		assert.ErrorIs(t, iokit.ErrWrite, err)
	})

	t.Run("error - writing padding", func(t *testing.T) {
		// --- Given ---
		src := smplWithoutLoopsWithData(t)
		skip4B(t, src) // Skip chunk ID.

		ch := SMPL()
		must.Value(ch.ReadFrom(src))

		// --- When ---
		_, err := ch.WriteTo(iokit.ErrWriter(&bytes.Buffer{}, 48))

		// --- Then ---
		assert.ErrorEqual(t, "error encoding smpl chunk: write error", err)
		assert.ErrorIs(t, iokit.ErrWrite, err)
	})
}

func Test_ChunkSMPL_WriteTo_tabular(t *testing.T) {
	tt := []struct {
		testN string

		n  int64
		ch func(tester.T) io.Reader
	}{
		{"smplWithoutLoopsNoData", 44, smplWithoutLoopsNoData},
		{"smplWithoutLoopsWithData", 48, smplWithoutLoopsWithData},
		{"smplWithLoopsNoData", 68, smplWithLoopsNoData},
		{"smplWithLoopsWithData", 72, smplWithLoopsWithData},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			src := tc.ch(t)
			skip4B(t, src) // Skip chunk ID.

			ch := SMPL()
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

func Test_ChunkSMPL_Reset(t *testing.T) {
	// --- Given ---
	ch := SMPL()
	ch.Manufacturer = 2
	ch.Product = 3
	ch.SamplePeriod = 4
	ch.MIDIUnityNote = 5
	ch.MIDIPitchFraction = 6
	ch.SMPTEFormat = 7
	ch.SMPTEOffset = 8
	ch.SampleLoopCnt = 9
	ch.SamplerDataCnt = 10
	ch.SampleLoops = []*SampleLoop{
		{
			CuePointID: 11,
			Type:       12,
			Start:      13,
			End:        14,
			Fraction:   15,
			PlayCnt:    16,
		},
	}
	ch.sampleData = []byte{0, 1, 2}

	// --- When ---
	ch.Reset()

	// --- Then ---
	assert.Equal(t, uint32(0), ch.Manufacturer)
	assert.Equal(t, uint32(0), ch.Product)
	assert.Equal(t, uint32(0), ch.SamplePeriod)
	assert.Equal(t, uint32(0), ch.MIDIUnityNote)
	assert.Equal(t, uint32(0), ch.MIDIPitchFraction)
	assert.Equal(t, uint32(0), ch.SMPTEFormat)
	assert.Equal(t, uint32(0), ch.SMPTEOffset)
	assert.Equal(t, uint32(0), ch.SampleLoopCnt)
	assert.Equal(t, uint32(0), ch.SamplerDataCnt)
	assert.Len(t, 0, ch.SampleLoops)
	assert.Equal(t, []byte{}, must.Value(io.ReadAll(ch.SamplerData())))
}
