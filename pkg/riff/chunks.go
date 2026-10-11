// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"io"
)

// Chunks represents a slice of chunks.
type Chunks []Chunk

// First returns the first chunk with the given ID, or nil if there is none.
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

// IDs returns the chunk IDs in order.
func (chs Chunks) IDs() []uint32 {
	ids := make([]uint32, len(chs))
	for i, ch := range chs {
		ids[i] = ch.ID()
	}
	return ids
}

// Size returns the encoded size of all the chunks, counting their ID and size
// fields and padding bytes.
func (chs Chunks) Size() uint32 {
	var size uint32
	for _, ch := range chs {
		size += RealSize(ch.Size()) + 8 // Add 8 for chunk ID and size fields.
	}
	return size
}

// size64 returns the encoded size like [Chunks.Size], computed in 64 bits so
// it does not wrap. LIST sub-chunks are summed the same way.
func (chs Chunks) size64() uint64 {
	var size uint64
	for _, ch := range chs {
		cs := uint64(ch.Size())
		if lst, ok := ch.(*ChunkLIST); ok {
			cs = uint64(ListTypeSize) + lst.chunks.size64()
		}
		size += cs + cs%2 + 8 // Add padding byte, chunk ID and size fields.
	}
	return size
}

// WriteTo writes all the chunks in the collection to w.
func (chs Chunks) WriteTo(w io.Writer) (int64, error) {
	var sum int64
	for _, ch := range chs {
		n, err := ch.WriteTo(w)
		sum += n
		if err != nil {
			return sum, err
		}
	}
	return sum, nil
}

// Remove returns the chunks without the chunks with the given id, keeping the
// order. It returns the receiver when no chunk has the id and never modifies
// it.
func (chs Chunks) Remove(id uint32) Chunks {
	if chs.Count(id) == 0 {
		return chs
	}
	out := make(Chunks, 0, len(chs))
	for _, ch := range chs {
		if ch.ID() != id {
			out = append(out, ch)
		}
	}
	return out
}
