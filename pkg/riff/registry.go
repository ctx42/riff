// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

// subKey identifies decoders a LIST type maker made: by list type, sub-chunk
// ID, and the options of the list that made them.
type subKey struct {
	typ  uint32
	id   uint32
	opts Opts
}

// Registry represents a registry of chunk decoders.
type Registry struct {
	// Map of chunk IDs and decoder [Maker] functions.
	makers map[uint32]Maker

	// Pool of instantiated chunk decoders that can be reused.
	pool map[uint32][]Chunk

	// Pool of decoders the raw maker made for top-level chunks.
	rawPool map[uint32][]Chunk

	// Pool of decoders LIST type makers made for sub-chunks.
	subPool map[subKey][]Chunk

	// Maker for raw chunk decoder.
	raw IDMaker
}

// NewRegistry returns a new instance of [Registry].
func NewRegistry(raw IDMaker) *Registry {
	return &Registry{
		makers:  make(map[uint32]Maker, 4),
		pool:    make(map[uint32][]Chunk, 4),
		rawPool: make(map[uint32][]Chunk, 4),
		subPool: make(map[subKey][]Chunk, 4),
		raw:     raw,
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

// Put resets the chunk decoder ch and returns it to the pool for reuse. A
// decoder for an unregistered ID is dropped: it was made by a raw or a LIST
// type specific maker, so reusing it for the same ID elsewhere would decode
// the chunk with the wrong decoder.
func (reg *Registry) Put(ch Chunk) { reg.tryPut(ch) }

// tryPut returns ch to the pool when its ID is registered and reports whether
// it did.
func (reg *Registry) tryPut(ch Chunk) bool {
	id := ch.ID()
	if _, ok := reg.makers[id]; !ok {
		return false
	}
	ch.Reset()
	reg.pool[id] = append(reg.pool[id], ch)
	return true
}

// release returns a top-level decoder: one for a registered ID to the pool, a
// raw one to the raw pool, from which only [Registry.Get] takes.
func (reg *Registry) release(ch Chunk) {
	id := ch.ID()
	ch.Reset()
	if _, ok := reg.makers[id]; ok {
		reg.pool[id] = append(reg.pool[id], ch)
		return
	}
	reg.rawPool[id] = append(reg.rawPool[id], ch)
}

// Get returns a decoder for id from the pool, or a new one from its maker.
// For an unregistered id it returns a raw decoder.
func (reg *Registry) Get(id uint32) Chunk {
	if ch := reg.GetNoRaw(id); ch != nil {
		return ch
	}
	if chs := reg.rawPool[id]; len(chs) > 0 {
		reg.rawPool[id] = chs[:len(chs)-1]
		return chs[len(chs)-1]
	}
	return reg.raw(id)
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

// putSub returns ch, a decoder made by the type maker of a list of type typ
// with options opts, to the pool of sub-chunk decoders.
func (reg *Registry) putSub(typ uint32, opts Opts, ch Chunk) {
	ch.Reset()
	key := subKey{typ: typ, id: ch.ID(), opts: opts}
	reg.subPool[key] = append(reg.subPool[key], ch)
}

// getSub returns a pooled sub-chunk decoder for id made for a list of type typ
// with options opts, or nil when there is none.
func (reg *Registry) getSub(typ, id uint32, opts Opts) Chunk {
	key := subKey{typ: typ, id: id, opts: opts}
	chs := reg.subPool[key]
	if len(chs) == 0 {
		return nil
	}
	reg.subPool[key] = chs[:len(chs)-1]
	return chs[len(chs)-1]
}
