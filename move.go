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

import (
// "fmt"
)

const (
	QUIET = (Move(EMPTY) << 15) | (Move(EMPTY) << 18)
)

type Move uint32

// To to fit into transposition table entries, moves are encoded using 21 bits as follows (in LSB order):
// From square - first 6 bits
// To square - next 6 bits
// Piece - next 3 bits
// Captured piece - next 3 bits
// promoted to - next 3 bits

func is_valid_move(brd *Board, m Move, d int) bool {
	if m == 0 {
		// fmt.Printf("Warning: Invalid move %s at depth %d for key %#x\n", m.ToString(), d, brd.hash_key)
		return false
	}
	// determine if there really is a piece of this type on the from square.
	piece := m.Piece()
	from := m.From()
	if brd.TypeAt(from) != piece {
		// fmt.Printf("Warning: Invalid move  %s at depth %d  for key %#x\n", m.ToString(), d, brd.hash_key)
		return false
	}
	captured_piece := m.CapturedPiece()
	to := m.To()
	if captured_piece != EMPTY {
		if !(brd.TypeAt(to) == captured_piece || (piece == PAWN && brd.enp_target == uint8(to))) {
			// fmt.Printf("Warning: Invalid move  %s at depth %d  for key %#x\n", m.ToString(), d, brd.hash_key)
			return false
		}
	}
	promoted_to := m.PromotedTo()
	if promoted_to != EMPTY {
		if piece != PAWN || promoted_to == PAWN {
			// fmt.Printf("Warning: Invalid move  %s  at depth %d for key %#x\n", m.ToString(), d, brd.hash_key)
			return false
		}
	}
	return true
}

func (m Move) From() int {
	return int(uint32(m) & uint32(63))
}

func (m Move) To() int {
	return int((uint32(m) >> 6) & uint32(63))
}

func (m Move) Piece() Piece {
	return Piece((uint32(m) >> 12) & uint32(7))
}

func (m Move) CapturedPiece() Piece {
	return Piece((uint32(m) >> 15) & uint32(7))
}

func (m Move) PromotedTo() Piece {
	return Piece((uint32(m) >> 18) & uint32(7))
}

func (m Move) IsQuiet() bool {
	if ((uint32(m) >> 15) & uint32(63)) == uint32(QUIET) {
		return true
	} else {
		return false
	}
}

func (m Move) IsCapture() bool {
	return m.CapturedPiece() != EMPTY
}

func (m Move) IsPromotion() bool {
	return m.PromotedTo() != EMPTY
}

var piece_chars = [6]string{"p", "n", "b", "r", "q", "k"}

func (m Move) ToString() string {
	var str string

	from_row := row(m.From())
	from_col := column(m.From())
	str += ParseCoordinates(from_row, from_col)
	to_row := row(m.To())
	to_col := column(m.To())
	str += ParseCoordinates(to_row, to_col)

	if m.PromotedTo() != EMPTY {
		str += piece_chars[m.PromotedTo()]
	}
	return str
}

// func (m Move) IsLegal(brd *Board) bool {
// 	if m.Piece() == KING {
// 		return !is_attacked_by(brd, m.To(), brd.c, brd.Enemy()) {
// 	} else {
// 		pinned := is_pinned(brd, m.From(), brd.c, brd.Enemy())
// 		return !(pinned > 0 && ((^pinned)&sq_mask_on[m.To()]) > 0)
// 	}
// }

func NewMove(from, to int, piece, captured_piece, promoted_to Piece) Move {
	return Move(from) | (Move(to) << 6) | (Move(piece) << 12) | (Move(captured_piece) << 15) | (Move(promoted_to) << 18)
}

func NewRegularMove(from, to int, piece Piece) Move {
	return Move(from) | (Move(to) << 6) | (Move(piece) << 12) | (Move(EMPTY) << 15) | (Move(EMPTY) << 18)
}

func NewCapture(from, to int, piece, captured_piece Piece) Move {
	return Move(from) | (Move(to) << 6) | (Move(piece) << 12) | (Move(captured_piece) << 15) | (Move(EMPTY) << 18)
}

// since moving piece is always PAWN (0) for promotions, no need to merge in the moving piece.
func NewPromotion(from, to int, piece, promoted_to Piece) Move {
	return Move(from) | (Move(to) << 6) | (Move(piece) << 12) | (Move(EMPTY) << 15) | (Move(promoted_to) << 18)
}
