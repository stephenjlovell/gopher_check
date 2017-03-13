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
// 3. During s.quiescence search, SEE is used to prune losing captures. This provides a very low-risk
//    way of reducing the size of the q-search without impacting playing strength.
func GetSee(brd *Board, from, to int, capturedPiece Piece) int {

	tempColor := brd.Enemy()
	// get initial map of all squares directly attacking this square (does not include 'discovered'/hidden attacks)
	bAttackers := brd.pieces[WHITE][BISHOP] | brd.pieces[BLACK][BISHOP] |
		brd.pieces[WHITE][QUEEN] | brd.pieces[BLACK][QUEEN]
	rAttackers := brd.pieces[WHITE][ROOK] | brd.pieces[BLACK][ROOK] |
		brd.pieces[WHITE][QUEEN] | brd.pieces[BLACK][QUEEN]

	tempOcc := brd.AllOccupied()
	tempMap := AttackMap(brd, tempOcc, to)

	var tempPieces BB

	var pieceList [20]int
	count := 1

	// if capturedPiece == KING {
	// 	// King captures should never be generated. If we reach this line, it means we've introduced a
	// 	// bug somewhere in move generation or tree traversal.
	// 	fmt.Printf("info string king capture detected in getSee(): %s\n", BoardToFEN(brd))
	// 	return SEE_MIN
	// }

	pc := brd.TypeAt(from)
	if pc == KING { // Only commit to the attack if target piece is undefended.
		if tempMap&brd.occupied[tempColor] > 0 {
			return SEE_MIN
		} else {
			return pieceValues[capturedPiece]
		}
	}
	// before entering the main loop, perform each step once for the initial attacking piece.
	// This ensures that the moved piece is the first to capture.
	pieceList[0] = pieceValues[capturedPiece]
	nextVictim := brd.ValueAt(from)

	tempOcc.Clear(from)
	if pc != KNIGHT && pc != KING { // if the attacker was a pawn, bishop, rook, or queen, re-scan for hidden attacks:
		if pc == PAWN || pc == BISHOP || pc == QUEEN {
			tempMap |= BishopAttacks(tempOcc, to) & bAttackers
		}
		if pc == PAWN || pc == ROOK || pc == QUEEN {
			tempMap |= RookAttacks(tempOcc, to) & rAttackers
		}
	}

	for tempMap &= tempOcc; tempMap > 0; tempMap &= tempOcc {
		for pc = PAWN; pc <= KING; pc++ { // loop over piece ts in order of value.
			tempPieces = brd.pieces[tempColor][pc] & tempMap
			if tempPieces > 0 {
				break
			} // stop as soon as a match is found.
		}
		if pc >= KING {
			if pc == KING {
				if tempMap&brd.occupied[tempColor^1] > 0 {
					break // only commit a king to the attack if the other side has no defenders left.
				}
			}
			break
		}

		pieceList[count] = nextVictim - pieceList[count-1]
		nextVictim = pieceValues[pc]

		count++

		if (pieceList[count-1] - nextVictim) > 0 { // TODO: validate this.
			break
		}

		tempOcc ^= (tempPieces & -tempPieces) // merge the first set bit of temp_pieces into temp_occ
		if pc != KNIGHT && pc != KING {
			if pc == PAWN || pc == BISHOP || pc == QUEEN {
				tempMap |= (BishopAttacks(tempOcc, to) & bAttackers)
			}
			if pc == ROOK || pc == QUEEN {
				tempMap |= (RookAttacks(tempOcc, to) & rAttackers)
			}
		}
		tempColor ^= 1
	}

	for count-1 > 0 {
		count--
		pieceList[count-1] = -Max(-pieceList[count-1], pieceList[count])
	}
	// fmt.Printf(" %d ", piece_list[0])
	return pieceList[0]
}
