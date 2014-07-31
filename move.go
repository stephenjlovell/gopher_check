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

// define an interface shared by all moves.
type AnyMove interface {
	From() int
	To() int
	Piece() Piece
	CapturedPiece() Piece
	PromotedTo() Piece
}

type Move uint32

// To to fit into transposition table entries, moves are encoded using 21 bits as follows (in LSB order):
// From square - first 6 bits
// To square - next 6 bits
// Piece - next 3 bits
// Captured piece - next 3 bits
// promoted to - next 3 bits

func (m Move) From() int {
	return int(uint32(m) & uint32(63))
}

func (m Move) To() int {
	return int((uint32(m) >> 6) & uint32(63))
}

func (m Move) Piece() Piece {
	return Piece((uint32(m) >> 12) & uint32(7))
}

func (m Move) CapturedPiece() Piece {
	return Piece((uint32(m) >> 15) & uint32(7))
}

func (m Move) PromotedTo() Piece {
	return Piece((uint32(m) >> 18) & uint32(7))
}

func (m Move) IsQuiet() bool {
	if ((uint32(m) >> 15) & uint32(63)) > 0 { return false } else { return true }
}

func NewMove(from, to int, piece, captured_piece, promoted_to Piece) Move {
	return Move(from) | (Move(to) << 6) | (Move(piece) << 12) | (Move(captured_piece) << 15) | (Move(promoted_to) << 18)
}

func NewRegularMove(from, to int, piece Piece) Move {
	return Move(from) | (Move(to) << 6) | (Move(piece) << 12) | (Move(EMPTY) << 15) | (Move(EMPTY) << 18)
}

func NewCapture(from, to int, piece, captured_piece Piece) Move {
	return Move(from) | (Move(to) << 6) | (Move(piece) << 12) | (Move(captured_piece) << 15) | (Move(EMPTY) << 18)
}

// Generate moves in batches to save effort on move generation when cutoffs occur.

// Ordering: PV/hash, promotions, winning captures, killers, losing captures, quiet moves

// # Moves are ordered based on expected subtree value. Better move ordering produces a greater
// # number of alpha/beta cutoffs during search, reducing the size of the actual search tree toward the minimal tree.
// def get_moves(depth, enhanced_sort=false, in_check=false)
//   promotions, captures, moves = [], [], []

//   if in_check
//     MoveGen::get_evasions(@pieces, @side_to_move, @board.squares, @enp_target, promotions, captures, moves)
//   else
//     MoveGen::get_captures(@pieces, @side_to_move, @board.squares, @enp_target, captures, promotions)
//     MoveGen::get_non_captures(@pieces, @side_to_move, @castle, moves, in_check)
//   end

//   if enhanced_sort  # At higher depths, expend additional effort on move ordering.
//     enhanced_sort(promotions, captures, moves, depth)
//   else
//     promotions + sort_captures_by_see!(captures) + history_sort!(moves)
//   end
// end

// # Generate only moves that create big swings in material balance, i.e. captures and promotions.
// # Used during Quiescence search to seek out positions from which a stable static evaluation can
// # be performed.
// def get_captures(evade_check)
//   # During quiesence search, sorting captures by SEE has the added benefit of enabling the pruning of bad
//   # captures (those with SEE < 0). In practice, this reduced the average number of q-nodes by around half.
//   promotions, captures = [], []
//   if evade_check
//     moves = []
//     MoveGen::get_evasions(@pieces, @side_to_move, @board.squares, @enp_target, promotions, captures, moves)
//     promotions + sort_captures_by_see!(captures) + history_sort!(moves)
//   else
//     MoveGen::get_winning_captures(@pieces, @side_to_move, @board.squares, @enp_target, captures, promotions)
//     promotions + sort_winning_captures_by_see!(captures)
//   end
// end
