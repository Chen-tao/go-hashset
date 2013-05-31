// Package hashset implements a "set" type store for hashes.
package hashset

import (
	"bytes"
	"encoding/binary"
	"sort"
)

type hashSorter struct {
	buf, hashes []byte
	size        int
}

func (p *hashSorter) Len() int {
	return len(p.hashes) / p.size
}

func (p *hashSorter) Less(i, j int) bool {
	ioff := i * p.size
	joff := j * p.size
	return bytes.Compare(p.hashes[ioff:ioff+p.size], p.hashes[joff:joff+p.size]) < 0
}

func (p *hashSorter) Swap(i, j int) {
	ioff := i * p.size
	joff := j * p.size
	copy(p.buf, p.hashes[ioff:ioff+p.size])
	copy(p.hashes[ioff:ioff+p.size], p.hashes[joff:joff+p.size])
	copy(p.hashes[joff:joff+p.size], p.buf)
}

// Store the hashes.
type Hashset struct {
	things  [65536][]byte
	sortbuf []byte
	size    int
}

// Add a hash to the Hashset.Add
//
// This is the []byte representation of a hash.  You *can* hex encode
// it, but you probably shouldn't.
func (hs *Hashset) Add(h []byte) {
	if hs.Contains(h) {
		return
	}

	if hs.size == 0 {
		hs.size = len(h)
		hs.sortbuf = make([]byte, hs.size)
	} else if hs.size != len(h) {
		panic("inconsistent size")
	}

	n := int(binary.BigEndian.Uint16(h))

	hs.things[n] = append(hs.things[n], h[2:]...)
	sorter := hashSorter{hs.sortbuf, hs.things[n], hs.size - 2}
	sort.Sort(&sorter)
}

// Return true if the given hash is in this Hashset.
func (hs *Hashset) Contains(h []byte) bool {
	n := int(binary.BigEndian.Uint16(h))
	bin := hs.things[n]
	if len(bin) == 0 {
		return false
	}
	sub := h[2:]
	pos := sort.Search(len(bin)/(hs.size-2), func(i int) bool {
		off := i * (hs.size - 2)
		rv := bytes.Compare(bin[off:off+hs.size-2], sub) >= 0
		return rv
	})
	off := pos * (hs.size - 2)
	return off < len(bin) && bytes.Equal(sub, bin[off:off+hs.size-2])
}

// How many things we've got.
func (hs *Hashset) Len() int {
	rv := 0
	for _, a := range hs.things {
		rv += (len(a) / (hs.size - 2))
	}
	return rv
}
