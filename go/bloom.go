// This code was adapted to Go from the libbloom C library - https://github.com/jvirkki/libbloom
//
// Original copyright note from libbloom:
//
//     Copyright (c) 2012, Jyri J. Virkki
//     All rights reserved.
//     This file is under BSD license. See LICENSE file.

// Package inbloom implements a portable bloom filter that can export and import data to and from
// implementations of the same library in different languages.
package inbloom

import (
	"errors"
	"fmt"
	"math"

	"github.com/EverythingMe/inbloom-go/gomurmur"
)

const denom = 0.480453013918201

// BloomFilter is our implementation of a simple dynamically sized bloom filter.
//
// This code was adapted to Go from the libbloom C library - https://github.com/jvirkki/libbloom
type BloomFilter struct {

	// These fields are part of the public interface of this structure.
	// Client code may read these values if desired. Client code MUST NOT
	// modify any of these.

	entries   int
	errorRate float64
	bits      int
	bytes     int
	hashes    int

	// Fields below are private to the implementation. These may go away or
	// change incompatibly at any moment. Client code MUST NOT access or rely
	// on these.

	bpe float64
	bf  []byte
}

// NewFilter creates an empty bloom filter, with the given expected number of entries, and desired error rate.
// The number of hash functions and size of the filter are calculated from these 2 parameters
func NewFilter(entries int, errorRate float64) (*BloomFilter, error) {
	return NewFilterFromData(nil, entries, errorRate)
}

// NewFilterFromData creates a bloom filter from an existing data buffer, created by another instance of this library (probably in another language).
//
// If the length of the data does not fit the number of entries and error rate, we return an error. If data is nil we allocate a new filter
func NewFilterFromData(data []byte, entries int, errorRate float64) (*BloomFilter, error) {

	if entries < 1 || errorRate == 0 {
		return nil, errors.New("Invalid params for bloom filter")
	}

	bpe := -(math.Log(errorRate) / denom)
	bits := int(float64(entries) * bpe)

	flt := &BloomFilter{
		entries:   entries,
		errorRate: errorRate,
		bpe:       bpe,
		bits:      bits,
		bytes:     (bits / 8),
		hashes:    int(math.Ceil(0.693147180559945 * bpe)), // ln(2)
	}

	if flt.bits%8 != 0 {
		flt.bytes++
	}

	if data != nil {
		if flt.bytes != len(data) {
			return nil, fmt.Errorf("Expected %d bytes, got %d", flt.bytes, len(data))
		}
		flt.bf = data
	} else {
		flt.bf = make([]byte, flt.bytes)
	}
	return flt, nil
}

// checkAdd checks existence or adds a key to the filter
func (f *BloomFilter) checkAdd(key []byte, add bool) bool {

	hits := 0
	a, _ := gomurmur.Sum32(key, 0x9747b28c)
	b, _ := gomurmur.Sum32(key, a)

	for i := 0; i < f.hashes; i++ {
		x := (a + uint32(i)*b) % uint32(f.bits)
		bt := x >> 3

		c := f.bf[bt] // expensive memory access
		mask := byte(1) << (x % 8)

		if (c & mask) != 0 {
			hits++
		} else {
			if add {
				f.bf[bt] = byte(c | mask)
			}
		}

	}

	return hits == f.hashes
}

// Contains returns true if a key exists in the filter
func (f *BloomFilter) Contains(key string) bool {
	return f.checkAdd([]byte(key), false)
}

// Add adds a key to the filter
func (f *BloomFilter) Add(key string) bool {
	return f.checkAdd([]byte(key), true)
}

// Len returns the number of BYTES in the filter
func (f *BloomFilter) Len() int {
	return f.bytes
}

// TODO: add serialization
//func (f *BloomFilter) Marshal() []byte {
//	return f.bf
//}
