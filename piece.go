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

const ( // type
	PAWN = iota
	KNIGHT
	BISHOP
	ROOK
	QUEEN
	KING
	EMPTY // no piece located at this square
)

const (
	PAWN_VALUE   = 100 // piece values are given in centipawns
	KNIGHT_VALUE = 320
	BISHOP_VALUE = 333
	ROOK_VALUE   = 510
	QUEEN_VALUE  = 880
	KING_VALUE   = 5000
)

type Piece uint8

var piece_values = [8]int{PAWN_VALUE, KNIGHT_VALUE, BISHOP_VALUE, ROOK_VALUE, QUEEN_VALUE, KING_VALUE} // default piece values

var promote_values = [8]int{0, KNIGHT_VALUE - PAWN_VALUE, BISHOP_VALUE - PAWN_VALUE, ROOK_VALUE - PAWN_VALUE,
	QUEEN_VALUE - PAWN_VALUE}

func (pc Piece) Value() int { return piece_values[pc] }

func (pc Piece) PromoteValue() int { return promote_values[pc] }
