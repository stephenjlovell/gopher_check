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

type Piece uint8

func (pc Piece) Value() int        { return piece_values[pc] }
func (pc Piece) PromoteValue() int { return promote_values[pc] }

// When spawning new goroutines for subtree search, a deep copy of the Board struct will have to be made
// and passed to the new goroutine.  Keep this struct as small as possible.
type Board struct {
	pieces          [2][6]BB  // 768 bits
	squares         [64]Piece // 512 bits
	occupied        [2]BB     // 128 bits
	material        [2]int32  // 64  bits
	hash_key        uint64    // 64  bits
	pawn_hash_key   uint64    // 64  bits
	c               uint8     // 8   bits
	castle          uint8     // 8   bits
	enp_target      uint8     // 8 	bits
	halfmove_clock  uint8     // 8 	bits
	endgame_counter uint8     // 8 	bits
}

type BoardMemento struct { // memento object used to store board state to unmake later.
	hash_key       uint64
	pawn_hash_key  uint64
	castle         uint8
	enp_target     uint8
	halfmove_clock uint8
}

func (brd *Board) NewMemento() *BoardMemento {
	return &BoardMemento{
		hash_key:       brd.hash_key,
		pawn_hash_key:  brd.pawn_hash_key,
		castle:         brd.castle,
		enp_target:     brd.enp_target,
		halfmove_clock: brd.halfmove_clock,
	}
}

// Used to verify that a killer or hash move is legal. 
func (brd *Board) LegalMove(m Move, in_check bool) bool {
	if in_check {
		return brd.EvadesCheck(m)
	} else {
		return brd.PseudolegalAvoidsCheck(m)
	}
}

// moves generated while in check should already be legal.
func (brd *Board) AvoidsCheck(m Move, in_check bool) bool {
	return in_check || brd.PseudolegalAvoidsCheck(m)
}

func (brd *Board) PseudolegalAvoidsCheck(m Move) bool {
	switch m.Piece() {
	case PAWN:
		if m.CapturedPiece() == PAWN && brd.TypeAt(m.To()) == EMPTY { // En-passant
			return pinned_can_move(brd, m.From(), m.To(), brd.c, brd.Enemy()) &&
				is_pinned(brd, int(brd.enp_target), brd.c, brd.Enemy()) & sq_mask_on[m.To()] > 0 
		} else {
			return pinned_can_move(brd, m.From(), m.To(), brd.c, brd.Enemy())
		}
	case KING:
		return !is_attacked_by(brd, brd.AllOccupied(), m.To(), brd.Enemy(), brd.c)
	default:
		return pinned_can_move(brd, m.From(), m.To(), brd.c, brd.Enemy())
	}
}

// Only called when in check
func (brd *Board) EvadesCheck(m Move) bool {
	piece, from, to := m.Piece(), m.From(), m.To()
	c, e := brd.c, brd.Enemy()
	if brd.pieces[c][KING] == 0 {
		return false
	}
	occ := brd.AllOccupied()
	if piece == KING {
		return !is_attacked_by(brd, occ_after_move(brd.AllOccupied(), from, to), to, e, c)
	} else {
		king_sq := furthest_forward(c, brd.pieces[c][KING])
		threats := color_attack_map(brd, occ, king_sq, e, c)
		if pop_count(threats) > 1 {
			return false // only king moves can escape from double check.
		}
		if (threats|intervening[furthest_forward(e, threats)][king_sq]) & sq_mask_on[to] == 0 {
			return false // the moving piece must kill or block the attacking piece.
		}
		if !pinned_can_move(brd, from, to, c, e) {
			return false // the moving piece can't be pinned to the king.
		}
		if brd.enp_target != SQ_INVALID && piece == PAWN && m.CapturedPiece() == PAWN && // En-passant
			brd.TypeAt(to) == EMPTY {
			occ = occ_after_move(brd.AllOccupied(), from, to)&sq_mask_off[brd.enp_target]
			return attacks_after_move(brd, occ, occ&brd.occupied[e], king_sq, e, c) == 0
		}
	}
	return true
}


func (brd *Board) ValidMove(m Move, in_check bool) bool {
	if !m.IsMove() {
		return false
	}

	c, e := brd.c, brd.Enemy()
	piece, from, to, captured_piece := m.Piece(), m.From(), m.To(), m.CapturedPiece()

	if brd.TypeAt(from) != piece || brd.pieces[c][piece]&sq_mask_on[from] == 0 {
		// fmt.Printf("No piece of this type available at from square!{%s}", m.ToString())
		return false
	}
	if sq_mask_on[to]&brd.occupied[c] > 0 {
		// fmt.Printf("To square occupied by own piece!{%s}", m.ToString())
		return false
	}
	if captured_piece == KING || brd.pieces[c][KING] == 0 {
		// fmt.Printf("King capture detected!{%s}", m.ToString())
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
			// fmt.Printf("Invalid pawn movement direction!{%s}", m.ToString())
			return false
		} else if diff == 8 {
			return brd.TypeAt(to) == EMPTY
		} else if diff == 16 {
			return brd.TypeAt(to) == EMPTY && brd.TypeAt(get_offset(c, from, 8)) == EMPTY 
		} else if captured_piece == EMPTY {
			// fmt.Printf("Invalid pawn move!{%s}", m.ToString())
			return false
		} else { 
			if captured_piece == PAWN && brd.TypeAt(to) == EMPTY {
				if brd.enp_target != SQ_INVALID && get_offset(c, to, -8) == int(brd.enp_target) {
					return true
				} else {
					// fmt.Printf("Invalid En-passant move!{%s}", m.ToString())
					return false
				}
			} else {
				return brd.TypeAt(to) == captured_piece
			}
		}
	case KING:
		if abs(to-from) == 2 { // validate castle moves
			occ := brd.AllOccupied()
			if c == WHITE && (brd.castle&12) > 0 {
				switch to {
				case C1:
					if !in_check && (brd.castle&C_WQ > uint8(0)) && castle_queenside_intervening[WHITE]&brd.AllOccupied() == 0 &&
						!is_attacked_by(brd, occ, B1, e, c) && !is_attacked_by(brd, occ, C1, e, c) && !is_attacked_by(brd, occ, D1, e, c) {
						return true
					}
				case G1:
					if !in_check && (brd.castle&C_WK > uint8(0)) && castle_kingside_intervening[WHITE]&brd.AllOccupied() == 0 &&
						!is_attacked_by(brd, occ, F1, e, c) && !is_attacked_by(brd, occ, G1, e, c) {
						return true
					}
				}
			} else if c == BLACK && (brd.castle&3) > 0 {
				switch to {
				case C8:
					if !in_check && (brd.castle&C_BQ > uint8(0)) && castle_queenside_intervening[BLACK]&brd.AllOccupied() == 0 &&
						!is_attacked_by(brd, occ, B8, e, c) && !is_attacked_by(brd, occ, C8, e, c) && !is_attacked_by(brd, occ, D8, e, c) {
						return true
					}
				case G8:
					if !in_check && (brd.castle&C_BK > uint8(0)) && castle_kingside_intervening[BLACK]&brd.AllOccupied() == 0 &&
						!is_attacked_by(brd, occ, F8, e, c) && !is_attacked_by(brd, occ, G8, e, c) {
						return true
					}
				}
			}
			// fmt.Printf("Invalid castle move!{%s}. Castle rights: %d", m.ToString(), brd.castle)
			return false
		}
	case KNIGHT: // no special treatment needed for knights.

	default:
		if sliding_attacks(piece, brd.AllOccupied(), from) & sq_mask_on[to] == 0 {
			// fmt.Printf("Invalid sliding attack!{%s}", m.ToString())
			return false
		}
	}

	if brd.TypeAt(to) != captured_piece {
		// fmt.Printf("Captured piece not found on to square!{%s}", m.ToString())
		return false
	}

	return true
}


func (brd *Board) ValueAt(sq int) int {
	return brd.squares[sq].Value()
}

func (brd *Board) TypeAt(sq int) Piece {
	return brd.squares[sq]
}

func (brd *Board) Enemy() uint8 {
	return brd.c ^ 1
}

func (brd *Board) AllOccupied() BB { return brd.occupied[0] | brd.occupied[1] }

func (brd *Board) Placement(c uint8) BB { return brd.occupied[c] }

func (brd *Board) pawns_only() bool {
	return brd.occupied[brd.c] == brd.pieces[brd.c][PAWN]|brd.pieces[brd.c][KING]
}

func (brd *Board) Copy() *Board {
	return &Board{
		pieces:          brd.pieces,
		squares:         brd.squares,
		occupied:        brd.occupied,
		material:        brd.material,
		hash_key:        brd.hash_key,
		pawn_hash_key:   brd.pawn_hash_key,
		c:               brd.c,
		castle:          brd.castle,
		enp_target:      brd.enp_target,
		halfmove_clock:  brd.halfmove_clock,
		endgame_counter: brd.endgame_counter,
	}
}

func (brd *Board) PrintDetails() {
	fmt.Printf("hash_key: %d, pawn_hash_key: %d\n", brd.hash_key, brd.pawn_hash_key)
	fmt.Printf("castle: %d, enp_target: %d, halfmove_clock: %d\noccupied:\n", brd.castle, brd.enp_target, brd.halfmove_clock)
	for i := 0; i < 2; i++ {
		fmt.Printf("side: %d, material: %d\n", i, brd.material[i])
		brd.occupied[i].Print()
		for pc := 0; pc < 6; pc++ {
			fmt.Printf("piece: %d\n", pc)
			brd.pieces[i][pc].Print()
		}
	}
	brd.Print()
}

func (brd *Board) Print() {
	fmt.Printf("\n    A   B   C   D   E   F   G   H\n")
	fmt.Printf("  ---------------------------------\n")
	row := brd.squares[56:]
	fmt.Printf("8 ")
	brd.print_row(56, row)

	for i := 48; i >= 0; i -= 8 {
		row = brd.squares[i : i+8]
		fmt.Printf("%v ", 1+(i/8))
		brd.print_row(i, row)
	}
	fmt.Printf("    A   B   C   D   E   F   G   H\n")
}

func (brd *Board) print_row(start int, row []Piece) {
	fmt.Printf("| ")
	for i, piece := range row {
		if piece == EMPTY {
			fmt.Printf("  | ")
		} else {
			if brd.occupied[WHITE]&sq_mask_on[start+i] > 0 {
				fmt.Printf("%v | ", piece_graphics[WHITE][piece])
			} else {
				fmt.Printf("%v | ", piece_graphics[BLACK][piece])
			}
		}
	}
	fmt.Printf("\n  ---------------------------------\n")
}

func EmptyBoard() *Board {
	brd := &Board{enp_target: SQ_INVALID}
	for sq := 0; sq < 64; sq++ {
		brd.squares[sq] = EMPTY
	}
	return brd
}

var piece_graphics = [2][6]string{
	{"\u265F", "\u265E", "\u265D", "\u265C", "\u265B", "\u265A"},
	{"\u2659", "\u2658", "\u2657", "\u2656", "\u2655", "\u2654"},
}

const (
	A1 = iota
	B1
	C1
	D1
	E1
	F1
	G1
	H1
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A8
	B8
	C8
	D8
	E8
	F8
	G8
	H8
	SQ_INVALID
)
