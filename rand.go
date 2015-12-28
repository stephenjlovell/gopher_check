//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//-----------------------------------------------------------------------------------

package main

import (
	"math/rand"
)

const (
	MAX_RAND = (1 << 32) - 1
)

func random_key64() uint64 { // create a pseudorandom 64-bit unsigned int key
	return (uint64(rand.Int63n(MAX_RAND)) << 32) | uint64(rand.Int63n(MAX_RAND))
}

func random_key32() uint32 {
	return uint32(rand.Int63n(MAX_RAND))
}

func setup_rand() {
	rand.Seed(4129246945) // keep the same seed each time for debugging purposes.
}

// RngKiss uses Bob Jenkins' pseudorandom number approach, which is well-suited for generating
// magic number candidates:  https://chessprogramming.wikispaces.com/Bob+Jenkins
type RngKiss struct {
	a        BB
	b        BB
	c        BB
	d        BB
	boosters [8]BB
}

func NewRngKiss(seed int) *RngKiss {
	r := &RngKiss{}
	r.Setup(seed)
	return r
}

func (r *RngKiss) Setup(seed int) {
	r.boosters = [8]BB{3101, 552, 3555, 926, 834, 26, 2131, 1117}
	r.a = 0xF1EA5EED
	r.b, r.c, r.d = 0xD4E12C77, 0xD4E12C77, 0xD4E12C77
	for i := 0; i < seed; i++ {
		_ = r.rand()
	}
}

func (r *RngKiss) RandomMagic(sq int) BB {
	return r.RandomBB(r.boosters[row(sq)])
}

func (r *RngKiss) RandomUint64(sq int) uint64 {
	return uint64(r.RandomBB(r.boosters[row(sq)]))
}

func (r *RngKiss) RandomUint32(sq int) uint32 {
	return uint32(r.RandomBB(r.boosters[row(sq)]))
}

func (r *RngKiss) RandomBB(booster BB) BB {
	return r.rotate((r.rotate(r.rand(), booster&63) & r.rand()), ((booster >> 6) & 63 & r.rand()))
}


func (r *RngKiss) rand() BB {
	e := r.a - r.rotate(r.b, 7)
	r.a = r.b ^ r.rotate(r.c, 13)
	r.b = r.c + r.rotate(r.d, 37)
	r.c = r.d + e
	r.d = e + r.a
	return r.d
}

func (r *RngKiss) rotate(x, k BB) BB {
	return (x << k) | (x >> (64 - k))
}
