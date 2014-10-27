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

func (m Move) IsValid(brd *Board) bool {
	if m == 0 {
		return false
	}
	// return true

	c, e := brd.c, brd.Enemy()
	piece := m.Piece()
	from, to := m.From(), m.To()
	captured_piece := m.CapturedPiece()

	if brd.pieces[c][piece]&sq_mask_on[from] == 0 { // Check that there's really a piece of this type on the from square.
		// brd.pieces[c][piece].Print()
		// sq_mask_on[from].Print()
		fmt.Printf("No piece of type %s available at from square.\n", piece_chars[piece])
		return false
	}

	if sq_mask_on[to]&brd.occupied[c] > 0 {
		fmt.Println("To square occupied by own piece.")
		return false
	}

	switch piece {
	case PAWN:
		var diff int
		if c == WHITE {
			diff = to - from
		} else {
			diff = from - to
		}
		if diff < 0 {
			fmt.Println("Invalid pawn movement direction.")
			return false
		}
		if diff == 8 {
			if brd.TypeAt(to) == EMPTY {
				return true
			} else {
				fmt.Println("Pawn forward movement blocked.")
				return false
			}
		} else if diff == 16 {
			if brd.TypeAt(to) == EMPTY && brd.TypeAt(get_offset(c, from, 8)) == EMPTY {
				return true
			} else {
				fmt.Println("Pawn forward movement blocked.")
				return false
			}
		} else if captured_piece == EMPTY {
			fmt.Println("Invalid pawn move.")
			return false
		}
		if brd.enp_target != SQ_INVALID && captured_piece == PAWN &&
			brd.TypeAt(to) == EMPTY && get_offset(c, to, -8) == int(brd.enp_target) {
			return true
		}
	case KING:
		if abs(to-from) == 2 { // validate castle moves
			if c == WHITE && (brd.castle & 12) > 0 {
				switch to {
				case C1:
					if !((brd.castle&C_WQ > uint8(0)) && castle_queenside_intervening[WHITE]&brd.AllOccupied() == 0 &&
						!is_attacked_by(brd, B1, e, c) && !is_attacked_by(brd, C1, e, c) && !is_attacked_by(brd, D1, e, c)) {
						fmt.Println("Invalid castle move.")
						return false
					}
				case G1:
					if !((brd.castle&C_WK > uint8(0)) && castle_kingside_intervening[WHITE]&brd.AllOccupied() == 0 &&
						!is_attacked_by(brd, F1, e, c) && !is_attacked_by(brd, G1, e, c)) {
						fmt.Println("Invalid castle move.")
						return false
					}
				}
			} else if c == BLACK && (brd.castle & 3) > 0 {
				switch to {
				case C8:
					if !((brd.castle&C_BQ > uint8(0)) && castle_queenside_intervening[BLACK]&brd.AllOccupied() == 0 &&
						!is_attacked_by(brd, B8, e, c) && !is_attacked_by(brd, C8, e, c) && !is_attacked_by(brd, D8, e, c)) {
						fmt.Println("Invalid castle move.")
						return false
					}
				case G8:
					if !((brd.castle&C_BK > uint8(0)) && castle_kingside_intervening[BLACK]&brd.AllOccupied() == 0 &&
						!is_attacked_by(brd, F8, e, c) && !is_attacked_by(brd, G8, e, c)) {
						fmt.Println("Invalid castle move.")
						return false
					}
				}
			} else {
				fmt.Println("Invalid castle move.")
				return false
			}
		}
	case KNIGHT:

	default:
		if intervening[from][to]&brd.AllOccupied() > 0 { // check intervening squares are empty for sliding attacks.
			fmt.Println("Sliding piece blocked by intervening pieces.")
			return false
		}
	}

	if captured_piece != EMPTY && brd.pieces[e][captured_piece]&sq_mask_on[to] == 0 {
		fmt.Println("No piece of type %s available at to capture at to square\n", piece_chars[piece])
		return false
	}
	// if captured_piece == KING {  // if illegal move was actually generated by search to get to this position,
	// 	return false							 // this could prevent their detection by king capture...
	// }
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
	return m.IsCapture() || m.IsPromotion()
}

func (m Move) IsCapture() bool {
	return m.CapturedPiece() != EMPTY
}

func (m Move) IsPromotion() bool {
	return m.PromotedTo() != EMPTY
}

var piece_chars = [6]string{"p", "n", "b", "r", "q", "k"}


func (m Move) ToString() string { // string representation used for debugging only.
	var str string
	str += piece_chars[m.Piece()] + " "
	str += ParseCoordinates(row(m.From()), column(m.From()))
	str += ParseCoordinates(row(m.To()), column(m.To()))
	if m.IsCapture() {
		str += " x " + piece_chars[m.CapturedPiece()]
	}
	if m.IsPromotion() {
		str += " promoted to " + piece_chars[m.PromotedTo()]
	}
	return str
}

func (m Move) ToUCI() string {
	var str string
	str += ParseCoordinates(row(m.From()), column(m.From()))
	str += ParseCoordinates(row(m.To()), column(m.To()))
	if m.PromotedTo() != EMPTY {
		str += piece_chars[m.PromotedTo()]
	}
	return str
}

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
