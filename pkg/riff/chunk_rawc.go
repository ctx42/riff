// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"fmt"
	"io"
)

// Compile time checks.
var _ Chunk = (*ChunkRAWC)(nil)

// idRAWC is the "RAWC" ID naming raw chunks in error messages.
const idRAWC uint32 = 0x52415743

// ChunkRAWC is a decoder for unknown (not registered) chunks.
type ChunkRAWC struct {
	// Chunk ID.
	id uint32

	// Chunk size in bytes.
	// The ID, size field, and padding byte are not counted in the chunk size.
	size uint32

	// Buffer to read the chunk data into.
	data []byte

	// Reader of the chunk content, reused across decodes.
	cr ChunkReader

	// Chunk processing options.
	opts Opts
}

// RAWCMake returns [IDMaker] function for creating [ChunkRAWC] instances.
func RAWCMake(opts ...OptsFn) IDMaker {
	return func(id uint32) Chunk {
		return RAWC(id, opts...)
	}
}

// RAWC returns a new instance of [ChunkRAWC] for the given ID.
func RAWC(id uint32, opts ...OptsFn) *ChunkRAWC {
	ch := &ChunkRAWC{
		id:   id,
		opts: NewOpts(opts...),
	}

	if ch.opts.LoadData() {
		ch.data = make([]byte, 0, 1<<8)
	}

	return ch
}

func (ch *ChunkRAWC) ID() uint32     { return ch.id }
func (ch *ChunkRAWC) Size() uint32   { return ch.size }
func (ch *ChunkRAWC) Type() uint32   { return 0 }
func (ch *ChunkRAWC) Multi() bool    { return true }
func (ch *ChunkRAWC) Chunks() Chunks { return nil }
func (ch *ChunkRAWC) Raw() bool      { return true }

// Body returns a reader for the chunk body. Without [WithLoadData] the reader
// is empty.
func (ch *ChunkRAWC) Body() io.Reader {
	return bytes.NewReader(ch.data)
}

func (ch *ChunkRAWC) ReadFrom(r io.Reader) (int64, error) {
	cr := &ch.cr
	size, err := ch.opts.startChunk(cr, r)
	if err != nil {
		return 0, ch.errDecode(err)
	}
	ch.size = size

	if !ch.opts.LoadData() {
		if err = cr.skipBody(ch.size); err != nil {
			return cr.Count(), ch.errDecode(err)
		}
		return cr.Count(), nil
	}

	ch.data, err = readBytes(cr, ch.data, int(ch.size))
	if err != nil {
		return cr.Count(), ch.errDecode(err)
	}

	_, err = ReadPaddingIf(cr, ch.size)
	if err != nil {
		return cr.Count(), ch.errDecode(err)
	}

	return cr.Count(), nil
}

// errDecode returns err wrapped with the chunk name. The name is built only
// for an error, so decoding allocates nothing for it.
func (ch *ChunkRAWC) errDecode(err error) error {
	return fmt.Errorf(errFmtDecode, linkids(idRAWC, ch.id), err)
}

func (ch *ChunkRAWC) WriteTo(w io.Writer) (int64, error) {
	if !ch.opts.LoadData() {
		return 0, ErrSkipDataMode
	}

	var sum int64

	b := ch.cr.buf[:] // The reader is idle while encoding.
	n, err := writeIDAndSize(w, b, ch.id, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(idRAWC, ch.id), err)
	}

	in, err := w.Write(ch.data)
	sum += int64(in)
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(idRAWC, ch.id), err)
	}

	n, err = writePaddingIf(w, b, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(idRAWC, ch.id), err)
	}

	return sum, nil
}

func (ch *ChunkRAWC) Reset() {
	ch.size = 0
	ch.data = ch.data[:0]
}
