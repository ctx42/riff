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
var _ Chunk = (*ChunkSMPL)(nil)

// IDsmpl represents "smpl" chunk ID.
const IDsmpl uint32 = 0x736d706c

// SMPLChunkSize is the size in bytes of the static part of the smpl chunk, not
// counting the ID, the sample loops, and the sampler data.
const SMPLChunkSize uint32 = 36

// SampleLoopSize is the size of a single sample loop in bytes.
const SampleLoopSize uint32 = 24

// smplStatic represents chunk static data (always there).
// This struct is defined separately to decode and encode it as one
// fixed-size block.
type smplStatic struct {
	// The manufacturer field specifies the MIDI Manufacturer's
	// Association (MMA) Manufacturer code for the sampler intended
	// to receive this file's waveform. Each manufacturer of a MIDI
	// product is assigned a unique ID which identifies the company.
	// If no particular manufacturer is to be specified, a value of
	// 0 should be used.
	//
	// The value is stored with some extra information to enable
	// translation to the value used in a MIDI System Exclusive
	// transmission to the sampler. The high byte indicates the number
	// of low-order bytes (1 or 3) that are valid for the manufacturer
	// code. For example, the value for Digidesign will be 0x01000013 (0x13)
	// and the value for Microsoft will be 0x30000041 (0x00, 0x00, 0x41).
	//
	// See the MIDI Manufacturers List for a list:
	// https://www.midi.org/specifications-old/item/manufacturer-id-numbers
	Manufacturer uint32

	// The product field specifies the MIDI model ID defined by the
	// manufacturer corresponding to the Manufacturer field. Contact
	// the manufacturer of the sampler to get the model ID. If no
	// particular manufacturer's product is to be specified, a value
	// of 0 should be used.
	Product uint32

	// The sample period specifies the duration of time that passes during
	// the playback of one sample in nanoseconds (normally equal to
	// 1 / Samples Per Second, where Samples Per Second is the value
	// found in the ChunkFMT chunk).
	SamplePeriod uint32

	// The MIDI unity note value has the same meaning as the instrument
	// chunk's MIDI Unshifted Note field which specifies the musical note
	// at which the sample will be played at its original sample
	// rate (the sample rate specified in the ChunkFMT chunk).
	MIDIUnityNote uint32

	// The MIDI pitch fraction specifies the fraction of a semitone up
	// from the specified MIDI unity note field. A value of 0x80000000
	// means 1/2 semitone (50 cents), and a value of 0x00000000 means no
	// fine-tuning between semitones.
	MIDIPitchFraction uint32

	// The SMPTE format specifies the Society of Motion Pictures and
	// Television E time format used in the following SMPTE Offset field.
	// If a value of 0 is set, SMPTE Offset should also be set to 0.
	// SMPTE Format:
	//
	//     0 - no SMPTE offset
	//     24 - 24 frames per second
	//     25 - 25 frames per second
	//     29 - 30 frames per second with frame dropping (30 drop)
	//     30 - 30 frames per second
	//
	SMPTEFormat uint32

	// The SMPTE Offset value specifies the time offset to be used for
	// the synchronization / calibration to the first sample in the waveform.
	// This value uses a format of 0xhhmmssff where hh is a signed value
	// that specifies the number of hours (-23 to 23), mm is an unsigned
	// value that specifies the number of minutes (0 to 59), ss is an
	// unsigned value that specifies the number of seconds (0 to 59)
	// and ff is an unsigned value that specifies the number of
	// frames (0 to -1).
	SMPTEOffset uint32

	// The sample loops field specifies the number of Sample Loop definitions
	// in the following list. This value may be set to 0, meaning that no
	// sample loops follow.
	SampleLoopCnt uint32

	// The sampler data value specifies the size in bytes of the optional
	// sampler specific data that follows the sample loops; it is 0 when there
	// is none. Decoding reads the sampler data by the chunk size instead, and
	// encoding sets the value from the sampler data.
	SamplerDataCnt uint32
}

// decode sets the fields from b, the little endian encoded static data.
func (sst *smplStatic) decode(b []byte) {
	sst.Manufacturer = le.Uint32(b[0:])
	sst.Product = le.Uint32(b[4:])
	sst.SamplePeriod = le.Uint32(b[8:])
	sst.MIDIUnityNote = le.Uint32(b[12:])
	sst.MIDIPitchFraction = le.Uint32(b[16:])
	sst.SMPTEFormat = le.Uint32(b[20:])
	sst.SMPTEOffset = le.Uint32(b[24:])
	sst.SampleLoopCnt = le.Uint32(b[28:])
	sst.SamplerDataCnt = le.Uint32(b[32:])
}

// encode writes the fields to b, little endian encoded.
func (sst *smplStatic) encode(b []byte) {
	le.PutUint32(b[0:], sst.Manufacturer)
	le.PutUint32(b[4:], sst.Product)
	le.PutUint32(b[8:], sst.SamplePeriod)
	le.PutUint32(b[12:], sst.MIDIUnityNote)
	le.PutUint32(b[16:], sst.MIDIPitchFraction)
	le.PutUint32(b[20:], sst.SMPTEFormat)
	le.PutUint32(b[24:], sst.SMPTEOffset)
	le.PutUint32(b[28:], sst.SampleLoopCnt)
	le.PutUint32(b[32:], sst.SamplerDataCnt)
}

// ChunkSMPL represents the "smpl" (sampler) chunk.
//
// Source:
// https://sites.google.com/site/musicgapi/technical-documents/wav-file-format
type ChunkSMPL struct {
	smplStatic

	// A list of sample loops is simply a set of consecutive loop descriptions.
	// The sample loops do not have to be in any particular order because
	// each sample loop associated cue point position is used to determine
	// the play order. The sampler chunk is optional.
	SampleLoops []*SampleLoop

	// Optional sampler-specific data.
	sampleData []byte

	// Reader of the chunk content, reused across decodes.
	cr ChunkReader

	// Chunk processing options.
	opts Opts
}

// SMPLMake returns a [Maker] function for [ChunkSMPL] instances.
func SMPLMake(opts ...OptsFn) Maker {
	return func() Chunk {
		return SMPL(opts...)
	}
}

// SMPL returns a new instance of [ChunkSMPL].
func SMPL(opts ...OptsFn) *ChunkSMPL {
	return &ChunkSMPL{
		opts: NewOpts(opts...),
	}
}

func (ch *ChunkSMPL) ID() uint32     { return IDsmpl }
func (ch *ChunkSMPL) Type() uint32   { return 0 }
func (ch *ChunkSMPL) Multi() bool    { return false }
func (ch *ChunkSMPL) Chunks() Chunks { return nil }
func (ch *ChunkSMPL) Raw() bool      { return false }

// Size returns the chunk size computed from its current content, capped at
// [math.MaxUint32]. [ChunkSMPL.WriteTo] returns [ErrTooLarge] for content
// that does not fit.
func (ch *ChunkSMPL) Size() uint32 {
	size := ch.size64()
	if size > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(size)
}

// size64 returns the chunk size like [ChunkSMPL.Size], computed in 64 bits
// so it does not wrap.
func (ch *ChunkSMPL) size64() uint64 {
	return uint64(SMPLChunkSize) +
		uint64(len(ch.SampleLoops))*uint64(SampleLoopSize) +
		uint64(len(ch.sampleData))
}

// SamplerData returns a reader for the sampler-specific data.
func (ch *ChunkSMPL) SamplerData() io.Reader {
	return bytes.NewReader(ch.sampleData)
}

// SetSamplerData sets the sampler-specific data that follows the sample
// loops. It returns [ErrTooLarge] when the data does not fit the 32-bit chunk
// size.
func (ch *ChunkSMPL) SetSamplerData(data []byte) error {
	var err error
	room := math.MaxUint32 - SMPLChunkSize
	ch.sampleData, err = setBytes(ch.sampleData, data, room)
	return err
}

func (ch *ChunkSMPL) ReadFrom(r io.Reader) (int64, error) {
	ch.Reset() // Decoding replaces any previous content.

	cr := &ch.cr
	size, err := ch.opts.startChunk(cr, r)
	if err != nil {
		return 0, fmt.Errorf(errFmtDecode, Uint32(IDsmpl), err)
	}

	if size < SMPLChunkSize {
		return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDsmpl), ErrTooShort)
	}

	b, err := cr.readFull(int(SMPLChunkSize))
	if err != nil {
		return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDsmpl), err)
	}
	ch.decode(b)

	// We trust size more than SamplerDataCnt.
	loops := int64(ch.SampleLoopCnt) * int64(SampleLoopSize)
	extra := int64(size) - int64(SMPLChunkSize) - loops
	if extra < 0 {
		return cr.Count(), fmt.Errorf(
			errFmtDecode,
			Uint32(IDsmpl),
			ErrChunkSizeMismatch,
		)
	}

	// The loops share one backing array that grows as they are read, so a
	// large declared count cannot force a large allocation up front.
	var vals []SampleLoop
	for range ch.SampleLoopCnt {
		if b, err = cr.readFull(int(SampleLoopSize)); err != nil {
			return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDsmpl), err)
		}
		var slp SampleLoop
		slp.decode(b)
		vals = append(vals, slp)
	}
	for i := range vals {
		ch.SampleLoops = append(ch.SampleLoops, &vals[i])
	}

	ch.sampleData, err = readBytes(cr, ch.sampleData, int(extra))
	if err != nil {
		return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDsmpl), err)
	}

	_, err = ReadPaddingIf(cr, size)
	if err != nil {
		return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDsmpl), err)
	}

	return cr.Count(), nil
}

func (ch *ChunkSMPL) WriteTo(w io.Writer) (int64, error) {
	if ch.size64() > math.MaxUint32 {
		return 0, fmt.Errorf(errFmtEncode, Uint32(IDsmpl), ErrTooLarge)
	}

	var sum int64

	size := ch.Size()

	ch.SampleLoopCnt = size32(len(ch.SampleLoops))
	ch.SamplerDataCnt = size32(len(ch.sampleData))

	b := ch.cr.buf[:] // The reader is idle while encoding.
	n, err := writeIDAndSize(w, b, IDsmpl, size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDsmpl), err)
	}

	ch.encode(b)
	if _, err = w.Write(b[:SMPLChunkSize]); err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDsmpl), err)
	}
	sum += int64(SMPLChunkSize)

	for _, slp := range ch.SampleLoops {
		slp.encode(b)
		if _, err = w.Write(b[:SampleLoopSize]); err != nil {
			return sum, fmt.Errorf(errFmtEncode, Uint32(IDsmpl), err)
		}
		sum += int64(SampleLoopSize)
	}

	if len(ch.sampleData) > 0 {
		var in int
		in, err = w.Write(ch.sampleData)
		sum += int64(in)
		if err != nil {
			return sum, fmt.Errorf(errFmtEncode, Uint32(IDsmpl), err)
		}
	}

	n, err = writePaddingIf(w, b, size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDsmpl), err)
	}

	return sum, nil
}

func (ch *ChunkSMPL) Reset() {
	ch.Manufacturer = 0
	ch.Product = 0
	ch.SamplePeriod = 0
	ch.MIDIUnityNote = 0
	ch.MIDIPitchFraction = 0
	ch.SMPTEFormat = 0
	ch.SMPTEOffset = 0
	ch.SampleLoopCnt = 0
	ch.SamplerDataCnt = 0
	ch.SampleLoops = ch.SampleLoops[:0]
	ch.sampleData = ch.sampleData[:0]
}
