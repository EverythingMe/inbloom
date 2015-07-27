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

package gomurmur

import (
	"bytes"
	"hash"
	"testing"
)

type golden struct {
	sum  []byte
	text string
}

var golden32 = []golden{
	{[]byte{0x00, 0x00, 0x00, 0x00}, ""},
	{[]byte{0x4b, 0x41, 0x75, 0x7c}, "a"},
	{[]byte{0xe3, 0xb5, 0x4d, 0xfb}, "ab"},
	{[]byte{0x7b, 0x0c, 0xc4, 0x28}, "abc"},
	{[]byte{0xef, 0x6a, 0x86, 0xaf}, "abcd"},
	{[]byte{0x9a, 0x26, 0x3e, 0xda}, "abcde"},
	{[]byte{0xe0, 0xba, 0xdc, 0x96}, "abcdef"},
	{[]byte{0xeb, 0xa7, 0x46, 0xf2}, "abcdefg"},
}

func TestGolden32(t *testing.T) {
	testGolden(t, New32(), golden32)
}

func testGolden(t *testing.T, hash hash.Hash, gold []golden) {
	for _, g := range gold {
		hash.Reset()
		done, error := hash.Write([]byte(g.text))
		if error != nil {
			t.Fatalf("write error: %s", error)
		}
		if done != len(g.text) {
			t.Fatalf("wrote only %d out of %d bytes", done, len(g.text))
		}
		if actual := hash.Sum(nil); !bytes.Equal(g.sum, actual) {
			t.Errorf("hash(%q) = 0x%x want 0x%x", g.text, actual, g.sum)
		}
	}
}
