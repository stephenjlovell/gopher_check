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

// when unmaking a move from the transposition table

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

	relocate_piece(brd, piece, from, to, c) // Relocate the piece making the move

  update_castle := brd.castle > 0
	if update_castle {
		update_castle_rights(brd, from)
	}

  captured_piece := move.CapturedPiece()

	switch piece {
	case PAWN:
		promoted_piece := move.PromotedTo()
		if promoted_piece != EMPTY {
      
      // handle promotions

		}
    

		// handle en passant advances and en passant captures

		brd.halfmove_clock = 0 // All pawn moves are irreversible.

	case KING:

		// determine if the king is castling.

    if captured_piece != EMPTY {
      if update_castle {
        update_castle_rights(brd, to) //
      }
      remove_piece(brd, captured_piece, to, brd.Enemy())
      brd.halfmove_clock = 0 // All capture moves are irreversible.
    } else {
      brd.halfmove_clock += 1
    }
	
  default:
    if captured_piece != EMPTY {
      if update_castle {
        update_castle_rights(brd, to) //
      }
      remove_piece(brd, captured_piece, to, brd.Enemy())
      brd.halfmove_clock = 0 // All capture moves are irreversible.
    } else {
      brd.halfmove_clock += 1
    }
	}

}

// Whenever a king or rook moves off its initial square or is captured,
// update castle rights via the procedure associated with that square.
func update_castle_rights(brd *Board, sq int) {
	switch sq {
	case A1:
		brd.castle &= (^C_WQ)
	case E1: // white king starting position
		brd.castle &= (^(C_WK | C_WQ))
	case H1:
		brd.castle &= (^C_WK)
	case A8:
		brd.castle &= (^C_BQ)
	case E8: // black king starting position
		brd.castle &= (^(C_BK | C_BQ))
	case H8:
		brd.castle &= (^C_BK)
	}
}

// As long as Board.squares is only used to determine the type of piece when the occupancy is already
// known, it may be possible to only update it for the 'to' square during relocate_piece() 

func relocate_piece(brd *Board, piece Piece, from, to int, c uint8) {
	from_to := (sq_mask_on[from] | sq_mask_on[to])
	brd.pieces[c][piece] ^= from_to
	brd.occupied[c] ^= from_to
	brd.squares[from] = EMPTY
	brd.squares[to] = piece
	// XOR out the key for piece at from, and XOR in the key for piece at to.
	brd.hash_key ^= (zobrist(piece, from, c) ^ zobrist(piece, to, c))
}

func remove_piece(brd *Board, captured_piece Piece, sq int, e uint8) {
	mask := sq_mask_on[sq]
	brd.pieces[e][captured_piece] ^= mask
	brd.occupied[e] ^= mask
  brd.material[e] -= captured_piece.Value()
	// captured_piece is 'removed' from brd.squares by being overwritten by the attacking piece.
	brd.hash_key ^= zobrist(captured_piece, sq, e) // XOR out the captured piece
}

func add_piece(brd *Board, added_piece Piece, sq int, c uint8) {
  mask := sq_mask_on[sq]
  brd.pieces[c][added_piece] ^= mask
  brd.occupied[c] ^= mask
  brd.squares[sq] = added_piece
  brd.hash_key ^= zobrist(added_piece, sq, c)
}

func unmake_move(brd *Board, move Move) {

}
