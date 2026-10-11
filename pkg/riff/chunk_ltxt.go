// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"fmt"
	"io"
)

// Compile time checks.
var _ Chunk = (*ChunkLTXT)(nil)

// IDltxt represents LIST sub-chunk ID "ltxt".
const IDltxt uint32 = 0x6C747874

// LTXTChunkSize is the size in bytes of the static part of the ltxt chunk, not
// counting the ID and the text bytes.
const LTXTChunkSize uint32 = 20

// ltxtStatic represents chunk static data (always there).
// This struct is defined separately to decode and encode it as one
// fixed-size block.
type ltxtStatic struct {
	// Cue Point ID specifies the starting sample point that corresponds to
	// this text label by providing the ID of a Cue Point defined in the
	// Cue Point List. The ID that associates this label with a Cue Point
	// must be unique to all other note chunk Cue Point IDs.
	CuePointID uint32

	// The sample length defines how many samples from the cue point the
	// region or section spans.
	SamLen uint32

	// The purpose field specifies what the text is used for. For example,
	// a value of "scrp" means script text, and "capt" means close-caption.
	// There are several more purpose IDs, but they are meant to be used
	// with other types of RIFF files (not usually found in WAVE files).
	PurID uint32

	// Information about the location used by the text, typically used
	// for queries to collect information from the operating system.
	Country uint16

	// Information about the language used by the text, typically used
	// for queries to collect information from the operating system.
	Language uint16

	// Information about the dialect used by the text, typically used
	// for queries to collect information from the operating system.
	Dialect uint16

	// Information about the code page used by the text, typically used
	// for queries to collect information from the operating system.
	CodePage uint16
}

// decode sets the fields from b, the little endian encoded static data.
func (lts *ltxtStatic) decode(b []byte) {
	lts.CuePointID = le.Uint32(b[0:])
	lts.SamLen = le.Uint32(b[4:])
	lts.PurID = le.Uint32(b[8:])
	lts.Country = le.Uint16(b[12:])
	lts.Language = le.Uint16(b[14:])
	lts.Dialect = le.Uint16(b[16:])
	lts.CodePage = le.Uint16(b[18:])
}

// encode writes the fields to b, little endian encoded.
func (lts *ltxtStatic) encode(b []byte) {
	le.PutUint32(b[0:], lts.CuePointID)
	le.PutUint32(b[4:], lts.SamLen)
	le.PutUint32(b[8:], lts.PurID)
	le.PutUint16(b[12:], lts.Country)
	le.PutUint16(b[14:], lts.Language)
	le.PutUint16(b[16:], lts.Dialect)
	le.PutUint16(b[18:], lts.CodePage)
}

// ChunkLTXT represents the labeled text chunk. It is used to associate a
// text label with a region or section of waveform data. This information
// is often displayed in marked regions of a waveform in digital audio editors.
type ChunkLTXT struct {
	// Chunk size in bytes.
	// The ID, size field, and padding byte are not counted in the chunk size.
	size uint32

	ltxtStatic

	// Text is a null-terminated string of characters. If the number of
	// characters in the string is not even, padding must be appended to
	// the string. The appended padding is not considered in the note
	// chunk's chunk size field.
	text []byte

	// Reader of the chunk content, reused across decodes.
	cr ChunkReader

	// Chunk processing options.
	opts Opts
}

// LTXTMake returns a [Maker] function for [ChunkLTXT] instances.
func LTXTMake(opts ...OptsFn) Maker {
	return func() Chunk {
		return LTXT(opts...)
	}
}

// LTXT returns a new instance of [ChunkLTXT].
func LTXT(opts ...OptsFn) *ChunkLTXT {
	return &ChunkLTXT{
		opts: NewOpts(opts...),
	}
}

func (ch *ChunkLTXT) ID() uint32     { return IDltxt }
func (ch *ChunkLTXT) Size() uint32   { return ch.size }
func (ch *ChunkLTXT) Type() uint32   { return 0 }
func (ch *ChunkLTXT) Multi() bool    { return true }
func (ch *ChunkLTXT) Chunks() Chunks { return nil }
func (ch *ChunkLTXT) Raw() bool      { return false }

// Text returns a reader for the text without trailing zero bytes.
func (ch *ChunkLTXT) Text() io.Reader {
	return bytes.NewReader(TrimZeroRight(ch.text))
}

func (ch *ChunkLTXT) ReadFrom(r io.Reader) (int64, error) {
	cr := &ch.cr
	size, err := ch.opts.startChunk(cr, r)
	if err != nil {
		return 0, ch.errDecode(err)
	}
	ch.size = size

	if ch.size < LTXTChunkSize {
		return cr.Count(), ch.errDecode(ErrTooShort)
	}

	b, err := cr.readFull(int(LTXTChunkSize))
	if err != nil {
		return cr.Count(), ch.errDecode(err)
	}
	ch.decode(b)

	tl := int(ch.size - LTXTChunkSize) // Subtract ltxtStatic fields size.
	ch.text, err = readBytes(cr, ch.text, tl)
	if err != nil {
		return cr.Count(), ch.errDecode(err)
	}

	// If the length of text bytes is odd, it means the padding byte was added
	// to the end.
	_, err = ReadPaddingIf(cr, ch.size)
	if err != nil {
		return cr.Count(), ch.errDecode(err)
	}
	return cr.Count(), nil
}

// errDecode returns err wrapped with the chunk name. The name is built only
// for an error, so decoding allocates nothing for it.
func (ch *ChunkLTXT) errDecode(err error) error {
	return fmt.Errorf(errFmtDecode, linkids(IDadtl, IDltxt), err)
}

func (ch *ChunkLTXT) WriteTo(w io.Writer) (int64, error) {
	var sum int64

	b := ch.cr.buf[:] // The reader is idle while encoding.
	n, err := writeIDAndSize(w, b, IDltxt, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDadtl, IDltxt), err)
	}

	ch.encode(b)
	if _, err = w.Write(b[:LTXTChunkSize]); err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDadtl, IDltxt), err)
	}
	sum += int64(LTXTChunkSize)

	in, err := w.Write(ch.text)
	sum += int64(in)
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDadtl, IDltxt), err)
	}

	n, err = writePaddingIf(w, b, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDadtl, IDltxt), err)
	}

	return sum, nil
}

func (ch *ChunkLTXT) Reset() {
	ch.size = 0
	ch.CuePointID = 0
	ch.SamLen = 0
	ch.PurID = 0
	ch.Country = 0
	ch.Language = 0
	ch.Dialect = 0
	ch.CodePage = 0
	ch.text = ch.text[:0]
}
