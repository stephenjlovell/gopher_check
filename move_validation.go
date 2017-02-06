//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import "fmt"

// Used to verify that a killer or hash move is legal.
func (brd *Board) LegalMove(m Move, inCheck bool) bool {
	if inCheck {
		return brd.EvadesCheck(m)
	} else {
		return brd.PseudolegalAvoidsCheck(m)
	}
}

// Moves generated while in check should already be legal, since we determine this
// as a side-effect of generating evasions.
func (brd *Board) AvoidsCheck(m Move, inCheck bool) bool {
	return inCheck || brd.PseudolegalAvoidsCheck(m)
}

func (brd *Board) PseudolegalAvoidsCheck(m Move) bool {
	switch m.Piece() {
	case PAWN:
		if m.CapturedPiece() == PAWN && brd.TypeAt(m.To()) == EMPTY { // En-passant
			return pinnedCanMove(brd, m.From(), m.To(), brd.c, brd.Enemy()) &&
				isPinned(brd, int(brd.enpTarget), brd.c, brd.Enemy())&sqMaskOn[m.To()] > 0
		} else {
			return pinnedCanMove(brd, m.From(), m.To(), brd.c, brd.Enemy())
		}
	case KING:
		return !isAttackedBy(brd, brd.AllOccupied(), m.To(), brd.Enemy(), brd.c)
	default:
		return pinnedCanMove(brd, m.From(), m.To(), brd.c, brd.Enemy())
	}
}

// Only called when in check
func (brd *Board) EvadesCheck(m Move) bool {
	piece, from, to := m.Piece(), m.From(), m.To()
	c, e := brd.c, brd.Enemy()

	if piece == KING {
		return !isAttackedBy(brd, occAfterMove(brd.AllOccupied(), from, to), to, e, c)
	} else {
		occ := brd.AllOccupied()
		kingSq := brd.KingSq(c)
		threats := colorAttackMap(brd, occ, kingSq, e, c)

		// TODO: EvadesCheck() called from non-check position in rare cases. Examples:
		// 5r1k/1b3p1p/pp3p1q/3n4/1P2R3/P2B1PP1/7P/6K1 w - - 0 1
		// 8/PPKR4/1Bn4P/3P3R/8/2p4r/pp4p1/r6k w - - 5 2  (r h3h5 x r)...?

		if threats == 0 {
			fmt.Println("info string EvadesCheck() called from non-check position!")
			// m.Print()
			// brd.PrintDetails()
			return true // no threats to evade.
		}

		if popCount(threats) > 1 {
			return false // only king moves can escape from double check.
		}
		if (threats|intervening[furthestForward(e, threats)][kingSq])&sqMaskOn[to] == 0 {
			return false // the moving piece must kill or block the attacking piece.
		}
		if !pinnedCanMove(brd, from, to, c, e) {
			return false // the moving piece can't be pinned to the king.
		}
		if brd.enpTarget != SQ_INVALID && piece == PAWN && m.CapturedPiece() == PAWN && // En-passant
			brd.TypeAt(to) == EMPTY {
			occ = occAfterMove(occ, from, to) & sqMaskOff[brd.enpTarget]

			return colorAttackMap(brd, occ, kingSq, e, c) == 0
			// return attacks_after_move(brd, occ, occ&brd.occupied[e], king_sq, e, c) == 0
		}
	}
	return true
}

func (brd *Board) ValidMove(m Move, inCheck bool) bool {
	if !m.IsMove() {
		return false
	}
	c, e := brd.c, brd.Enemy()
	piece, from, to, capturedPiece := m.Piece(), m.From(), m.To(), m.CapturedPiece()

	if brd.TypeAt(from) != piece || brd.pieces[c][piece]&sqMaskOn[from] == 0 {
		// fmt.Printf("No piece of this type available at from square!{%s}", m.ToString())
		return false
	}
	if sqMaskOn[to]&brd.occupied[c] > 0 {
		// fmt.Printf("To square occupied by own piece!{%s}", m.ToString())
		return false
	}
	if capturedPiece == KING || brd.pieces[c][KING] == 0 {
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
			return brd.TypeAt(to) == EMPTY && brd.TypeAt(pawnStopSq[c][from]) == EMPTY
		} else if capturedPiece == EMPTY {
			// fmt.Printf("Invalid pawn move!{%s}", m.ToString())
			return false
		} else {
			if capturedPiece == PAWN && brd.TypeAt(to) == EMPTY {
				if brd.enpTarget != SQ_INVALID && pawnStopSq[e][to] == int(brd.enpTarget) {
					return true
				} else {
					// fmt.Printf("Invalid En-passant move!{%s}", m.ToString())
					return false
				}
			} else {
				return brd.TypeAt(to) == capturedPiece
			}
		}
	case KING:
		if abs(to-from) == 2 { // validate castle moves
			occ := brd.AllOccupied()
			if c == WHITE && !inCheck && (brd.castle&12) > 0 {
				switch to {
				case C1:
					if (brd.castle&C_WQ > uint8(0)) && castleQueensideIntervening[WHITE]&occ == 0 &&
						!isAttackedBy(brd, occ, B1, e, c) && !isAttackedBy(brd, occ, C1, e, c) &&
						!isAttackedBy(brd, occ, D1, e, c) {
						return true
					}
				case G1:
					if (brd.castle&C_WK > uint8(0)) && castleKingsideIntervening[WHITE]&occ == 0 &&
						!isAttackedBy(brd, occ, F1, e, c) && !isAttackedBy(brd, occ, G1, e, c) {
						return true
					}
				}
			} else if c == BLACK && !inCheck && (brd.castle&3) > 0 {
				switch to {
				case C8:
					if (brd.castle&C_BQ > uint8(0)) && castleQueensideIntervening[BLACK]&occ == 0 &&
						!isAttackedBy(brd, occ, B8, e, c) && !isAttackedBy(brd, occ, C8, e, c) &&
						!isAttackedBy(brd, occ, D8, e, c) {
						return true
					}
				case G8:
					if (brd.castle&C_BK > uint8(0)) && castleKingsideIntervening[BLACK]&occ == 0 &&
						!isAttackedBy(brd, occ, F8, e, c) && !isAttackedBy(brd, occ, G8, e, c) {
						return true
					}
				}
			}
			return false
		}
	case KNIGHT:
		// no special treatment needed for knights.
	default:
		if slidingAttacks(piece, brd.AllOccupied(), from)&sqMaskOn[to] == 0 {
			return false
		}
	}

	return brd.TypeAt(to) == capturedPiece
}
