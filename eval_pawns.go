//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

// "fmt"

const (
	DOUBLED_PENALTY  = 20
	ISOLATED_PENALTY = 12
	BACKWARD_PENALTY = 4
)

var passedPawnBonus = [2][8]int{
	{0, 192, 96, 48, 24, 12, 6, 0},
	{0, 6, 12, 24, 48, 96, 192, 0},
}
var tarraschBonus = [2][8]int{
	{0, 12, 8, 4, 2, 0, 0, 0},
	{0, 0, 0, 2, 4, 8, 12, 0},
}
var defenseBonus = [2][8]int{
	{0, 12, 8, 6, 5, 4, 3, 0},
	{0, 3, 4, 5, 6, 8, 12, 0},
}
var duoBonus = [2][8]int{
	{0, 0, 2, 1, 1, 1, 0, 0},
	{0, 0, 1, 1, 1, 2, 0, 0},
}

var promoteRow = [2][2]int{
	{1, 2},
	{6, 5},
}

// PAWN EVALUATION
// Good structures:
//   -Passed pawns - Bonus for pawns unblocked by an enemy pawn on the same or adjacent file.
//                   May eventually get promoted.
//   -Pawn duos - Pawns side by side to another friendly pawn receive a small bonus
// Bad structures:
//   -Isolated pawns - Penalty for any pawn without friendly pawns on adjacent files.
//   -Double/tripled pawns - Penalty for having multiple pawns on the same file.
//   -Backward pawns

func setPawnStructure(brd *Board, pentry *PawnEntry) {
	pentry.key = brd.pawnHashKey
	setPawnMaps(brd, pentry, WHITE)
	setPawnMaps(brd, pentry, BLACK)
	pentry.value[WHITE] = pawnStructure(brd, pentry, WHITE, BLACK) -
		pawnStructure(brd, pentry, BLACK, WHITE)
	pentry.value[BLACK] = -pentry.value[WHITE]
}

func setPawnMaps(brd *Board, pentry *PawnEntry, c uint8) {
	pentry.leftAttacks[c], pentry.rightAttacks[c] = pawnAttacks(brd, c)
	pentry.allAttacks[c] = pentry.leftAttacks[c] | pentry.rightAttacks[c]
	pentry.count[c] = uint8(popCount(brd.pieces[c][PAWN]))
	pentry.passedPawns[c] = 0
}

// pawnStructure() sets the remaining pentry attributes for side c
func pawnStructure(brd *Board, pentry *PawnEntry, c, e uint8) int {

	var value, sq, sqRow int
	ownPawns, enemyPawns := brd.pieces[c][PAWN], brd.pieces[e][PAWN]
	for b := ownPawns; b > 0; b.Clear(sq) {
		sq = furthestForward(c, b)
		sqRow = row(sq)

		if (pawnAttackMasks[e][sq])&ownPawns > 0 { // defended pawns
			value += defenseBonus[c][sqRow]
		}
		if (pawnSideMasks[sq] & ownPawns) > 0 { // pawn duos
			value += duoBonus[c][sqRow]
		}

		if pawnDoubledMasks[sq]&ownPawns > 0 { // doubled or tripled pawns
			value -= DOUBLED_PENALTY
		}

		if pawnPassedMasks[c][sq]&enemyPawns == 0 { // passed pawns
			value += passedPawnBonus[c][sqRow]
			pentry.passedPawns[c].Add(sq) // note the passed pawn location in the pawn hash entry.
		} else { // don't penalize passed pawns for being isolated.
			if pawnIsolatedMasks[sq]&ownPawns == 0 {
				value -= ISOLATED_PENALTY // isolated pawns
			}
		}

		// https://chessprogramming.wikispaces.com/Backward+Pawn
		// backward pawns:
		// 1. cannot be defended by friendly pawns,
		// 2. their stop square is defended by an enemy sentry pawn,
		// 3. their stop square is not defended by a friendly pawn
		if (pawnBackwardSpans[c][sq]&ownPawns == 0) &&
			(pentry.allAttacks[e]&pawnStopMasks[c][sq] > 0) {
			value -= BACKWARD_PENALTY
		}
	}
	return value
}

func netPawnPlacement(brd *Board, pentry *PawnEntry, c, e uint8) int {
	return pentry.value[c] + netPassedPawns(brd, pentry, c, e)
}

func netPassedPawns(brd *Board, pentry *PawnEntry, c, e uint8) int {
	return evalPassedPawns(brd, c, e, pentry.passedPawns[c]) -
		evalPassedPawns(brd, e, c, pentry.passedPawns[e])
}

func evalPassedPawns(brd *Board, c, e uint8, passedPawns BB) int {
	var value, sq int
	enemyKingSq := brd.KingSq(e)
	for ; passedPawns > 0; passedPawns.Clear(sq) {
		sq = furthestForward(c, passedPawns)
		// Tarrasch rule: assign small bonus for friendly rook behind the passed pawn
		if pawnFrontSpans[e][sq]&brd.pieces[c][ROOK] > 0 {
			value += tarraschBonus[c][row(sq)]
		}
		// pawn race: Assign a bonus if the pawn is closer to its promote square than the enemy king.
		promoteSquare := pawnPromoteSq[c][sq]
		if brd.c == c {
			if chebyshevDistance(sq, promoteSquare) < (chebyshevDistance(enemyKingSq, promoteSquare)) {
				value += passedPawnBonus[c][row(sq)]
			}
		} else {
			if chebyshevDistance(sq, promoteSquare) < (chebyshevDistance(enemyKingSq, promoteSquare) - 1) {
				value += passedPawnBonus[c][row(sq)]
			}
		}
	}
	return value
}
