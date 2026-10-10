// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testkit/pkg/oskit"
)

func Test_Chunks_First(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		// --- Given ---
		rif := New()
		must.Value(rif.ReadFrom(oskit.Open(t, "testdata/bwf.wav")))
		chs := rif.Chunks()

		// --- When ---
		have := chs.First(IDfmt)

		// --- Then ---
		assert.Equal(t, uint32(IDfmt), have.ID())
		assert.Equal(t, uint32(0x10), have.Size())
	})

	t.Run("not present", func(t *testing.T) {
		// --- Given ---
		rif := New()
		must.Value(rif.ReadFrom(oskit.Open(t, "testdata/bwf.wav")))
		chs := rif.Chunks()

		// --- When ---
		have := chs.First(IDUNKN)

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_Chunks_Count(t *testing.T) {
	// --- Given ---
	rif := New()
	must.Value(rif.ReadFrom(oskit.Open(t, "testdata/bwf.wav")))
	chs := rif.Chunks()

	// --- When ---
	have := chs.Count(IDJUNK)

	// --- Then ---
	assert.Equal(t, 7, have)
}

func Test_Chunks_IDs(t *testing.T) {
	// --- Given ---
	rif := New()
	must.Value(rif.ReadFrom(oskit.Open(t, "testdata/bwf.wav")))
	chs := rif.Chunks()

	// --- When ---
	have := chs.IDs()

	// --- Then ---
	want := []uint32{
		0x62657874, // bext
		IDfmt,      // fmt
		IDdata,     // data
		0x4146416e, // AFAn
		IDJUNK,     // JUNK
		IDJUNK,     // JUNK
		IDJUNK,     // JUNK
		IDJUNK,     // JUNK
		IDJUNK,     // JUNK
		IDJUNK,     // JUNK
		IDJUNK,     // JUNK
		IDLIST,     // LIST
		0x41466d64, // AFmd
		IDID3,      // ID3+
	}
	assert.Equal(t, want, have)
}

func Test_Chunks_Size(t *testing.T) {
	// --- Given ---
	rif := New()
	must.Value(rif.ReadFrom(oskit.Open(t, "testdata/bwf.wav")))
	chs := rif.Chunks()

	// --- When ---
	have := chs.Size()

	// --- Then ---
	// The 27074 is the file have, we add 12 for ID, type and have fields.
	assert.Equal(t, uint32(27074), have+12)
}

func Test_Chunks_Remove(t *testing.T) {
	t.Run("removes chunk", func(t *testing.T) {
		// --- Given ---
		rif := New()
		must.Value(rif.ReadFrom(oskit.Open(t, "testdata/bwf.wav")))

		el := len(rif.Chunks().IDs())

		// --- When ---
		have := rif.Chunks().Remove(IDfmt).IDs()

		// --- Then ---
		assert.Equal(t, el-1, len(have))
		assert.Equal(t, 14, el)
	})

	t.Run("right order", func(t *testing.T) {
		// --- Given ---
		rif := New()
		must.Value(rif.ReadFrom(oskit.Open(t, "testdata/bwf.wav")))

		// --- When ---
		have := rif.Chunks().Remove(IDfmt).IDs()

		// --- Then ---
		want := []uint32{
			0x62657874, // bext
			IDdata,     // data
			0x4146416e, // AFAn
			IDJUNK,     // JUNK
			IDJUNK,     // JUNK
			IDJUNK,     // JUNK
			IDJUNK,     // JUNK
			IDJUNK,     // JUNK
			IDJUNK,     // JUNK
			IDJUNK,     // JUNK
			IDLIST,     // LIST
			0x41466d64, // AFmd
			IDID3,      // ID3+
		}
		assert.Equal(t, want, have)
	})

	t.Run("all chunks with ID", func(t *testing.T) {
		// --- Given ---
		chs := Chunks{RAWC(1), RAWC(2), RAWC(1)}

		// --- When ---
		have := chs.Remove(1)

		// --- Then ---
		assert.Equal(t, []uint32{2}, have.IDs())
	})

	t.Run("receiver unchanged", func(t *testing.T) {
		// --- Given ---
		chs := Chunks{RAWC(1), RAWC(2), RAWC(3)}

		// --- When ---
		have := chs.Remove(2)

		// --- Then ---
		assert.Equal(t, []uint32{1, 3}, have.IDs())
		assert.Equal(t, []uint32{1, 2, 3}, chs.IDs())
	})

	t.Run("key does not exist", func(t *testing.T) {
		// --- Given ---
		rif := New()
		must.Value(rif.ReadFrom(oskit.Open(t, "testdata/bwf.wav")))
		chsIDs := rif.Chunks().IDs()
		id := uint32(1)

		// --- When ---
		have := rif.Chunks().Remove(id).IDs()

		// --- Then ---
		assert.Equal(t, len(chsIDs), len(have))
	})
}
