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

	c, e := brd.c, brd.Enemy()
	piece := m.Piece()
	captured_piece := m.CapturedPiece()

	if brd.pieces[c][piece]&sq_mask_on[m.From()] == 0 { // Check that there's really a piece of this type on the from square.
		return false
	}

	if sq_mask_on[m.To()]&brd.occupied[c] > 0 {
		return false
	}

	switch piece {
	case PAWN:
		diff := m.To() - m.From()
		if captured_piece != EMPTY {
			if !(brd.pieces[e][captured_piece]&sq_mask_on[m.To()] > 0 ||
				(brd.enp_target != SQ_INVALID && pawn_side_masks[m.From()]&sq_mask_on[brd.enp_target] > 0)) {
				return false
			}
			if brd.c == WHITE {
				if !(diff == 7 || diff == 9) {
					return false
				}
			} else {
				if !(diff == -7 || diff == -9) {
					return false
				}
			}
		} else {
			if brd.c == WHITE {
				if diff == 8 {
					if brd.TypeAt(m.To()) != EMPTY {
						return false
					}
				} else if diff == 16 {
					if brd.TypeAt(m.To()) != EMPTY || brd.TypeAt(m.From()+8) != EMPTY {
						return false
					}
				} else {
					return false
				}
			} else {
				if diff == -8 {
					if brd.TypeAt(m.To()) != EMPTY {
						return false
					}
				} else if diff == -16 {
					if brd.TypeAt(m.To()) != EMPTY || brd.TypeAt(m.From()-8) != EMPTY {
						return false
					}
				} else {
					return false
				}
			}
		}

	case KING:
		if captured_piece != EMPTY && brd.pieces[e][captured_piece]&sq_mask_on[m.To()] == 0 {
			return false
		}
		if abs(m.To()-m.From()) == 2 { // validate castle moves
			if brd.c == WHITE {
				switch m.To() {
				case C1:
					if !((brd.castle&C_WQ > uint8(0)) && castle_queenside_intervening[WHITE]&brd.AllOccupied() == 0 &&
						!is_attacked_by(brd, B1, e, c) && !is_attacked_by(brd, C1, e, c) && !is_attacked_by(brd, D1, e, c)) {
						return false
					}
				case G1:
					if !((brd.castle&C_WK > uint8(0)) && castle_kingside_intervening[WHITE]&brd.AllOccupied() == 0 &&
						!is_attacked_by(brd, F1, e, c) && !is_attacked_by(brd, G1, e, c)) {
						return false
					}
				}
			} else {
				switch m.To() {
				case C8:
					if !((brd.castle&C_BQ > uint8(0)) && castle_queenside_intervening[BLACK]&brd.AllOccupied() == 0 &&
						!is_attacked_by(brd, B8, e, c) && !is_attacked_by(brd, C8, e, c) && !is_attacked_by(brd, D8, e, c)) {
						return false
					}
				case G8:
					if !((brd.castle&C_BK > uint8(0)) && castle_kingside_intervening[BLACK]&brd.AllOccupied() == 0 &&
						!is_attacked_by(brd, F8, e, c) && !is_attacked_by(brd, G8, e, c)) {
						return false
					}
				}
			}
		}
	case KNIGHT:
		if captured_piece != EMPTY && brd.pieces[e][captured_piece]&sq_mask_on[m.To()] == 0 {
			return false
		}
	default:
		if captured_piece != EMPTY && brd.pieces[e][captured_piece]&sq_mask_on[m.To()] == 0 {
			return false
		}
		if intervening[m.From()][m.To()]&brd.AllOccupied() > 0 { // check intervening squares are empty for sliding attacks.
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
	return m.IsCapture() || m.IsPromotion()
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
