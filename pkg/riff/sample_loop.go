// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

// SampleLoop represents a sample loop used in [ChunkSMPL].
//
// Source:
// https://sites.google.com/site/musicgapi/technical-documents/wav-file-format
type SampleLoop struct {
	// The Cue Point ID specifies the unique ID that corresponds to one
	// of the defined cue points in the cue point list. Furthermore,
	// this ID corresponds to any labels defined in the associated data
	// list chunk, which allows text labels to be assigned to the various
	// sample loops.
	CuePointID uint32

	// The type field defines how the waveform samples will be looped.
	// Loop types:
	//
	// - 0 - Loop forward (normal)
	// - 1 - Alternating loop (forward/backward, also known as Ping-Pong)
	// - 2 - Loop backward (reverse)
	// - 3 - 31 Reserved for future standard types
	// - 32 - 0xFFFFFFFF Sampler specific types (defined by a manufacturer)
	//
	Type uint32

	// The start value specifies the byte offset into the waveform data
	// of the first sample to be played in the loop.
	Start uint32

	// The end value specifies the byte offset into the waveform data of
	// the last sample to be played in the loop.
	End uint32

	// The fractional value specifies a fraction of a sample at which to loop.
	// This allows a loop to be fine-tuned at a resolution greater than one
	// sample. The value can range from 0x00000000 to 0xFFFFFFFF.
	// A value of 0 means no fraction,
	// a value of 0x80000000 means 1/2 of a sample length.
	// The 0xFFFFFFFF is the smallest fraction of a sample that
	// can be represented.
	Fraction uint32

	// The play count value determines the number of times to play the loop.
	// A value of 0 specifies an infinite sustain loop. An infinite sustain
	// loop will continue looping until some external force interrupts
	// playback, such as the musician releasing the key that triggered the
	// wave's playback. All other values specify an absolute number of
	// times to loop.
	PlayCnt uint32
}

// decode sets the fields from b, the little endian encoded sample loop.
func (slp *SampleLoop) decode(b []byte) {
	slp.CuePointID = le.Uint32(b[0:])
	slp.Type = le.Uint32(b[4:])
	slp.Start = le.Uint32(b[8:])
	slp.End = le.Uint32(b[12:])
	slp.Fraction = le.Uint32(b[16:])
	slp.PlayCnt = le.Uint32(b[20:])
}

// encode writes the fields to b, little endian encoded.
func (slp *SampleLoop) encode(b []byte) {
	le.PutUint32(b[0:], slp.CuePointID)
	le.PutUint32(b[4:], slp.Type)
	le.PutUint32(b[8:], slp.Start)
	le.PutUint32(b[12:], slp.End)
	le.PutUint32(b[16:], slp.Fraction)
	le.PutUint32(b[20:], slp.PlayCnt)
}

// Reset sets all fields to zero.
func (slp *SampleLoop) Reset() {
	slp.CuePointID = 0
	slp.Type = 0
	slp.Start = 0
	slp.End = 0
	slp.Fraction = 0
	slp.PlayCnt = 0
}
