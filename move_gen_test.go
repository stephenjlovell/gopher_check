package main

import (
	"fmt"
	// "strconv"
	"testing"
	"time"
	"github.com/stephenjlovell/gopher_check/load_balancer"
)

// func TestSetup(t *testing.T) {
// 	setup()
// 	brd := StartPos()
// 	brd.PrintDetails()
// }

var legal_max_tree = [10]int{1, 20, 400, 8902, 197281, 4865609, 119060324, 3195901860, 84998978956, 2439530234167}

// func TestLegalMoveGen(t *testing.T) {
// 	setup()
// 	brd := StartPos()
// 	copy := brd.Copy()
// 	depth := 6
// 	start := time.Now()
// 	sum := PerftLegal(brd, depth)
// 	elapsed := time.Since(start)
// 	nps := int64(float64(sum) / elapsed.Seconds())

// 	fmt.Printf("%d nodes at depth %d. %d NPS\n", sum, depth, nps)

// 	fmt.Printf("%d total nodes in check\n", check_count)
// 	fmt.Printf("%d total capture nodes\n", capture_count)

// 	CompareBoards(copy, brd)
// 	Assert(*brd == *copy, "move generation did not return to initial board state.")
// 	Assert(sum == legal_max_tree[depth], "Expected "+strconv.Itoa(legal_max_tree[depth])+
// 		" nodes, got "+strconv.Itoa(sum))
// }

// func TestMoveGen(t *testing.T) {
// 	setup()
// 	brd := StartPos()
// 	copy := brd.Copy()
// 	depth := 5
// 	start := time.Now()
// 	sum := Perft(brd, depth)
// 	elapsed := time.Since(start)
// 	nps := int64(float64(sum)/elapsed.Seconds())

// 	fmt.Printf("%d nodes at depth %d. %d NPS\n", sum, depth, nps)

// 	fmt.Printf("%d total nodes in check\n", check_count)
// 	fmt.Printf("%d total capture nodes\n", capture_count)

// 	CompareBoards(copy, brd)
// 	Assert(*brd == *copy, "move generation did not return to initial board state.")
// }

func TestParallelMoveGen(t *testing.T) {
	setup()
	brd := StartPos()
	copy := brd.Copy()
	depth := 5
	balancer := load_balancer.NewBalancer(work)
	balancer.Setup(work)

	go func() {
		for _ = range time.Tick(time.Second) {  
			balancer.Print()  // periodically print out the number of pending tasks assigned to each worker.
		}
	}()

	start := time.Now()

	cancel_child := make(chan bool)
	update_child := make(chan int)

	sum := PerftParallel(brd, depth, cancel_child, update_child)
	elapsed := time.Since(start)
	nps := int64(float64(sum)/elapsed.Seconds())

	fmt.Printf("\n%d nodes at depth %d. %d NPS\n", sum, depth, nps)

	fmt.Printf("%d total nodes in check\n", check_count)
	fmt.Printf("%d total capture nodes\n", capture_count)

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
		castle:         uint8(15),
		enp_target:     SQ_INVALID,
		halfmove_clock: uint8(0),
	}
	for sq := 0; sq < 64; sq++ {
		brd.squares[sq] = EMPTY
	}
	for sq := A2; sq <= H2; sq++ {
		add_piece(brd, PAWN, sq, WHITE)
	}
	for sq := A7; sq <= H7; sq++ {
		add_piece(brd, PAWN, sq, BLACK)
	}
	add_piece(brd, ROOK, A1, WHITE)
	add_piece(brd, KNIGHT, B1, WHITE)
	add_piece(brd, BISHOP, C1, WHITE)
	add_piece(brd, QUEEN, D1, WHITE)
	add_piece(brd, KING, E1, WHITE)
	add_piece(brd, BISHOP, F1, WHITE)
	add_piece(brd, KNIGHT, G1, WHITE)
	add_piece(brd, ROOK, H1, WHITE)
	add_piece(brd, ROOK, A8, BLACK)
	add_piece(brd, KNIGHT, B8, BLACK)
	add_piece(brd, BISHOP, C8, BLACK)
	add_piece(brd, QUEEN, D8, BLACK)
	add_piece(brd, KING, E8, BLACK)
	add_piece(brd, BISHOP, F8, BLACK)
	add_piece(brd, KNIGHT, G8, BLACK)
	add_piece(brd, ROOK, H8, BLACK)

	return brd
}

var check_count int = 0
var capture_count int = 0

func PerftLegal(brd *Board, depth int) int {
	if depth == 0 {
		return 1
	}
	sum := 0
	in_check := is_in_check(brd)
	if in_check {
		check_count += 1
	}
	best_moves, remaining_moves := get_all_moves(brd, in_check, 0)
	for _, item := range *best_moves {
		// if in_check || avoids_check(brd, item.move) {
		sum += PerftLegal_make_unmake(brd, item.move, depth-1)
		// }
	}
	for _, item := range *remaining_moves {
		// if in_check || avoids_check(brd, item.move) {
		sum += PerftLegal_make_unmake(brd, item.move, depth-1)
		// }
	}
	return sum
}

func PerftLegal_make_unmake(brd *Board, m Move, depth int) int {

	Assert(m != 0, "invalid move generated.")

	if m.IsCapture() {
		capture_count += 1
	}
	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock
	sum := 0
	make_move(brd, m) // to do: make move
	if !enemy_in_check(brd) {
		sum = PerftLegal(brd, depth)
	}
	unmake_move(brd, m, enp_target) // to do: unmake move
	brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
	brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
	return sum
}

func Perft(brd *Board, depth int) int {
	if depth == 0 {
		return 1
	}
	sum := 0
	in_check := is_in_check(brd)
	if in_check {
		check_count += 1
	}
	best_moves, remaining_moves := get_all_moves(brd, in_check, 0)
	for _, item := range *best_moves {
		sum += Perft_make_unmake(brd, item.move, depth-1)
	}
	for _, item := range *remaining_moves {
		sum += Perft_make_unmake(brd, item.move, depth-1)
	}
	return sum
}

func Perft_make_unmake(brd *Board, m Move, depth int) int {

	Assert(m != 0, "invalid move generated.")

	if m.IsCapture() {
		capture_count += 1
	}
	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock
	make_move(brd, m) // to do: make move
	sum := Perft(brd, depth)
	unmake_move(brd, m, enp_target) // to do: unmake move
	brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
	brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
	return sum
}



func PerftParallel(brd *Board, depth int, cancel chan bool, update chan int) int {
	if depth == 0 {
		return 1
	}
	sum := 0
	in_check := is_in_check(brd)
	if in_check {
		check_count += 1
	}
	cancel_child := make(chan bool)
	update_child := make(chan int)

	best_moves, remaining_moves := get_best_moves(brd, in_check, 0)
	for _, item := range *best_moves {
		if is_cancelled(cancel, cancel_child, update_child) {
			return 0
		} // make sure the job hasn't been cancelled.
		sum += PerftParallel_make_unmake(brd, item.move, depth-1, cancel_child, update_child)
	}

	get_remaining_moves(brd, in_check, remaining_moves, 0)  // search remaining nodes in parallel
	result_child := make(chan int, 30)
	child_counter := 0
	for _, item := range *remaining_moves {
		m := item.move
		new_brd := brd.Copy() // create a locally scoped deep copy of the board.

		req := load_balancer.Request{ // package the subtree search into a Request object
			Cancel: cancel_child,
			Result: result_child,
			Size:   (3 << uint(depth-1)), // estimate of the number of main search leaf nodes remaining
			Fn: func() int {
				return PerftParallel_make_unmake(new_brd, m, depth-1, cancel_child, update_child)
			},
		}
		work <- req // pipe the new request object to the load balancer to execute in parallel.
		child_counter++
	}

	// fmt.Printf("%d nodes spawned in parallel at depth %d\n", child_counter, depth)

	if child_counter > 0 { // wait for a message to come in on one of the channels
remaining_pieces:
		for {
			select {
			case <-cancel: // task was cancelled.
				println("task cancelled")
				cancel_work(cancel_child, update_child)
				return 0
			case child_sum := <-result_child: // one of the child subtrees has been completely searched.
				// println("response received.")
				sum += child_sum
				child_counter--
				if child_counter == 0 {
					break remaining_pieces // exit the for loop
				}
			// default:
			// 	if child_counter == 0 {
			// 		break remaining_pieces // exit the for loop
			// 	}
			// }
			}
		}
	}
	return sum
}

func PerftParallel_make_unmake(brd *Board, m Move, depth int, cancel chan bool, update chan int) int {

	Assert(m != 0, "invalid move generated.")

	if m.IsCapture() {
		capture_count += 1
	}
	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock
	make_move(brd, m) // to do: make move
	sum := PerftParallel(brd, depth, cancel, update)
	unmake_move(brd, m, enp_target) // to do: unmake move
	brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
	brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
	return sum
}












