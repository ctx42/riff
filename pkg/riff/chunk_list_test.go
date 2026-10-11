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

func listChunkTypeINFO(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDLIST)                // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 16)                    // ( 4) 4 - Chunk size
	writeUint32BE(t, src, IDINFO)                // ( 8) 4 - Type
	writeUint32BE(t, src, LabIART)               // (12) 4 - Chunk ID
	writeUint32LE(t, src, 4)                     // (16) 4 - Chunk size
	writeBytes(t, src, []byte{'a', 'b', 'c', 0}) // (20) 4 - Text
	// Total length: 24
	return src
}

func listChunkTypeAdtl(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDLIST)                // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 52)                    // ( 4) 4 - Chunk size
	writeUint32BE(t, src, IDadtl)                // ( 8) 4 - Type
	writeUint32BE(t, src, IDlabl)                // (12) 4 - Chunk ID
	writeUint32LE(t, src, 8)                     // (16) 4 - Chunk size
	writeUint32LE(t, src, 123)                   // (20) 4 - Cue ID
	writeBytes(t, src, []byte{'a', 'b', 'c', 0}) // (24) 4 - Text
	writeUint32BE(t, src, IDltxt)                // (28) 4 - Chunk ID
	writeUint32LE(t, src, 24)                    // (32) 4 - Chunk size
	writeUint32LE(t, src, 1)                     // (36) 4 - PID
	writeUint32LE(t, src, 2)                     // (40) 4 - SamLen
	writeUint32LE(t, src, 3)                     // (44) 4 - PurID
	writeUint16LE(t, src, 4)                     // (48) 2 - Country
	writeUint16LE(t, src, 5)                     // (50) 2 - Language
	writeUint16LE(t, src, 6)                     // (52) 2 - Dialect
	writeUint16LE(t, src, 7)                     // (54) 2 - CodePage
	writeBytes(t, src, []byte{'a', 'b', 'c', 0}) // (56) 4 - Text
	// Total length: 60
	return src
}

func listChunkTypeUnknown(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDLIST)                // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 16)                    // ( 4) 4 - Chunk size
	writeUint32BE(t, src, IDUNKN)                // ( 8) 4 - Type
	writeUint32BE(t, src, LabIART)               // (12) 4 - Chunk ID
	writeUint32LE(t, src, 4)                     // (16) 4 - Chunk size
	writeBytes(t, src, []byte{'a', 'b', 'c', 0}) // (20) 4 - Text
	// Total length: 24
	return src
}

func Test_LIST(t *testing.T) {
	// --- Given ---
	reg := NewRegistry(RAWCMake(WithLoadData()))

	// --- When ---
	have := LIST(reg, WithLoadData())

	// --- Then ---
	assert.Equal(t, IDLIST, have.ID())
	assert.Equal(t, uint32(4), have.Size())
	assert.Equal(t, uint32(0), have.Type())
	assert.True(t, have.Multi())
	assert.Nil(t, have.Chunks())
	assert.False(t, have.Raw())
}

func Test_ChunkLIST_Size(t *testing.T) {
	t.Run("no chunks", func(t *testing.T) {
		// --- Given ---
		ch := LIST(nil)

		// --- When ---
		have := ch.Size()

		// --- Then ---
		assert.Equal(t, uint32(4), have)
	})

	t.Run("edited sub-chunk", func(t *testing.T) {
		// --- Given ---
		fmtc := FMT()
		ch := LIST(nil)
		ch.Modify(Chunks{fmtc})
		fmtc.SetExtra([]byte{1, 2})

		// --- When ---
		have := ch.Size()

		// --- Then ---
		assert.Equal(t, uint32(4+8+16+4), have)
	})
}

func Test_ChunkLIST_ReadFrom(t *testing.T) {
	t.Run("missing final padding at end of input", func(t *testing.T) {
		// --- Given ---
		src := &bytes.Buffer{}
		writeUint32LE(t, src, 4+8+3+1)          // Chunk size.
		writeUint32BE(t, src, IDINFO)           // List type.
		writeUint32BE(t, src, LabIART)          // Sub-chunk ID.
		writeUint32LE(t, src, 3)                // Sub-chunk size.
		writeBytes(t, src, []byte{'a', 'b', 0}) // Text, no padding byte.

		ch := LIST(nil)

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(19), n)
		assert.Equal(t, uint32(16), ch.Size())
	})

	t.Run("error - input ends before declared size", func(t *testing.T) {
		// --- Given ---
		src := &bytes.Buffer{}
		writeUint32LE(t, src, 4+8+3+1+8)        // Chunk size.
		writeUint32BE(t, src, IDINFO)           // List type.
		writeUint32BE(t, src, LabIART)          // Sub-chunk ID.
		writeUint32LE(t, src, 3)                // Sub-chunk size.
		writeBytes(t, src, []byte{'a', 'b', 0}) // Text, no padding byte.

		ch := LIST(nil)

		// --- When ---
		_, err := ch.ReadFrom(src)

		// --- Then ---
		want := "error decoding LIST:INFO chunk: unexpected EOF"
		assert.ErrorEqual(t, want, err)
		assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
	})
	t.Run("INFO list", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake(WithLoadData()))

		src := listChunkTypeINFO(t)
		skip4B(t, src) // Skip chunk ID.

		ch := LIST(reg, WithLoadData())

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(20), n)
		assert.Equal(t, IDINFO, ch.Type())
		assert.Len(t, 1, ch.Chunks())
		sub := ch.Chunks()[0]
		assert.SameType(t, &ChunkINFO{}, sub)
		assert.Equal(t, LabIART, sub.ID())
	})

	t.Run("adtl list", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake(WithLoadData()))

		src := listChunkTypeAdtl(t)
		skip4B(t, src) // Skip chunk ID.

		ch := LIST(reg, WithLoadData())

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(56), n)
		assert.Equal(t, IDadtl, ch.Type())
		assert.Len(t, 2, ch.Chunks())
		sub := ch.Chunks()[0]
		assert.SameType(t, &ChunkLABL{}, sub)
		assert.Equal(t, IDlabl, sub.ID())
		sub = ch.Chunks()[1]
		assert.SameType(t, &ChunkLTXT{}, sub)
		assert.Equal(t, IDltxt, sub.ID())
	})

	t.Run("adtl list with registered decoder", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake(WithLoadData()))
		reg.Register(IDlabl, func() Chunk {
			return RAWC(IDlabl, WithLoadData())
		})

		src := listChunkTypeAdtl(t)
		skip4B(t, src) // Skip chunk ID.

		ch := LIST(reg, WithLoadData())

		// --- When ---
		_, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Len(t, 2, ch.Chunks())
		assert.SameType(t, &ChunkRAWC{}, ch.Chunks()[0])
		assert.SameType(t, &ChunkLTXT{}, ch.Chunks()[1])
		assert.False(t, reg.Has(IDltxt))
	})

	t.Run("unknown list type", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake(WithLoadData()))

		src := listChunkTypeUnknown(t)
		skip4B(t, src) // Skip chunk ID.

		ch := LIST(reg, WithLoadData())

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(20), n)
		assert.Equal(t, IDUNKN, ch.Type())
		assert.Len(t, 1, ch.Chunks())
		sub := ch.Chunks()[0]
		assert.SameType(t, &ChunkRAWC{}, sub)
		assert.Equal(t, LabIART, sub.ID())
	})

	t.Run("empty list", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake(WithLoadData()))

		src := &bytes.Buffer{}
		writeUint32LE(t, src, 4)      // ( 0) 4 - Chunk size
		writeUint32BE(t, src, IDINFO) // ( 4) 4 - Type

		ch := LIST(reg, WithLoadData())

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(8), n)
		assert.Equal(t, IDINFO, ch.Type())
		assert.Len(t, 0, ch.Chunks())
	})

	t.Run("nil registry", func(t *testing.T) {
		// --- Given ---
		src := listChunkTypeINFO(t)
		skip4B(t, src) // Skip chunk ID.

		ch := LIST(nil, WithLoadData())

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(20), n)
		assert.Len(t, 1, ch.Chunks())
		ch.Reset()
	})

	t.Run("error - sub-chunks overrun size", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake(WithLoadData()))

		src := &bytes.Buffer{}
		writeUint32LE(t, src, 13)        // ( 0) 4 - Chunk size
		writeUint32BE(t, src, IDUNKN)    // ( 4) 4 - Type
		writeUint32BE(t, src, IDUNKN)    // ( 8) 4 - Chunk ID
		writeUint32LE(t, src, 1)         // (12) 4 - Chunk size
		writeBytes(t, src, []byte{1, 0}) // (16) 2 - Data and padding

		ch := LIST(reg, WithLoadData())

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		want := "error decoding LIST:ABCD chunk: chunk size mismatch"
		assert.ErrorEqual(t, want, err)
		assert.ErrorIs(t, ErrChunkSizeMismatch, err)
		assert.Equal(t, int64(18), n)
	})

	t.Run("error - size limit", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake(WithLoadData()))

		src := listChunkTypeAdtl(t)
		skip4B(t, src) // Skip chunk ID.

		ch := LIST(reg, WithLoadData(), WithSizeLimit(51))

		// --- When ---
		_, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrTooLarge, err)
	})

	t.Run("error - truncated input", func(t *testing.T) {
		// Reading less than 20 bytes should always result in an error.
		for i := 1; i < 20; i++ {
			// --- Given ---
			reg := NewRegistry(RAWCMake(WithLoadData()))
			src := listChunkTypeINFO(t)
			skip4B(t, src) // Skip chunk ID.

			ch := LIST(reg, WithLoadData())

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
		reg := NewRegistry(RAWCMake(WithLoadData()))
		src := &bytes.Buffer{}
		writeUint32LE(t, src, 3)
		ch := LIST(reg, WithLoadData())

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrTooShort, err)
		assert.ErrorEqual(t, "error decoding LIST chunk: length too short", err)
		assert.Equal(t, int64(4), n)
	})
}

func Test_ChunkLIST_WriteTo(t *testing.T) {
	t.Run("error - sub-chunks over 4 GiB", func(t *testing.T) {
		// --- Given ---
		ch := LIST(nil, WithLoadData())
		ch.Modify(Chunks{&sizedChunk{1 << 31}, &sizedChunk{1 << 31}})
		dst := &bytes.Buffer{}

		// --- When ---
		n, err := ch.WriteTo(dst)

		// --- Then ---
		assert.ErrorEqual(t, "error encoding LIST chunk: size too large", err)
		assert.ErrorIs(t, ErrTooLarge, err)
		assert.Equal(t, int64(0), n)
		assert.Equal(t, 0, dst.Len())
	})
	t.Run("error - write", func(t *testing.T) {
		// Writing less than 60 bytes should always result in an error.
		for i := 60; i > 0; i-- {
			// --- Given ---
			reg := NewRegistry(RAWCMake(WithLoadData()))

			src := listChunkTypeAdtl(t)
			skip4B(t, src) // Skip chunk ID.

			ch := LIST(reg, WithLoadData())
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

	t.Run("error - skip data mode", func(t *testing.T) {
		// --- Given ---
		ch := LIST(nil)

		// --- When ---
		n, err := ch.WriteTo(&bytes.Buffer{})

		// --- Then ---
		assert.ErrorIs(t, ErrSkipDataMode, err)
		assert.Equal(t, int64(0), n)
	})
}

func Test_ChunkLIST_WriteTo_tabular(t *testing.T) {
	tt := []struct {
		testN string

		n  int64
		ch func(tester.T) io.Reader
	}{
		{"listChunkTypeINFO", 24, listChunkTypeINFO},
		{"listChunkTypeAdtl", 60, listChunkTypeAdtl},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			reg := NewRegistry(RAWCMake(WithLoadData()))

			src := tc.ch(t)
			skip4B(t, src) // Skip chunk ID.

			ch := LIST(reg, WithLoadData())
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

func Test_ChunkLIST_Reset(t *testing.T) {
	t.Run("decoded chunks", func(t *testing.T) {
		reg := NewRegistry(RAWCMake(WithLoadData()))

		src := listChunkTypeINFO(t)
		skip4B(t, src) // Skip chunk ID.

		ch := LIST(reg, WithLoadData())
		must.Value(ch.ReadFrom(src))

		// --- When ---
		ch.Reset()

		// --- Then ---
		assert.Equal(t, uint32(4), ch.Size())
		assert.Equal(t, uint32(0), ch.ListType)
		assert.Len(t, 0, ch.Chunks())
	})

	t.Run("modified chunks not pooled", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake(WithLoadData()))
		ch := INFO(LabIART)

		lst := LIST(reg, WithLoadData())
		lst.Modify(Chunks{ch})

		// --- When ---
		lst.Reset()

		// --- Then ---
		assert.Nil(t, reg.GetNoRaw(LabIART))
	})
}

func Test_ChunkLIST_subMaker_tabular(t *testing.T) {
	tt := []struct {
		testN string

		typ  uint32
		want Chunk
	}{
		{"INFO", IDINFO, &ChunkINFO{}},
		{"adtl", IDadtl, &ChunkLABL{}},
		{"other", IDUNKN, &ChunkRAWC{}},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			ch := LIST(nil)
			ch.ListType = tc.typ

			// --- When ---
			have := ch.subMaker()

			// --- Then ---
			assert.SameType(t, tc.want, have(IDlabl))
		})
	}
}

func Test_adtlMake_tabular(t *testing.T) {
	tt := []struct {
		testN string

		id   uint32
		want Chunk
	}{
		{"labl", IDlabl, &ChunkLABL{}},
		{"ltxt", IDltxt, &ChunkLTXT{}},
		{"other", IDUNKN, &ChunkRAWC{}},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			mkr := adtlMake(WithLoadData())

			// --- When ---
			have := mkr(tc.id)

			// --- Then ---
			assert.SameType(t, tc.want, have)
			assert.Equal(t, tc.id, have.ID())
		})
	}
}

func Test_ChunkLIST_Modify(t *testing.T) {
	t.Run("replaces chunks", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake(WithLoadData()))

		src := listChunkTypeAdtl(t)
		skip4B(t, src) // Skip chunk ID.

		ch := LIST(reg, WithLoadData())
		_, _ = ch.ReadFrom(src)

		size := ch.Size()
		chs := ch.Chunks()
		txt := chs.First(IDltxt)

		// --- When ---
		ch.Modify(chs.Remove(IDltxt))

		// --- Then ---
		assert.Equal(t, size-(txt.Size()+8), ch.Size())
		assert.Equal(t, 1, len(ch.Chunks().IDs()))
		assert.Nil(t, ch.Chunks().First(IDltxt))
	})
}
