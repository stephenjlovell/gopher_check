//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------
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

import (
// "fmt"
)

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
	brd.hash_key ^= enp_zobrist(enp_target) // XOR out old en passant target.
	brd.enp_target = SQ_INVALID

	assert(captured_piece != KING, "Illegal king capture detected during make()")

	switch piece {
	case PAWN:
		brd.halfmove_clock = 0 // All pawn moves are irreversible.
		brd.pawn_hash_key ^= (pawn_zobrist(from, c) ^ pawn_zobrist(to, c))
		switch captured_piece {
		case EMPTY:
			if abs(to-from) == 16 { // handle en passant advances
				brd.enp_target = uint8(to)
				brd.hash_key ^= enp_zobrist(uint8(to)) // XOR in new en passant target
			}
		case PAWN: // Destination square will be empty if en passant capture
			if enp_target != SQ_INVALID && brd.TypeAt(to) == EMPTY {
				brd.pawn_hash_key ^= pawn_zobrist(int(enp_target), brd.Enemy())
				remove_piece(brd, PAWN, int(enp_target), brd.Enemy())
				brd.squares[enp_target] = EMPTY
			} else {
				brd.pawn_hash_key ^= pawn_zobrist(to, brd.Enemy())
				remove_piece(brd, PAWN, to, brd.Enemy())
			}
		case ROOK:
			if brd.castle > 0 {
				update_castle_rights(brd, to)
			}
			remove_piece(brd, captured_piece, to, brd.Enemy())
		default: // any non-pawn piece is captured
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

	// to do: update pawn hash key...

	case KING:
		switch captured_piece {
		case ROOK:
			if brd.castle > 0 {
				update_castle_rights(brd, from)
				update_castle_rights(brd, to) //
			}
			remove_piece(brd, captured_piece, to, brd.Enemy())
			brd.halfmove_clock = 0 // All capture moves are irreversible.
		case EMPTY:
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
		case PAWN:
			if brd.castle > 0 {
				update_castle_rights(brd, from)
			}
			remove_piece(brd, captured_piece, to, brd.Enemy())
			brd.pawn_hash_key ^= pawn_zobrist(to, brd.Enemy())
			brd.halfmove_clock = 0 // All capture moves are irreversible.
		default:
			if brd.castle > 0 {
				update_castle_rights(brd, from)
			}
			remove_piece(brd, captured_piece, to, brd.Enemy())
			brd.halfmove_clock = 0 // All capture moves are irreversible.
		}
		relocate_king(brd, KING, captured_piece, from, to, c)

	case ROOK:
		switch captured_piece {
		case ROOK:
			if brd.castle > 0 {
				update_castle_rights(brd, from)
				update_castle_rights(brd, to)
			}
			remove_piece(brd, captured_piece, to, brd.Enemy())
			brd.halfmove_clock = 0 // All capture moves are irreversible.
		case EMPTY:
			if brd.castle > 0 {
				update_castle_rights(brd, from)
			}
			brd.halfmove_clock += 1
		case PAWN:
			if brd.castle > 0 {
				update_castle_rights(brd, from)
			}
			remove_piece(brd, captured_piece, to, brd.Enemy())
			brd.halfmove_clock = 0 // All capture moves are irreversible.
			brd.pawn_hash_key ^= pawn_zobrist(to, brd.Enemy())
		default:
			if brd.castle > 0 {
				update_castle_rights(brd, from)
			}
			remove_piece(brd, captured_piece, to, brd.Enemy())
			brd.halfmove_clock = 0 // All capture moves are irreversible.
		}
		relocate_piece(brd, ROOK, from, to, c)

	default:
		switch captured_piece {
		case ROOK:
			if brd.castle > 0 {
				update_castle_rights(brd, to) //
			}
			remove_piece(brd, captured_piece, to, brd.Enemy())
			brd.halfmove_clock = 0 // All capture moves are irreversible.
		case EMPTY:
			brd.halfmove_clock += 1
		case PAWN:
			remove_piece(brd, captured_piece, to, brd.Enemy())
			brd.halfmove_clock = 0 // All capture moves are irreversible.
			brd.pawn_hash_key ^= pawn_zobrist(to, brd.Enemy())
		default:
			remove_piece(brd, captured_piece, to, brd.Enemy())
			brd.halfmove_clock = 0 // All capture moves are irreversible.
		}
		relocate_piece(brd, piece, from, to, c)
	}

	brd.c ^= 1 // flip the current side to move.
	brd.hash_key ^= side_key64
}

// Castle flag, enp target, hash key, pawn hash key, and halfmove clock are all restored during search
func unmake_move(brd *Board, move Move, memento *BoardMemento) {



	brd.c ^= 1 // flip the current side to move.

	c := brd.c
	piece := move.Piece()
	from := move.From()
	to := move.To()
	captured_piece := move.CapturedPiece()
	enp_target := memento.enp_target

	switch piece {
	case PAWN:
		if move.PromotedTo() != EMPTY {
			unmake_remove_piece(brd, move.PromotedTo(), to, c)
			brd.squares[to] = captured_piece
			unmake_add_piece(brd, piece, from, c)
		} else {
			unmake_relocate_piece(brd, piece, to, from, c)
		}
		switch captured_piece {
		case PAWN:
			if enp_target != SQ_INVALID {
				if c == WHITE {
					if to == int(enp_target)+8 {
						unmake_add_piece(brd, PAWN, int(enp_target), brd.Enemy())
					} else {
						unmake_add_piece(brd, PAWN, to, brd.Enemy())
					}
				} else {
					if to == int(enp_target)-8 {
						unmake_add_piece(brd, PAWN, int(enp_target), brd.Enemy())
					} else {
						unmake_add_piece(brd, PAWN, to, brd.Enemy())
					}
				}
			} else {
				unmake_add_piece(brd, PAWN, to, brd.Enemy())
			}
		case EMPTY:
		default: // any non-pawn piece was captured
			unmake_add_piece(brd, captured_piece, to, brd.Enemy())
		}

	case KING:
		unmake_relocate_king(brd, piece, captured_piece, to, from, c)
		if captured_piece != EMPTY {
			unmake_add_piece(brd, captured_piece, to, brd.Enemy())
		} else if abs(to-from) == 2 { // king castled.
			if c == WHITE {
				if to == G1 {
					unmake_relocate_piece(brd, ROOK, F1, H1, WHITE)
				} else {
					unmake_relocate_piece(brd, ROOK, D1, A1, WHITE)
				}
			} else {
				if to == G8 {
					unmake_relocate_piece(brd, ROOK, F8, H8, BLACK)
				} else {
					unmake_relocate_piece(brd, ROOK, D8, A8, BLACK)
				}
			}
		}

	default:
		unmake_relocate_piece(brd, piece, to, from, c)
		if captured_piece != EMPTY {
			unmake_add_piece(brd, captured_piece, to, brd.Enemy())
		}
	}

	brd.hash_key, brd.pawn_hash_key = memento.hash_key, memento.pawn_hash_key
	brd.castle, brd.enp_target = memento.castle, memento.enp_target
	brd.halfmove_clock = memento.halfmove_clock
}

// Whenever a king or rook moves off its initial square or is captured,
// update castle rights via the procedure associated with that square.
func update_castle_rights(brd *Board, sq int) {
	switch sq { // if brd.castle remains unchanged, hash key will be unchanged.
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

// do not use for en-passant captures.
func remove_piece(brd *Board, removed_piece Piece, sq int, e uint8) {
	brd.pieces[e][removed_piece].Clear(sq)
	brd.occupied[e].Clear(sq)
	brd.material[e] -= int32(removed_piece.Value() + main_pst[e][removed_piece][sq])
	brd.endgame_counter -= endgame_count_values[removed_piece]
	brd.hash_key ^= zobrist(removed_piece, sq, e) // XOR out the captured piece
}
func unmake_remove_piece(brd *Board, removed_piece Piece, sq int, e uint8) {
	brd.pieces[e][removed_piece].Clear(sq)
	brd.occupied[e].Clear(sq)
	brd.material[e] -= int32(removed_piece.Value() + main_pst[e][removed_piece][sq])
	brd.endgame_counter -= endgame_count_values[removed_piece]
}

func add_piece(brd *Board, added_piece Piece, sq int, c uint8) {
	brd.pieces[c][added_piece].Add(sq)
	brd.squares[sq] = added_piece
	brd.occupied[c].Add(sq)
	brd.material[c] += int32(added_piece.Value() + main_pst[c][added_piece][sq])
	brd.endgame_counter += endgame_count_values[added_piece]
	brd.hash_key ^= zobrist(added_piece, sq, c) // XOR in key for added_piece
}
func unmake_add_piece(brd *Board, added_piece Piece, sq int, c uint8) {
	brd.pieces[c][added_piece].Add(sq)
	brd.squares[sq] = added_piece
	brd.occupied[c].Add(sq)
	brd.material[c] += int32(added_piece.Value() + main_pst[c][added_piece][sq])
	brd.endgame_counter += endgame_count_values[added_piece]
}

func relocate_piece(brd *Board, piece Piece, from, to int, c uint8) {
	from_to := (sq_mask_on[from] | sq_mask_on[to])
	brd.pieces[c][piece] ^= from_to
	brd.occupied[c] ^= from_to
	brd.squares[from] = EMPTY
	brd.squares[to] = piece
	brd.material[c] += int32(main_pst[c][piece][to] - main_pst[c][piece][from])
	// XOR out the key for piece at from, and XOR in the key for piece at to.
	brd.hash_key ^= (zobrist(piece, from, c) ^ zobrist(piece, to, c))
}
func unmake_relocate_piece(brd *Board, piece Piece, from, to int, c uint8) {
	from_to := (sq_mask_on[from] | sq_mask_on[to])
	brd.pieces[c][piece] ^= from_to
	brd.occupied[c] ^= from_to
	brd.squares[from] = EMPTY
	brd.squares[to] = piece
	brd.material[c] += int32(main_pst[c][piece][to] - main_pst[c][piece][from])
}

func relocate_king(brd *Board, piece, captured_piece Piece, from, to int, c uint8) {
	from_to := (sq_mask_on[from] | sq_mask_on[to])
	brd.pieces[c][piece] ^= from_to
	brd.occupied[c] ^= from_to
	brd.squares[from] = EMPTY
	brd.squares[to] = piece
	// XOR out the key for piece at from, and XOR in the key for piece at to.
	brd.hash_key ^= (zobrist(piece, from, c) ^ zobrist(piece, to, c))
}
func unmake_relocate_king(brd *Board, piece, captured_piece Piece, from, to int, c uint8) {
	from_to := (sq_mask_on[from] | sq_mask_on[to])
	brd.pieces[c][piece] ^= from_to
	brd.occupied[c] ^= from_to
	brd.squares[from] = EMPTY
	brd.squares[to] = piece
}
