/*
* Copyright (c) 2013-2016, Sureshkumar Nedunchezhian
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions are met:
*
*   * Redistributions of source code must retain the above copyright notice,
*     this list of conditions and the following disclaimer.
*   * Redistributions in binary form must reproduce the above copyright
*     notice, this list of conditions and the following disclaimer in
*     the documentation and/or other materials provided with the distribution.
*
* THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
* AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO,
* THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
* PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS
* BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY,
* OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
* SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
* INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
* CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
* ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF
* THE POSSIBILITY OF SUCH DAMAGE.
 */

/*
 * "Murmur" hash provided by Austin, tanjent@gmail.com
 * http://murmurhash.googlepages.com/
 *
 * Note - This code makes a few assumptions about how your machine behaves -
 *
 * 1. We can read a 4-byte value from any address without crashing
 * 2. sizeof(int) == 4
 *
 * And it has a few limitations -
 * 1. It will not work incrementally.
 * 2. It will not produce the same results on little-endian and big-endian
 *  machines. */

package gomurmur

import (
	"bytes"
	"encoding/binary"
	"hash"
)

type (
	sum32 uint32
)

const (
	m = 0x5bd1e995
	r = 24
)

func Sum32(b []byte, seed uint32) (uint32, error) {
	var s sum32 = 0
	h := &s

	if _, err := h.WriteSeed(b, seed); err != nil {
		return 0, err
	}
	return uint32(*h), nil
}

// New32 returns a new 32-bit FNV-1 hash.Hash.
func New32() hash.Hash32 {
	var s sum32 = 0
	return &s
}

func (s *sum32) Reset() { *s = 0 }
func (s *sum32) Sum32() uint32 {
	return uint32(*s)
}

const defaultSeed uint32 = 0x9747b28c

func (s *sum32) Write(data []byte) (int, error) {
	return s.WriteSeed(data, defaultSeed)
}

func (s *sum32) WriteSeed(data []byte, seed uint32) (int, error) {
	var length = uint32(len(data))

	/* Initialize the hash to a 'random' value */
	h := *s
	h = sum32(seed ^ length)

	/* Mix 4 bytes at a time into the hash */
	var i int = 0

	for length >= 4 {
		var k uint32
		buf := bytes.NewBuffer(data[i : i+4])
		err := binary.Read(buf, binary.LittleEndian, &k)
		if err != nil {
			return 0, err
		}
		k *= m
		k ^= k >> r
		k *= m

		h *= m
		h ^= sum32(k)

		i += 4
		length -= 4
	}
	switch length {
	case 3:
		h ^= sum32((uint32)(data[i+2]) << 16)
		fallthrough
	case 2:
		h ^= sum32((uint32)(data[i+1]) << 8)
		fallthrough
	case 1:
		h ^= sum32((uint32)(data[i]))
		h *= m
	default:
	}
	h ^= h >> 13
	h *= m
	h ^= h >> 15
	*s = h

	return len(data), nil
}

func (s *sum32) Size() int { return 4 }

func (s *sum32) BlockSize() int { return 1 }

func (s *sum32) Sum(in []byte) []byte {
	v := uint32(*s)
	return append(in, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}
