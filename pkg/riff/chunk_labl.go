// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"fmt"
	"io"
	"math"
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

	// Reader of the chunk content, reused across decodes.
	cr ChunkReader

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
func (ch *ChunkLABL) Type() uint32   { return 0 }
func (ch *ChunkLABL) Multi() bool    { return true }
func (ch *ChunkLABL) Chunks() Chunks { return nil }
func (ch *ChunkLABL) Raw() bool      { return false }

// Size returns the chunk size computed from its current content.
func (ch *ChunkLABL) Size() uint32 {
	return LABLChunkSize + size32(len(ch.label))
}

// Label returns a reader for the label without trailing zero bytes.
func (ch *ChunkLABL) Label() io.Reader {
	return bytes.NewReader(TrimZeroRight(ch.label))
}

// SetLabel sets the label. The label is stored as a zero-terminated string,
// so a zero byte is appended when label does not end with one. It returns
// [ErrTooLarge] when the label does not fit the 32-bit chunk size.
func (ch *ChunkLABL) SetLabel(label []byte) error {
	var err error
	ch.label, err = setZSTR(ch.label, label, math.MaxUint32-LABLChunkSize)
	return err
}

func (ch *ChunkLABL) ReadFrom(r io.Reader) (int64, error) {
	cr := &ch.cr
	size, err := ch.opts.startChunk(cr, r)
	if err != nil {
		return 0, ch.errDecode(err)
	}

	if size < LABLChunkSize {
		return cr.Count(), ch.errDecode(ErrTooShort)
	}

	if ch.CuePointID, err = cr.readUint32LE(); err != nil {
		return cr.Count(), ch.errDecode(err)
	}

	// Subtract the cue point ID field size.
	ch.label, err = readBytes(cr, ch.label, int(size-LABLChunkSize))
	if err != nil {
		return cr.Count(), ch.errDecode(err)
	}

	// If the length of label bytes is odd, it means the padding byte was added
	// to the end.
	_, err = ReadPaddingIf(cr, size)
	if err != nil {
		return cr.Count(), ch.errDecode(err)
	}

	return cr.Count(), nil
}

// errDecode returns err wrapped with the chunk name. The name is built only
// for an error, so decoding allocates nothing for it.
func (ch *ChunkLABL) errDecode(err error) error {
	return fmt.Errorf(errFmtDecode, linkids(IDadtl, IDlabl), err)
}

func (ch *ChunkLABL) WriteTo(w io.Writer) (int64, error) {
	var sum int64

	size := ch.Size()
	b := ch.cr.buf[:] // The reader is idle while encoding.
	n, err := writeIDAndSize(w, b, IDlabl, size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDadtl, IDlabl), err)
	}

	le.PutUint32(b, ch.CuePointID)
	if _, err = w.Write(b[:4]); err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDadtl, IDlabl), err)
	}
	sum += int64(LABLChunkSize)

	in, err := w.Write(ch.label)
	sum += int64(in)
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDadtl, IDlabl), err)
	}

	n, err = writePaddingIf(w, b, size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDadtl, IDlabl), err)
	}

	return sum, nil
}

func (ch *ChunkLABL) Reset() {
	ch.CuePointID = 0
	ch.label = ch.label[:0]
}
