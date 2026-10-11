// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/ctx42/testing/pkg/tester"
)

// IDUNKN is a test chunk ID "ABCD" which is unlikely to be in the wild.
const IDUNKN uint32 = 0x41424344

// writeUint32LE writes v to dst in little endian. Fails the test on error.
func writeUint32LE(t tester.T, dst io.Writer, v uint32) {
	t.Helper()
	if err := binary.Write(dst, binary.LittleEndian, v); err != nil {
		t.Fatal(err)
	}
}

// writeUint32BE writes v to dst in big endian, the byte order of chunk IDs.
// Fails the test on error.
func writeUint32BE(t tester.T, dst io.Writer, v uint32) {
	t.Helper()
	if err := binary.Write(dst, binary.BigEndian, v); err != nil {
		t.Fatal(err)
	}
}

// writeUint16LE writes v to dst in little endian. Fails the test on error.
func writeUint16LE(t tester.T, dst io.Writer, v uint16) {
	t.Helper()
	if err := binary.Write(dst, binary.LittleEndian, v); err != nil {
		t.Fatal(err)
	}
}

// writeBytes writes p to dst. Fails the test on error.
func writeBytes(t tester.T, dst io.Writer, p []byte) {
	t.Helper()
	if _, err := dst.Write(p); err != nil {
		t.Fatal(err)
	}
}

// skip4B reads and discards exactly 4 bytes from r. Fails the test on error.
func skip4B(t tester.T, r io.Reader) {
	t.Helper()
	if _, err := io.ReadFull(r, make([]byte, 4)); err != nil {
		t.Fatal(err)
	}
}

// isAllRead returns true if r has no bytes left to read.
func isAllRead(r io.Reader) bool {
	_, err := r.Read(make([]byte, 1))
	return err == io.EOF
}

// kitchenSink returns a RIFF file with every built-in chunk type, odd sized
// chunks, and an unknown chunk.
func kitchenSink(t tester.T) []byte {
	t.Helper()

	fmtc := &bytes.Buffer{}
	writeUint32BE(t, fmtc, IDfmt)           // Chunk ID.
	writeUint32LE(t, fmtc, 20)              // Chunk size.
	writeBytes(t, fmtc, make([]byte, 16))   // Static fields.
	writeUint16LE(t, fmtc, 2)               // Extra size.
	writeBytes(t, fmtc, []byte{0x01, 0x02}) // Extra.

	smpl := &bytes.Buffer{}
	writeUint32BE(t, smpl, IDsmpl)                // Chunk ID.
	writeUint32LE(t, smpl, 36+24+3)               // Chunk size.
	writeBytes(t, smpl, make([]byte, 28))         // Static fields.
	writeUint32LE(t, smpl, 1)                     // Sample loop count.
	writeUint32LE(t, smpl, 3)                     // Sampler data size.
	writeBytes(t, smpl, make([]byte, 24))         // Sample loop.
	writeBytes(t, smpl, []byte{0x01, 0x02, 0x03}) // Sampler data.
	writeBytes(t, smpl, []byte{0})                // Padding byte.

	info := &bytes.Buffer{}
	writeUint32BE(t, info, IDLIST)                // Chunk ID.
	writeUint32LE(t, info, 4+12+12)               // Chunk size.
	writeUint32BE(t, info, IDINFO)                // List type.
	writeUint32BE(t, info, LabIART)               // Sub-chunk ID.
	writeUint32LE(t, info, 3)                     // Sub-chunk size.
	writeBytes(t, info, []byte{'a', 'b', 0, 0})   // Text and padding byte.
	writeUint32BE(t, info, LabISFT)               // Sub-chunk ID.
	writeUint32LE(t, info, 4)                     // Sub-chunk size.
	writeBytes(t, info, []byte{'a', 'b', 'c', 0}) // Text.

	adtl := &bytes.Buffer{}
	writeUint32BE(t, adtl, IDLIST)              // Chunk ID.
	writeUint32LE(t, adtl, 4+16+30)             // Chunk size.
	writeUint32BE(t, adtl, IDadtl)              // List type.
	writeUint32BE(t, adtl, IDlabl)              // Sub-chunk ID.
	writeUint32LE(t, adtl, 7)                   // Sub-chunk size.
	writeUint32LE(t, adtl, 1)                   // Cue point ID.
	writeBytes(t, adtl, []byte{'a', 'b', 0, 0}) // Label and padding byte.
	writeUint32BE(t, adtl, IDltxt)              // Sub-chunk ID.
	writeUint32LE(t, adtl, 22)                  // Sub-chunk size.
	writeBytes(t, adtl, make([]byte, 20))       // Static fields.
	writeBytes(t, adtl, []byte{'a', 0})         // Text.

	rawc := &bytes.Buffer{}
	writeUint32BE(t, rawc, IDUNKN)       // Chunk ID.
	writeUint32LE(t, rawc, 1)            // Chunk size.
	writeBytes(t, rawc, []byte{0x01, 0}) // Body and padding byte.

	data := &bytes.Buffer{}
	writeUint32BE(t, data, IDdata)                // Chunk ID.
	writeUint32LE(t, data, 3)                     // Chunk size.
	writeBytes(t, data, []byte{0x01, 0x02, 0x03}) // Samples.
	writeBytes(t, data, []byte{0})                // Padding byte.

	body := &bytes.Buffer{}
	for _, ch := range []*bytes.Buffer{fmtc, smpl, info, adtl, rawc, data} {
		writeBytes(t, body, ch.Bytes())
	}

	src := &bytes.Buffer{}
	writeUint32BE(t, src, IDRIFF)               // Chunk ID.
	writeUint32LE(t, src, uint32(4+body.Len())) // Chunk size.
	writeUint32BE(t, src, TypeWAVE)             // RIFF type.
	writeBytes(t, src, body.Bytes())            // Sub-chunks.
	return src.Bytes()
}

// sparseReader reads size bytes of zeros overlaid with segments keyed by
// offset, and seeks without reading. It simulates large inputs without
// allocating them.
type sparseReader struct {
	size int64
	pos  int64
	segs map[int64][]byte
}

func (rdr *sparseReader) Read(p []byte) (int, error) {
	if rdr.pos >= rdr.size {
		return 0, io.EOF
	}
	n := min(int64(len(p)), rdr.size-rdr.pos)
	clear(p[:n])
	for off, seg := range rdr.segs {
		lo := max(off, rdr.pos)
		hi := min(off+int64(len(seg)), rdr.pos+n)
		if lo < hi {
			copy(p[lo-rdr.pos:hi-rdr.pos], seg[lo-off:hi-off])
		}
	}
	rdr.pos += n
	return int(n), nil
}

func (rdr *sparseReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		rdr.pos = offset

	case io.SeekCurrent:
		rdr.pos += offset

	case io.SeekEnd:
		rdr.pos = rdr.size + offset
	}
	return rdr.pos, nil
}

// sizedChunk is a chunk reporting a size without holding any data.
type sizedChunk struct{ size uint32 }

func (ch *sizedChunk) ID() uint32                        { return IDUNKN }
func (ch *sizedChunk) Size() uint32                      { return ch.size }
func (ch *sizedChunk) Type() uint32                      { return 0 }
func (ch *sizedChunk) Multi() bool                       { return true }
func (ch *sizedChunk) Chunks() Chunks                    { return nil }
func (ch *sizedChunk) Raw() bool                         { return false }
func (ch *sizedChunk) ReadFrom(io.Reader) (int64, error) { return 0, nil }
func (ch *sizedChunk) WriteTo(io.Writer) (int64, error)  { return 0, nil }
func (ch *sizedChunk) Reset()                            {}

// nonSeeker hides the [io.Seeker] implementation of the wrapped reader.
type nonSeeker struct{ io.Reader }
