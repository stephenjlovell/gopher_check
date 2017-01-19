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
	c := brd.c
	piece := move.Piece()
	from := move.From()
	to := move.To()
	capturedPiece := move.CapturedPiece()

	enpTarget := brd.enpTarget
	brd.hashKey ^= enpZobrist(enpTarget) // XOR out old en passant target.
	brd.enpTarget = SQ_INVALID

	// assert(capturedPiece != KING, "Illegal king capture detected during makeMove()")

	switch piece {
	case PAWN:
		brd.halfmoveClock = 0 // All pawn moves are irreversible.
		brd.pawnHashKey ^= pawnZobrist(from, c)
		switch capturedPiece {
		case EMPTY:
			if abs(to-from) == 16 { // handle en passant advances
				brd.enpTarget = uint8(to)
				brd.hashKey ^= enpZobrist(uint8(to)) // XOR in new en passant target
			}
		case PAWN: // Destination square will be empty if en passant capture
			if enpTarget != SQ_INVALID && brd.TypeAt(to) == EMPTY {
				// fmt.Println(move.ToString())
				// brd.Print()

				brd.pawnHashKey ^= pawnZobrist(int(enpTarget), brd.Enemy())
				removePiece(brd, PAWN, int(enpTarget), brd.Enemy())
				brd.squares[enpTarget] = EMPTY
			} else {
				brd.pawnHashKey ^= pawnZobrist(to, brd.Enemy())
				removePiece(brd, PAWN, to, brd.Enemy())
			}
		case ROOK:
			if brd.castle > 0 {
				updateCastleRights(brd, to)
			}
			removePiece(brd, capturedPiece, to, brd.Enemy())
		default: // any non-pawn piece is captured
			removePiece(brd, capturedPiece, to, brd.Enemy())
		}
		promotedPiece := move.PromotedTo()
		if promotedPiece != EMPTY {
			removePiece(brd, PAWN, from, c)
			brd.squares[from] = EMPTY
			addPiece(brd, promotedPiece, to, c)
		} else {
			brd.pawnHashKey ^= pawnZobrist(to, c)
			relocatePiece(brd, PAWN, from, to, c)
		}

	case KING:
		switch capturedPiece {
		case ROOK:
			if brd.castle > 0 {
				updateCastleRights(brd, from)
				updateCastleRights(brd, to) //
			}
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
		case EMPTY:
			brd.halfmoveClock += 1
			if brd.castle > 0 {
				updateCastleRights(brd, from)
				if abs(to-from) == 2 { // king is castling.
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
			}
		case PAWN:
			if brd.castle > 0 {
				updateCastleRights(brd, from)
			}
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.pawnHashKey ^= pawnZobrist(to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
		default:
			if brd.castle > 0 {
				updateCastleRights(brd, from)
			}
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
		}
		relocateKing(brd, KING, capturedPiece, from, to, c)

	case ROOK:
		switch capturedPiece {
		case ROOK:
			if brd.castle > 0 {
				updateCastleRights(brd, from)
				updateCastleRights(brd, to)
			}
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
		case EMPTY:
			if brd.castle > 0 {
				updateCastleRights(brd, from)
			}
			brd.halfmoveClock += 1
		case PAWN:
			if brd.castle > 0 {
				updateCastleRights(brd, from)
			}
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
			brd.pawnHashKey ^= pawnZobrist(to, brd.Enemy())
		default:
			if brd.castle > 0 {
				updateCastleRights(brd, from)
			}
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
		}
		relocatePiece(brd, ROOK, from, to, c)

	default:
		switch capturedPiece {
		case ROOK:
			if brd.castle > 0 {
				updateCastleRights(brd, to) //
			}
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
		case EMPTY:
			brd.halfmoveClock += 1
		case PAWN:
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
			brd.pawnHashKey ^= pawnZobrist(to, brd.Enemy())
		default:
			removePiece(brd, capturedPiece, to, brd.Enemy())
			brd.halfmoveClock = 0 // All capture moves are irreversible.
		}
		relocatePiece(brd, piece, from, to, c)
	}

	brd.c ^= 1 // flip the current side to move.
	brd.hashKey ^= sideKey64
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
		if move.PromotedTo() != EMPTY {
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
		case EMPTY:
		default: // any non-pawn piece was captured
			unmakeAddPiece(brd, capturedPiece, to, brd.Enemy())
		}

	case KING:
		unmakeRelocateKing(brd, piece, capturedPiece, to, from, c)
		if capturedPiece != EMPTY {
			unmakeAddPiece(brd, capturedPiece, to, brd.Enemy())
		} else if abs(to-from) == 2 { // king castled.
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
		if capturedPiece != EMPTY {
			unmakeAddPiece(brd, capturedPiece, to, brd.Enemy())
		}
	}

	brd.hashKey, brd.pawnHashKey = memento.hashKey, memento.pawnHashKey
	brd.castle, brd.enpTarget = memento.castle, memento.enpTarget
	brd.halfmoveClock = memento.halfmoveClock
}

// Whenever a king or rook moves off its initial square or is captured,
// update castle rights via the procedure associated with that square.
func updateCastleRights(brd *Board, sq int) {
	switch sq { // if brd.castle remains unchanged, hash key will be unchanged.
	case A1:
		brd.hashKey ^= castleZobrist(brd.castle)
		brd.castle &= (^uint8(C_WQ))
		brd.hashKey ^= castleZobrist(brd.castle)
	case E1: // white king starting position
		brd.hashKey ^= castleZobrist(brd.castle)
		brd.castle &= (^uint8(C_WK | C_WQ))
		brd.hashKey ^= castleZobrist(brd.castle)
	case H1:
		brd.hashKey ^= castleZobrist(brd.castle)
		brd.castle &= (^uint8(C_WK))
		brd.hashKey ^= castleZobrist(brd.castle)
	case A8:
		brd.hashKey ^= castleZobrist(brd.castle)
		brd.castle &= (^uint8(C_BQ))
		brd.hashKey ^= castleZobrist(brd.castle)
	case E8: // black king starting position
		brd.hashKey ^= castleZobrist(brd.castle)
		brd.castle &= (^uint8(C_BK | C_BQ))
		brd.hashKey ^= castleZobrist(brd.castle)
	case H8:
		brd.hashKey ^= castleZobrist(brd.castle)
		brd.castle &= (^uint8(C_BK))
		brd.hashKey ^= castleZobrist(brd.castle)
	}
}

func removePiece(brd *Board, removedPiece Piece, sq int, e uint8) {
	brd.pieces[e][removedPiece].Clear(sq)
	brd.occupied[e].Clear(sq)
	brd.material[e] -= int32(removedPiece.Value() + mainPst[e][removedPiece][sq])
	brd.endgameCounter -= endgameCountValues[removedPiece]
	brd.hashKey ^= zobrist(removedPiece, sq, e) // XOR out the captured piece
}
func unmakeRemovePiece(brd *Board, removedPiece Piece, sq int, e uint8) {
	brd.pieces[e][removedPiece].Clear(sq)
	brd.occupied[e].Clear(sq)
	brd.material[e] -= int32(removedPiece.Value() + mainPst[e][removedPiece][sq])
	brd.endgameCounter -= endgameCountValues[removedPiece]
}

func addPiece(brd *Board, addedPiece Piece, sq int, c uint8) {
	brd.pieces[c][addedPiece].Add(sq)
	brd.squares[sq] = addedPiece
	brd.occupied[c].Add(sq)
	brd.material[c] += int32(addedPiece.Value() + mainPst[c][addedPiece][sq])
	brd.endgameCounter += endgameCountValues[addedPiece]
	brd.hashKey ^= zobrist(addedPiece, sq, c) // XOR in key for addedPiece
}
func unmakeAddPiece(brd *Board, addedPiece Piece, sq int, c uint8) {
	brd.pieces[c][addedPiece].Add(sq)
	brd.squares[sq] = addedPiece
	brd.occupied[c].Add(sq)
	brd.material[c] += int32(addedPiece.Value() + mainPst[c][addedPiece][sq])
	brd.endgameCounter += endgameCountValues[addedPiece]
}

func relocatePiece(brd *Board, piece Piece, from, to int, c uint8) {
	fromTo := (sqMaskOn[from] | sqMaskOn[to])
	brd.pieces[c][piece] ^= fromTo
	brd.occupied[c] ^= fromTo
	brd.squares[from] = EMPTY
	brd.squares[to] = piece
	brd.material[c] += int32(mainPst[c][piece][to] - mainPst[c][piece][from])
	// XOR out the key for piece at from, and XOR in the key for piece at to.
	brd.hashKey ^= (zobrist(piece, from, c) ^ zobrist(piece, to, c))
}
func unmakeRelocatePiece(brd *Board, piece Piece, from, to int, c uint8) {
	fromTo := (sqMaskOn[from] | sqMaskOn[to])
	brd.pieces[c][piece] ^= fromTo
	brd.occupied[c] ^= fromTo
	brd.squares[from] = EMPTY
	brd.squares[to] = piece
	brd.material[c] += int32(mainPst[c][piece][to] - mainPst[c][piece][from])
}

func relocateKing(brd *Board, piece, capturedPiece Piece, from, to int, c uint8) {
	fromTo := (sqMaskOn[from] | sqMaskOn[to])
	brd.pieces[c][piece] ^= fromTo
	brd.occupied[c] ^= fromTo
	brd.squares[from] = EMPTY
	brd.squares[to] = piece
	// XOR out the key for piece at from, and XOR in the key for piece at to.
	brd.hashKey ^= (zobrist(piece, from, c) ^ zobrist(piece, to, c))
}
func unmakeRelocateKing(brd *Board, piece, capturedPiece Piece, from, to int, c uint8) {
	fromTo := (sqMaskOn[from] | sqMaskOn[to])
	brd.pieces[c][piece] ^= fromTo
	brd.occupied[c] ^= fromTo
	brd.squares[from] = EMPTY
	brd.squares[to] = piece
}
