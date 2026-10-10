// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

// Command generate writes the generated WAVE test files to the directory
// given as the first argument, the current directory by default. Run it from
// the pkg/riff directory:
//
//	go run ./testdata/generate testdata
//
// The files carry the chunk layout of the encoders they model, with silent
// (zero) sample data. They are encoded without the riff package so its tests
// decode input it did not write.
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// codec describes a WAVE file holding a fmt, an optional fact, and a data
// chunk.
type codec struct {
	name       string // File name.
	code       uint16 // Compression code.
	channels   uint16 // Channel count.
	sampleRate uint32 // Sample rate in Hz.
	byteRate   uint32 // Average byte rate.
	blockAlign uint16 // Block align.
	bits       uint16 // Bits per sample.
	extra      []byte // Extra format bytes; nil writes no extra size field.
	samples    uint32 // Sample count of the fact chunk; zero writes none.
	dataSize   uint32 // Size of the data chunk.
}

// codecs model files encoded with Windows Audio Compression Manager codecs.
var codecs = []codec{
	{"11k16bitpcm.wav", 0x01, 1, 11025, 22050, 2, 16, nil, 0, 304534},
	{"11k8bitpcm.wav", 0x01, 1, 11025, 11025, 1, 8, nil, 0, 152267},
	{"11kadpcm.wav", 0x11, 1, 11025, 5588, 256, 4, []byte{0xf9, 0x01}, 152267, 77192},
	{"11kgsm.wav", 0x31, 1, 11025, 2239, 65, 0, []byte{0x40, 0x01}, 152267, 30940},
	{"11kulaw.wav", 0x07, 1, 11025, 11025, 1, 8, []byte{}, 152267, 152267},
	{"8k16bitpcm.wav", 0x01, 1, 8000, 16000, 2, 16, nil, 0, 220982},
	{"8k8bitpcm.wav", 0x01, 1, 8000, 8000, 1, 8, nil, 0, 110488},
	{"8kadpcm.wav", 0x11, 1, 8000, 4055, 256, 4, []byte{0xf9, 0x01}, 110488, 56012},
	{"8kcelp.wav", 0x70, 1, 8000, 600, 12, 16, []byte{}, 110488, 8292},
	{"8kgsm.wav", 0x31, 1, 8000, 1625, 65, 0, []byte{0x40, 0x01}, 110488, 22490},
	{
		"8kmp316.wav", 0x55, 1, 8000, 2000, 1, 0,
		[]byte{1, 0, 2, 0, 0, 0, 0x90, 0, 1, 0, 0x71, 0x05},
		110488, 27315,
	},
	{
		"8kmp38.wav", 0x55, 1, 8000, 1000, 1, 0,
		[]byte{1, 0, 2, 0, 0, 0, 0x48, 0, 1, 0, 0x71, 0x05},
		110488, 13514,
	},
	{"8ksbc12.wav", 0x72, 1, 8000, 2312, 37, 16, []byte{}, 110488, 19600},
	{
		"8ktruespeech.wav", 0x22, 1, 8000, 1067, 32, 1,
		append([]byte{1, 0, 0xf0, 0}, make([]byte, 28)...),
		110488, 14752,
	},
	{"8kulaw.wav", 0x07, 1, 8000, 8000, 1, 8, []byte{}, 110488, 110488},
}

// infoTags are the INFO tags of sample16bit.wav, in file order.
var infoTags = []struct {
	id   string
	text string
}{
	{"IART", "Artist"},
	{"ICMT", "Comment"},
	{"ICRD", "1797"},
	{"IGNR", "Classical"},
	{"INAM", "Title"},
	{"IPRD", "Album"},
	{"ITRK", "1/1"},
}

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	if err := run(dir); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run writes all generated files to dir.
func run(dir string) error {
	for _, cdc := range codecs {
		if err := write(dir, cdc.name, cdc.encode()); err != nil {
			return err
		}
	}
	return write(dir, "sample16bit.wav", sample16bit())
}

// encode returns the WAVE file the codec describes.
func (cdc codec) encode() []byte {
	fmtc := &bytes.Buffer{}
	le(fmtc, cdc.code, cdc.channels, cdc.sampleRate, cdc.byteRate)
	le(fmtc, cdc.blockAlign, cdc.bits)
	if cdc.extra != nil {
		le(fmtc, uint16(len(cdc.extra))) //nolint:gosec
		fmtc.Write(cdc.extra)
	}

	body := &bytes.Buffer{}
	chunk(body, "fmt ", fmtc.Bytes())
	if cdc.samples != 0 {
		fact := &bytes.Buffer{}
		le(fact, cdc.samples)
		chunk(body, "fact", fact.Bytes())
	}
	chunk(body, "data", make([]byte, cdc.dataSize))
	return wave(body.Bytes())
}

// sample16bit returns five seconds of 16-bit stereo silence at 44.1 kHz
// followed by an INFO list.
func sample16bit() []byte {
	fmtc := &bytes.Buffer{}
	le(fmtc, uint16(1), uint16(2), uint32(44100), uint32(176400))
	le(fmtc, uint16(4), uint16(16))

	info := &bytes.Buffer{}
	info.WriteString("INFO")
	for _, tag := range infoTags {
		chunk(info, tag.id, append([]byte(tag.text), 0))
	}

	body := &bytes.Buffer{}
	chunk(body, "fmt ", fmtc.Bytes())
	chunk(body, "data", make([]byte, 5*176400))
	chunk(body, "LIST", info.Bytes())
	return wave(body.Bytes())
}

// wave returns body wrapped in a RIFF chunk of the WAVE type.
func wave(body []byte) []byte {
	buf := &bytes.Buffer{}
	buf.WriteString("RIFF")
	le(buf, uint32(4+len(body))) //nolint:gosec
	buf.WriteString("WAVE")
	buf.Write(body)
	return buf.Bytes()
}

// chunk writes a chunk with the given ID and data to buf, followed by a zero
// pad byte when the data size is odd.
func chunk(buf *bytes.Buffer, id string, data []byte) {
	buf.WriteString(id)
	le(buf, uint32(len(data))) //nolint:gosec
	buf.Write(data)
	if len(data)%2 == 1 {
		buf.WriteByte(0)
	}
}

// le writes vs to buf in little endian.
func le(buf *bytes.Buffer, vs ...any) {
	for _, v := range vs {
		_ = binary.Write(buf, binary.LittleEndian, v)
	}
}

// write writes data to the file name in dir.
func write(dir, name string, data []byte) error {
	return os.WriteFile(filepath.Join(dir, name), data, 0o600) //nolint:gosec
}
