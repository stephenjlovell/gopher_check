//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

const (
	C_WQ = 8 // White castle queen side
	C_WK = 4 // White castle king side
	C_BQ = 2 // Black castle queen side
	C_BK = 1 // Black castle king side
)

func makeMove(brd *Board, move Move) {
	from := move.From()
	to := move.To()
	updateCastleRights(brd, from, to)

	capturedPiece := move.CapturedPiece()
	c := brd.c
	piece := move.Piece()
	enpTarget := brd.enpTarget
	brd.hashKey ^= EnpZobrist(enpTarget) // XOR out old en passant target.
	brd.enpTarget = SQ_INVALID

	switch piece {
	case PAWN:
		brd.halfmoveClock = 0 // All pawn moves are irreversible.
		brd.pawnHashKey ^= PawnZobrist(from, c)
		switch capturedPiece {
		case NO_PIECE:
			if Abs(to-from) == 16 { // handle en passant advances
				brd.enpTarget = uint8(to)
				brd.hashKey ^= EnpZobrist(uint8(to)) // XOR in new en passant target
			}
		case PAWN: // Destination square will be empty if en passant capture
			if enpTarget != SQ_INVALID && brd.TypeAt(to) == NO_PIECE {
				brd.pawnHashKey ^= PawnZobrist(int(enpTarget), brd.Enemy())
				removePiece(brd, PAWN, int(enpTarget), brd.Enemy())
				brd.squares[enpTarget] = NO_PIECE
			} else {
				brd.pawnHashKey ^= PawnZobrist(to, brd.Enemy())
				removePiece(brd, PAWN, to, brd.Enemy())
			}
		default: // any non-pawn piece is captured
			removePiece(brd, capturedPiece, to, brd.Enemy())
		}
		promotedPiece := move.PromotedTo()
		if promotedPiece != NO_PIECE {
			removePiece(brd, PAWN, from, c)
			brd.squares[from] = NO_PIECE
			addPiece(brd, promotedPiece, to, c)
		} else {
			brd.pawnHashKey ^= PawnZobrist(to, c)
			relocatePiece(brd, PAWN, from, to, c)
		}

	case KING:
		switch capturedPiece {
		case NO_PIECE:
			brd.halfmoveClock += 1
			if Abs(to-from) == 2 { // king is castling.
				brd.halfmoveClock = 0
				if c == WHITE {
					if to == G1 {
						relocatePiece(brd, ROOK, H1, F1, c)
					} else {
						relocatePiece(brd, ROOK, A1, D1, c)
					}
				} else {
					if to == G8 {
						relocatePiece(brd, ROOK, H8, F8, c)
					} else {
						relocatePiece(brd, ROOK, A8, D8, c)
					}
				}
			}
		case PAWN:
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.pawnHashKey ^= PawnZobrist(to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
		default:
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
		}
		relocateKing(brd, KING, capturedPiece, from, to, c)

	case ROOK:
		switch capturedPiece {
		case ROOK:
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
		case NO_PIECE:
			brd.halfmoveClock += 1
		case PAWN:
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
			brd.pawnHashKey ^= PawnZobrist(to, brd.Enemy())
		default:
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
		}
		relocatePiece(brd, ROOK, from, to, c)

	default:
		switch capturedPiece {
		case ROOK:
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
		case NO_PIECE:
			brd.halfmoveClock += 1
		case PAWN:
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
			brd.pawnHashKey ^= PawnZobrist(to, brd.Enemy())
		default:
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
		}
		relocatePiece(brd, piece, from, to, c)
	}

	brd.c ^= 1 // flip the current side to move.
	brd.hashKey ^= sideKey
}

// Castle flag, enp target, hash key, pawn hash key, and halfmove clock are all restored during search
func unmakeMove(brd *Board, move Move, memento *BoardMemento) {
	brd.c ^= 1 // flip the current side to move.

	c := brd.c
	piece := move.Piece()
	from := move.From()
	to := move.To()
	capturedPiece := move.CapturedPiece()
	enpTarget := memento.enpTarget

	switch piece {
	case PAWN:
		if move.PromotedTo() != NO_PIECE {
			unmakeRemovePiece(brd, move.PromotedTo(), to, c)
			brd.squares[to] = capturedPiece
			unmakeAddPiece(brd, piece, from, c)
		} else {
			unmakeRelocatePiece(brd, piece, to, from, c)
		}
		switch capturedPiece {
		case PAWN:
			if enpTarget != SQ_INVALID {
				if c == WHITE {
					if to == int(enpTarget)+8 {
						unmakeAddPiece(brd, PAWN, int(enpTarget), brd.Enemy())
					} else {
						unmakeAddPiece(brd, PAWN, to, brd.Enemy())
					}
				} else {
					if to == int(enpTarget)-8 {
						unmakeAddPiece(brd, PAWN, int(enpTarget), brd.Enemy())
					} else {
						unmakeAddPiece(brd, PAWN, to, brd.Enemy())
					}
				}
			} else {
				unmakeAddPiece(brd, PAWN, to, brd.Enemy())
			}
		case NO_PIECE:
		default: // any non-pawn piece was captured
			unmakeAddPiece(brd, capturedPiece, to, brd.Enemy())
		}

	case KING:
		unmakeRelocateKing(brd, piece, capturedPiece, to, from, c)
		if capturedPiece != NO_PIECE {
			unmakeAddPiece(brd, capturedPiece, to, brd.Enemy())
		} else if Abs(to-from) == 2 { // king castled.
			if c == WHITE {
				if to == G1 {
					unmakeRelocatePiece(brd, ROOK, F1, H1, WHITE)
				} else {
					unmakeRelocatePiece(brd, ROOK, D1, A1, WHITE)
				}
			} else {
				if to == G8 {
					unmakeRelocatePiece(brd, ROOK, F8, H8, BLACK)
				} else {
					unmakeRelocatePiece(brd, ROOK, D8, A8, BLACK)
				}
			}
		}

	default:
		unmakeRelocatePiece(brd, piece, to, from, c)
		if capturedPiece != NO_PIECE {
			unmakeAddPiece(brd, capturedPiece, to, brd.Enemy())
		}
	}

	brd.hashKey, brd.pawnHashKey = memento.hashKey, memento.pawnHashKey
	brd.castle, brd.enpTarget = memento.castle, memento.enpTarget
	brd.halfmoveClock = memento.halfmoveClock
}

// Update castling rights whenever a piece moves from or to a square associated with the
// current castling rights.
func updateCastleRights(brd *Board, from, to int) {
	if castle := brd.castle; castle > 0 && (sqMaskOn[from]|sqMaskOn[to])&castleMasks[castle] > 0 {
		updateCasleRightsForSq(brd, from)
		updateCasleRightsForSq(brd, to)
		brd.hashKey ^= CastleZobrist(castle)
		brd.hashKey ^= CastleZobrist(brd.castle)
	}
}

func updateCasleRightsForSq(brd *Board, sq int) {
	switch sq { // if brd.castle remains unchanged, hash key will be unchanged.
	case A1:
		brd.castle &= (^uint8(C_WQ))
	case E1: // white king starting position
		brd.castle &= (^uint8(C_WK | C_WQ))
	case H1:
		brd.castle &= (^uint8(C_WK))
	case A8:
		brd.castle &= (^uint8(C_BQ))
	case E8: // black king starting position
		brd.castle &= (^uint8(C_BK | C_BQ))
	case H8:
		brd.castle &= (^uint8(C_BK))
	default:
	}
}

func removePiece(brd *Board, removedPiece Piece, sq int, e uint8) {
	unmakeRemovePiece(brd, removedPiece, sq, e)
	brd.hashKey ^= Zobrist(removedPiece, sq, e) // XOR out the captured piece
}

func unmakeRemovePiece(brd *Board, removedPiece Piece, sq int, e uint8) {
	brd.pieces[e][removedPiece].Clear(sq)
	brd.occupied[e].Clear(sq)
	brd.material[e] -= int16(removedPiece.Value() + mainPst[e][removedPiece][sq])
	brd.endgameCounter -= endgameCountValues[removedPiece]
}

func addPiece(brd *Board, addedPiece Piece, sq int, c uint8) {
	unmakeAddPiece(brd, addedPiece, sq, c)
	brd.hashKey ^= Zobrist(addedPiece, sq, c) // XOR in key for added_piece
}

func unmakeAddPiece(brd *Board, addedPiece Piece, sq int, c uint8) {
	brd.pieces[c][addedPiece].Add(sq)
	brd.squares[sq] = addedPiece
	brd.occupied[c].Add(sq)
	brd.material[c] += int16(addedPiece.Value() + mainPst[c][addedPiece][sq])
	brd.endgameCounter += endgameCountValues[addedPiece]
}

func relocatePiece(brd *Board, piece Piece, from, to int, c uint8) {
	unmakeRelocatePiece(brd, piece, from, to, c)
	// XOR out the key for piece at from, and XOR in the key for piece at to.
	brd.hashKey ^= (Zobrist(piece, from, c) ^ Zobrist(piece, to, c))
}

func unmakeRelocatePiece(brd *Board, piece Piece, from, to int, c uint8) {
	fromTo := (sqMaskOn[from] | sqMaskOn[to])
	brd.pieces[c][piece] ^= fromTo
	brd.occupied[c] ^= fromTo
	brd.squares[from] = NO_PIECE
	brd.squares[to] = piece
	brd.material[c] += int16(mainPst[c][piece][to] - mainPst[c][piece][from])
}

func relocateKing(brd *Board, piece, capturedPiece Piece, from, to int, c uint8) {
	unmakeRelocateKing(brd, piece, capturedPiece, from, to, c)
	// XOR out the key for piece at from, and XOR in the key for piece at to.
	brd.hashKey ^= (Zobrist(piece, from, c) ^ Zobrist(piece, to, c))
}

func unmakeRelocateKing(brd *Board, piece, capturedPiece Piece, from, to int, c uint8) {
	fromTo := (sqMaskOn[from] | sqMaskOn[to])
	brd.pieces[c][piece] ^= fromTo
	brd.occupied[c] ^= fromTo
	brd.squares[from] = NO_PIECE
	brd.squares[to] = piece
	brd.kingSq[c] = uint8(to)
	// Since king PST is weighted by endgame phase, it cannot be easily incrementally updated.
}
