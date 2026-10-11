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

		// --- When ---
		reg.Put(ch)

		// --- Then ---
		assert.Same(t, ch, reg.GetNoRaw(IDfmt))
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

func Test_Registry_Get(t *testing.T) {
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
