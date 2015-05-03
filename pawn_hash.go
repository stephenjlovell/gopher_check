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

const (
  PAWN_ENTRY_COUNT = 16384
  // PAWN_ENTRY_COUNT = 32768
  PAWN_TT_MASK = PAWN_ENTRY_COUNT - 1
)

var main_pawn_tt PawnTT

type PawnTT [PAWN_ENTRY_COUNT]*PawnEntry

type PawnEntry struct {
  passed_pawns BB
  value int
  key uint32
}

func setup_pawn_tt() {
  for i, _ := range main_pawn_tt {
    main_pawn_tt[i] = &PawnEntry{
      value: NO_SCORE,
    }
  } 
}

// Typical hit rate is around 97 %
func (ptt *PawnTT) Probe(key uint32) *PawnEntry {
  return ptt[key & PAWN_TT_MASK]
}

func (entry *PawnEntry) Store(key uint32, value int, passed_pawns BB) {
  entry.passed_pawns = passed_pawns
  entry.value = value
  entry.key = key
}





















