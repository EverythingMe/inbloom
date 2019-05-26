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
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"math"
	"unsafe"

	"github.com/EverythingMe/inbloom/go/internal/gomurmur"
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
	return newFilterFromData(nil, entries, errorRate)
}

// NewFilterFromData creates a bloom filter from an existing data buffer, created by another instance of this library (probably in another language).
//
// If the length of the data does not fit the number of entries and error rate, we return an error. If data is nil we allocate a new filter
func newFilterFromData(data []byte, entries int, errorRate float64) (*BloomFilter, error) {

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

// checksum returns a 16 bit checksum of the data (using xor folded crc32 checksum)
func (f *BloomFilter) checksum() uint16 {

	checksum32 := crc32.ChecksumIEEE(f.bf)
	return uint16(checksum32&0xFFFF) ^ uint16(checksum32>>16)

}

// The structure of a marshaled binary filter is:
//    checksum uint16
//    error_rate uint16
//    cardinality uint32
//    data []byte

// Marshal dumps the filter to a byte array, with a header containing the error rate, cardinality and a checksum.
// This data can be passed to another inbloom filter over the network, and thus the other end can open the data
// without the user having to pass the filter size explicitly. See Unmarshal for reading these dumpss
func (f *BloomFilter) Marshal() []byte {

	buf := bytes.NewBuffer(make([]byte, 0, len(f.bf)+int(unsafe.Sizeof(uint16(0))*2)+int(unsafe.Sizeof(uint32(0)))))
	binary.Write(buf, binary.BigEndian, f.checksum())

	errs := uint16(1 / f.errorRate)
	binary.Write(buf, binary.BigEndian, errs)
	binary.Write(buf, binary.BigEndian, uint32(f.entries))
	buf.Write(f.bf)
	return buf.Bytes()
}

// MarshalBase64 is a convenience method that dumps the filter's data to a base64 encoded string.
// By default uses URLEncoding which ready to be passed as a GET/POST parameter.
// Pass an encoding param to use different encoding.
func (f *BloomFilter) MarshalBase64(encoding ...*base64.Encoding) string {
	if len(encoding) > 1 {
		panic(fmt.Sprintf("Expected at most 1 encoding, got %d", len(encoding)))
	} else if len(encoding) == 1 {
		return encoding[0].EncodeToString(f.Marshal())
	} else {
		return base64.URLEncoding.EncodeToString(f.Marshal())
	}
}

// UnmarshalBase64 is a convenience function that unmarshals a filter that has been encoded into base64.
// Uses URLEncoding by default, pass an encoding param to use different encoding.
func UnmarshalBase64(b64 string, encoding ...*base64.Encoding) (*BloomFilter, error) {
	selectedEncoding := base64.URLEncoding
	if len(encoding) > 1 {
		panic(fmt.Sprintf("Expected at most 1 encoding, got %d", len(encoding)))
	} else if len(encoding) == 1 {
		selectedEncoding = encoding[0]
	}
	if b, err := selectedEncoding.DecodeString(b64); err != nil {
		return nil, fmt.Errorf("bloom: could not decode base64 data: %s", err)
	} else {
		return Unmarshal(b)
	}

}

// Unmarshal reads a binary dump of an inbloom filter with its header, and returns the resulting filter.
// Since this is a dump containing size and precisin metadata, you do not need to specify them.
//
// If the data is corrupt or the buffer is not complete, we return an error
func Unmarshal(data []byte) (*BloomFilter, error) {

	if data == nil || len(data) <= int(unsafe.Sizeof(uint16(0))*2)+int(unsafe.Sizeof(uint32(0))) {
		return nil, errors.New("Invalid buffer size")
	}
	buf := bytes.NewBuffer(data)
	var checksum, errRate uint16
	var entries uint32

	if err := binary.Read(buf, binary.BigEndian, &checksum); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &errRate); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &entries); err != nil {
		return nil, err
	}

	if errRate == 0 {
		return nil, errors.New("Error rate cannot be 0")
	}

	// Read the data
	bf := make([]byte, len(data))
	if n, err := buf.Read(bf); err != nil {
		return nil, err
	} else {
		bf = bf[:n]
	}

	// Create a new filter from the data we read
	ret, err := newFilterFromData(bf, int(entries), 1/float64(errRate))
	if err != nil {
		return nil, err
	}

	// Verify checksum
	if ret.checksum() != checksum {
		return nil, errors.New("Bad checksum")
	}

	return ret, nil
}
