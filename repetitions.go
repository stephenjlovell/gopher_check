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
// "fmt"
)

type RepList struct {
	key    uint32
	parent *RepList
}

func (l *RepList) Scan(big_key uint64) bool {
	p := l.parent
	key := uint32(big_key)
	repetition_count := 0
	for p != nil {
		if p.key == key {
			repetition_count += 1
			if repetition_count == 2 {
				// fmt.Println("Repetition found.")
				return true
			}
		}
		if p.parent == nil {
			break
		}
		p = p.parent.parent
	}
	return false
}

func (l *RepList) Len() int {
	if l == nil {
		return 0
	}
	sum := 0
	p := l.parent
	for p != nil {
		sum += 1
		p = p.parent
	}
	return sum
}