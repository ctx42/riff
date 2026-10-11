// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testkit/pkg/iokit"
	"github.com/ctx42/testkit/pkg/oskit"
	"github.com/ctx42/testkit/pkg/testkit"
)

func Test_New(t *testing.T) {
	// --- When ---
	have := New(WithLoadData())

	// --- Then ---
	assert.Equal(t, IDRIFF, have.ID())
	assert.Equal(t, uint32(4), have.Size())
	assert.Equal(t, uint32(0), have.Type())
	assert.False(t, have.Multi())
	assert.False(t, have.Raw())
	assert.True(t, have.IsRegistered(IDfmt))
	assert.True(t, have.IsRegistered(IDdata))
	assert.False(t, have.IsRegistered(0))
}

func Test_Bare(t *testing.T) {
	t.Run("registry", func(t *testing.T) {
		reg := NewRegistry(RAWCMake(WithLoadData()))

		// --- When ---
		rif := Bare(reg)

		// --- Then ---
		assert.Equal(t, IDRIFF, rif.ID())
		assert.Equal(t, uint32(4), rif.Size())
		assert.Equal(t, uint32(0), rif.Type())
		assert.False(t, rif.Multi())
		assert.False(t, rif.IsRegistered(IDfmt))
		assert.False(t, rif.IsRegistered(IDdata))
	})

	t.Run("nil registry", func(t *testing.T) {
		// --- When ---
		have := Bare(nil)

		// --- Then ---
		assert.Equal(t, IDRIFF, have.ID())
		assert.Equal(t, uint32(4), have.Size())
		assert.Equal(t, uint32(0), have.Type())
		assert.False(t, have.Multi())
		assert.False(t, have.IsRegistered(IDfmt))
		assert.False(t, have.IsRegistered(IDdata))
	})

	t.Run("nil registry with options", func(t *testing.T) {
		// --- Given ---
		want := kitchenSink(t)
		rif := Bare(nil, WithLoadData())
		must.Value(rif.ReadFrom(bytes.NewReader(want)))
		dst := &bytes.Buffer{}

		// --- When ---
		_, err := rif.WriteTo(dst)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, dst.Bytes())
	})
}

func Test_Compose(t *testing.T) {
	t.Run("without removed chunk", func(t *testing.T) {
		// --- Given ---
		rif := New()
		must.Value(rif.ReadFrom(oskit.Open(t, "testdata/bwf.wav")))
		chs := rif.Chunks()

		// --- When ---
		have := Compose(chs.Remove(IDfmt))

		// --- Then ---
		assert.Equal(t, 14, len(chs.IDs()))
		assert.Equal(t, 13, len(have.Chunks()))
	})

	t.Run("same size", func(t *testing.T) {
		// --- Given ---
		rif := New()
		must.Value(rif.ReadFrom(oskit.Open(t, "testdata/bwf.wav")))
		chs := rif.Chunks()

		// --- When ---
		have := Compose(chs)

		// --- Then ---
		assert.Equal(t, rif.Size(), have.Size())
		assert.Equal(t, len(chs.IDs()), len(have.Chunks().IDs()))
	})
}

func Test_RIFF_Size(t *testing.T) {
	t.Run("no chunks", func(t *testing.T) {
		// --- Given ---
		rif := New()

		// --- When ---
		have := rif.Size()

		// --- Then ---
		assert.Equal(t, uint32(4), have)
	})

	t.Run("edited chunk", func(t *testing.T) {
		// --- Given ---
		rif := New(WithLoadData())
		must.Value(rif.ReadFrom(oskit.Open(t, "testdata/kick.wav")))
		rif.Chunks().First(IDfmt).(*ChunkFMT).SetExtra([]byte{1, 2})

		// --- When ---
		have := rif.Size()

		// --- Then ---
		assert.Equal(t, uint32(9004+4), have)
	})
}

func Test_RIFF_SetType(t *testing.T) {
	// --- Given ---
	rif := New()

	// --- When ---
	rif.SetType(TypeAVI)

	// --- Then ---
	assert.Equal(t, TypeAVI, rif.Type())
}

func Test_RIFF_ReadFrom(t *testing.T) {
	t.Run("error - empty input", func(t *testing.T) {
		// --- Given ---
		rif := New()

		// --- When ---
		n, err := rif.ReadFrom(bytes.NewReader(nil))

		// --- Then ---
		assert.ErrorEqual(t, "error decoding RIFF chunk: EOF", err)
		assert.ErrorIs(t, io.EOF, err)
		assert.Equal(t, int64(0), n)
	})

	t.Run("error - not RIFF", func(t *testing.T) {
		// --- Given ---
		rif := New()

		// --- When ---
		n, err := rif.ReadFrom(bytes.NewReader([]byte("RIFX")))

		// --- Then ---
		assert.ErrorIs(t, ErrNotRIFF, err)
		assert.Equal(t, int64(4), n)
	})

	t.Run("zero value", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewReader(must.Value(os.ReadFile("testdata/kick.wav")))
		rif := &RIFF{}

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(9012), n)
		assert.Equal(t, []uint32{IDfmt, IDdata}, rif.Chunks().IDs())
		assert.False(t, rif.IsRegistered(IDfmt))
	})

	t.Run("extra bytes - respect declared size", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		writeUint32BE(t, src, IDRIFF)
		writeUint32LE(t, src, 4)
		writeUint32BE(t, src, TypeWAVE)
		writeUint32LE(t, src, 0x64646464)
		writeUint32LE(t, src, 4)
		writeUint32LE(t, src, 1000)

		rif := New(WithLoadData())

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(12), n)
		assert.Equal(t, uint32(4), rif.Size())
		assert.Len(t, 0, rif.Chunks())
	})

	t.Run("extra bytes - ignore declared size", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		writeUint32BE(t, src, IDRIFF)
		writeUint32LE(t, src, 4)
		writeUint32BE(t, src, TypeWAVE)
		writeUint32LE(t, src, 0x64646464)
		writeUint32LE(t, src, 4)
		writeUint32LE(t, src, 1000)

		rif := New(WithLoadData(), WithIgnoreSize())

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(24), n)
		assert.Equal(t, uint32(16), rif.Size()) // Size corrected.
		assert.Len(t, 1, rif.Chunks())
	})

	t.Run("error - noise at the end with ignored size", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		writeUint32BE(t, src, IDRIFF)
		writeUint32LE(t, src, 4)
		writeUint32BE(t, src, TypeWAVE)
		writeUint32LE(t, src, 0x64646464)
		writeUint32LE(t, src, 8000)
		writeUint32LE(t, src, 7000)

		rif := New(WithLoadData(), WithIgnoreSize())

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		want := "" +
			"error reading RIFF offset 24: " +
			"error decoding RAWC:dddd chunk: " +
			"unexpected EOF"
		assert.ErrorEqual(t, want, err)
		assert.Equal(t, int64(24), n)
		assert.Equal(t, uint32(4), rif.Size())
		assert.Len(t, 0, rif.Chunks())
	})

	t.Run("error - fewer bytes than declared", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		writeUint32BE(t, src, IDRIFF)
		writeUint32LE(t, src, 20)
		writeUint32BE(t, src, TypeWAVE)
		writeUint32LE(t, src, 0x64646464)
		writeUint32LE(t, src, 4)
		writeUint32LE(t, src, 1000)

		rif := New(WithLoadData())

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		want := "RIFF declared 20 bytes, decoder read 16 bytes: unexpected EOF"
		assert.ErrorEqual(t, want, err)
		assert.Equal(t, int64(24), n)
		assert.Equal(t, uint32(16), rif.Size())
		assert.Len(t, 1, rif.Chunks())
	})

	t.Run("error - chunk name cut in half", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		writeUint32BE(t, src, IDRIFF)
		writeUint32LE(t, src, 20)
		writeUint32BE(t, src, TypeWAVE)
		writeUint16LE(t, src, 0x6464)

		rif := New(WithLoadData())

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		want := "error reading RIFF offset 14: unexpected EOF"
		assert.ErrorEqual(t, want, err)
		assert.Equal(t, int64(14), n)
		assert.Equal(t, uint32(4), rif.Size())
		assert.Len(t, 0, rif.Chunks())
	})

	t.Run("within size limit", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		writeUint32BE(t, src, IDRIFF)
		writeUint32LE(t, src, 16)
		writeUint32BE(t, src, TypeWAVE)
		writeUint32LE(t, src, 0x64646464)
		writeUint32LE(t, src, 4)
		writeUint32LE(t, src, 1000)

		rif := New(WithLoadData(), WithSizeLimit(16))

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(24), n)
		assert.Equal(t, uint32(16), rif.Size())
		assert.Len(t, 1, rif.Chunks())
	})

	t.Run("error - main chunk size limit", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		writeUint32BE(t, src, IDRIFF)
		writeUint32LE(t, src, 16)
		writeUint32BE(t, src, TypeWAVE)
		writeUint32LE(t, src, 0x64646464)
		writeUint32LE(t, src, 4)
		writeUint32LE(t, src, 1000)

		rif := New(WithLoadData(), WithSizeLimit(15))

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		want := "error decoding RIFF chunk: size too large"
		assert.ErrorEqual(t, want, err)
		assert.Equal(t, int64(4), n)
		assert.Equal(t, uint32(4), rif.Size())
		assert.Len(t, 0, rif.Chunks())
	})

	t.Run("error - sub-chunk size limit", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		writeUint32BE(t, src, IDRIFF)
		writeUint32LE(t, src, 16)
		writeUint32BE(t, src, TypeWAVE)
		writeUint32LE(t, src, 0x64646464)
		writeUint32LE(t, src, 4)
		writeUint32LE(t, src, 1000)

		reg := NewRegistry(RAWCMake(WithLoadData(), WithSizeLimit(3)))
		rif := Bare(reg)

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		want := "" +
			"error reading RIFF offset 20: " +
			"error decoding RAWC:dddd chunk: " +
			"size too large"
		assert.ErrorEqual(t, want, err)
		assert.Equal(t, int64(20), n)
		assert.Equal(t, uint32(4), rif.Size())
		assert.Len(t, 0, rif.Chunks())
	})

	t.Run("error - size limit with ignored size", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		writeUint32BE(t, src, IDRIFF)
		writeUint32LE(t, src, 4)
		writeUint32BE(t, src, TypeWAVE)
		writeUint32LE(t, src, 0x64646464)
		writeUint32LE(t, src, 4)
		writeUint32LE(t, src, 1000)

		rif := New(WithLoadData(), WithIgnoreSize(), WithSizeLimit(15))

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		want := "" +
			"error reading RIFF offset 23: " +
			"error decoding RAWC:dddd chunk: " +
			"unexpected EOF"
		assert.ErrorEqual(t, want, err)
		assert.Equal(t, int64(23), n)
		assert.Equal(t, uint32(4), rif.Size())
		assert.Len(t, 0, rif.Chunks())
	})

	t.Run("error - truncated type", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		writeUint32BE(t, src, IDRIFF)
		writeUint32LE(t, src, 4)
		writeBytes(t, src, []byte{'W', 'A'})

		rif := New()

		// --- When ---
		_, err := rif.ReadFrom(src)

		// --- Then ---
		assert.ErrorEqual(t, "error decoding RIFF chunk: unexpected EOF", err)
		assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
	})

	t.Run("error - repeated single chunk", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		writeUint32BE(t, src, IDRIFF)
		writeUint32LE(t, src, 52)
		writeUint32BE(t, src, TypeWAVE)
		for range 2 {
			writeUint32BE(t, src, IDfmt)
			writeUint32LE(t, src, FMTChunkSize)
			writeBytes(t, src, make([]byte, FMTChunkSize))
		}

		rif := New()

		// --- When ---
		_, err := rif.ReadFrom(src)

		// --- Then ---
		want := "" +
			"error reading RIFF offset 40: " +
			"chunk fmt  (0x666d7420) already seen"
		assert.ErrorEqual(t, want, err)
	})

	t.Run("error - over 4 GiB with ignored size", func(t *testing.T) {
		// --- Given ---
		head := &bytes.Buffer{}
		writeUint32BE(t, head, IDRIFF)
		writeUint32LE(t, head, 4)
		writeUint32BE(t, head, TypeWAVE)
		writeUint32BE(t, head, IDJUNK)
		writeUint32LE(t, head, math.MaxUint32-1)

		tail := &bytes.Buffer{}
		writeUint32BE(t, tail, IDJUNK)
		writeUint32LE(t, tail, 16)

		end := int64(20) + math.MaxUint32 - 1
		src := &sparseReader{
			size: end + 8 + 16,
			segs: map[int64][]byte{0: head.Bytes(), end: tail.Bytes()},
		}

		rif := New(WithIgnoreSize())

		// --- When ---
		_, err := rif.ReadFrom(src)

		// --- Then ---
		assert.ErrorEqual(t, "error decoding RIFF chunk: size too large", err)
		assert.ErrorIs(t, ErrTooLarge, err)
	})

	t.Run("missing final pad re-encodes with pad", func(t *testing.T) {
		// --- Given ---
		want := kitchenSink(t)
		rif := New(WithLoadData())
		must.Value(rif.ReadFrom(bytes.NewReader(want[:len(want)-1])))
		dst := &bytes.Buffer{}

		// --- When ---
		_, err := rif.WriteTo(dst)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, dst.Bytes())
	})

	t.Run("missing final pad in last list", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		writeUint32BE(t, src, IDRIFF)
		writeUint32LE(t, src, 4+8+16)
		writeUint32BE(t, src, TypeWAVE)
		writeUint32BE(t, src, IDLIST)
		writeUint32LE(t, src, 4+8+3+1)
		writeUint32BE(t, src, IDINFO)
		writeUint32BE(t, src, LabIART)
		writeUint32LE(t, src, 3)
		writeBytes(t, src, []byte{'a', 'b', 0}) // No pad byte.

		rif := New()

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(35), n)
		assert.Equal(t, uint32(28), rif.Size())
	})

	t.Run("error - missing pad mid-file", func(t *testing.T) {
		// --- Given ---
		data := kitchenSink(t)
		data = append(data[:215:215], data[216:]...) // Drop raw pad byte.
		le.PutUint32(data[4:], 219)

		rif := New(WithLoadData())

		// --- When ---
		_, err := rif.ReadFrom(bytes.NewReader(data))

		// --- Then ---
		assert.ErrorIs(t, ErrPadByte, err)
	})

	t.Run("error - missing final pad and more", func(t *testing.T) {
		// --- Given ---
		data := kitchenSink(t)
		data = data[:len(data)-1] // Drop the data chunk pad byte.
		le.PutUint32(data[4:], 222)

		rif := New(WithLoadData())

		// --- When ---
		_, err := rif.ReadFrom(bytes.NewReader(data))

		// --- Then ---
		want := "RIFF declared 222 bytes, decoder read 219 bytes: " +
			"unexpected EOF"
		assert.ErrorEqual(t, want, err)
	})

	t.Run("odd RIFF size rounded up", func(t *testing.T) {
		// --- Given ---
		data := kitchenSink(t)
		// Final pad byte present but not counted, as in testdata/bwf.wav.
		le.PutUint32(data[4:], 219)

		rif := New()

		// --- When ---
		n, err := rif.ReadFrom(bytes.NewReader(data))

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(228), n)
		assert.Equal(t, uint32(220), rif.Size())
	})

	t.Run("reuse - raw and list sub-chunk decoders", func(t *testing.T) {
		// --- Given ---
		rif := New()
		must.Value(rif.ReadFrom(bytes.NewReader(kitchenSink(t))))
		raw := rif.Chunks().First(IDUNKN)
		labl := rif.Chunks()[3].Chunks().First(IDlabl)
		iart := rif.Chunks()[2].Chunks().First(LabIART)

		// --- When ---
		_, err := rif.ReadFrom(bytes.NewReader(kitchenSink(t)))

		// --- Then ---
		assert.NoError(t, err)
		assert.Same(t, raw, rif.Chunks().First(IDUNKN))
		assert.Same(t, labl, rif.Chunks()[3].Chunks().First(IDlabl))
		assert.Same(t, iart, rif.Chunks()[2].Chunks().First(LabIART))
	})

	t.Run("reuse - sub-chunk decoder of other list type", func(t *testing.T) {
		// --- Given ---
		rif := New()
		must.Value(rif.ReadFrom(bytes.NewReader(kitchenSink(t))))

		src := bytes.NewBuffer(nil)
		writeUint32BE(t, src, IDRIFF)
		writeUint32LE(t, src, 26)
		writeUint32BE(t, src, TypeWAVE)
		writeUint32BE(t, src, IDLIST)
		writeUint32LE(t, src, 14)
		writeUint32BE(t, src, IDINFO)
		writeUint32BE(t, src, IDltxt)
		writeUint32LE(t, src, 2)
		writeBytes(t, src, []byte{'a', 0})

		// --- When ---
		_, err := rif.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		lst := rif.Chunks().First(IDLIST)
		assert.SameType(t, &ChunkINFO{}, lst.Chunks().First(IDltxt))
	})

	t.Run("reuse - top-level chunk with list sub-chunk ID", func(t *testing.T) {
		// --- Given ---
		rif := New()
		must.Value(rif.ReadFrom(bytes.NewReader(kitchenSink(t))))

		src := bytes.NewBuffer(nil)
		writeUint32BE(t, src, IDRIFF)
		writeUint32LE(t, src, 38)
		writeUint32BE(t, src, TypeWAVE)
		for range 2 { // Reusing both lists pools their sub-chunks.
			writeUint32BE(t, src, IDLIST)
			writeUint32LE(t, src, 4)
			writeUint32BE(t, src, IDINFO)
		}
		writeUint32BE(t, src, LabIART)
		writeUint32LE(t, src, 2)
		writeBytes(t, src, []byte{'a', 0})

		// --- When ---
		_, err := rif.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.SameType(t, &ChunkRAWC{}, rif.Chunks().First(LabIART))
	})
}

func Test_RIFF_ReadFrom_missingFinalPad_tabular(t *testing.T) {
	tt := []struct {
		testN string

		size uint32 // Declared RIFF size.
		opts OptsFn // Decoding options.
		seek bool   // Source implements io.Seeker.
	}{
		{"pad counted load data", 220, WithLoadData(), true},
		{"pad counted skip seeker", 220, WithOpts(Opts{}), true},
		{"pad counted skip no seeker", 220, WithOpts(Opts{}), false},
		{"pad not counted load data", 219, WithLoadData(), true},
		{"pad not counted skip seeker", 219, WithOpts(Opts{}), true},
		{"pad not counted skip no seeker", 219, WithOpts(Opts{}), false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			data := kitchenSink(t)
			data = data[:len(data)-1] // Drop the data chunk pad byte.
			le.PutUint32(data[4:], tc.size)

			var src io.Reader = bytes.NewReader(data)
			if !tc.seek {
				src = nonSeeker{src}
			}

			rif := New(tc.opts)

			// --- When ---
			n, err := rif.ReadFrom(src)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, int64(227), n)
			assert.Equal(t, uint32(220), rif.Size())
			assert.Len(t, 6, rif.Chunks())
		})
	}
}

func Test_RIFF_ReadFrom_padByte_tabular(t *testing.T) {
	tt := []struct {
		testN string

		off  int    // Offset of the pad byte in kitchenSink.
		opts OptsFn // Decoding options.
		seek bool   // Source implements io.Seeker.
	}{
		{"smpl", 111, WithLoadData(), true},
		{"INFO", 135, WithLoadData(), true},
		{"labl", 175, WithLoadData(), true},
		{"raw load data", 215, WithLoadData(), true},
		{"raw skip seeker", 215, WithOpts(Opts{}), true},
		{"raw skip no seeker", 215, WithOpts(Opts{}), false},
		{"data load data", 227, WithLoadData(), true},
		{"data skip seeker", 227, WithOpts(Opts{}), true},
		{"data skip no seeker", 227, WithOpts(Opts{}), false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			data := kitchenSink(t)
			data[tc.off] = 0xFF

			var src io.Reader = bytes.NewReader(data)
			if !tc.seek {
				src = nonSeeker{src}
			}

			rif := New(tc.opts)

			// --- When ---
			_, err := rif.ReadFrom(src)

			// --- Then ---
			assert.ErrorIs(t, ErrPadByte, err)
		})
	}
}

func Test_RIFF_ReadFrom_tabular(t *testing.T) {
	rif := New()

	tt := []struct {
		pth    string
		typ    uint32
		size   int64
		chunks int
	}{
		{"testdata/11k16bitpcm.wav", TypeWAVE, 304578, 2},
		{"testdata/11k8bitpcm.wav", TypeWAVE, 152312, 2},
		{"testdata/11kadpcm.wav", TypeWAVE, 77252, 3},
		{"testdata/11kgsm.wav", TypeWAVE, 31000, 3},
		{"testdata/11kulaw.wav", TypeWAVE, 152326, 3},
		{"testdata/32bit.wav", TypeWAVE, 352844, 2},
		{"testdata/8bit.wav", TypeWAVE, 88244, 2},
		{"testdata/8k16bitpcm.wav", TypeWAVE, 221026, 2},
		{"testdata/8k8bitpcm.wav", TypeWAVE, 110532, 2},
		{"testdata/8kadpcm.wav", TypeWAVE, 56072, 3},
		{"testdata/8kcelp.wav", TypeWAVE, 8350, 3},
		{"testdata/8kgsm.wav", TypeWAVE, 22550, 3},
		{"testdata/8kmp316.wav", TypeWAVE, 27386, 3},
		{"testdata/8kmp38.wav", TypeWAVE, 13584, 3},
		{"testdata/8ksbc12.wav", TypeWAVE, 19658, 3},
		{"testdata/8ktruespeech.wav", TypeWAVE, 14842, 3},
		{"testdata/8kulaw.wav", TypeWAVE, 110546, 3},
		{"testdata/bass.wav", TypeWAVE, 143786, 2},
		{"testdata/bwf.wav", TypeWAVE, 27074, 14},
		{"testdata/dirty-kick-24b441k.wav", TypeWAVE, 132136, 3},
		{"testdata/flloop.wav", TypeWAVE, 434838, 7},
		{"testdata/junkKick.wav", TypeWAVE, 83084, 9},
		{"testdata/kick-16b441k.wav", TypeWAVE, 31692, 3},
		{"testdata/kick.wav", TypeWAVE, 9012, 2},
		{"testdata/listChunkInHeader.wav", TypeWAVE, 104196, 4},
		{"testdata/listinfo.wav", TypeWAVE, 176720, 4},
		{"testdata/misaligned-chunk.wav", TypeWAVE, 3441572, 8},
		{"testdata/padded24b.wav", TypeWAVE, 24908, 5},
		{"testdata/sample16bit.wav", TypeWAVE, 882160, 3},
		{"testdata/sample.avi", TypeAVI, 230264, 4}, // --
		{"testdata/sample.rmi", TypeRMID, 29640, 4}, // --
		{"testdata/sample.wav", TypeWAVE, 54002, 2},
	}

	for _, tc := range tt {
		t.Run(tc.pth, func(t *testing.T) {
			// --- Given ---
			fil := oskit.Open(t, tc.pth)

			// --- When ---
			n, err := rif.ReadFrom(fil)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.size, n)
			assert.Equal(t, uint32(tc.size-8), rif.Size())
			assert.Equal(t, tc.typ, rif.Type())
			assert.Len(t, tc.chunks, rif.Chunks())
		})
	}
}

func Test_RIFF_ReadFrom_allocs_tabular(t *testing.T) {
	tt := []struct {
		testN string

		pth  string
		opts []OptsFn
		want float64
	}{
		{"kick meta", "testdata/kick.wav", nil, 0},
		{"bwf meta", "testdata/bwf.wav", nil, 0},
		{"listinfo meta", "testdata/listinfo.wav", nil, 0},
		{"flloop meta", "testdata/flloop.wav", nil, 1}, // Sample loops.
		{"misaligned meta", "testdata/misaligned-chunk.wav", nil, 0},
		{"kick load", "testdata/kick.wav", []OptsFn{WithLoadData()}, 0},
		{"bwf load", "testdata/bwf.wav", []OptsFn{WithLoadData()}, 0},
		{"flloop load", "testdata/flloop.wav", []OptsFn{WithLoadData()}, 1},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			data := must.Value(os.ReadFile(tc.pth))
			src := bytes.NewReader(data)
			rif := New(tc.opts...)
			must.Value(rif.ReadFrom(src)) // Fill the decoder pools.

			// --- When ---
			have := testing.AllocsPerRun(10, func() {
				src.Reset(data)
				_, _ = rif.ReadFrom(src)
			})

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_RIFF_WriteTo(t *testing.T) {
	t.Run("corrects size", func(t *testing.T) {
		// --- Given ---
		rif := New(WithLoadData())
		must.Value(rif.ReadFrom(oskit.Open(t, "testdata/bwf.wav")))
		dst := &bytes.Buffer{}

		// --- When ---
		n, err := rif.WriteTo(dst)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(27074), n)
		assert.Equal(t, uint32(27074), rif.Size()+8)
		want := "ebdf4fefcbbb804b44cddbc50521ba29f833e05c"
		assert.Equal(t, want, testkit.SHA1Reader(dst))
	})

	t.Run("error - chunks over 4 GiB", func(t *testing.T) {
		// --- Given ---
		rif := Compose(Chunks{&sizedChunk{1 << 31}, &sizedChunk{1 << 31}})
		dst := &bytes.Buffer{}

		// --- When ---
		n, err := rif.WriteTo(dst)

		// --- Then ---
		assert.ErrorEqual(t, "error encoding RIFF chunk: size too large", err)
		assert.ErrorIs(t, ErrTooLarge, err)
		assert.Equal(t, int64(0), n)
		assert.Equal(t, 0, dst.Len())
	})

	t.Run("error - list over 4 GiB", func(t *testing.T) {
		// --- Given ---
		lst := LIST(nil, WithLoadData())
		lst.Modify(Chunks{&sizedChunk{1 << 31}, &sizedChunk{1 << 31}})
		rif := Compose(Chunks{lst})
		dst := &bytes.Buffer{}

		// --- When ---
		n, err := rif.WriteTo(dst)

		// --- Then ---
		assert.ErrorEqual(t, "error encoding RIFF chunk: size too large", err)
		assert.ErrorIs(t, ErrTooLarge, err)
		assert.Equal(t, int64(0), n)
		assert.Equal(t, 0, dst.Len())
	})

	t.Run("error - writing header", func(t *testing.T) {
		// --- Given ---
		rif := New(WithLoadData())
		must.Value(rif.ReadFrom(oskit.Open(t, "testdata/kick.wav")))

		// --- When ---
		_, err := rif.WriteTo(iokit.ErrWriter(&bytes.Buffer{}, 4))

		// --- Then ---
		assert.ErrorEqual(t, "error encoding RIFF chunk: write error", err)
		assert.ErrorIs(t, iokit.ErrWrite, err)
	})

	t.Run("error - writing type", func(t *testing.T) {
		// --- Given ---
		rif := New(WithLoadData())
		must.Value(rif.ReadFrom(oskit.Open(t, "testdata/kick.wav")))

		// --- When ---
		_, err := rif.WriteTo(iokit.ErrWriter(&bytes.Buffer{}, 12))

		// --- Then ---
		assert.ErrorEqual(t, "error encoding RIFF chunk: write error", err)
		assert.ErrorIs(t, iokit.ErrWrite, err)
	})

	t.Run("error - writing chunks", func(t *testing.T) {
		// --- Given ---
		rif := New(WithLoadData())
		must.Value(rif.ReadFrom(oskit.Open(t, "testdata/kick.wav")))

		// --- When ---
		_, err := rif.WriteTo(iokit.ErrWriter(&bytes.Buffer{}, 16))

		// --- Then ---
		assert.ErrorEqual(t, "error encoding fmt  chunk: write error", err)
		assert.ErrorIs(t, iokit.ErrWrite, err)
	})
}

func Test_RIFF_WriteTo_tabular(t *testing.T) {
	rif := New(WithLoadData())

	tt := []struct {
		pth  string
		size int64
		hash string
	}{
		{
			"testdata/11k16bitpcm.wav",
			304578,
			"5f90c25e3993819fc8765e87cb262751548ce71c",
		},
		{
			"testdata/11k8bitpcm.wav",
			152312,
			"96ae9bec8dfcd366918387e05b21b114fd3e54e4",
		},
		{
			"testdata/11kadpcm.wav",
			77252,
			"3be2a0c1ffebae0636fc77f8ef8f701be8a7de9e",
		},
		{
			"testdata/11kgsm.wav",
			31000,
			"019860d50c93ac790f0bcffaa23df418236da075",
		},
		{
			"testdata/11kulaw.wav",
			152326,
			"8288b2123bfc102e97ed849c659b29df1e5519b7",
		},
		{
			"testdata/32bit.wav",
			352844,
			"46bc96f0953df4f6f36053e65733f5e32d276167",
		},
		{
			"testdata/8bit.wav",
			88244,
			"ef28c1a4fc75b0e4630499b1899bef244c837424",
		},
		{
			"testdata/8k16bitpcm.wav",
			221026,
			"c853aa623e1a3b717fa452665f078b5b0dbf3139",
		},
		{
			"testdata/8k8bitpcm.wav",
			110532,
			"118c7c2d55a6bbaf55543cd8e405be146f41666d",
		},
		{
			"testdata/8kadpcm.wav",
			56072,
			"119f54e8cf5469d6eba0cb7a97a91ca2597faeec",
		},
		{
			"testdata/8kcelp.wav",
			8350,
			"e534f24ce1177a9db1d538ddb5c523125854b7a5",
		},
		{
			"testdata/8kgsm.wav",
			22550,
			"6b97716f68ec2d11d2fc5ef28642a2a78f39c376",
		},
		{
			"testdata/8kmp316.wav",
			27386,
			"da10b9c9c4f71b7f661feebf023fdb22f492b437",
		},
		{
			"testdata/8kmp38.wav",
			13584,
			"85da88b8fa13ea4bbee491389125729b40203682",
		},
		{
			"testdata/8ksbc12.wav",
			19658,
			"35036ceba255622883b5b3c948e13eb1e7184a4e",
		},
		{
			"testdata/8ktruespeech.wav",
			14842,
			"1be8b839cbf0135d6b330b5a5b3d63f65c1daae9",
		},
		{
			"testdata/8kulaw.wav",
			110546,
			"5205e7e836c2f50da1261c32649532e4fee2cdaf",
		},
		{
			"testdata/bass.wav",
			143786,
			"5c547df0f3d24da873d92e28ca3ecd22b83ad92b",
		},
		{
			"testdata/dirty-kick-24b441k.wav",
			132136,
			"61b6cebfe073f34a81bc4352e69b5a044b07e188",
		},
		{
			"testdata/flloop.wav",
			434838,
			"673928ba8a28bced469c35bc64c26530b89b91b7",
		},
		{
			"testdata/junkKick.wav",
			83084,
			"5974547313bcb8804618a3048f6902a429d62a4e",
		},
		{
			"testdata/kick-16b441k.wav",
			31692,
			"1d7dbd0fe12ce2f8ec33ef5f90e271125a83ea94",
		},
		{"testdata/kick.wav", 9012, "cfa0812a881c3d3f2a2783b5b7a6d2ba62f7a1aa"},
		{
			"testdata/listChunkInHeader.wav",
			104196,
			"9d722b63acf00c0031a5aaf7b3f73291321aafe3",
		},
		{
			"testdata/listinfo.wav",
			176720,
			"8dd085ecf49c1ca781857dd1f26b16b636b7ab07",
		},
		{
			"testdata/misaligned-chunk.wav",
			3441572,
			"dda435791002e772da033e3e7f09a3854c9e8d76",
		},
		{
			"testdata/padded24b.wav",
			24908,
			"25b88e9afdc533be201b4393bc489cbfa61ec2e7",
		},
		{
			"testdata/sample.avi",
			230264,
			"8f30db3104fafec017241b63fbba6588ee8cd5b4",
		},
		{
			"testdata/sample.rmi",
			29640,
			"aa4cdab911e956439f8553772b56cd01993e9e12",
		},
	}

	for _, tc := range tt {
		t.Run(tc.pth, func(t *testing.T) {
			// --- Given ---
			must.Value(rif.ReadFrom(oskit.Open(t, tc.pth)))
			buf := &bytes.Buffer{}

			// --- When ---
			n, err := rif.WriteTo(buf)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.size, n)
			size := int64(RealSize(rif.Size()) + 8)
			assert.Equal(t, tc.size, size)
			assert.Equal(t, tc.hash, testkit.SHA1Reader(buf))
		})
	}
}

func Test_RIFF_WriteTo_allocs_tabular(t *testing.T) {
	for _, pth := range benchFiles {
		t.Run(filepath.Base(pth), func(t *testing.T) {
			// --- Given ---
			data := must.Value(os.ReadFile(pth))
			rif := New(WithLoadData())
			must.Value(rif.ReadFrom(bytes.NewReader(data)))

			// --- When ---
			have := testing.AllocsPerRun(10, func() {
				_, _ = rif.WriteTo(io.Discard)
			})

			// --- Then ---
			assert.Equal(t, float64(0), have)
		})
	}
}

func Test_RIFF_Reset(t *testing.T) {
	t.Run("modified chunks not pooled", func(t *testing.T) {
		// --- Given ---
		ch := DATA(WithLoadData())
		must.Nil(ch.SetData([]byte{1, 2}))

		rif := New(WithLoadData())
		rif.Modify(Chunks{ch})

		// --- When ---
		rif.Reset()

		// --- Then ---
		assert.NotSame(t, ch, rif.reg.Get(IDdata))
		assert.Equal(t, uint32(2), ch.Size())
	})
}

func Test_RIFF_Modify(t *testing.T) {
	t.Run("removes chunk", func(t *testing.T) {
		// --- Given ---
		rif := New()
		must.Value(rif.ReadFrom(oskit.Open(t, "testdata/bwf.wav")))
		s := rif.Size()
		chs := rif.Chunks()
		f := chs.First(IDfmt)

		// --- When ---
		rif.Modify(chs.Remove(IDfmt))

		// --- Then ---
		assert.Equal(t, s-(f.Size()+8), rif.Size())
		assert.Equal(t, 13, len(rif.Chunks().IDs()))
	})
}

// benchFiles are the sample files the benchmarks decode and encode.
var benchFiles = []string{
	"testdata/kick.wav",             // fmt and data.
	"testdata/bwf.wav",              // 14 chunks, LIST and raw chunks.
	"testdata/listinfo.wav",         // LIST INFO.
	"testdata/flloop.wav",           // smpl.
	"testdata/misaligned-chunk.wav", // 3.4 MB.
}

func Benchmark_RIFF_ReadFrom(b *testing.B) {
	modes := []struct {
		name string
		opts []OptsFn
	}{
		{"meta", nil},
		{"load", []OptsFn{WithLoadData()}},
	}

	for _, mode := range modes {
		for _, pth := range benchFiles {
			b.Run(mode.name+"/"+filepath.Base(pth), func(b *testing.B) {
				data := must.Value(os.ReadFile(pth))
				src := bytes.NewReader(data)
				rif := New(mode.opts...)

				b.ReportAllocs()
				for b.Loop() {
					src.Reset(data)
					if _, err := rif.ReadFrom(src); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func Benchmark_RIFF_WriteTo(b *testing.B) {
	for _, pth := range benchFiles {
		b.Run(filepath.Base(pth), func(b *testing.B) {
			data := must.Value(os.ReadFile(pth))
			rif := New(WithLoadData())
			must.Value(rif.ReadFrom(bytes.NewReader(data)))

			b.ReportAllocs()
			for b.Loop() {
				if _, err := rif.WriteTo(io.Discard); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
