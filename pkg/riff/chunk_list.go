// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Compile time checks.
var _ Chunk = (*ChunkLIST)(nil)

// IDLIST represents "LIST" chunk ID.
const IDLIST uint32 = 0x4c495354

// ListTypeSize is the size of the LIST chunk type field in bytes.
const ListTypeSize uint32 = 4

// LIST chunk types and sub-chunk IDs.
const (
	// IDlabl represents the "labl" sub-chunk ID of the "adtl" LIST chunk.
	IDlabl uint32 = 0x6C61626C

	// IDadtl represents the "adtl" LIST chunk type.
	IDadtl uint32 = 0x6164746C
)

// ChunkLIST represents the "LIST" chunk holding a list of sub-chunks.
type ChunkLIST struct {
	// List type.
	ListType uint32

	// Sub-chunks.
	chunks Chunks

	// Registered chunk decoders.
	reg *Registry

	// Set when chunks were supplied by the caller through Modify. Reset does
	// not return those to the registry pool, so later decoding never
	// overwrites them.
	borrowed bool

	// Chunk processing options.
	opts Opts
}

// LISTMake returns a [Maker] function for [ChunkLIST] instances.
func LISTMake(reg *Registry, opts ...OptsFn) Maker {
	return func() Chunk {
		return LIST(reg, opts...)
	}
}

// LIST returns a new instance of [ChunkLIST] decoding sub-chunks with
// decoders from reg. When reg is nil, a registry with only the raw chunk
// decoder is used.
func LIST(reg *Registry, opts ...OptsFn) *ChunkLIST {
	ch := &ChunkLIST{
		reg:  reg,
		opts: NewOpts(opts...),
	}
	return ch
}

func (ch *ChunkLIST) ID() uint32     { return IDLIST }
func (ch *ChunkLIST) Type() uint32   { return ch.ListType }
func (ch *ChunkLIST) Multi() bool    { return true }
func (ch *ChunkLIST) Chunks() Chunks { return ch.chunks }
func (ch *ChunkLIST) Raw() bool      { return false }

// Size returns the size of the list type field and the sub-chunks, computed
// from the sub-chunks on every call.
func (ch *ChunkLIST) Size() uint32 { return ListTypeSize + ch.chunks.Size() }

func (ch *ChunkLIST) ReadFrom(r io.Reader) (int64, error) {
	cr, size, err := ch.opts.ChunkReader(r)
	if err != nil {
		return 0, fmt.Errorf(errFmtDecode, Uint32(IDLIST), err)
	}
	if size < ListTypeSize {
		return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDLIST), ErrTooShort)
	}

	if err = binary.Read(cr, be, &ch.ListType); err != nil {
		return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDLIST), err)
	}
	name := linkids(IDLIST, ch.ListType)

	mkr := ch.subMaker()

	// Decode sub-chunks until all bytes declared in size are read.
	for cr.Count()-4 < int64(size) {
		var id uint32
		if err = ReadChunkID(cr, &id); err != nil {
			return cr.Count(), fmt.Errorf(errFmtDecode, name, err)
		}

		dec := ch.registry().GetNoRaw(id)
		if dec == nil {
			dec = mkr(id)
		}
		dec.Reset()

		_, err = dec.ReadFrom(cr)
		if err != nil {
			sub := linkids(IDLIST, id)
			return cr.Count(), fmt.Errorf(errFmtDecode, sub, err)
		}
		ch.chunks = append(ch.chunks, dec)

		// Sub-chunks take an even number of bytes, so an odd count means the
		// input ended without the final padding byte.
		if (cr.Count()-4)%2 != 0 {
			break
		}
	}

	if cr.Count()-4 > int64(size) {
		return cr.Count(), fmt.Errorf(errFmtDecode, name, ErrChunkSizeMismatch)
	}
	if shortRead(int64(size), cr.Count()-4) {
		return cr.Count(), fmt.Errorf(errFmtDecode, name, io.ErrUnexpectedEOF)
	}

	return cr.Count(), nil
}

func (ch *ChunkLIST) WriteTo(w io.Writer) (int64, error) {
	if !ch.opts.LoadData() {
		return 0, ErrSkipDataMode
	}

	size := uint64(ListTypeSize) + ch.chunks.size64()
	if size > math.MaxUint32 {
		return 0, fmt.Errorf(errFmtEncode, Uint32(IDLIST), ErrTooLarge)
	}

	var sum int64

	n, err := WriteIDAndSize(w, IDLIST, uint32(size))
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDLIST), err)
	}

	if err = binary.Write(w, be, ch.ListType); err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDLIST), err)
	}
	sum += 4

	n, err = ch.chunks.WriteTo(w)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDLIST), err)
	}

	return sum, nil
}

func (ch *ChunkLIST) Reset() {
	ch.ListType = 0
	if ch.borrowed {
		ch.chunks = nil
		ch.borrowed = false
		return
	}
	for _, dec := range ch.chunks {
		ch.registry().Put(dec)
	}
	ch.chunks = ch.chunks[:0]
}

// registry returns the chunk decoder registry, creating an empty one when
// the chunk has none (nil registry or zero value).
func (ch *ChunkLIST) registry() *Registry {
	if ch.reg == nil {
		ch.reg = NewRegistry(RAWCMake(WithOpts(ch.opts)))
	}
	return ch.reg
}

// subMaker returns the maker of sub-chunk decoders for the list type, used for
// IDs without a registered decoder.
func (ch *ChunkLIST) subMaker() IDMaker {
	switch ch.ListType {
	case IDINFO:
		return INFOMake(WithOpts(ch.opts))

	case IDadtl:
		return adtlMake(WithOpts(ch.opts))

	default:
		return RAWCMake(WithOpts(ch.opts))
	}
}

// adtlMake returns an [IDMaker] creating decoders for sub-chunks of the
// "adtl" LIST chunk: [ChunkLABL], [ChunkLTXT], and [ChunkRAWC] for the rest.
func adtlMake(opts ...OptsFn) IDMaker {
	return func(id uint32) Chunk {
		switch id {
		case IDlabl:
			return LABL(opts...)

		case IDltxt:
			return LTXT(opts...)

		default:
			return RAWC(id, opts...)
		}
	}
}

// Modify replaces the sub-chunks with chs. [ChunkLIST.Reset] drops chs
// instead of returning them to the registry pool.
func (ch *ChunkLIST) Modify(chs Chunks) {
	ch.chunks = chs
	ch.borrowed = true
}
