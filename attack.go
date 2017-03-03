//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import "fmt"

func attackMap(brd *Board, occ BB, sq int) BB {
	bb := ((pawnAttackMasks[BLACK][sq] & brd.pieces[WHITE][PAWN]) |
		(pawnAttackMasks[WHITE][sq] & brd.pieces[BLACK][PAWN])) | // Pawns
		(knightMasks[sq] & (brd.pieces[WHITE][KNIGHT] | brd.pieces[BLACK][KNIGHT])) | // Knights
		(kingMasks[sq] & (brd.pieces[WHITE][KING] | brd.pieces[BLACK][KING])) // Kings
	if bSliders := (brd.pieces[WHITE][BISHOP] | brd.pieces[BLACK][BISHOP] | brd.pieces[WHITE][QUEEN] | brd.pieces[BLACK][QUEEN]); bSliders&bishopMasks[sq] > 0 {
		bb |= (bishopAttacks(occ, sq) & bSliders) // Bishops and Queens
	}
	if rSliders := (brd.pieces[WHITE][ROOK] | brd.pieces[BLACK][ROOK] | brd.pieces[WHITE][QUEEN] | brd.pieces[BLACK][QUEEN]); rSliders&rookMasks[sq] > 0 {
		bb |= (rookAttacks(occ, sq) & rSliders) // Rooks and Queens
	}
	return bb
}

func colorAttackMap(brd *Board, occ BB, sq int, c, e uint8) BB {
	bb := (pawnAttackMasks[e][sq] & brd.pieces[c][PAWN]) | // Pawns
		(knightMasks[sq] & brd.pieces[c][KNIGHT]) | // Knights
		(kingMasks[sq] & brd.pieces[c][KING]) // Kings
	if bSliders := (brd.pieces[c][BISHOP] | brd.pieces[c][QUEEN]); bSliders&bishopMasks[sq] > 0 {
		bb |= (bishopAttacks(occ, sq) & bSliders) // Bishops and Queens
	}
	if rSliders := (brd.pieces[c][ROOK] | brd.pieces[c][QUEEN]); rSliders&rookMasks[sq] > 0 {
		bb |= (rookAttacks(occ, sq) & rSliders) // Rooks and Queens
	}
	return bb
}

func isAttackedBy(brd *Board, occ BB, sq int, attacker, defender uint8) bool {
	return (pawnAttackMasks[defender][sq]&brd.pieces[attacker][PAWN] > 0) || // Pawns
		(knightMasks[sq]&(brd.pieces[attacker][KNIGHT]) > 0) || // Knights
		(kingMasks[sq]&(brd.pieces[attacker][KING]) > 0) || // Kings
		(bishopAttacks(occ, sq)&(brd.pieces[attacker][BISHOP]|brd.pieces[attacker][QUEEN]) > 0) || // Bishops and Queens
		(rookAttacks(occ, sq)&(brd.pieces[attacker][ROOK]|brd.pieces[attacker][QUEEN]) > 0) // Rooks and Queens
}

func pinnedCanMove(brd *Board, from, to int, c, e uint8) bool {
	return isPinned(brd, brd.AllOccupied(), from, c, e)&sqMaskOn[to] > 0
}

// Determines if a piece is blocking a ray attack to its king, and cannot move off this ray
// without placing its king in check.
// Returns the area to which the piece can move without leaving its king in check.
// 1. Find the displacement vector between the piece at sq and its own king and determine if it
//    lies along a valid ray attack.  If the vector is a valid ray attack:
// 2. Scan toward the king to see if there are any other pieces blocking this route to the king.
// 3. Scan in the opposite direction to see detect any potential threats along this ray.

// Return a bitboard of locations the piece at sq can move to without leaving the king in check.

func isPinned(brd *Board, occ BB, sq int, c, e uint8) BB {
	var line, attacks, threat BB
	kingSq := brd.KingSq(c)
	line = lineMasks[sq][kingSq]
	if line > 0 { // can only be pinned if on a ray to the king.
		if directions[sq][kingSq] < NORTH { // direction toward king
			attacks = bishopAttacks(occ, sq)
			threat = line & attacks & (brd.pieces[e][BISHOP] | brd.pieces[e][QUEEN])
		} else {
			attacks = rookAttacks(occ, sq)
			threat = line & attacks & (brd.pieces[e][ROOK] | brd.pieces[e][QUEEN])
		}
		if threat > 0 && (attacks&brd.pieces[c][KING]) > 0 {
			return line & attacks
		}
	}
	return BB(ANY_SQUARE_MASK)
}

// The Static Exchange Evaluation (SEE) heuristic provides a way to determine if a capture
// is a 'winning' or 'losing' capture.
// 1. When a capture results in an exchange of pieces by both sides, SEE is used to determine the
//    net gain/loss in material for the side initiating the exchange.
// 2. SEE scoring of moves is used for move ordering of captures at critical nodes.
// 3. During s.quiescence search, SEE is used to prune losing captures. This provides a very low-risk
//    way of reducing the size of the q-search without impacting playing strength.
const (
	SEE_MIN = -780 // worst possible outcome (trading a queen for a pawn)
	// SEE_MAX = 880  // best outcome (capturing an undefended queen)
)

func getSee(brd *Board, from, to int, capturedPiece Piece) int {
	var nextVictim int
	var t Piece
	// var t, last_t Piece
	tempColor := brd.Enemy()
	// get initial map of all squares directly attacking this square (does not include 'discovered'/hidden attacks)
	bAttackers := brd.pieces[WHITE][BISHOP] | brd.pieces[BLACK][BISHOP] |
		brd.pieces[WHITE][QUEEN] | brd.pieces[BLACK][QUEEN]
	rAttackers := brd.pieces[WHITE][ROOK] | brd.pieces[BLACK][ROOK] |
		brd.pieces[WHITE][QUEEN] | brd.pieces[BLACK][QUEEN]

	tempOcc := brd.AllOccupied()
	tempMap := attackMap(brd, tempOcc, to)

	var tempPieces BB

	var pieceList [20]int
	count := 1

	if capturedPiece == KING {
		// this move is illegal and will be discarded by the move selector. Return the lowest possible
		// SEE value so that this move will be put at end of list.  If cutoff occurs before then,
		// the cost of detecting the illegal move will be saved.
		fmt.Println("info string king capture detected in getSee(): %s", BoardToFEN(brd))
		return SEE_MIN
	}
	t = brd.TypeAt(from)
	if t == KING { // Only commit to the attack if target piece is undefended.
		if tempMap&brd.occupied[tempColor] > 0 {
			return SEE_MIN
		} else {
			return pieceValues[capturedPiece]
		}
	}
	// before entering the main loop, perform each step once for the initial attacking piece.
	// This ensures that the moved piece is the first to capture.
	pieceList[0] = pieceValues[capturedPiece]
	nextVictim = brd.ValueAt(from)

	tempOcc.Clear(from)
	if t != KNIGHT && t != KING { // if the attacker was a pawn, bishop, rook, or queen, re-scan for hidden attacks:
		if t == PAWN || t == BISHOP || t == QUEEN {
			tempMap |= bishopAttacks(tempOcc, to) & bAttackers
		}
		if t == PAWN || t == ROOK || t == QUEEN {
			tempMap |= rookAttacks(tempOcc, to) & rAttackers
		}
	}

	for tempMap &= tempOcc; tempMap > 0; tempMap &= tempOcc {
		for t = PAWN; t <= KING; t++ { // loop over piece ts in order of value.
			tempPieces = brd.pieces[tempColor][t] & tempMap
			if tempPieces > 0 {
				break
			} // stop as soon as a match is found.
		}
		if t >= KING {
			if t == KING {
				if tempMap&brd.occupied[tempColor^1] > 0 {
					break // only commit a king to the attack if the other side has no defenders left.
				}
			}
			break
		}

		pieceList[count] = nextVictim - pieceList[count-1]
		nextVictim = pieceValues[t]

		count++

		if (pieceList[count-1] - nextVictim) > 0 { // validate this.
			break
		}

		tempOcc ^= (tempPieces & -tempPieces) // merge the first set bit of temp_pieces into temp_occ
		if t != KNIGHT && t != KING {
			if t == PAWN || t == BISHOP || t == QUEEN {
				tempMap |= (bishopAttacks(tempOcc, to) & bAttackers)
			}
			if t == ROOK || t == QUEEN {
				tempMap |= (rookAttacks(tempOcc, to) & rAttackers)
			}
		}
		tempColor ^= 1
	}

	for count-1 > 0 {
		count--
		pieceList[count-1] = -max(-pieceList[count-1], pieceList[count])
	}
	// fmt.Printf(" %d ", piece_list[0])
	return pieceList[0]
}

// TODO: handle other pieces getting king out of check...
func isCheckmate(brd *Board, inCheck bool) bool {
	if !inCheck {
		return false
	}
	c := brd.c
	e := brd.Enemy()
	var to int
	from := brd.KingSq(c)
	occ := brd.AllOccupied()
	for t := kingMasks[from] & (^brd.occupied[c]); t > 0; t.Clear(to) { // generate to squares
		to = furthestForward(c, t)
		if !isAttackedBy(brd, occAfterMove(occ, from, to), to, e, c) {
			return false
		}
	}
	return true
}

func occAfterMove(occ BB, from, to int) BB {
	return (occ | sqMaskOn[to]) & sqMaskOff[from]
}
