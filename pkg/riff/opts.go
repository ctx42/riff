// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
)

// Compile time checks.
var _ io.Reader = (*ChunkReader)(nil)

// OptsFn is a function setting [Opts] options.
type OptsFn func(*Opts)

// WithSizeLimit sets the maximum chunk size in bytes.
func WithSizeLimit(size uint32) OptsFn {
	return func(opt *Opts) { opt.sizeLimit = size }
}

// WithLoadData makes decoders load chunk content, not only metadata.
func WithLoadData() OptsFn {
	return func(opt *Opts) { opt.loadData = true }
}

// WithIgnoreSize makes decoders ignore declared chunk sizes and read past them.
// The RIFF size is then corrected to the number of bytes read.
func WithIgnoreSize() OptsFn {
	return func(opt *Opts) { opt.ignoreDeclaredSize = true }
}

// WithOpts copies all values from opts.
func WithOpts(opts Opts) OptsFn {
	return func(opt *Opts) {
		opt.sizeLimit = opts.sizeLimit
		opt.ignoreDeclaredSize = opts.ignoreDeclaredSize
		opt.loadData = opts.loadData
	}
}

// Opts represents options for chunks handling.
type Opts struct {
	// Maximum chunk size in bytes. Zero means unlimited.
	sizeLimit uint32

	// When true, allows reading more data than the chunk declares.
	// By default, it is set to false.
	ignoreDeclaredSize bool

	// Controls how chunks are processed. If set to false, only chunk
	// metadata is read and the rest is skipped. This improves
	// performance in cases when a user is only interested in metadata and
	// isn't going to modify or write the RIFF file.
	// It's up to the chunk decoder to decide what is considered data vs.
	// metadata.
	// By default, it is set to false.
	loadData bool
}

// LoadData reports whether chunk content is loaded.
func (opt *Opts) LoadData() bool {
	return opt.loadData
}

// IgnoreDeclaredSize reports whether declared chunk sizes are ignored.
func (opt *Opts) IgnoreDeclaredSize() bool {
	return opt.ignoreDeclaredSize
}

// NewOpts returns [Opts] with fns applied.
func NewOpts(fns ...OptsFn) Opts {
	opt := Opts{}
	for _, fn := range fns {
		fn(&opt)
	}
	return opt
}

// ChunkReader reads the four-byte chunk size from r, checks it against the
// options, and returns the reader for the chunk content along with the size.
// It returns [ErrTooLarge] when the size exceeds the limit or cannot be padded.
func (opt *Opts) ChunkReader(r io.Reader) (*ChunkReader, uint32, error) {
	var size uint32
	if err := binary.Read(r, le, &size); err != nil {
		return nil, 0, err
	}

	// The padded size of an odd chunk at the uint32 maximum overflows.
	if size == math.MaxUint32 {
		return nil, 0, ErrTooLarge
	}
	if opt.sizeLimit != 0 && opt.sizeLimit < RealSize(size) {
		return nil, 0, ErrTooLarge
	}

	cr := NewChunkReader(r)
	cr.setOffset(4) // Already read chunk size.

	if !opt.IgnoreDeclaredSize() {
		cr.setLimit(int64(RealSize(size)))
	}
	if opt.IgnoreDeclaredSize() && opt.sizeLimit != 0 {
		cr.setLimit(int64(opt.sizeLimit))
	}

	return cr, size, nil
}

// ChunkReader wraps an [io.Reader], counts the bytes read, and limits how many
// can be read.
type ChunkReader struct {
	src    io.Reader // Original reader (may implement io.Seeker).
	rdr    io.Reader // Main reader (may have limit).
	cnt    int64     // Number of read bytes from reader.
	offset int64     // Shift added to the value Count returns.
}

// NewChunkReader returns a [ChunkReader] reading from r.
func NewChunkReader(r io.Reader) *ChunkReader {
	return &ChunkReader{src: r, rdr: r}
}

// setOffset sets the shift added to the value Count returns.
func (cr *ChunkReader) setOffset(n int) {
	cr.offset = int64(n)
}

// setLimit limits the number of bytes that can be read to n.
func (cr *ChunkReader) setLimit(n int64) {
	cr.rdr = io.LimitReader(cr.src, n)
}

func (cr *ChunkReader) Read(p []byte) (n int, err error) {
	n, err = cr.rdr.Read(p)
	cr.cnt += int64(n)
	return n, err
}

// SkipN skips n bytes. When the source implements [io.Seeker], SkipN seeks and
// reads the last skipped byte to detect truncated input; otherwise it reads
// and discards n bytes. It returns [io.ErrUnexpectedEOF] when fewer than n
// bytes are left.
func (cr *ChunkReader) SkipN(n uint32) error {
	num := int64(n)

	if !cr.canRead(num) {
		return io.ErrUnexpectedEOF
	}

	// Handle nested chunk readers.
	if pcr, ok := cr.src.(*ChunkReader); ok {
		err := pcr.SkipN(n)
		if err != nil {
			return err
		}
		cr.moveCursors(num)
		return nil
	}

	// When src implements Seeker, we can just skip n bytes. Seeking past the
	// end succeeds, so read the last skipped byte to prove it exists.
	if skr, ok := cr.src.(io.Seeker); ok {
		if num == 0 {
			return nil
		}
		if _, err := skr.Seek(num-1, io.SeekCurrent); err != nil {
			return err
		}
		if _, err := io.ReadFull(cr.src, make([]byte, 1)); err != nil {
			if errors.Is(err, io.EOF) {
				err = io.ErrUnexpectedEOF
			}
			return err
		}
		cr.moveCursors(num)
		return nil
	}

	// If we cannot seek, we read data to black hole.
	if _, err := io.CopyN(io.Discard, cr, num); err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
		return err
	}
	return nil
}

// canRead reports whether n bytes can be read without passing the limit.
func (cr *ChunkReader) canRead(n int64) bool {
	lr, ok := cr.rdr.(*io.LimitedReader)
	if !ok {
		return true
	}
	return lr.N >= n
}

// moveCursors pushes cursors forward n bytes. Shift on reader must be done
// separately.
func (cr *ChunkReader) moveCursors(n int64) {
	// Add skipped bytes to counter.
	cr.cnt += n

	lr, ok := cr.rdr.(*io.LimitedReader)
	if ok {
		// Subtract skipped bytes from reader limit.
		lr.N -= n
	}
}

// Count returns the number of bytes read plus the offset.
func (cr *ChunkReader) Count() int64 {
	return cr.cnt + cr.offset
}
