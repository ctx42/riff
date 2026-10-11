// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

// Package riff provides low level tools for working with files in Resource
// Interchange File Format (RIFF).
package riff

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// Compile time checks.
var _ Chunk = (*RIFF)(nil)

// IDRIFF represents "RIFF" chunk ID.
const IDRIFF uint32 = 0x52494646

// RIFF file types, the type field of the [RIFF] chunk.
const (
	// TypeWAVE represents the "WAVE" file type.
	TypeWAVE uint32 = 0x57415645

	// TypeAVI represents the "AVI " file type.
	TypeAVI uint32 = 0x41564920

	// TypeRMID represents the "RMID" file type.
	TypeRMID uint32 = 0x524d4944
)

// Common chunk IDs.
const (
	// IDJUNK represents "JUNK" chunk ID.
	IDJUNK uint32 = 0x4a554e4b

	// IDID3 represents "ID3 " chunk ID.
	IDID3 uint32 = 0x49443320
)

// Byte order shorthands.
var (
	le = binary.LittleEndian
	be = binary.BigEndian
)

// Errors.
var (
	// ErrNotRIFF is returned when a file is not in the RIFF format.
	ErrNotRIFF = errors.New("not RIFF file")

	// ErrTooShort is returned when a chunk or field is shorter than its
	// defined length.
	ErrTooShort = errors.New("length too short")

	// ErrTooLarge is returned when a chunk size is greater than the set
	// maximum.
	ErrTooLarge = errors.New("size too large")

	// ErrChunkSizeMismatch is returned when a chunk size does not match its
	// content.
	ErrChunkSizeMismatch = errors.New("chunk size mismatch")

	// ErrPadByte is returned when the pad byte after odd-sized chunk data is
	// not zero.
	ErrPadByte = errors.New("non-zero pad byte")

	// ErrSkipDataMode is returned when a chunk decoded without
	// [WithLoadData] is written or edited.
	ErrSkipDataMode = errors.New(
		"decoder in meta only mode used in write context",
	)
)

// Error format strings.
const (
	// errFmtDecode is the format of chunk decoding errors.
	errFmtDecode = "error decoding %s chunk: %w"

	// errFmtEncode is the format of chunk encoding errors.
	errFmtEncode = "error encoding %s chunk: %w"

	// errFmtReadingRIFF is the format of RIFF chunk reading errors.
	errFmtReadingRIFF = "error reading RIFF offset %d: %w"
)

// Maker returns a new chunk decoder.
type Maker func() Chunk

// IDMaker returns a new chunk decoder for the chunk id.
type IDMaker func(id uint32) Chunk

// RIFF represents a file in Resource Interchange File Format.
type RIFF struct {
	// Determines the type of the resource (e.g.: WAVE).
	riffType uint32

	// List of decoded file chunks in order they appeared in the file.
	chunks Chunks

	// Registered chunk decoders.
	reg *Registry

	// Set when chunks were supplied by the caller through Modify. Reset does
	// not return those to the registry pool, so later decoding never
	// overwrites them.
	borrowed bool

	// Reader of the chunk content, reused across decodes.
	cr ChunkReader

	// Main chunk processing options.
	opts Opts
}

// New returns a new instance of [RIFF] with the built-in chunk decoders
// registered: fmt, data, LIST, and smpl.
func New(opts ...OptsFn) *RIFF {
	reg := NewRegistry(RAWCMake(opts...))

	// Register "out of the box" chunk decoders.
	reg.Register(IDfmt, FMTMake(opts...))
	reg.Register(IDdata, DATAMake(opts...))
	reg.Register(IDLIST, LISTMake(reg, opts...))
	reg.Register(IDsmpl, SMPLMake(opts...))

	return Bare(reg, opts...)
}

// Bare returns a new instance of [RIFF] decoding chunks with decoders from reg.
// If reg is nil, a registry with only the raw chunk decoder ([ChunkRAWC]) made
// with opts is used.
func Bare(reg *Registry, opts ...OptsFn) *RIFF {
	if reg == nil {
		reg = NewRegistry(RAWCMake(opts...))
	}
	rif := &RIFF{
		chunks: make([]Chunk, 0, 4),
		reg:    reg,
		opts:   NewOpts(opts...),
	}
	return rif
}

// Compose returns a new instance of [RIFF] holding chs. Use it to build a
// modified copy of a decoded file.
func Compose(chs Chunks) *RIFF {
	rif := Bare(nil)
	rif.Modify(chs)

	return rif
}

func (rif *RIFF) ID() uint32     { return IDRIFF }
func (rif *RIFF) Type() uint32   { return rif.riffType }
func (rif *RIFF) Multi() bool    { return false }
func (rif *RIFF) Chunks() Chunks { return rif.chunks }
func (rif *RIFF) Raw() bool      { return false }

// Size returns the size of the RIFF type field and the chunks, computed from
// the chunks on every call.
func (rif *RIFF) Size() uint32 { return 4 + rif.chunks.Size() }

// SetType sets the RIFF file type, such as [TypeWAVE].
func (rif *RIFF) SetType(t uint32) { rif.riffType = t }

// IsRegistered reports whether a decoder for id is registered.
func (rif *RIFF) IsRegistered(id uint32) bool {
	return rif.registry().Has(id)
}

// ReadFrom resets the instance and decodes a RIFF file from r.
//
// Decoding follows the RIFF specification strictly, except for these
// deviations, written by some encoders, which it tolerates:
//
//   - A missing pad byte after the final chunk. The specification writes a
//     zero pad byte after odd-sized chunk data. When the input ends right
//     after the data of the last chunk, also inside a LIST, the missing byte
//     is accepted. A pad byte missing anywhere else is an error.
//   - An odd RIFF size. The RIFF chunk holds its type and whole, padded
//     chunks, so its size is even. Encoders that leave the final pad byte out
//     of it write an odd size, which is rounded up.
//
// [RIFF.WriteTo] always writes a conforming file, with every pad byte and an
// even RIFF size.
func (rif *RIFF) ReadFrom(r io.Reader) (int64, error) {
	rif.Reset()

	var err error
	var id uint32

	cr := &rif.cr
	if _, err = io.ReadFull(r, cr.buf[:4]); err != nil {
		return 0, fmt.Errorf(errFmtDecode, Uint32(IDRIFF), err)
	}
	id = be.Uint32(cr.buf[:4])

	if id != IDRIFF {
		return 4, ErrNotRIFF
	}

	size, err := rif.opts.startChunk(cr, r)
	if err != nil {
		return 4, fmt.Errorf(errFmtDecode, Uint32(IDRIFF), err)
	}
	cr.setOffset(int(cr.Count()) + 4) // IDRIFF was already read.

	// An odd RIFF size is tolerated and rounded up (see the godoc).
	size = RealSize(size)

	if rif.riffType, err = cr.readUint32BE(); err != nil {
		return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDRIFF), err)
	}

	for {
		if id, err = cr.readUint32BE(); err != nil {
			if !errors.Is(err, io.EOF) {
				return cr.Count(), fmt.Errorf(
					errFmtReadingRIFF,
					cr.Count(),
					err,
				)
			}
			break
		}

		_, err = rif.decodeChunk(id, cr)
		if err != nil {
			return cr.Count(), fmt.Errorf(errFmtReadingRIFF, cr.Count(), err)
		}
	}
	// EOF reached. With ignored sizes nothing limits the input, but a RIFF
	// size is a 32-bit value.
	read := cr.Count() - 8
	if read < 0 || read > math.MaxUint32 {
		return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDRIFF), ErrTooLarge)
	}
	readSize := uint32(read)

	ignore := rif.opts.IgnoreDeclaredSize()
	if !ignore && shortRead(int64(size), int64(readSize)) {
		return cr.Count(), fmt.Errorf(
			"RIFF declared %d bytes, decoder read %d bytes: %w",
			size,
			readSize,
			io.ErrUnexpectedEOF,
		)
	}

	return cr.Count(), nil
}

func (rif *RIFF) WriteTo(w io.Writer) (int64, error) {
	size := 4 + rif.chunks.size64()
	if size > math.MaxUint32 {
		return 0, fmt.Errorf(errFmtEncode, Uint32(IDRIFF), ErrTooLarge)
	}

	var sum int64

	n, err := WriteIDAndSize(w, IDRIFF, uint32(size))
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDRIFF), err)
	}

	if err = binary.Write(w, be, rif.riffType); err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDRIFF), err)
	}
	sum += 4

	n, err = rif.chunks.WriteTo(w)
	sum += n
	if err != nil {
		return sum, err
	}

	return sum, nil
}

// Reset resets the decoded chunks and returns them to the registry pools, so
// chunks taken from [RIFF.Chunks] must not be used afterwards; chunks set by
// [RIFF.Modify] are dropped instead.
func (rif *RIFF) Reset() {
	if rif.borrowed {
		rif.chunks = nil
		rif.borrowed = false
		return
	}
	reg := rif.registry()
	for _, ch := range rif.chunks {
		reg.release(ch)
	}
	rif.chunks = rif.chunks[:0]
}

// registry returns the chunk decoder registry, creating one with only the raw
// chunk decoder when the instance has none (zero value).
func (rif *RIFF) registry() *Registry {
	if rif.reg == nil {
		rif.reg = NewRegistry(RAWCMake(WithOpts(rif.opts)))
	}
	return rif.reg
}

// decodeChunk decodes a chunk with id.
func (rif *RIFF) decodeChunk(id uint32, r io.Reader) (int64, error) {
	if rif.chunks.Count(id) > 0 && !rif.chunks.First(id).Multi() {
		return 0, fmt.Errorf("chunk %s (0x%x) already seen", Uint32(id), id)
	}
	dec := rif.registry().Get(id)
	n, err := dec.ReadFrom(r)
	if err != nil {
		return n, err
	}
	rif.chunks = append(rif.chunks, dec)
	return n, nil
}

// Modify replaces the chunks with chs. [RIFF.Reset] drops chs instead of
// returning them to the registry pool.
func (rif *RIFF) Modify(chs Chunks) {
	rif.chunks = chs
	rif.borrowed = true
}
