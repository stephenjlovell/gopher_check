//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

const (
	SEE_MIN = -780 // worst possible outcome (trading a queen for a pawn)
	// SEE_MAX = 880  // best outcome (capturing an undefended queen)
)

// The Static Exchange Evaluation (SEE) heuristic provides a way to determine if a capture
// is a 'winning' or 'losing' capture.
// 1. When a capture results in an exchange of pieces by both sides, SEE is used to determine the
//    net gain/loss in material for the side initiating the exchange.
// 2. SEE scoring of moves is used for move ordering of captures at critical nodes.
// 3. During quiescence search, SEE is used to prune losing captures. This provides a very low-risk
//    way of reducing the size of the q-search without impacting playing strength.
func GetSee(brd *Board, tempOcc BB, from, to int, capturedPiece Piece) int {

	tempColor := brd.Enemy()
	// get initial map of all squares directly attacking this square (does not include 'discovered'/hidden attacks)
	bAttackers := brd.pieces[WHITE][BISHOP] | brd.pieces[BLACK][BISHOP] |
		brd.pieces[WHITE][QUEEN] | brd.pieces[BLACK][QUEEN]
	rAttackers := brd.pieces[WHITE][ROOK] | brd.pieces[BLACK][ROOK] |
		brd.pieces[WHITE][QUEEN] | brd.pieces[BLACK][QUEEN]

	// before entering the main loop, perform each step once for the initial attacking piece.
	// This ensures that the moved piece is the first to capture.
	pc := brd.TypeAt(from)
	tempOcc.Clear(from)

	if pc == KING {
		if IsAttackedBy(brd, tempOcc, to, tempColor, brd.c) {
			// If the captured piece is defended, this move leaves the king in check and would be discarded
			// by the move selector. Put this move toward the end of the move list to hopefully avoid
			// wasting time checking its legality.
			return SEE_MIN
		} else {
			return pieceValues[capturedPiece]
		}
	}
	// From square has already been cleared from tempOcc; no need to re-scan for hidden attackers.
	tempMap := AttackMap(brd, tempOcc, to)
	var pieceList [16]int
	pieceList[0] = pieceValues[capturedPiece]
	nextVictim := pieceValues[pc]
	count := 1
	var tempPieces BB

	for tempMap &= tempOcc; tempMap > 0; tempMap &= tempOcc {
		// Find the least valuable remaining attacking piece. Attacking side will always attack with its
		// least valuable pieces first.
		for pc = PAWN; pc <= KING; pc++ {
			tempPieces = brd.pieces[tempColor][pc] & tempMap
			if tempPieces > 0 {
				break
			}
		}

		if pc >= KING { // Only commit a king to the attack if the other side has no defenders left.
			if pc == KING && tempMap&brd.occupied[tempColor^1] == 0 {
				pieceList[count] = nextVictim - pieceList[count-1]
				count++
			}
			break
		}

		pieceList[count] = nextVictim - pieceList[count-1]
		nextVictim = pieceValues[pc]

		count++

		if (pieceList[count-1] - nextVictim) > 0 { // TODO: validate this.
			break
		}

		tempOcc ^= (tempPieces & -tempPieces) // Merge the first set bit of tempPieces into tempOcc.

		switch pc {
		case PAWN, BISHOP:
			tempMap |= (BishopAttacks(tempOcc, to) & bAttackers)
		case ROOK:
			tempMap |= (RookAttacks(tempOcc, to) & rAttackers)
		case QUEEN:
			tempMap |= (BishopAttacks(tempOcc, to) & bAttackers)
			tempMap |= (RookAttacks(tempOcc, to) & rAttackers)
		default:
		}

		tempColor ^= 1
	}

	for count >= 2 {
		count--
		pieceList[count-1] = -Max(-pieceList[count-1], pieceList[count])
	}
	return pieceList[0]
}
