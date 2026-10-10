// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"io"
	"runtime"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"
	"github.com/ctx42/testkit/pkg/iokit"
)

func infoChunkTextLenEven(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, LabIART)               // (0) 4 - Chunk ID
	writeUint32LE(t, src, 4)                     // (4) 4 - Chunk size
	writeBytes(t, src, []byte{'a', 'b', 'c', 0}) // (8) 4 - Text
	// Total length: 8+4=12
	return src
}

func infoChunkTextLenOdd(t tester.T) io.Reader {
	t.Helper()
	src := &bytes.Buffer{}
	writeUint32BE(t, src, LabIART)          // ( 0) 4 - Chunk ID
	writeUint32LE(t, src, 3)                // ( 4) 4 - Chunk size
	writeBytes(t, src, []byte{'a', 'b', 0}) // ( 8) 4 - Text
	writeBytes(t, src, []byte{0})           // (11) 1 - Padding byte
	// Total length: 8+3+1=12
	return src
}

func Test_INFO(t *testing.T) {
	// --- When ---
	have := INFO(LabIART)

	// --- Then ---
	assert.Equal(t, LabIART, have.ID())
	assert.Equal(t, uint32(0), have.Size())
	assert.Equal(t, uint32(0), have.Type())
	assert.True(t, have.Multi())
	assert.Nil(t, have.Chunks())
	assert.False(t, have.Raw())
}

func Test_ChunkINFO_ReadFrom(t *testing.T) {
	t.Run("even text length", func(t *testing.T) {
		// --- Given ---
		src := infoChunkTextLenEven(t)
		skip4B(t, src) // Skip chunk ID.

		ch := INFO(LabIART)

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(8), n)
		assert.Equal(t, LabIART, ch.ID())
		assert.Equal(t, uint32(4), ch.Size())
		assert.Equal(t, []byte{'a', 'b', 'c', 0}, ch.text)
		text := must.Value(io.ReadAll(ch.Text()))
		assert.Equal(t, []byte{'a', 'b', 'c'}, text)
		assert.True(t, isAllRead(src))
	})

	t.Run("odd text length", func(t *testing.T) {
		// --- Given ---
		src := infoChunkTextLenOdd(t)
		skip4B(t, src) // Skip chunk ID.

		ch := INFO(LabIART)

		// --- When ---
		n, err := ch.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(8), n)
		assert.Equal(t, LabIART, ch.ID())
		assert.Equal(t, uint32(3), ch.Size())
		assert.Equal(t, []byte{'a', 'b', 0}, ch.text)
		assert.Equal(t, []byte{'a', 'b'}, must.Value(io.ReadAll(ch.Text())))
		assert.True(t, isAllRead(src))
	})

	t.Run("error - size limit", func(t *testing.T) {
		// --- Given ---
		src := infoChunkTextLenOdd(t)
		skip4B(t, src) // Skip chunk ID.

		ch := INFO(LabIART, WithSizeLimit(3))

		// --- When ---
		_, err := ch.ReadFrom(src)

		// --- Then ---
		assert.ErrorIs(t, ErrTooLarge, err)
	})

	t.Run("error - truncated input", func(t *testing.T) {
		// Reading less than 8 bytes should always result in an error.
		for i := 1; i < 8; i++ {
			// --- Given ---
			src := infoChunkTextLenEven(t)
			skip4B(t, src) // Skip chunk ID.

			ch := INFO(LabIART)

			// --- When ---
			_, err := ch.ReadFrom(io.LimitReader(src, int64(i)))

			// --- Then ---
			if !assert.Error(t, err) {
				t.Logf("errro i=%d", i)
			}
		}
	})

	t.Run("error - declared size over input", func(t *testing.T) {
		// --- Given ---
		src := &bytes.Buffer{}
		writeUint32LE(t, src, 1<<30)         // (0) 4 - Chunk size
		writeBytes(t, src, []byte{'a', 'b'}) // (4) 2 - Text

		ch := INFO(LabIART)

		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)

		// --- When ---
		_, err := ch.ReadFrom(src)

		// --- Then ---
		runtime.ReadMemStats(&after)
		assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
		assert.True(t, after.TotalAlloc-before.TotalAlloc < 1<<20)
	})
}

func Test_ChunkINFO_WriteTo(t *testing.T) {
	t.Run("error - write", func(t *testing.T) {
		// Writing less then 12 bytes should always result in an error.
		for i := 12; i > 0; i-- {
			// --- Given ---
			src := infoChunkTextLenEven(t)
			skip4B(t, src) // Skip chunk ID.

			ch := INFO(LabIART)
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

func Test_ChunkINFO_WriteTo_tabular(t *testing.T) {
	tt := []struct {
		testN string

		id uint32
		n  int64
		ch func(tester.T) io.Reader
	}{
		{"infoChunkTextLenEven", LabIART, 12, infoChunkTextLenEven},
		{"infoChunkTextLenOdd", LabIART, 12, infoChunkTextLenOdd},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			src := tc.ch(t)
			skip4B(t, src) // Skip chunk ID.

			ch := INFO(tc.id)
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

func Test_ChunkINFO_Reset(t *testing.T) {
	// --- Given ---
	ch := INFO(LabIART)
	ch.size = 3
	ch.text = []byte{'a', 'b', 0}

	// --- When ---
	ch.Reset()

	// --- Then ---
	assert.Equal(t, uint32(LabIART), ch.ID())
	assert.Equal(t, uint32(0), ch.Size())
	assert.Equal(t, []byte{}, ch.text)
	assert.Equal(t, []byte{}, must.Value(io.ReadAll(ch.Text())))
}

func Test_InfoLabel_tabular(t *testing.T) {
	tt := []struct {
		label uint32
		code  string
		str   string
	}{
		{LabIARL, "IARL", "archival location"},
		{LabIART, "IART", "artist"},
		{LabICMS, "ICMS", "commissioned"},
		{LabICMT, "ICMT", "comments"},
		{LabICOP, "ICOP", "copyright"},
		{LabICRD, "ICRD", "creation date"},
		{LabIENG, "IENG", "engineer"},
		{LabIGNR, "IGNR", "genre"},
		{LabIKEY, "IKEY", "keywords"},
		{LabIMED, "IMED", "original medium"},
		{LabINAM, "INAM", "title"},
		{LabIPRD, "IPRD", "album"},
		{LabITRK, "ITRK", "track"},
		{LabISBJ, "ISBJ", "subject"},
		{LabISFT, "ISFT", "software"},
		{LabISRC, "ISRC", "source"},
		{LabISRF, "ISRF", "source form"},
		{LabITCH, "ITCH", "technician"},
		{IDINFO, "INFO", "INFO"},
	}

	for _, tc := range tt {
		t.Run(tc.str, func(t *testing.T) {
			t.Parallel()

			// --- When ---
			have := InfoLabel(tc.label)

			// --- Then ---
			assert.Equal(t, tc.str, have)

			assert.Equal(t, tc.code, Uint32(tc.label).String())
		})
	}
}
