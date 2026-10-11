// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"io"
)

// Chunk represents a Resource Interchange File Format (RIFF) chunk decoder.
type Chunk interface {
	// ID returns a four-byte (uint32) ASCII identifier of the chunk.
	ID() uint32

	// Size returns the chunk size in bytes. The chunk ID, size field, and
	// padding byte (if present) are not counted in the returned value.
	// If Chunk has been edited, this method should return the new size.
	Size() uint32

	// Type returns the chunk type, or zero if the chunk doesn't have the
	// type field.
	Type() uint32

	// Multi returns true if there can be more than one chunk with a given ID
	// in the RIFF file.
	Multi() bool

	// Chunks returns the sub-chunks of the chunk, or nil if the chunk doesn't
	// support sub-chunks.
	Chunks() Chunks

	// Raw returns true if the chunk was decoded by the raw chunk decoder
	// ([ChunkRAWC]).
	Raw() bool

	// ReadFrom reads bytes from r decoding and validating a chunk. The decoded
	// chunk replaces any content of a previous ReadFrom.
	// It expects r to be in a position right after the chunk ID.
	// It returns the actual number of bytes read and an error, if any.
	// If an error is returned, the number of read bytes might not be accurate.
	ReadFrom(r io.Reader) (int64, error)

	// WriteTo encodes the chunk to w. It returns the number of bytes written
	// and an error, if any.
	WriteTo(w io.Writer) (int64, error)

	// Reset resets the chunk so it can be reused.
	Reset()
}
