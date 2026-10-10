// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

// Registry represents a registry of chunk decoders.
type Registry struct {
	// Map of chunk IDs and decoder [Maker] functions.
	makers map[uint32]Maker

	// Pool of instantiated chunk decoders that can be reused.
	pool map[uint32][]Chunk

	// Maker for raw chunk decoder.
	raw IDMaker
}

// NewRegistry returns a new instance of [Registry].
func NewRegistry(raw IDMaker) *Registry {
	return &Registry{
		makers: make(map[uint32]Maker, 4),
		pool:   make(map[uint32][]Chunk, 4),
		raw:    raw,
	}
}

// Has reports whether a decoder for id is registered.
func (reg *Registry) Has(id uint32) bool {
	_, ok := reg.makers[id]
	return ok
}

// Register registers the decoder maker for id, replacing any previous one.
func (reg *Registry) Register(id uint32, maker Maker) {
	reg.makers[id] = maker
}

// Put returns the chunk decoder ch to the pool for reuse.
func (reg *Registry) Put(ch Chunk) {
	id := ch.ID()
	if _, ok := reg.pool[id]; !ok {
		reg.pool[id] = make([]Chunk, 0, 4)
	}
	reg.pool[id] = append(reg.pool[id], ch)
}

// Get returns a decoder for id from the pool, or a new one from its maker.
// For an unregistered id with no pooled decoder it returns a raw decoder.
func (reg *Registry) Get(id uint32) Chunk {
	ch := reg.GetNoRaw(id)
	if ch == nil {
		ch = reg.raw(id)
	}
	return ch
}

// GetNoRaw works like [Registry.Get] but returns nil instead of a raw
// decoder.
func (reg *Registry) GetNoRaw(id uint32) Chunk {
	chs, ok := reg.pool[id]
	cl := len(chs)
	if !ok || cl == 0 {
		if dec, ok := reg.makers[id]; ok {
			return dec()
		}
		return nil
	}
	ch := reg.pool[id][cl-1]
	reg.pool[id] = reg.pool[id][:cl-1]
	return ch
}
