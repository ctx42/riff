// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Compile time checks.
var _ Chunk = (*ChunkLABL)(nil)

// LABLChunkSize is the size in bytes of the static part of the labl chunk, not
// counting the ID and the label bytes.
const LABLChunkSize uint32 = 4

// ChunkLABL represents the "labl" chunk which is always contained inside an
// "adtl" LIST chunk. It's used to associate a text label with a Cue
// Point. This information is often displayed next to markers or flags in
// digital audio editors.
type ChunkLABL struct {
	// Chunk size in bytes.
	// The ID, size field, and padding byte are not counted in the chunk size.
	size uint32

	// The Cue Point ID specifies the sample point that corresponds to
	// this text label by providing the ID of a Cue Point defined in the
	// Cue Point List. The ID that associates this label with a Cue Point
	// must be unique to all other label Cue Point IDs.
	CuePointID uint32

	// The label is a null terminated string of characters. If the number of
	// characters in the string is not even, padding must be appended to
	// the string. The appended padding is not considered in the label
	// chunk's chunk size field.
	label []byte

	// Chunk processing options.
	opts Opts
}

// LABLMake returns a [Maker] function for [ChunkLABL] instances.
func LABLMake(opts ...OptsFn) Maker {
	return func() Chunk {
		return LABL(opts...)
	}
}

// LABL returns a new instance of [ChunkLABL].
func LABL(opts ...OptsFn) *ChunkLABL {
	return &ChunkLABL{
		opts: NewOpts(opts...),
	}
}

func (ch *ChunkLABL) ID() uint32     { return IDlabl }
func (ch *ChunkLABL) Size() uint32   { return ch.size }
func (ch *ChunkLABL) Type() uint32   { return 0 }
func (ch *ChunkLABL) Multi() bool    { return true }
func (ch *ChunkLABL) Chunks() Chunks { return nil }
func (ch *ChunkLABL) Raw() bool      { return false }

// Label returns a reader for the label without trailing zero bytes.
func (ch *ChunkLABL) Label() io.Reader {
	return bytes.NewReader(TrimZeroRight(ch.label))
}

func (ch *ChunkLABL) ReadFrom(r io.Reader) (int64, error) {
	cr, size, err := ch.opts.ChunkReader(r)
	if err != nil {
		return 0, fmt.Errorf(errFmtDecode, linkids(IDadtl, IDlabl), err)
	}
	ch.size = size

	if ch.size < LABLChunkSize {
		return cr.Count(), fmt.Errorf(errFmtDecode, linkids(IDadtl, IDlabl), ErrTooShort)
	}

	if err := binary.Read(cr, le, &ch.CuePointID); err != nil {
		return cr.Count(), fmt.Errorf(errFmtDecode, linkids(IDadtl, IDlabl), err)
	}

	// Subtract the cue point ID field size.
	ch.label, err = readBytes(cr, ch.label, int(ch.size-LABLChunkSize))
	if err != nil {
		return cr.Count(), fmt.Errorf(errFmtDecode, linkids(IDadtl, IDlabl), err)
	}

	// If the length of label bytes is odd, it means the padding byte was added
	// to the end.
	_, err = ReadPaddingIf(cr, ch.size)
	if err != nil {
		return cr.Count(), fmt.Errorf(errFmtDecode, linkids(IDadtl, IDlabl), err)
	}

	return cr.Count(), nil
}

func (ch *ChunkLABL) WriteTo(w io.Writer) (int64, error) {
	var sum int64

	n, err := WriteIDAndSize(w, IDlabl, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDadtl, IDlabl), err)
	}

	if err = binary.Write(w, le, ch.CuePointID); err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDadtl, IDlabl), err)
	}
	sum += int64(LABLChunkSize)

	in, err := w.Write(ch.label)
	sum += int64(in)
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDadtl, IDlabl), err)
	}

	n, err = WritePaddingIf(w, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDadtl, IDlabl), err)
	}

	return sum, nil
}

func (ch *ChunkLABL) Reset() {
	ch.size = 0
	ch.CuePointID = 0
	ch.label = ch.label[:0]
}
