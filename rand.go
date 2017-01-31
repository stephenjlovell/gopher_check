//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"math/rand"
)

func setupRand() {
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
