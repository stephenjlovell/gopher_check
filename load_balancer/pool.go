//-----------------------------------------------------------------------------------
// Copyright (c) 2014 Stephen J. Lovell
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

type Pool []*Worker // a heap implemented as a priority queue of pointers to worker objects.

// satisfy the container#heap interface...

func (p Pool) Len() int { return len(p) }

func (p Pool) Less(i, j int) bool { return p[i].pending < p[j].pending }

func (p Pool) Swap(i, j int) {
  p[i], p[j] = p[j], p[i]
  p[i].index, p[j].index = j, i
}

func (p *Pool) Push(w interface{}) {
  n := len(*p)
  item := w.(*Worker)
  item.index = n
  *p = append(*p, item)
}

func (p *Pool) Pop() interface{} {
  old := *p
  n := len(old)
  item := old[n-1]
  *p = old[0 : n-1]
  return item
}