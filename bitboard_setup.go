//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

// "fmt"

const ( // direction codes (0...8)
	NW = iota
	NE
	SE
	SW
	NORTH // 4
	EAST
	SOUTH
	WEST // 7
	DIR_INVALID
)

const (
	OFF_SINGLE = iota
	OFF_DOUBLE
	OFF_LEFT
	OFF_RIGHT
)

var pawnFromOffsets = [2][4]int{{8, 16, 9, 7}, {-8, -16, -7, -9}}
var knightOffsets = [8]int{-17, -15, -10, -6, 6, 10, 15, 17}
var bishopOffsets = [4]int{7, 9, -7, -9}
var rookOffsets = [4]int{8, 1, -8, -1}
var kingOffsets = [8]int{-9, -7, 7, 9, -8, -1, 1, 8}
var pawnAttackOffsets = [4]int{9, 7, -9, -7}

// var pawnAdvanceOffsets = [4]int{8, 16, -8, -16}

var directions [64][64]int

var oppositeDir = [16]int{SE, SW, NW, NE, SOUTH, WEST, NORTH, EAST, DIR_INVALID}

// var middle_rows BB

var maskOfLength [65]uint64

var rowMasks, columnMasks [8]BB

var pawnIsolatedMasks, pawnSideMasks, pawnDoubledMasks, knightMasks, bishopMasks, rookMasks,
	queenMasks, kingMasks, sqMaskOn, sqMaskOff [64]BB

var intervening, lineMasks [64][64]BB

var castleQueensideIntervening, castleKingsideIntervening [2]BB

var pawnAttackMasks, pawnPassedMasks, pawnAttackSpans, pawnBackwardSpans, pawnFrontSpans,
	pawnStopMasks, kingZoneMasks, kingShieldMasks [2][64]BB

var rayMasks [8][64]BB

var pawnStopSq, pawnPromoteSq [2][64]int

func manhattanDistance(from, to int) int {
	return abs(row(from)-row(to)) + abs(column(from)-column(to))
}

func setupSquareMasks() {
	for i := 0; i < 64; i++ {
		sqMaskOn[i] = BB(1 << uint(i))
		sqMaskOff[i] = (^sqMaskOn[i])
		maskOfLength[i] = uint64(sqMaskOn[i] - 1)
	}
}

func setupPawnMasks() {
	var sq int
	for i := 0; i < 64; i++ {
		pawnSideMasks[i] = (kingMasks[i] & rowMasks[row(i)])
		if i < 56 {
			pawnStopMasks[WHITE][i] = sqMaskOn[i] << 8
			pawnStopSq[WHITE][i] = i + 8
			for j := 0; j < 2; j++ {
				sq = i + pawnAttackOffsets[j]
				if manhattanDistance(sq, i) == 2 {
					pawnAttackMasks[WHITE][i].Add(sq)
				}
			}
		}
		if i > 7 {
			pawnStopMasks[BLACK][i] = sqMaskOn[i] >> 8
			pawnStopSq[BLACK][i] = i - 8
			for j := 2; j < 4; j++ {
				sq = i + pawnAttackOffsets[j]
				if manhattanDistance(sq, i) == 2 {
					pawnAttackMasks[BLACK][i].Add(sq)
				}
			}
		}
	}
}

func setupKnightMasks() {
	var sq int
	for i := 0; i < 64; i++ {
		for j := 0; j < 8; j++ {
			sq = i + knightOffsets[j]
			if onBoard(sq) && manhattanDistance(sq, i) == 3 {
				knightMasks[i] |= sqMaskOn[sq]
			}
		}
	}
}

func setupBishopMasks() {
	var previous, current, offset int
	for i := 0; i < 64; i++ {
		for j := 0; j < 4; j++ {
			previous = i
			offset = bishopOffsets[j]
			current = i + offset
			for onBoard(current) && manhattanDistance(current, previous) == 2 {
				rayMasks[j][i].Add(current)
				previous = current
				current += offset
			}
		}
		bishopMasks[i] = rayMasks[NW][i] | rayMasks[NE][i] | rayMasks[SE][i] | rayMasks[SW][i]
	}
}

func setupRookMasks() {
	var previous, current, offset int
	for i := 0; i < 64; i++ {
		for j := 0; j < 4; j++ {
			previous = i
			offset = rookOffsets[j]
			current = i + offset
			for onBoard(current) && manhattanDistance(current, previous) == 1 {
				rayMasks[j+4][i].Add(current)
				previous = current
				current += offset
			}
		}
		rookMasks[i] = rayMasks[NORTH][i] | rayMasks[SOUTH][i] | rayMasks[EAST][i] | rayMasks[WEST][i]
	}
}

func setupQueenMasks() {
	for i := 0; i < 64; i++ {
		queenMasks[i] = bishopMasks[i] | rookMasks[i]
	}
}

func setupKingMasks() {
	var sq int
	var center BB
	for i := 0; i < 64; i++ {
		for j := 0; j < 8; j++ {
			sq = i + kingOffsets[j]
			if onBoard(sq) && manhattanDistance(sq, i) <= 2 {
				kingMasks[i].Add(sq)
			}
		}
		center = kingMasks[i] | sqMaskOn[i]
		// The king zone is the 3 x 4 square area consisting of the squares around the king and
		// the squares facing the enemy side.
		kingZoneMasks[WHITE][i] = center | (center << 8)
		kingZoneMasks[BLACK][i] = center | (center >> 8)
		// The king shield is the three squares adjacent to the king and closest to the enemy side.
		kingShieldMasks[WHITE][i] = (kingZoneMasks[WHITE][i] ^ center) >> 8
		kingShieldMasks[BLACK][i] = (kingZoneMasks[BLACK][i] ^ center) << 8
	}

}

func setupRowMasks() {
	rowMasks[0] = 0xff // set the first row to binary 11111111, or 255.
	for i := 1; i < 8; i++ {
		rowMasks[i] = (rowMasks[i-1] << 8) // create the remaining rows by shifting the previous
	} // row up by 8 squares.
	// middle_rows = row_masks[2] | row_masks[3] | row_masks[4] | row_masks[5]
}

func setupColumnMasks() {
	columnMasks[0] = 1
	for i := 0; i < 8; i++ { // create the first column
		columnMasks[0] |= columnMasks[0] << 8
	}
	for i := 1; i < 8; i++ { // create the remaining columns by transposing the first column rightward.
		columnMasks[i] = (columnMasks[i-1] << 1)
	}
}

func setupDirections() {
	var ray BB
	for i := 0; i < 64; i++ {
		for j := 0; j < 64; j++ {
			directions[i][j] = DIR_INVALID // initialize array.
		}
	}
	for i := 0; i < 64; i++ {
		for j := 0; j < 64; j++ {
			for dir := 0; dir < 8; dir++ {
				ray = rayMasks[dir][i]
				if sqMaskOn[j]&ray > 0 {
					directions[i][j] = dir
					intervening[i][j] = ray ^ (rayMasks[dir][j] | sqMaskOn[j])
					lineMasks[i][j] = ray | rayMasks[oppositeDir[dir]][j]
				}
			}
		}
	}
}

func setupPawnStructureMasks() {
	var col int
	for i := 0; i < 64; i++ {
		col = column(i)
		pawnIsolatedMasks[i] = (kingMasks[i] & (^columnMasks[col]))

		pawnPassedMasks[WHITE][i] = rayMasks[NORTH][i]
		pawnPassedMasks[BLACK][i] = rayMasks[SOUTH][i]
		if col < 7 {
			pawnPassedMasks[WHITE][i] |= pawnPassedMasks[WHITE][i] << BB(1)
			pawnPassedMasks[BLACK][i] |= pawnPassedMasks[BLACK][i] << BB(1)
		}
		if col > 0 {
			pawnPassedMasks[WHITE][i] |= pawnPassedMasks[WHITE][i] >> BB(1)
			pawnPassedMasks[BLACK][i] |= pawnPassedMasks[BLACK][i] >> BB(1)
		}

		pawnAttackSpans[WHITE][i] = pawnPassedMasks[WHITE][i] & (^columnMasks[col])
		pawnAttackSpans[BLACK][i] = pawnPassedMasks[BLACK][i] & (^columnMasks[col])

		pawnBackwardSpans[WHITE][i] = pawnAttackSpans[BLACK][i] | pawnSideMasks[i]
		pawnBackwardSpans[BLACK][i] = pawnAttackSpans[WHITE][i] | pawnSideMasks[i]

		pawnFrontSpans[WHITE][i] = pawnPassedMasks[WHITE][i] & (columnMasks[col])
		pawnFrontSpans[BLACK][i] = pawnPassedMasks[BLACK][i] & (columnMasks[col])

		pawnDoubledMasks[i] = pawnFrontSpans[WHITE][i] | pawnFrontSpans[BLACK][i]

		pawnPromoteSq[WHITE][i] = msb(pawnFrontSpans[WHITE][i])
		pawnPromoteSq[BLACK][i] = lsb(pawnFrontSpans[BLACK][i])
	}
}

func setupCastleMasks() {
	castleQueensideIntervening[WHITE] |= (sqMaskOn[B1] | sqMaskOn[C1] | sqMaskOn[D1])
	castleKingsideIntervening[WHITE] |= (sqMaskOn[F1] | sqMaskOn[G1])
	castleQueensideIntervening[BLACK] = (castleQueensideIntervening[WHITE] << 56)
	castleKingsideIntervening[BLACK] = (castleKingsideIntervening[WHITE] << 56)
}

func setupMasks() {
	setupRowMasks() // Create bitboard masks for each row and column.
	setupColumnMasks()
	setupSquareMasks() // First set up masks used to add/remove bits by their index.
	setupKnightMasks() // For each square, calculate bitboard attack maps showing
	setupBishopMasks() // the squares to which the given piece type may move. These are
	setupRookMasks()   // used as bitmasks during move generation to find pseudolegal moves.
	setupQueenMasks()
	setupKingMasks()
	setupDirections()
	setupPawnMasks()
	setupPawnStructureMasks()
	setupCastleMasks()
}
