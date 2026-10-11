// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"io"
	"math"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/mock"
	"github.com/ctx42/testkit/pkg/iokit"
)

func Test_StrToID_tabular(t *testing.T) {
	tt := []struct {
		id   string
		want uint32
	}{
		{"RIFF", IDRIFF},
		{"fmt ", IDfmt},
		{"data", IDdata},
		{"fact", 0x66616374},
		{"WAVE", TypeWAVE},
		{"LIST", IDLIST},
		{"INFO", IDINFO},
		{"id3 ", 0x69643320},
		{"IARL", 0x4941524c},
		{"IART", 0x49415254},
		{"ICMS", 0x49434d53},
		{"ICMT", 0x49434d54},
		{"ICOP", 0x49434f50},
		{"ICRD", 0x49435244},
		{"IENG", 0x49454e47},
		{"IGNR", 0x49474e52},
		{"IKEY", 0x494b4559},
		{"IMED", 0x494d4544},
		{"INAM", 0x494e414d},
		{"IPRD", 0x49505244},
		{"ITRK", 0x4954524b},
		{"ISBJ", 0x4953424a},
		{"ISFT", 0x49534654},
		{"ISRC", 0x49535243},
		{"ISRF", 0x49535246},
		{"ITCH", 0x49544348},
		{"tlst", 0x746c7374},
		{"JUNK", IDJUNK},
		{"bext", 0x62657874},
		{"ABCD", IDUNKN},
		{"AVI ", TypeAVI},
		{"RMID", TypeRMID},

		{"ITCHxx", 0x49544348},
		{"I", 0x49202020},
	}

	for _, tc := range tt {
		t.Run(tc.id, func(t *testing.T) {
			// --- When ---
			have := StrToID(tc.id)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_ReadChunkID(t *testing.T) {
	t.Run("read", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewReader([]byte{0x41, 0x42, 0x43, 0x44})
		var id uint32

		// --- When ---
		err := ReadChunkID(src, &id)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, uint32(0x41424344), id)
	})

	t.Run("error - read", func(t *testing.T) {
		// --- Given ---
		src := iokit.NewReaderMock(t)
		src.OnRead(mock.Any).Return(0, iokit.ErrRead)
		var id uint32

		// --- When ---
		err := ReadChunkID(src, &id)

		// --- Then ---
		assert.ErrorIs(t, iokit.ErrRead, err)
		assert.Equal(t, uint32(0), id)
	})
}

func Test_RealSize(t *testing.T) {
	t.Run("odd size", func(t *testing.T) {
		// --- When ---
		have := RealSize(123)

		// --- Then ---
		assert.Equal(t, uint32(124), have)
	})

	t.Run("even size", func(t *testing.T) {
		// --- When ---
		have := RealSize(124)

		// --- Then ---
		assert.Equal(t, uint32(124), have)
	})
}

func Test_ReadPaddingIf(t *testing.T) {
	t.Run("odd size", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewReader([]byte{0, 1})

		// --- When ---
		n, err := ReadPaddingIf(src, 3)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(1), n)
	})

	t.Run("even size", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewReader([]byte{0, 1})

		// --- When ---
		n, err := ReadPaddingIf(src, 4)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(0), n)
	})

	t.Run("padding at end of input", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewReader([]byte{0})

		// --- When ---
		n, err := ReadPaddingIf(src, 3)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(1), n)
	})

	t.Run("error - missing padding", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewReader([]byte{})

		// --- When ---
		n, err := ReadPaddingIf(src, 3)

		// --- Then ---
		assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
		assert.Equal(t, int64(0), n)
	})
	t.Run("missing at end of chunk reader source", func(t *testing.T) {
		// --- Given ---
		cr := NewChunkReader(bytes.NewReader(nil))

		// --- When ---
		n, err := ReadPaddingIf(cr, 3)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(0), n)
	})

	t.Run("missing at end of parent chunk reader source", func(t *testing.T) {
		// --- Given ---
		cr := NewChunkReader(NewChunkReader(bytes.NewReader(nil)))

		// --- When ---
		n, err := ReadPaddingIf(cr, 3)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(0), n)
	})

	t.Run("error - missing at chunk reader limit", func(t *testing.T) {
		// --- Given ---
		cr := NewChunkReader(bytes.NewReader([]byte{0}))
		cr.setLimit(0)

		// --- When ---
		n, err := ReadPaddingIf(cr, 3)

		// --- Then ---
		assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
		assert.Equal(t, int64(0), n)
	})

	t.Run("error - missing at parent chunk reader limit", func(t *testing.T) {
		// --- Given ---
		pcr := NewChunkReader(bytes.NewReader([]byte{0}))
		pcr.setLimit(0)
		cr := NewChunkReader(pcr)

		// --- When ---
		n, err := ReadPaddingIf(cr, 3)

		// --- Then ---
		assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
		assert.Equal(t, int64(0), n)
	})

	t.Run("error - non-zero pad byte", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewReader([]byte{1})

		// --- When ---
		n, err := ReadPaddingIf(src, 3)

		// --- Then ---
		assert.ErrorIs(t, ErrPadByte, err)
		assert.Equal(t, int64(1), n)
	})
}

func Test_discardPaddingIf(t *testing.T) {
	t.Run("odd size any value", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewReader([]byte{0xFF, 1})

		// --- When ---
		err := discardPaddingIf(src, 3)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, 1, src.Len())
	})

	t.Run("even size", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewReader([]byte{0xFF})

		// --- When ---
		err := discardPaddingIf(src, 2)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, 1, src.Len())
	})

	t.Run("error - missing", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewReader(nil)

		// --- When ---
		err := discardPaddingIf(src, 3)

		// --- Then ---
		assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
	})
}

func Test_size32(t *testing.T) {
	t.Run("fits", func(t *testing.T) {
		// --- When ---
		have := size32(7)

		// --- Then ---
		assert.Equal(t, uint32(7), have)
	})

	t.Run("over 4 GiB", func(t *testing.T) {
		if math.MaxInt <= math.MaxUint32 {
			t.Skip("int cannot hold a length over 4 GiB")
		}

		// --- Given ---
		n := uint64(math.MaxUint32) + 1

		// --- When ---
		have := size32(int(n))

		// --- Then ---
		assert.Equal(t, uint32(math.MaxUint32), have)
	})
}

func Test_shortRead_tabular(t *testing.T) {
	tt := []struct {
		testN string

		size int64
		read int64
		want bool
	}{
		{"read all", 8, 8, false},
		{"read more", 8, 10, false},
		{"missing final pad", 8, 7, false},
		{"even read short", 8, 6, true},
		{"odd read short", 8, 5, true},
		{"odd size one short", 7, 6, true},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := shortRead(tc.size, tc.read)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_WriteIDAndSize(t *testing.T) {
	t.Run("write", func(t *testing.T) {
		// --- Given ---
		dst := &bytes.Buffer{}

		// --- When ---
		n, err := WriteIDAndSize(dst, IDRIFF, 124)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(8), n)
		want := []byte{0x52, 0x49, 0x46, 0x46, 0x7c, 0x0, 0x0, 0x0}
		assert.Equal(t, want, dst.Bytes())
	})

	t.Run("error - writing ID", func(t *testing.T) {
		// --- Given ---
		buf := &bytes.Buffer{}
		dst := iokit.ErrWriter(buf, 3)

		// --- When ---
		n, err := WriteIDAndSize(dst, IDRIFF, 124)

		// --- Then ---
		assert.ErrorIs(t, iokit.ErrWrite, err)
		assert.Equal(t, int64(0), n)
		want := []byte{0x52, 0x49, 0x46}
		assert.Equal(t, want, buf.Bytes())
	})

	t.Run("error - writing size", func(t *testing.T) {
		// --- Given ---
		buf := &bytes.Buffer{}
		dst := iokit.ErrWriter(buf, 5)

		// --- When ---
		n, err := WriteIDAndSize(dst, IDRIFF, 124)

		// --- Then ---
		assert.ErrorIs(t, iokit.ErrWrite, err)
		assert.Equal(t, int64(4), n)
		want := []byte{0x52, 0x49, 0x46, 0x46, 0x7c}
		assert.Equal(t, want, buf.Bytes())
	})
}

func Test_WritePaddingIf(t *testing.T) {
	t.Run("odd size", func(t *testing.T) {
		// --- Given ---
		dst := &bytes.Buffer{}

		// --- When ---
		n, err := WritePaddingIf(dst, 123)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(1), n)
		want := []byte{0x00}
		assert.Equal(t, want, dst.Bytes())
	})

	t.Run("even size", func(t *testing.T) {
		// --- Given ---
		dst := &bytes.Buffer{}

		// --- When ---
		n, err := WritePaddingIf(dst, 124)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(0), n)
		assert.Equal(t, []byte(nil), dst.Bytes())
	})

	t.Run("error - write", func(t *testing.T) {
		// --- Given ---
		buf := &bytes.Buffer{}
		dst := iokit.ErrWriter(buf, 0)

		// --- When ---
		n, err := WritePaddingIf(dst, 123)

		// --- Then ---
		assert.ErrorIs(t, iokit.ErrWrite, err)
		assert.Equal(t, int64(0), n)
		assert.Equal(t, []byte(nil), buf.Bytes())
	})
}

func Test_TrimZeroRight_tabular(t *testing.T) {
	tt := []struct {
		testN string

		in   []byte
		want []byte
	}{
		{"empty", []byte{}, []byte{}},
		{"zero", []byte{0}, []byte{}},
		{"zeros", []byte{0, 0}, []byte{}},
		{"no zeros", []byte{'a', 'b'}, []byte{'a', 'b'}},
		{"one trailing zero", []byte{'a', 'b', 0}, []byte{'a', 'b'}},
		{"two trailing zeros", []byte{'a', 'b', 0, 0}, []byte{'a', 'b'}},
		{"single byte", []byte{'a'}, []byte{'a'}},
		{"single byte and zero", []byte{'a', 0}, []byte{'a'}},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := TrimZeroRight(tc.in)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_linkids(t *testing.T) {
	// --- When ---
	have := linkids(idRAWC, IDRIFF)

	// --- Then ---
	assert.Equal(t, "RAWC:RIFF", have)
}

func Test_readBytes(t *testing.T) {
	t.Run("within capacity", func(t *testing.T) {
		// --- Given ---
		buf := make([]byte, 0, 8)

		// --- When ---
		have, err := readBytes(bytes.NewReader([]byte{1, 2, 3}), buf, 3)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, []byte{1, 2, 3}, have)
		assert.Equal(t, 8, cap(have))
	})

	t.Run("grows past capacity", func(t *testing.T) {
		// --- Given ---
		data := bytes.Repeat([]byte{7}, 3*minReadStep+1)

		// --- When ---
		have, err := readBytes(bytes.NewReader(data), nil, len(data))

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, data, have)
	})

	t.Run("zero bytes", func(t *testing.T) {
		// --- When ---
		have, err := readBytes(bytes.NewReader(nil), nil, 0)

		// --- Then ---
		assert.NoError(t, err)
		assert.Len(t, 0, have)
	})

	t.Run("error - short input", func(t *testing.T) {
		// --- When ---
		have, err := readBytes(bytes.NewReader([]byte{1, 2}), nil, 1<<20)

		// --- Then ---
		assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
		assert.Equal(t, []byte{1, 2}, have)
		assert.True(t, cap(have) <= minReadStep)
	})

	t.Run("error - short input after a full step", func(t *testing.T) {
		// --- Given ---
		data := bytes.Repeat([]byte{7}, minReadStep)

		// --- When ---
		have, err := readBytes(bytes.NewReader(data), nil, 1<<20)

		// --- Then ---
		assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
		assert.Equal(t, data, have)
	})

	t.Run("error - empty input", func(t *testing.T) {
		// --- When ---
		have, err := readBytes(bytes.NewReader(nil), nil, 10)

		// --- Then ---
		assert.ErrorIs(t, io.EOF, err)
		assert.Len(t, 0, have)
	})

	t.Run("error - read", func(t *testing.T) {
		// --- Given ---
		src := iokit.ErrReader(bytes.NewReader([]byte{1, 2, 3}), 1)

		// --- When ---
		have, err := readBytes(src, nil, 1<<20)

		// --- Then ---
		assert.ErrorIs(t, iokit.ErrRead, err)
		assert.Equal(t, []byte{1}, have)
	})
}
