package main

import (
	"fmt"
	"testing"
)

// func TestSetup(t *testing.T) {
// 	setup()
// 	brd := StartPos()
// 	brd.PrintDetails()
// }


// legal perft size:
var legal_max_tree = [10]int{1,20,400,8902,197281,4865609,119060324,3195901860,84998978956,2439530234167}

func TestMoveGen(t *testing.T) {
	setup()
	brd := StartPos()
	copy := brd.Copy()
	depth := 5
	sum := Perft(brd, depth)
	fmt.Printf("%d total nodes at depth %d\n", sum, depth)
	// if *brd != *copy {
	// 	brd.PrintDetails()
	// }
	CompareBoards(copy, brd)
	Assert(*brd == *copy, "move generation did not return to initial board state.")
}

func CompareBoards(brd, other *Board) {
	if brd.pieces != other.pieces {
		fmt.Println("Board.pieces unequal")
	}
	if brd.squares != other.squares {
		fmt.Println("Board.squares unequal")
		fmt.Println("original:")
		brd.Print()
		fmt.Println("new board:")
		other.Print()
	}
	if brd.occupied != other.occupied {
		fmt.Println("Board.occupied unequal")
		for i := 0; i < 2; i++ {
			fmt.Printf("side: %d\n", i)
			fmt.Println("original:")
			brd.occupied[i].Print()
			fmt.Println("new board:")
			other.occupied[i].Print()
		}
	}
	if brd.material != other.material {
		fmt.Println("Board.material unequal")
	}
	if brd.hash_key != other.hash_key {
		fmt.Println("Board.hash_key unequal")
	}
	if brd.pawn_hash_key != other.pawn_hash_key {
		fmt.Println("Board.pawn_hash_key unequal")
	}
	if brd.c != other.c {
		fmt.Println("Board.c unequal")
	}
	if brd.castle != other.castle {
		fmt.Println("Board.castle unequal")
	}
	if brd.enp_target != other.enp_target {
		fmt.Println("Board.enp_target unequal")
	}
	if brd.halfmove_clock != other.halfmove_clock {
		fmt.Println("Board.halfmove_clock unequal")
	}
}


func Assert(statement bool, failure_message string) {
	if !statement {
		fmt.Printf("F")
		panic("\nAssertion failed: " + failure_message + "\n")
	}
}

// pieces         [2][6]BB  // 768 bits
// squares        [64]Piece // 512 bits
// occupied       [2]BB     // 128 bits
// material       [2]int32  // 64 bits
// hash_key       uint64    // 64 bits
// pawn_hash_key  uint64    // 64 bits
// c              uint8     // 8 bits
// castle         uint8     // 8 bits
// enp_target     uint8     // 8 bits
// halfmove_clock uint8     // 8 bits

func StartPos() *Board {
	brd := &Board{
		c:              WHITE,
		castle:         uint8(8),
		enp_target:     SQ_INVALID,
		halfmove_clock: uint8(0),
	}
	for sq := 0; sq < 64; sq++ {
		brd.squares[sq] = EMPTY
	}
	for sq := A2; sq <= H2; sq++ {
		Add_piece(brd, PAWN, sq, WHITE)
	}
	for sq := A7; sq <= H7; sq++ {
		Add_piece(brd, PAWN, sq, BLACK)
	}
	Add_piece(brd, ROOK, A1, WHITE)
	Add_piece(brd, KNIGHT, B1, WHITE)
	Add_piece(brd, BISHOP, C1, WHITE)
	Add_piece(brd, QUEEN, D1, WHITE)
	Add_piece(brd, KING, E1, WHITE)
	Add_piece(brd, BISHOP, F1, WHITE)
	Add_piece(brd, KNIGHT, G1, WHITE)
	Add_piece(brd, ROOK, H1, WHITE)
	Add_piece(brd, ROOK, A8, BLACK)
	Add_piece(brd, KNIGHT, B8, BLACK)
	Add_piece(brd, BISHOP, C8, BLACK)
	Add_piece(brd, QUEEN, D8, BLACK)
	Add_piece(brd, KING, E8, BLACK)
	Add_piece(brd, BISHOP, F8, BLACK)
	Add_piece(brd, KNIGHT, G8, BLACK)
	Add_piece(brd, ROOK, H8, BLACK)

	return brd
}

func Perft(brd *Board, depth int) int {
	if depth == 0 {
		return 1
	}
	sum := 0
	in_check := is_in_check(brd)
	best_moves, remaining_moves := get_all_moves(brd, in_check, 0)
	for _, item := range *best_moves {
		sum += Perft_make_unmake(brd, item.move, depth)
	}
	for _, item := range *remaining_moves {
		sum += Perft_make_unmake(brd, item.move, depth)
	}
	return sum
}

func Perft_make_unmake(brd *Board, m Move, depth int) int {
	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock

	make_move(brd, m) // to do: make move
	sum := Perft(brd, depth-1)
	unmake_move(brd, m, enp_target) // to do: unmake move

	brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
	brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
	return sum
}
