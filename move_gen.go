//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

func getNonCaptures(brd *Board, htable *HistoryTable, remainingMoves *MoveList) {
	var from, to int
	var singleAdvances, doubleAdvances BB
	c := brd.c
	occ := brd.AllOccupied()
	empty := ^occ
	var m Move

	// Castles
	castle := brd.castle
	if castle > uint8(0) { // getNonCaptures is only called when not in check.
		e := brd.Enemy()
		if c == WHITE {
			if (castle&C_WQ > uint8(0)) && castleQueensideIntervening[WHITE]&occ == 0 &&
				!isAttackedBy(brd, occ, C1, e, c) && !isAttackedBy(brd, occ, D1, e, c) {
				m = NewRegularMove(E1, C1, KING)
				remainingMoves.Push(SortItem{htable.Probe(KING, c, C1) | 1, m})
			}
			if (castle&C_WK > uint8(0)) && castleKingsideIntervening[WHITE]&occ == 0 &&
				!isAttackedBy(brd, occ, F1, e, c) && !isAttackedBy(brd, occ, G1, e, c) {
				m = NewRegularMove(E1, G1, KING)
				remainingMoves.Push(SortItem{htable.Probe(KING, c, G1) | 1, m})
			}
		} else {
			if (castle&C_BQ > uint8(0)) && castleQueensideIntervening[BLACK]&occ == 0 &&
				!isAttackedBy(brd, occ, C8, e, c) && !isAttackedBy(brd, occ, D8, e, c) {
				m = NewRegularMove(E8, C8, KING)
				remainingMoves.Push(SortItem{htable.Probe(KING, c, C8) | 1, m})
			}
			if (castle&C_BK > uint8(0)) && castleKingsideIntervening[BLACK]&occ == 0 &&
				!isAttackedBy(brd, occ, F8, e, c) && !isAttackedBy(brd, occ, G8, e, c) {
				m = NewRegularMove(E8, G8, KING)
				remainingMoves.Push(SortItem{htable.Probe(KING, c, G8) | 1, m})
			}
		}
	}

	// Pawns
	//  Pawns behave differently than other pieces. They:
	//  1. can move only in one direction;
	//  2. can attack diagonally but can only advance on file (forward);
	//  3. can move an extra space from the starting square;
	//  4. can capture other pawns via the En-Passant Rule;
	//  5. are promoted to another piece type if they reach the enemy's back rank.
	if c > 0 { // white to move
		singleAdvances = (brd.pieces[WHITE][PAWN] << 8) & empty & (^rowMasks[7]) // promotions generated in getCaptures
		doubleAdvances = ((singleAdvances & rowMasks[2]) << 8) & empty
	} else { // black to move
		singleAdvances = (brd.pieces[BLACK][PAWN] >> 8) & empty & (^rowMasks[0])
		doubleAdvances = ((singleAdvances & rowMasks[5]) >> 8) & empty
	}
	for ; doubleAdvances > 0; doubleAdvances.Clear(to) {
		to = furthestForward(c, doubleAdvances)
		from = to + pawnFromOffsets[c][OFF_DOUBLE]
		m = NewRegularMove(from, to, PAWN)
		remainingMoves.Push(SortItem{htable.Probe(PAWN, c, to), m})
	}
	for ; singleAdvances > 0; singleAdvances.Clear(to) {
		to = furthestForward(c, singleAdvances)
		from = to + pawnFromOffsets[c][OFF_SINGLE]
		m = NewRegularMove(from, to, PAWN)
		remainingMoves.Push(SortItem{htable.Probe(PAWN, c, to), m})
	}
	// Knights
	for f := brd.pieces[c][KNIGHT]; f > 0; f.Clear(from) {
		from = furthestForward(c, f)                               // Locate each knight for the side to move.
		for t := (knightMasks[from] & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewRegularMove(from, to, KNIGHT)
			remainingMoves.Push(SortItem{htable.Probe(KNIGHT, c, to), m})
		}
	}
	// Bishops
	for f := brd.pieces[c][BISHOP]; f > 0; f.Clear(from) {
		from = furthestForward(c, f)
		for t := (bishopAttacks(occ, from) & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewRegularMove(from, to, BISHOP)
			remainingMoves.Push(SortItem{htable.Probe(BISHOP, c, to), m})
		}
	}
	// Rooks
	for f := brd.pieces[c][ROOK]; f > 0; f.Clear(from) {
		from = furthestForward(c, f)
		for t := (rookAttacks(occ, from) & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewRegularMove(from, to, ROOK)
			remainingMoves.Push(SortItem{htable.Probe(ROOK, c, to), m})
		}
	}
	// Queens
	for f := brd.pieces[c][QUEEN]; f > 0; f.Clear(from) {
		from = furthestForward(c, f)
		for t := (queenAttacks(occ, from) & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewRegularMove(from, to, QUEEN)
			remainingMoves.Push(SortItem{htable.Probe(QUEEN, c, to), m})
		}
	}
	// Kings
	for f := brd.pieces[c][KING]; f > 0; f.Clear(from) {
		from = brd.KingSq(c)
		for t := (kingMasks[from] & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewRegularMove(from, to, KING)
			remainingMoves.Push(SortItem{htable.Probe(KING, c, to), m})
		}
	}
}

// Pawn promotions are also generated during getCaptures routine.
func getCaptures(brd *Board, htable *HistoryTable, winning, losing *MoveList) {
	var from, to int
	var m Move

	c, e := brd.c, brd.Enemy()
	occ := brd.AllOccupied()
	enemy := brd.Placement(e)

	// Pawns
	var leftTemp, rightTemp BB
	var leftAttacks, rightAttacks BB
	var promotionCapturesLeft, promotionCapturesRight, promotionAdvances BB

	if c == WHITE { // white to move
		leftTemp = ((brd.pieces[c][PAWN] & (^columnMasks[0])) << 7) & enemy
		leftAttacks = leftTemp & (^rowMasks[7])
		promotionCapturesLeft = leftTemp & (rowMasks[7])

		rightTemp = ((brd.pieces[c][PAWN] & (^columnMasks[7])) << 9) & enemy
		rightAttacks = rightTemp & (^rowMasks[7])
		promotionCapturesRight = rightTemp & (rowMasks[7])
		promotionAdvances = ((brd.pieces[c][PAWN] << 8) & rowMasks[7]) & (^occ)

	} else { // black to move
		leftTemp = ((brd.pieces[c][PAWN] & (^columnMasks[0])) >> 9) & enemy
		leftAttacks = leftTemp & (^rowMasks[0])
		promotionCapturesLeft = leftTemp & (rowMasks[0])

		rightTemp = ((brd.pieces[c][PAWN] & (^columnMasks[7])) >> 7) & enemy
		rightAttacks = rightTemp & (^rowMasks[0])
		promotionCapturesRight = rightTemp & (rowMasks[0])
		promotionAdvances = ((brd.pieces[c][PAWN] >> 8) & rowMasks[0]) & (^occ)
	}

	// promotion captures
	for ; promotionCapturesLeft > 0; promotionCapturesLeft.Clear(to) {
		to = furthestForward(c, promotionCapturesLeft)
		from = to + pawnFromOffsets[c][OFF_LEFT]
		getPromotionCaptures(brd, winning, from, to, brd.squares[to])
	}

	for ; promotionCapturesRight > 0; promotionCapturesRight.Clear(to) {
		to = furthestForward(c, promotionCapturesRight)
		from = to + pawnFromOffsets[c][OFF_RIGHT]
		getPromotionCaptures(brd, winning, from, to, brd.squares[to])
	}

	// promotion advances
	for ; promotionAdvances > 0; promotionAdvances.Clear(to) {
		to = furthestForward(c, promotionAdvances)
		from = to + pawnFromOffsets[c][OFF_SINGLE]
		getPromotionAdvances(brd, winning, losing, from, to)
	}

	// regular pawn attacks
	for ; leftAttacks > 0; leftAttacks.Clear(to) {
		to = furthestForward(c, leftAttacks)
		from = to + pawnFromOffsets[c][OFF_LEFT]
		m = NewCapture(from, to, PAWN, brd.squares[to])
		winning.Push(SortItem{mvvLva(brd.squares[to], PAWN), m})
	}
	for ; rightAttacks > 0; rightAttacks.Clear(to) {
		to = furthestForward(c, rightAttacks)
		from = to + pawnFromOffsets[c][OFF_RIGHT]
		m = NewCapture(from, to, PAWN, brd.squares[to])
		winning.Push(SortItem{mvvLva(brd.squares[to], PAWN), m})
	}
	// en-passant captures
	if brd.enpTarget != SQ_INVALID {
		enpTarget := brd.enpTarget
		for f := (brd.pieces[c][PAWN] & pawnSideMasks[enpTarget]); f > 0; f.Clear(from) {
			from = furthestForward(c, f)
			if c == WHITE {
				to = int(enpTarget) + 8
			} else {
				to = int(enpTarget) - 8
			}
			m = NewCapture(from, to, PAWN, PAWN)
			winning.Push(SortItem{mvvLva(PAWN, PAWN), m})
		}
	}
	// Knights
	for f := brd.pieces[c][KNIGHT]; f > 0; f.Clear(from) {
		from = furthestForward(c, f)
		for t := (knightMasks[from] & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewCapture(from, to, KNIGHT, brd.squares[to])
			if getSee(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvvLva(brd.squares[to], KNIGHT), m})
			} else {
				losing.Push(SortItem{mvvLva(brd.squares[to], KNIGHT), m})
			}
		}
	}
	// Bishops
	for f := brd.pieces[c][BISHOP]; f > 0; f.Clear(from) {
		from = furthestForward(c, f)
		for t := (bishopAttacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewCapture(from, to, BISHOP, brd.squares[to])
			if getSee(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvvLva(brd.squares[to], BISHOP), m})
			} else {
				losing.Push(SortItem{mvvLva(brd.squares[to], BISHOP), m})
			}
		}
	}
	// Rooks
	for f := brd.pieces[c][ROOK]; f > 0; f.Clear(from) {
		from = furthestForward(c, f)
		for t := (rookAttacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewCapture(from, to, ROOK, brd.squares[to])
			if getSee(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvvLva(brd.squares[to], ROOK), m})
			} else {
				losing.Push(SortItem{mvvLva(brd.squares[to], ROOK), m})
			}
		}
	}
	// Queens
	for f := brd.pieces[c][QUEEN]; f > 0; f.Clear(from) {
		from = furthestForward(c, f)
		for t := (queenAttacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewCapture(from, to, QUEEN, brd.squares[to])
			if getSee(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvvLva(brd.squares[to], QUEEN), m})
			} else {
				losing.Push(SortItem{mvvLva(brd.squares[to], QUEEN), m})
			}
		}
	}
	// King
	for f := brd.pieces[c][KING]; f > 0; f.Clear(from) {
		from = brd.KingSq(c)
		for t := (kingMasks[from] & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewCapture(from, to, KING, brd.squares[to])
			if getSee(brd, from, to, brd.squares[to]) >= 0 { // Cannot move into check
				winning.Push(SortItem{mvvLva(brd.squares[to], KING), m})
			}
		}
	}
}

func getWinningCaptures(brd *Board, htable *HistoryTable, winning *MoveList) {
	var from, to int
	var m Move

	c, e := brd.c, brd.Enemy()
	occ := brd.AllOccupied()
	enemy := brd.Placement(e)

	// Pawns
	var leftTemp, rightTemp, leftAttacks, rightAttacks BB
	var promotionCapturesLeft, promotionCapturesRight BB
	var promotionAdvances BB

	if c == WHITE { // white to move
		leftTemp = ((brd.pieces[c][PAWN] & (^columnMasks[0])) << 7) & enemy
		leftAttacks = leftTemp & (^rowMasks[7])
		promotionCapturesLeft = leftTemp & (rowMasks[7])

		rightTemp = ((brd.pieces[c][PAWN] & (^columnMasks[7])) << 9) & enemy
		rightAttacks = rightTemp & (^rowMasks[7])
		promotionCapturesRight = rightTemp & (rowMasks[7])

		promotionAdvances = ((brd.pieces[c][PAWN] << 8) & rowMasks[7]) & (^occ)
	} else { // black to move
		leftTemp = ((brd.pieces[c][PAWN] & (^columnMasks[0])) >> 9) & enemy
		leftAttacks = leftTemp & (^rowMasks[0])
		promotionCapturesLeft = leftTemp & (rowMasks[0])

		rightTemp = ((brd.pieces[c][PAWN] & (^columnMasks[7])) >> 7) & enemy
		rightAttacks = rightTemp & (^rowMasks[0])
		promotionCapturesRight = rightTemp & (rowMasks[0])

		promotionAdvances = ((brd.pieces[c][PAWN] >> 8) & rowMasks[0]) & (^occ)
	}

	// promotion captures
	for ; promotionCapturesLeft > 0; promotionCapturesLeft.Clear(to) {
		to = furthestForward(c, promotionCapturesLeft)
		from = to + pawnFromOffsets[c][OFF_LEFT]
		getPromotionCaptures(brd, winning, from, to, brd.squares[to])
	}

	for ; promotionCapturesRight > 0; promotionCapturesRight.Clear(to) {
		to = furthestForward(c, promotionCapturesRight)
		from = to + pawnFromOffsets[c][OFF_RIGHT]
		getPromotionCaptures(brd, winning, from, to, brd.squares[to])
	}

	// promotion advances
	for ; promotionAdvances > 0; promotionAdvances.Clear(to) {
		to = furthestForward(c, promotionAdvances)
		from = to + pawnFromOffsets[c][OFF_SINGLE]
		getPromotionAdvances(brd, winning, winning, from, to)
	}

	// regular pawn attacks
	for ; leftAttacks > 0; leftAttacks.Clear(to) {
		to = furthestForward(c, leftAttacks)
		from = to + pawnFromOffsets[c][OFF_LEFT]
		m = NewCapture(from, to, PAWN, brd.squares[to])
		winning.Push(SortItem{mvvLva(brd.squares[to], PAWN), m})
	}
	for ; rightAttacks > 0; rightAttacks.Clear(to) {
		to = furthestForward(c, rightAttacks)
		from = to + pawnFromOffsets[c][OFF_RIGHT]
		m = NewCapture(from, to, PAWN, brd.squares[to])
		winning.Push(SortItem{mvvLva(brd.squares[to], PAWN), m})
	}
	// en-passant captures
	if brd.enpTarget != SQ_INVALID {
		enpTarget := brd.enpTarget
		for f := (brd.pieces[c][PAWN] & pawnSideMasks[enpTarget]); f > 0; f.Clear(from) {
			from = furthestForward(c, f)
			if c == WHITE {
				to = int(enpTarget) + 8
			} else {
				to = int(enpTarget) - 8
			}
			m = NewCapture(from, to, PAWN, PAWN)
			winning.Push(SortItem{mvvLva(PAWN, PAWN), m})
		}
	}
	// Knights
	for f := brd.pieces[c][KNIGHT]; f > 0; f.Clear(from) {
		from = furthestForward(c, f)
		for t := (knightMasks[from] & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewCapture(from, to, KNIGHT, brd.squares[to])
			if getSee(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvvLva(brd.squares[to], KNIGHT), m})
			}
		}
	}
	// Bishops
	for f := brd.pieces[c][BISHOP]; f > 0; f.Clear(from) {
		from = furthestForward(c, f)
		for t := (bishopAttacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewCapture(from, to, BISHOP, brd.squares[to])
			if getSee(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvvLva(brd.squares[to], BISHOP), m})
			}
		}
	}
	// Rooks
	for f := brd.pieces[c][ROOK]; f > 0; f.Clear(from) {
		from = furthestForward(c, f)
		for t := (rookAttacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewCapture(from, to, ROOK, brd.squares[to])
			if getSee(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvvLva(brd.squares[to], ROOK), m})
			}
		}
	}
	// Queens
	for f := brd.pieces[c][QUEEN]; f > 0; f.Clear(from) {
		from = furthestForward(c, f)
		for t := (queenAttacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewCapture(from, to, QUEEN, brd.squares[to])
			if getSee(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvvLva(brd.squares[to], QUEEN), m})
			}
		}
	}
	// King
	for f := brd.pieces[c][KING]; f > 0; f.Clear(from) {
		from = brd.KingSq(c)
		for t := (kingMasks[from] & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewCapture(from, to, KING, brd.squares[to])
			if getSee(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvvLva(brd.squares[to], KING), m})
			}
		}
	}
}

func getEvasions(brd *Board, htable *HistoryTable, winning, losing, remainingMoves *MoveList) {
	c, e := brd.c, brd.Enemy()

	var defenseMap BB
	var from, to, threatSq1, threatSq2 int

	threatDir1, threatDir2 := DIR_INVALID, DIR_INVALID
	occ := brd.AllOccupied()
	empty := ^occ
	enemy := brd.Placement(e)

	kingSq := brd.KingSq(c)
	threats := colorAttackMap(brd, occ, kingSq, e, c) // find any enemy pieces that attack the king.
	threatCount := popCount(threats)

	// Get direction of the attacker(s) and any intervening squares between the attacker and the king.
	if threatCount == 1 {
		threatSq1 = lsb(threats)
		if brd.TypeAt(threatSq1) != PAWN {
			threatDir1 = directions[threatSq1][kingSq]
		}
		defenseMap |= (intervening[threatSq1][kingSq] | threats)
	} else {
		threatSq1 = lsb(threats)
		if brd.TypeAt(threatSq1) != PAWN {
			threatDir1 = directions[threatSq1][kingSq]
		}
		threatSq2 = msb(threats)
		if brd.TypeAt(threatSq2) != PAWN {
			threatDir2 = directions[threatSq2][kingSq]
		}
	}

	var m Move
	if threatCount == 1 { // Attempt to capture or block the attack with any piece if there's only one attacker.
		// Pawns
		var singleAdvances, doubleAdvances, leftTemp, rightTemp, leftAttacks, rightAttacks BB
		var promotionCapturesLeft, promotionCapturesRight BB
		var promotionAdvances BB

		if c > 0 { // white to move
			singleAdvances = (brd.pieces[WHITE][PAWN] << 8) & empty & (^rowMasks[7])
			doubleAdvances = ((singleAdvances & rowMasks[2]) << 8) & empty & defenseMap
			promotionAdvances = ((brd.pieces[c][PAWN] << 8) & rowMasks[7]) & empty & defenseMap

			leftTemp = ((brd.pieces[WHITE][PAWN] & (^columnMasks[0])) << 7) & enemy & defenseMap
			leftAttacks = leftTemp & (^rowMasks[7])
			promotionCapturesLeft = leftTemp & (rowMasks[7])

			rightTemp = ((brd.pieces[c][PAWN] & (^columnMasks[7])) << 9) & enemy & defenseMap
			rightAttacks = rightTemp & (^rowMasks[7])
			promotionCapturesRight = rightTemp & (rowMasks[7])
		} else { // black to move
			singleAdvances = (brd.pieces[BLACK][PAWN] >> 8) & empty & (^rowMasks[0])
			doubleAdvances = ((singleAdvances & rowMasks[5]) >> 8) & empty & defenseMap
			promotionAdvances = ((brd.pieces[BLACK][PAWN] >> 8) & rowMasks[0]) & empty & defenseMap

			leftTemp = ((brd.pieces[BLACK][PAWN] & (^columnMasks[0])) >> 9) & enemy & defenseMap
			leftAttacks = leftTemp & (^rowMasks[0])
			promotionCapturesLeft = leftTemp & (rowMasks[0])

			rightTemp = ((brd.pieces[BLACK][PAWN] & (^columnMasks[7])) >> 7) & enemy & defenseMap
			rightAttacks = rightTemp & (^rowMasks[0])
			promotionCapturesRight = rightTemp & (rowMasks[0])
		}
		singleAdvances &= defenseMap

		// promotion captures
		for ; promotionCapturesLeft > 0; promotionCapturesLeft.Clear(to) {
			to = furthestForward(c, promotionCapturesLeft)
			from = to + pawnFromOffsets[c][OFF_LEFT]
			if pinnedCanMove(brd, from, to, c, e) {
				getPromotionCaptures(brd, winning, from, to, brd.squares[to])
			}
		}
		for ; promotionCapturesRight > 0; promotionCapturesRight.Clear(to) {
			to = furthestForward(c, promotionCapturesRight)
			from = to + pawnFromOffsets[c][OFF_RIGHT]
			if pinnedCanMove(brd, from, to, c, e) {
				getPromotionCaptures(brd, winning, from, to, brd.squares[to])
			}
		}
		// promotion advances
		for ; promotionAdvances > 0; promotionAdvances.Clear(to) {
			to = furthestForward(c, promotionAdvances)
			from = to + pawnFromOffsets[c][OFF_SINGLE]
			if pinnedCanMove(brd, from, to, c, e) {
				getPromotionAdvances(brd, winning, remainingMoves, from, to)
			}
		}
		// regular pawn attacks
		for ; leftAttacks > 0; leftAttacks.Clear(to) {
			to = furthestForward(c, leftAttacks)
			from = to + pawnFromOffsets[c][OFF_LEFT]
			if pinnedCanMove(brd, from, to, c, e) {
				m = NewCapture(from, to, PAWN, brd.squares[to])
				winning.Push(SortItem{mvvLva(brd.squares[to], PAWN), m})
			}
		}
		for ; rightAttacks > 0; rightAttacks.Clear(to) {
			to = furthestForward(c, rightAttacks)
			from = to + pawnFromOffsets[c][OFF_RIGHT]
			if pinnedCanMove(brd, from, to, c, e) {
				m = NewCapture(from, to, PAWN, brd.squares[to])
				winning.Push(SortItem{mvvLva(brd.squares[to], PAWN), m})
			}
		}
		// en-passant captures
		if brd.enpTarget != SQ_INVALID {
			enpTarget := brd.enpTarget
			for f := (brd.pieces[c][PAWN] & pawnSideMasks[enpTarget]); f > 0; f.Clear(from) {
				from = furthestForward(c, f)
				if c == WHITE {
					to = int(enpTarget) + 8
				} else {
					to = int(enpTarget) - 8
				}
				// In addition to making sure this capture will get the king out of check and that
				// the piece is not pinned, verify that removing the enemy pawn does not leave the
				// king in check.
				if (sqMaskOn[to]&defenseMap) > 0 && pinnedCanMove(brd, from, to, c, e) &&
					isPinned(brd, int(enpTarget), c, e)&sqMaskOn[to] > 0 {

					m = NewCapture(from, to, PAWN, PAWN)
					winning.Push(SortItem{mvvLva(PAWN, PAWN), m})
				}
			}
		}
		// double advances
		for ; doubleAdvances > 0; doubleAdvances.Clear(to) {
			to = furthestForward(c, doubleAdvances)
			from = to + pawnFromOffsets[c][OFF_DOUBLE]
			if pinnedCanMove(brd, from, to, c, e) {
				m = NewRegularMove(from, to, PAWN)
				remainingMoves.Push(SortItem{htable.Probe(PAWN, c, to), m})
			}
		}
		// single advances
		for ; singleAdvances > 0; singleAdvances.Clear(to) {
			to = furthestForward(c, singleAdvances)
			from = to + pawnFromOffsets[c][OFF_SINGLE]
			if pinnedCanMove(brd, from, to, c, e) {
				m = NewRegularMove(from, to, PAWN)
				remainingMoves.Push(SortItem{htable.Probe(PAWN, c, to), m})
			}
		}
		// Knights
		for f := brd.pieces[c][KNIGHT]; f > 0; f.Clear(from) {
			from = furthestForward(c, f) // Locate each knight for the side to move.
			// Knights cannot move if pinned by a sliding piece, since they can't move along the ray between
			// the threat piece and their own king.
			if isPinned(brd, from, c, e) == BB(ANY_SQUARE_MASK) {
				for t := (knightMasks[from] & defenseMap); t > 0; t.Clear(to) { // generate to squares
					to = furthestForward(c, t)
					if sqMaskOn[to]&enemy > 0 {
						m = NewCapture(from, to, KNIGHT, brd.squares[to])
						if getSee(brd, from, to, brd.squares[to]) >= 0 {
							winning.Push(SortItem{mvvLva(brd.squares[to], KNIGHT), m})
						} else {
							losing.Push(SortItem{mvvLva(brd.squares[to], KNIGHT), m})
						}
					} else {
						m = NewRegularMove(from, to, KNIGHT)
						remainingMoves.Push(SortItem{htable.Probe(KNIGHT, c, to), m})
					}
				}
			}
		}
		// Bishops
		for f := brd.pieces[c][BISHOP]; f > 0; f.Clear(from) {
			from = furthestForward(c, f)
			for t := (bishopAttacks(occ, from) & defenseMap); t > 0; t.Clear(to) { // generate to squares
				to = furthestForward(c, t)
				if pinnedCanMove(brd, from, to, c, e) {
					if sqMaskOn[to]&enemy > 0 {
						m = NewCapture(from, to, BISHOP, brd.squares[to])
						if getSee(brd, from, to, brd.squares[to]) >= 0 {
							winning.Push(SortItem{mvvLva(brd.squares[to], BISHOP), m})
						} else {
							losing.Push(SortItem{mvvLva(brd.squares[to], BISHOP), m})
						}
					} else {
						m = NewRegularMove(from, to, BISHOP)
						remainingMoves.Push(SortItem{htable.Probe(BISHOP, c, to), m})
					}
				}
			}
		}
		// Rooks
		for f := brd.pieces[c][ROOK]; f > 0; f.Clear(from) {
			from = furthestForward(c, f)
			for t := (rookAttacks(occ, from) & defenseMap); t > 0; t.Clear(to) { // generate to squares
				to = furthestForward(c, t)
				if pinnedCanMove(brd, from, to, c, e) {
					if sqMaskOn[to]&enemy > 0 {
						m = NewCapture(from, to, ROOK, brd.squares[to])
						if getSee(brd, from, to, brd.squares[to]) >= 0 {
							winning.Push(SortItem{mvvLva(brd.squares[to], ROOK), m})
						} else {
							losing.Push(SortItem{mvvLva(brd.squares[to], ROOK), m})
						}
					} else {
						m = NewRegularMove(from, to, ROOK)
						remainingMoves.Push(SortItem{htable.Probe(ROOK, c, to), m})
					}
				}
			}
		}
		// Queens
		for f := brd.pieces[c][QUEEN]; f > 0; f.Clear(from) {
			from = furthestForward(c, f)
			for t := (queenAttacks(occ, from) & defenseMap); t > 0; t.Clear(to) { // generate to squares
				to = furthestForward(c, t)
				if pinnedCanMove(brd, from, to, c, e) {
					if sqMaskOn[to]&enemy > 0 {
						m = NewCapture(from, to, QUEEN, brd.squares[to])
						if getSee(brd, from, to, brd.squares[to]) >= 0 {
							winning.Push(SortItem{mvvLva(brd.squares[to], QUEEN), m})
						} else {
							losing.Push(SortItem{mvvLva(brd.squares[to], QUEEN), m})
						}
					} else {
						m = NewRegularMove(from, to, QUEEN)
						remainingMoves.Push(SortItem{htable.Probe(QUEEN, c, to), m})
					}
				}
			}
		}
	}
	// If there's more than one attacking piece, the only way out is to move the king.
	// King captures
	for t := (kingMasks[kingSq] & enemy); t > 0; t.Clear(to) { // generate to squares
		to = furthestForward(c, t)
		if !isAttackedBy(brd, occ, to, e, c) && threatDir1 != directions[kingSq][to] &&
			threatDir2 != directions[kingSq][to] {
			m = NewCapture(kingSq, to, KING, brd.squares[to])
			winning.Push(SortItem{mvvLva(brd.squares[to], KING), m})
		}
	}
	// King moves
	for t := (kingMasks[kingSq] & empty); t > 0; t.Clear(to) { // generate to squares
		to = furthestForward(c, t)
		if !isAttackedBy(brd, occ, to, e, c) && threatDir1 != directions[kingSq][to] &&
			threatDir2 != directions[kingSq][to] {
			m = NewRegularMove(kingSq, to, KING)
			remainingMoves.Push(SortItem{htable.Probe(KING, c, to), m})
		}
	}
}

func getChecks(brd *Board, htable *HistoryTable, remainingMoves *MoveList) {
	c, e := brd.c, brd.Enemy()
	kingSq := brd.KingSq(e)
	var f, t, singleAdvances, target, queenTarget BB
	var from, to int
	var m Move
	occ := brd.AllOccupied()
	empty := ^occ
	// Pawn direct checks
	if c > 0 { // white to move
		singleAdvances = (brd.pieces[WHITE][PAWN] << 8) & empty
	} else { // black to move
		singleAdvances = (brd.pieces[BLACK][PAWN] >> 8) & empty
	}
	target = pawnAttackMasks[e][kingSq]
	for t = singleAdvances & target; t > 0; t.Clear(to) {
		to = furthestForward(c, t)
		from = to + pawnFromOffsets[c][OFF_SINGLE]
		if getSee(brd, from, to, EMPTY) >= 0 { // make sure the checking piece won't be immediately recaptured
			m = NewRegularMove(from, to, PAWN)
			remainingMoves.Push(SortItem{htable.Probe(PAWN, c, to), m})
		}
	}
	// Knight direct checks
	target = knightMasks[kingSq] & empty
	for f = brd.pieces[c][KNIGHT]; f > 0; f.Clear(from) {
		from = furthestForward(c, f) // Locate each knight for the side to move.
		for t = (knightMasks[from] & target); t > 0; t.Clear(to) {
			to = furthestForward(c, t)
			if getSee(brd, from, to, EMPTY) >= 0 {
				m = NewRegularMove(from, to, KNIGHT)
				remainingMoves.Push(SortItem{htable.Probe(KNIGHT, c, to), m})
			}
		}
	}
	// Bishop direct checks
	target = bishopAttacks(occ, kingSq) & empty
	queenTarget = target
	for f = brd.pieces[c][BISHOP]; f > 0; f.Clear(from) {
		from = furthestForward(c, f)
		for t = (bishopAttacks(occ, from) & target); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			if getSee(brd, from, to, EMPTY) >= 0 {
				m = NewRegularMove(from, to, BISHOP)
				remainingMoves.Push(SortItem{htable.Probe(BISHOP, c, to), m})
			}
		}
	}
	// Rook direct checks
	target = rookAttacks(occ, kingSq) & empty
	queenTarget |= target
	for f = brd.pieces[c][ROOK]; f > 0; f.Clear(from) {
		from = furthestForward(c, f)
		for t = (rookAttacks(occ, from) & target); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			if getSee(brd, from, to, EMPTY) >= 0 {
				m = NewRegularMove(from, to, ROOK)
				remainingMoves.Push(SortItem{htable.Probe(ROOK, c, to), m})
			}
		}
	}
	// Queen direct checks
	for f = brd.pieces[c][QUEEN]; f > 0; f.Clear(from) {
		from = furthestForward(c, f)
		for t = (queenAttacks(occ, from) & queenTarget); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			if getSee(brd, from, to, EMPTY) >= 0 {
				m = NewRegularMove(from, to, QUEEN)
				remainingMoves.Push(SortItem{htable.Probe(QUEEN, c, to), m})
			}
		}
	}

	// indirect (discovered) checks
	var rookBlockers, bishopBlockers BB

	rookBlockers = rookAttacks(occ, kingSq) & (brd.pieces[c][BISHOP] |
		brd.pieces[c][KNIGHT] | brd.pieces[c][PAWN])
	bishopBlockers = bishopAttacks(occ, kingSq) & (brd.pieces[c][ROOK] |
		brd.pieces[c][KNIGHT] | brd.pieces[c][PAWN])
	if rookBlockers > 0 {
		rookAttackers := rookAttacks(occ^rookBlockers, kingSq) & (brd.pieces[c][ROOK] | brd.pieces[c][QUEEN])
		for dir := NORTH; dir <= WEST; dir++ {
			if rayMasks[dir][kingSq]&rookAttackers == 0 {
				rookBlockers &= (^rayMasks[dir][kingSq])
			}
		}
	}
	if bishopBlockers > 0 {
		bishopAttackers := bishopAttacks(occ^bishopBlockers, kingSq) & (brd.pieces[c][BISHOP] | brd.pieces[c][QUEEN])
		for dir := NW; dir <= SW; dir++ {
			if rayMasks[dir][kingSq]&bishopAttackers == 0 {
				bishopBlockers &= (^rayMasks[dir][kingSq])
			}
		}
	}

	var unblockPath BB // blockers must move off the path of attack.

	// don't bother with double advances.
	for t = singleAdvances & (bishopBlockers | rookBlockers); t > 0; t.Clear(to) {
		to = furthestForward(c, t)
		from = to + pawnFromOffsets[c][OFF_SINGLE]
		m = NewRegularMove(from, to, PAWN)
		remainingMoves.Push(SortItem{htable.Probe(PAWN, c, to), m})
	}
	// Knights
	for f = brd.pieces[c][KNIGHT] & (bishopBlockers | rookBlockers); f > 0; f.Clear(from) {
		from = furthestForward(c, f) // Locate each knight for the side to move.
		for t = (knightMasks[from] & empty); t > 0; t.Clear(to) {
			to = furthestForward(c, t)
			m = NewRegularMove(from, to, KNIGHT)
			remainingMoves.Push(SortItem{htable.Probe(KNIGHT, c, to), m})
		}
	}
	// Bishops
	for f = brd.pieces[c][BISHOP] & rookBlockers; f > 0; f.Clear(from) {
		from = furthestForward(c, f)
		unblockPath = (^intervening[kingSq][from]) & empty
		for t = (bishopAttacks(occ, from) & unblockPath); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewRegularMove(from, to, BISHOP)
			remainingMoves.Push(SortItem{htable.Probe(BISHOP, c, to), m})
		}
	}
	// Rooks
	for f = brd.pieces[c][ROOK] & bishopBlockers; f > 0; f.Clear(from) {
		from = furthestForward(c, f)
		unblockPath = (^intervening[kingSq][from]) & empty
		for t = (rookAttacks(occ, from) & unblockPath); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewRegularMove(from, to, ROOK)
			remainingMoves.Push(SortItem{htable.Probe(ROOK, c, to), m})
		}
	}
	// Queens cannot give discovered check, since the enemy king would already be in check.

	// Kings
	for f := brd.pieces[c][KING] & (bishopBlockers | rookBlockers); f > 0; f.Clear(from) {
		from = brd.KingSq(c)
		unblockPath = (^intervening[kingSq][from]) & empty
		for t := (kingMasks[from] & unblockPath); t > 0; t.Clear(to) { // generate to squares
			to = furthestForward(c, t)
			m = NewRegularMove(from, to, KING)
			remainingMoves.Push(SortItem{htable.Probe(KING, c, to), m})
		}
	}

}

// // uncomment for movegen testing.

// func getPromotionAdvances(brd *Board, winning, losing *MoveList, from, to int) {
// 	var m Move
// 	var sort uint64
// 	for pc := Piece(QUEEN); pc >= KNIGHT; pc-- {
// 		m = NewMove(from, to, PAWN, EMPTY, pc)
// 		sort = sortPromotionAdvances(brd, from, to, pc)
// 		if sort >= SORT_WINNING_PROMOTION {
// 			winning.Push(SortItem{sort})
// 		} else {
// 			losing.Push(SortItem{sort})
// 		}
// 	}
// }
//
// func getPromotionCaptures(brd *Board, winning *MoveList, from, to int, capturedPiece Piece) {
// 	var m Move
// 	for pc := Piece(QUEEN); pc >= KNIGHT; pc-- {
// 		m = NewMove(from, to, PAWN, capturedPiece, pc)
// 		winning.Push(SortItem{sortPromotionCaptures(brd, from, to, capturedPiece, pc)})
// 	}
// }

func getPromotionAdvances(brd *Board, winning, losing *MoveList, from, to int) {
	var m Move
	var sort uint64
	m = NewMove(from, to, PAWN, EMPTY, QUEEN)
	sort = sortPromotionAdvances(brd, from, to, QUEEN)
	if sort >= SORT_WINNING_PROMOTION {
		winning.Push(SortItem{sort, m})
	} else {
		losing.Push(SortItem{sort, m})
	}
	m = NewMove(from, to, PAWN, EMPTY, KNIGHT)
	sort = sortPromotionAdvances(brd, from, to, KNIGHT)
	if sort >= SORT_WINNING_PROMOTION {
		winning.Push(SortItem{sort, m})
	} else {
		losing.Push(SortItem{sort, m})
	}
}

func getPromotionCaptures(brd *Board, winning *MoveList, from, to int, capturedPiece Piece) {
	winning.Push(SortItem{sortPromotionCaptures(brd, from, to, capturedPiece, QUEEN),
		NewMove(from, to, PAWN, capturedPiece, QUEEN)})
	winning.Push(SortItem{sortPromotionCaptures(brd, from, to, capturedPiece, KNIGHT),
		NewMove(from, to, PAWN, capturedPiece, KNIGHT)})
}
