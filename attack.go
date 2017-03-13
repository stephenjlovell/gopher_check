//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

func AttackMap(brd *Board, occ BB, sq int) BB {
	bb := ((pawnAttackMasks[BLACK][sq] & brd.pieces[WHITE][PAWN]) |
		(pawnAttackMasks[WHITE][sq] & brd.pieces[BLACK][PAWN])) | // Pawns
		(knightMasks[sq] & (brd.pieces[WHITE][KNIGHT] | brd.pieces[BLACK][KNIGHT])) | // Knights
		(kingMasks[sq] & (brd.pieces[WHITE][KING] | brd.pieces[BLACK][KING])) // Kings
	if bSliders := (brd.pieces[WHITE][BISHOP] | brd.pieces[BLACK][BISHOP] | brd.pieces[WHITE][QUEEN] | brd.pieces[BLACK][QUEEN]); bSliders&bishopMasks[sq] > 0 {
		bb |= (BishopAttacks(occ, sq) & bSliders) // Bishops and Queens
	}
	if rSliders := (brd.pieces[WHITE][ROOK] | brd.pieces[BLACK][ROOK] | brd.pieces[WHITE][QUEEN] | brd.pieces[BLACK][QUEEN]); rSliders&rookMasks[sq] > 0 {
		bb |= (RookAttacks(occ, sq) & rSliders) // Rooks and Queens
	}
	return bb
}

func ColorAttackMap(brd *Board, occ BB, sq int, c, e uint8) BB {
	bb := (pawnAttackMasks[e][sq] & brd.pieces[c][PAWN]) | // Pawns
		(knightMasks[sq] & brd.pieces[c][KNIGHT]) | // Knights
		(kingMasks[sq] & brd.pieces[c][KING]) // Kings
	if bSliders := (brd.pieces[c][BISHOP] | brd.pieces[c][QUEEN]); bSliders&bishopMasks[sq] > 0 {
		bb |= (BishopAttacks(occ, sq) & bSliders) // Bishops and Queens
	}
	if rSliders := (brd.pieces[c][ROOK] | brd.pieces[c][QUEEN]); rSliders&rookMasks[sq] > 0 {
		bb |= (RookAttacks(occ, sq) & rSliders) // Rooks and Queens
	}
	return bb
}

func IsAttackedBy(brd *Board, occ BB, sq int, attacker, defender uint8) bool {
	return (pawnAttackMasks[defender][sq]&brd.pieces[attacker][PAWN] > 0) || // Pawns
		(knightMasks[sq]&(brd.pieces[attacker][KNIGHT]) > 0) || // Knights
		(kingMasks[sq]&(brd.pieces[attacker][KING]) > 0) || // Kings
		(BishopAttacks(occ, sq)&(brd.pieces[attacker][BISHOP]|brd.pieces[attacker][QUEEN]) > 0) || // Bishops and Queens
		(RookAttacks(occ, sq)&(brd.pieces[attacker][ROOK]|brd.pieces[attacker][QUEEN]) > 0) // Rooks and Queens
}

func PinnedCanMove(brd *Board, from, to int, c, e uint8) bool {
	return IsPinned(brd, brd.AllOccupied(), from, c, e)&sqMaskOn[to] > 0
}

// Determines if a piece is blocking a ray attack to its king, and cannot move off this ray
// without placing its king in check.
// Returns the area to which the piece can move without leaving its king in check.
// 1. Find the displacement vector between the piece at sq and its own king and determine if it
//    lies along a valid ray attack.  If the vector is a valid ray attack:
// 2. Scan toward the king to see if there are any other pieces blocking this route to the king.
// 3. Scan in the opposite direction to see detect any potential threats along this ray.

// Return a bitboard of locations the piece at sq can move to without leaving the king in check.

func IsPinned(brd *Board, occ BB, sq int, c, e uint8) BB {
	var line, attacks, threat BB
	kingSq := brd.KingSq(c)
	line = lineMasks[sq][kingSq]
	if line > 0 { // can only be pinned if on a ray to the king.
		if directions[sq][kingSq] < NORTH { // direction toward king
			attacks = BishopAttacks(occ, sq)
			threat = line & attacks & (brd.pieces[e][BISHOP] | brd.pieces[e][QUEEN])
		} else {
			attacks = RookAttacks(occ, sq)
			threat = line & attacks & (brd.pieces[e][ROOK] | brd.pieces[e][QUEEN])
		}
		if threat > 0 && (attacks&brd.pieces[c][KING]) > 0 {
			return line & attacks
		}
	}
	return BB(ANY_SQUARE_MASK)
}

// TODO: handle other pieces getting king out of check...
func IsCheckmate(brd *Board, inCheck bool) bool {
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
		if !IsAttackedBy(brd, OccAfterMove(occ, from, to), to, e, c) {
			return false
		}
	}
	return true
}

func OccAfterMove(occ BB, from, to int) BB {
	return (occ | sqMaskOn[to]) & sqMaskOff[from]
}
