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
  "fmt"
)

var main_ktable KTable

type KTable [64][2]Move

func (k *KTable) Store(m Move, ply int) {
	if m != k[ply][0] {
		k[ply][1] = k[ply][0]
		k[ply][0] = m
	}
}

func (k *KTable) Clear() {
	for ply := 0; ply < 64; ply++ {
		for m := 0; m < 2; m++ {
			k[ply][m] = 0
		}
	}
}

func (k *KTable) Print() {
  for ply := 0; ply < 64; ply++ {
    fmt.Printf("Ply %d, K1: %s, K2: %s\n", ply, k[ply][0].ToString(), k[ply][1].ToString())
  }
}



