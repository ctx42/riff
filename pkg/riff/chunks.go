// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"io"
)

// Chunks represent slice of sub-chunks.
type Chunks []Chunk

// First returns the first chunk with the given ID.
func (chs Chunks) First(id uint32) Chunk {
	for _, ch := range chs {
		if ch.ID() == id {
			return ch
		}
	}
	return nil
}

// Count returns the number of chunks with the given ID.
func (chs Chunks) Count(id uint32) int {
	var cnt int
	for _, ch := range chs {
		if ch.ID() == id {
			cnt++
		}
	}
	return cnt
}

// IDs return chunk IDs in order they were seen in the file.
func (chs Chunks) IDs() []uint32 {
	ids := make([]uint32, len(chs))
	for i, ch := range chs {
		ids[i] = ch.ID()
	}
	return ids
}

// Size returns size (with padding bytes) of all the chunks in the collection.
func (chs Chunks) Size() uint32 {
	var size uint32
	for _, ch := range chs {
		size += RealSize(ch.Size()) + 8 // Add 8 for chunk ID and size fields.
	}
	return size
}

// WriteTo writes all the chunks in the collection to w.
func (chs Chunks) WriteTo(w io.Writer) (n int64, err error) {
	var sum int64
	for _, ch := range chs {
		n, err = ch.WriteTo(w)
		sum += n
		if err != nil {
			return sum, err
		}
	}
	return sum, nil
}

// Remove returns chunks without given id.
// Keeps the previous order and leaves the receiver unchanged.
func (chs Chunks) Remove(id uint32) Chunks {
	for i, ch := range chs {
		if ch.ID() == id {
			out := make(Chunks, 0, len(chs)-1)
			out = append(out, chs[:i]...)
			return append(out, chs[i+1:]...)
		}
	}
	return chs
}
