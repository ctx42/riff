// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"time"
)

// Compile time checks.
var _ Chunk = (*ChunkDATA)(nil)

// IDdata represents "data" chunk ID.
const IDdata uint32 = 0x64617461

// ChunkDATA represents the "data" chunk decoder.
type ChunkDATA struct {
	// Chunk size in bytes.
	// The ID, size field, and padding byte are not counted in the chunk size.
	size uint32

	// Buffer the chunk data is read into.
	data []byte

	// Chunk processing options.
	opts Opts
}

func (ch *ChunkDATA) ID() uint32     { return IDdata }
func (ch *ChunkDATA) Size() uint32   { return ch.size }
func (ch *ChunkDATA) Type() uint32   { return 0 }
func (ch *ChunkDATA) Multi() bool    { return false }
func (ch *ChunkDATA) Chunks() Chunks { return nil }
func (ch *ChunkDATA) Raw() bool      { return false }

// DATAMake returns a [Maker] function for [ChunkDATA] instances.
func DATAMake(opts ...OptsFn) Maker {
	return func() Chunk {
		return DATA(opts...)
	}
}

// DATA returns a new instance of [ChunkDATA]. Without [WithLoadData] the chunk
// data is skipped instead of loaded into memory, which reduces the memory
// footprint when only metadata is of interest.
func DATA(opts ...OptsFn) *ChunkDATA {
	ch := &ChunkDATA{
		opts: NewOpts(opts...),
	}
	if ch.opts.LoadData() {
		ch.data = make([]byte, 0, 1<<15)
	}
	return ch
}

// Data returns a reader for the chunk data. Without [WithLoadData] the reader
// is empty.
func (ch *ChunkDATA) Data() io.Reader {
	return bytes.NewReader(ch.data)
}

// SetData sets the chunk data. It returns [ErrSkipDataMode] when the chunk was
// created without [WithLoadData] and [ErrTooLarge] when data does not fit the
// 32-bit chunk size.
func (ch *ChunkDATA) SetData(data []byte) error {
	if !ch.opts.LoadData() {
		return ErrSkipDataMode
	}
	l := len(data)
	if uint64(l) > math.MaxUint32 {
		return ErrTooLarge
	}
	ch.data = grow(ch.data, l)
	copy(ch.data, data)
	ch.size = uint32(l)
	return nil
}

// Duration returns file duration given average byte rate abr. It returns
// zero when abr is zero.
func (ch *ChunkDATA) Duration(abr uint32) time.Duration {
	if abr == 0 {
		return 0
	}
	dur := float64(ch.size) / float64(abr)
	dur *= float64(time.Second)
	return time.Duration(dur)
}

func (ch *ChunkDATA) ReadFrom(r io.Reader) (int64, error) {
	cr, size, err := ch.opts.ChunkReader(r)
	if err != nil {
		return 0, fmt.Errorf(errFmtDecode, Uint32(IDdata), err)
	}
	ch.size = size

	if !ch.opts.LoadData() {
		if err = cr.skipBody(ch.size); err != nil {
			return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDdata), err)
		}
		return cr.Count(), nil
	}

	ch.data, err = readBytes(cr, ch.data, int(ch.size))
	if err != nil {
		return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDdata), err)
	}

	_, err = ReadPaddingIf(cr, ch.size)
	if err != nil {
		return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDdata), err)
	}

	return cr.Count(), nil
}

func (ch *ChunkDATA) WriteTo(w io.Writer) (int64, error) {
	if !ch.opts.LoadData() {
		return 0, ErrSkipDataMode
	}

	var sum int64

	n, err := WriteIDAndSize(w, IDdata, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDdata), err)
	}

	in, err := w.Write(ch.data)
	sum += int64(in)
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDdata), err)
	}

	n, err = WritePaddingIf(w, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDdata), err)
	}

	return sum, nil
}

func (ch *ChunkDATA) Reset() {
	ch.size = 0
	ch.data = ch.data[:0]
}
