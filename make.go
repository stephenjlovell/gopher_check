//-----------------------------------------------------------------------------------
// Copyright (c) 2014 Stephen J. Lovell
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//-----------------------------------------------------------------------------------

package main

var C_WQ uint8 = 8 // White castle queen side
var C_WK uint8 = 4 // White castle king side
var C_BQ uint8 = 2 // Black castle queen side
var C_BK uint8 = 1 // Black castle king side

func make_move(brd *Board, move Move) {
	c := brd.c
	piece := move.Piece()
	from := move.From()
	to := move.To()
	captured_piece := move.CapturedPiece()

	enp_target := brd.enp_target
	brd.enp_target = SQ_INVALID

	switch piece {
	case PAWN:
		brd.halfmove_clock = 0 // All pawn moves are irreversible.

		switch captured_piece {
		case PAWN: // Destination square will be empty if en passant capture
			if enp_target != SQ_INVALID && brd.TypeAt(to) == EMPTY {
				remove_piece(brd, PAWN, int(enp_target), brd.Enemy())
				brd.squares[enp_target] = EMPTY
			} else {
				remove_piece(brd, PAWN, to, brd.Enemy())
			}
		case EMPTY:
			if abs(to-from) == 16 { // handle en passant advances
				brd.enp_target = uint8(to)
			}
		default: // any non-pawn piece is captured
			if brd.castle > 0 {
				update_castle_rights(brd, to) //
			}
			remove_piece(brd, captured_piece, to, brd.Enemy())
		}

		promoted_piece := move.PromotedTo()
		if promoted_piece != EMPTY {
			remove_piece(brd, PAWN, from, c)
			brd.squares[from] = EMPTY
			add_piece(brd, promoted_piece, to, c)
		} else {
			relocate_piece(brd, PAWN, from, to, c)
		}

		// update the pawn hash key.

	case KING:
		if captured_piece != EMPTY {
			if brd.castle > 0 {
				update_castle_rights(brd, from)
				update_castle_rights(brd, to) //
			}
			remove_piece(brd, captured_piece, to, brd.Enemy())
			brd.halfmove_clock = 0 // All capture moves are irreversible.
		} else {
			brd.halfmove_clock += 1
			if brd.castle > 0 {
				update_castle_rights(brd, from)
				if abs(to-from) == 2 { // king is castling.
					brd.halfmove_clock = 0
					if c == WHITE {
						if to == G1 {
							relocate_piece(brd, ROOK, H1, F1, c)
						} else {
							relocate_piece(brd, ROOK, A1, D1, c)
						}
					} else {
						if to == G8 {
							relocate_piece(brd, ROOK, H8, F8, c)
						} else {
							relocate_piece(brd, ROOK, A8, D8, c)
						}
					}
				}
			}
		}
		relocate_piece(brd, KING, from, to, c)
	case ROOK:
		if captured_piece != EMPTY {
			if brd.castle > 0 {
				update_castle_rights(brd, from) // only need to check this for rooks and kings
				update_castle_rights(brd, to)   //
			}
			remove_piece(brd, captured_piece, to, brd.Enemy())
			brd.halfmove_clock = 0 // All capture moves are irreversible.
		} else {
			if brd.castle > 0 {
				update_castle_rights(brd, from) // only need to check this for rooks and kings
			}
			brd.halfmove_clock += 1
		}
		relocate_piece(brd, ROOK, from, to, c)
	default:
		if captured_piece != EMPTY {
			if brd.castle > 0 {
				update_castle_rights(brd, to) //
			}
			remove_piece(brd, captured_piece, to, brd.Enemy())
			brd.halfmove_clock = 0 // All capture moves are irreversible.
		} else {
			brd.halfmove_clock += 1
		}
		relocate_piece(brd, piece, from, to, c)
	}

	brd.c ^= 1 // flip the current side to move.
}

// Castle flag, enp target, hash key, pawn hash key, and halfmove clock are all restored during search
func unmake_move(brd *Board, move Move, enp_target uint8) {
	brd.c ^= 1 // flip the current side to move.

	c := brd.c
	piece := move.Piece()
	from := move.From()
	to := move.To()
	captured_piece := move.CapturedPiece()

	switch piece {
	case PAWN:

		if move.PromotedTo() != EMPTY {
			remove_piece(brd, move.PromotedTo(), to, c)
			brd.squares[to] = captured_piece
			add_piece(brd, piece, from, c)
		} else {
			relocate_piece(brd, piece, to, from, c)
		}

		switch captured_piece {
		case PAWN:
			if enp_target != SQ_INVALID {
				if c == WHITE {
					if to == int(enp_target)+8 {
						add_piece(brd, PAWN, int(enp_target), brd.Enemy())
					} else {
						add_piece(brd, PAWN, to, brd.Enemy())
					}
				} else {
					if to == int(enp_target)-8 {
						add_piece(brd, PAWN, int(enp_target), brd.Enemy())
					} else {
						add_piece(brd, PAWN, to, brd.Enemy())
					}
				}
			} else {
				add_piece(brd, PAWN, to, brd.Enemy())
			}
		case EMPTY:
		default: // any non-pawn piece was captured
			add_piece(brd, captured_piece, to, brd.Enemy())
		}

	case KING:
		relocate_piece(brd, piece, to, from, c)
		if captured_piece != EMPTY {
			add_piece(brd, captured_piece, to, brd.Enemy())
		} else {
			if abs(to-from) == 2 { // king castled.
				if c == WHITE {
					if to == G1 {
						relocate_piece(brd, ROOK, F1, H1, WHITE)
					} else {
						relocate_piece(brd, ROOK, D1, A1, WHITE)
					}
				} else {
					if to == G8 {
						relocate_piece(brd, ROOK, F8, H8, BLACK)
					} else {
						relocate_piece(brd, ROOK, D8, A8, BLACK)
					}
				}
			}
		}

	default:
		relocate_piece(brd, piece, to, from, c)
		if captured_piece != EMPTY {
			add_piece(brd, captured_piece, to, brd.Enemy())
		}
	}

}

// Whenever a king or rook moves off its initial square or is captured,
// update castle rights via the procedure associated with that square.
func update_castle_rights(brd *Board, sq int) {
	switch sq {
	case A1:
		brd.hash_key ^= castle_zobrist(brd.castle)
		brd.castle &= (^C_WQ)
		brd.hash_key ^= castle_zobrist(brd.castle)
	case E1: // white king starting position
		brd.hash_key ^= castle_zobrist(brd.castle)
		brd.castle &= (^(C_WK | C_WQ))
		brd.hash_key ^= castle_zobrist(brd.castle)
	case H1:
		brd.hash_key ^= castle_zobrist(brd.castle)
		brd.castle &= (^C_WK)
		brd.hash_key ^= castle_zobrist(brd.castle)
	case A8:
		brd.hash_key ^= castle_zobrist(brd.castle)
		brd.castle &= (^C_BQ)
		brd.hash_key ^= castle_zobrist(brd.castle)
	case E8: // black king starting position
		brd.hash_key ^= castle_zobrist(brd.castle)
		brd.castle &= (^(C_BK | C_BQ))
		brd.hash_key ^= castle_zobrist(brd.castle)
	case H8:
		brd.hash_key ^= castle_zobrist(brd.castle)
		brd.castle &= (^C_BK)
		brd.hash_key ^= castle_zobrist(brd.castle)
	}
}

func relocate_piece(brd *Board, piece Piece, from, to int, c uint8) {
	from_to := (sq_mask_on[from] | sq_mask_on[to])
	brd.pieces[c][piece] ^= from_to
	brd.occupied[c] ^= from_to
	brd.squares[from] = EMPTY
	brd.squares[to] = piece
	// XOR out the key for piece at from, and XOR in the key for piece at to.
	brd.hash_key ^= (zobrist(piece, from, c) ^ zobrist(piece, to, c))
}

// do not use for en-passant captures.
func remove_piece(brd *Board, removed_piece Piece, sq int, e uint8) {
	brd.pieces[e][removed_piece].Clear(sq)
	// brd.squares[sq] = EMPTY
	brd.occupied[e].Clear(sq)
	brd.material[e] -= int32(removed_piece.Value())
	brd.hash_key ^= zobrist(removed_piece, sq, e) // XOR out the captured piece
}

func add_piece(brd *Board, added_piece Piece, sq int, c uint8) {
	brd.pieces[c][added_piece].Add(sq)
	brd.squares[sq] = added_piece
	brd.occupied[c].Add(sq)
	brd.material[c] += int32(added_piece.Value())
	brd.hash_key ^= zobrist(added_piece, sq, c) // XOR in key for added_piece
}
