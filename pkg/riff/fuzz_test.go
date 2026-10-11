// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"
)

// fuzzSeedMax is the largest sample file, after shrinking its data chunk,
// used as a fuzz seed.
const fuzzSeedMax = 1 << 16

// nonSeeker hides the [io.Seeker] implementation of the wrapped reader.
type nonSeeker struct{ io.Reader }

// fuzzSeeds returns the kitchen sink file and the sample files with their
// data chunk shrunk to a few bytes.
func fuzzSeeds(t tester.T) [][]byte {
	t.Helper()

	seeds := [][]byte{kitchenSink(t)}
	pths := must.Value(filepath.Glob("testdata/*.*"))
	for _, pth := range pths {
		rif := New(WithLoadData())
		src := must.Value(os.Open(pth))
		_, err := rif.ReadFrom(src)
		_ = src.Close()
		if err != nil {
			continue
		}
		if ch, ok := rif.Chunks().First(IDdata).(*ChunkDATA); ok {
			data := must.Value(io.ReadAll(ch.Data()))
			must.Nil(ch.SetData(data[:min(len(data), 64)]))
		}
		dst := &bytes.Buffer{}
		must.Value(rif.WriteTo(dst))
		if dst.Len() <= fuzzSeedMax {
			seeds = append(seeds, dst.Bytes())
		}
	}
	return seeds
}

// encode returns the bytes rif encodes to and the error.
func encode(rif *RIFF) ([]byte, error) {
	dst := &bytes.Buffer{}
	_, err := rif.WriteTo(dst)
	return dst.Bytes(), err
}

// errStr returns the error message or an empty string when err is nil.
func errStr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// FuzzRIFF_ReadFrom_seeker checks decoding gives the same result whether the
// source implements [io.Seeker] or not.
func FuzzRIFF_ReadFrom_seeker(f *testing.F) {
	for _, seed := range fuzzSeeds(f) {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// --- Given ---
		want := New()
		wn, wErr := want.ReadFrom(nonSeeker{bytes.NewReader(data)})

		have := New()

		// --- When ---
		hn, hErr := have.ReadFrom(bytes.NewReader(data))

		// --- Then ---
		// The offset in a truncated input error depends on the reader.
		assert.Equal(t, wErr == nil, hErr == nil)
		wEOF := errors.Is(wErr, io.ErrUnexpectedEOF)
		assert.Equal(t, wEOF, errors.Is(hErr, io.ErrUnexpectedEOF))
		if wErr == nil && hErr == nil {
			assert.Equal(t, wn, hn)
			assert.Equal(t, want.Chunks().IDs(), have.Chunks().IDs())
		}
	})
}

// FuzzRIFF_ReadFrom_loadData checks decoding only metadata accepts the same
// files and finds the same chunks as decoding with [WithLoadData].
func FuzzRIFF_ReadFrom_loadData(f *testing.F) {
	for _, seed := range fuzzSeeds(f) {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// --- Given ---
		want := New(WithLoadData())
		wn, wErr := want.ReadFrom(bytes.NewReader(data))

		have := New()

		// --- When ---
		hn, hErr := have.ReadFrom(bytes.NewReader(data))

		// --- Then ---
		assert.Equal(t, wErr == nil, hErr == nil)
		if wErr == nil && hErr == nil {
			assert.Equal(t, wn, hn)
			assert.Equal(t, want.Chunks().IDs(), have.Chunks().IDs())
		}
	})
}

// FuzzRIFF_WriteTo checks a decoded file encodes to the size it reports and
// that its encoding decodes and encodes back to the same bytes.
func FuzzRIFF_WriteTo(f *testing.F) {
	for _, seed := range fuzzSeeds(f) {
		f.Add(seed, false)
		f.Add(seed, true)
	}

	f.Fuzz(func(t *testing.T, data []byte, ignore bool) {
		// --- Given ---
		opts := []OptsFn{WithLoadData()}
		if ignore {
			opts = append(opts, WithIgnoreSize())
		}

		rif := New(opts...)
		if _, err := rif.ReadFrom(bytes.NewReader(data)); err != nil {
			t.Skip()
		}

		// --- When ---
		have, err := encode(rif)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, uint32(len(have)), rif.Size()+8)

		again := New(opts...)
		_, err = again.ReadFrom(bytes.NewReader(have))
		assert.NoError(t, err)
		assert.Equal(t, have, must.Value(encode(again)))
	})
}

// FuzzRIFF_ReadFrom_reuse checks an instance that decoded another file
// decodes the same as a new one.
func FuzzRIFF_ReadFrom_reuse(f *testing.F) {
	seeds := fuzzSeeds(f)
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// --- Given ---
		rif := New(WithLoadData())
		must.Value(rif.ReadFrom(bytes.NewReader(seeds[0])))

		fresh := New(WithLoadData())
		_, wErr := fresh.ReadFrom(bytes.NewReader(data))

		// --- When ---
		_, hErr := rif.ReadFrom(bytes.NewReader(data))

		// --- Then ---
		assert.Equal(t, errStr(wErr), errStr(hErr))
		if wErr == nil && hErr == nil {
			want, wEncErr := encode(fresh)
			have, hEncErr := encode(rif)
			assert.Equal(t, errStr(wEncErr), errStr(hEncErr))
			assert.Equal(t, want, have)
		}
	})
}
