// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_Registry_Has(t *testing.T) {
	// --- Given ---
	reg := NewRegistry(RAWCMake(WithLoadData()))
	reg.Register(IDfmt, FMTMake())

	// --- Then ---
	assert.True(t, reg.Has(IDfmt))
	assert.False(t, reg.Has(IDLIST))
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

func Test_Registry_Get(t *testing.T) {
	t.Run("reuse pooled decoder", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry(RAWCMake(WithLoadData()))
		reg.Register(IDfmt, FMTMake())

		// --- When ---
		ch0 := reg.Get(IDfmt)
		reg.Put(ch0)
		ch1 := reg.Get(IDfmt)

		// --- Then ---
		assert.Same(t, ch0, ch1)
	})
}

func Test_Registry_GetNoRaw(t *testing.T) {
	// --- Given ---
	reg := NewRegistry(RAWCMake(WithLoadData()))

	// --- When ---
	ch0 := reg.GetNoRaw(IDfmt)

	// --- Then ---
	assert.Nil(t, ch0)
}
