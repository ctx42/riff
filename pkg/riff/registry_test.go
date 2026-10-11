// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_Registry_Has(t *testing.T) {
	t.Run("registered", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake(WithLoadData()))
		reg.Register(IDfmt, FMTMake())

		// --- When ---
		have := reg.Has(IDfmt)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not registered", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake(WithLoadData()))
		reg.Register(IDfmt, FMTMake())

		// --- When ---
		have := reg.Has(IDLIST)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Registry_Register(t *testing.T) {
	// --- Given ---
	reg := NewRegistry(RAWCMake(WithLoadData()))

	// --- When ---
	reg.Register(IDfmt, FMTMake())

	// --- Then ---
	assert.SameType(t, &ChunkFMT{}, reg.Get(IDfmt))
	assert.SameType(t, &ChunkRAWC{}, reg.Get(IDUNKN))
}

func Test_Registry_Put(t *testing.T) {
	t.Run("registered", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake(WithLoadData()))
		reg.Register(IDfmt, FMTMake())
		ch := FMT()
		ch.SampleRate = 44100

		// --- When ---
		reg.Put(ch)

		// --- Then ---
		assert.Same(t, ch, reg.GetNoRaw(IDfmt))
		assert.Equal(t, uint32(0), ch.SampleRate)
	})

	t.Run("not registered", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake(WithLoadData()))
		ch := RAWC(IDUNKN)

		// --- When ---
		reg.Put(ch)

		// --- Then ---
		assert.Nil(t, reg.GetNoRaw(IDUNKN))
	})
}

func Test_Registry_tryPut(t *testing.T) {
	t.Run("registered", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake())
		reg.Register(IDfmt, FMTMake())
		ch := FMT()

		// --- When ---
		have := reg.tryPut(ch)

		// --- Then ---
		assert.True(t, have)
		assert.Same(t, ch, reg.GetNoRaw(IDfmt))
	})

	t.Run("not registered", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake())

		// --- When ---
		have := reg.tryPut(RAWC(IDUNKN))

		// --- Then ---
		assert.False(t, have)
		assert.Nil(t, reg.GetNoRaw(IDUNKN))
	})
}

func Test_Registry_release(t *testing.T) {
	t.Run("registered", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake())
		reg.Register(IDfmt, FMTMake())
		ch := FMT()

		// --- When ---
		reg.release(ch)

		// --- Then ---
		assert.Same(t, ch, reg.GetNoRaw(IDfmt))
	})

	t.Run("raw", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake())
		ch := RAWC(IDUNKN)

		// --- When ---
		reg.release(ch)

		// --- Then ---
		assert.Nil(t, reg.GetNoRaw(IDUNKN))
		assert.Same(t, ch, reg.Get(IDUNKN))
	})
}

func Test_Registry_Get(t *testing.T) {
	t.Run("reuse released raw decoder", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake())
		ch := reg.Get(IDUNKN)
		reg.release(ch)

		// --- When ---
		have := reg.Get(IDUNKN)

		// --- Then ---
		assert.Same(t, ch, have)
		assert.NotSame(t, ch, reg.Get(IDUNKN))
	})

	t.Run("reuse pooled decoder", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake(WithLoadData()))
		reg.Register(IDfmt, FMTMake())

		ch := reg.Get(IDfmt)
		reg.Put(ch)

		// --- When ---
		have := reg.Get(IDfmt)

		// --- Then ---
		assert.Same(t, ch, have)
	})
}

func Test_Registry_GetNoRaw(t *testing.T) {
	// --- Given ---
	reg := NewRegistry(RAWCMake(WithLoadData()))

	// --- When ---
	have := reg.GetNoRaw(IDfmt)

	// --- Then ---
	assert.Nil(t, have)
}

func Test_Registry_putSub(t *testing.T) {
	// --- Given ---
	reg := NewRegistry(RAWCMake())
	ch := INFO(LabIART)
	ch.text = []byte{'a', 'b', 0}

	// --- When ---
	reg.putSub(IDINFO, Opts{}, ch)

	// --- Then ---
	assert.Equal(t, uint32(0), ch.Size())
	assert.Same(t, ch, reg.getSub(IDINFO, LabIART, Opts{}))
}

func Test_Registry_getSub(t *testing.T) {
	t.Run("empty pool", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake())

		// --- When ---
		have := reg.getSub(IDINFO, LabIART, Opts{})

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("other list type", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake())
		reg.putSub(IDadtl, Opts{}, RAWC(LabIART))

		// --- When ---
		have := reg.getSub(IDINFO, LabIART, Opts{})

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("other options", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake())
		reg.putSub(IDINFO, Opts{}, INFO(LabIART))

		// --- When ---
		have := reg.getSub(IDINFO, LabIART, NewOpts(WithLoadData()))

		// --- Then ---
		assert.Nil(t, have)
	})
}
