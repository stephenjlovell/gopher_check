//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"fmt"
	"time"
)

func RunTestSuite(test_suite string, depth, timeout int) {
	test := load_epd_file(test_suite)
	var move_str string
	sum, score := 0, 0
	var gt *GameTimer
	var search *Search

	start := time.Now()
	for i, epd := range test {
		gt = NewGameTimer(0, epd.brd.c)
		gt.SetMoveTime(time.Duration(timeout) * time.Millisecond)
		search = NewSearch(SearchParams{depth, false, false, false}, gt, nil, nil)
		search.Start(epd.brd)

		move_str = ToSAN(epd.brd, search.best_move)
		if correct_move(epd, move_str) {
			score += 1
			fmt.Printf("-")
		} else {
			fmt.Printf("%d.", i+1)
		}
		sum += search.nodes
	}
	seconds_elapsed := time.Since(start).Seconds()
	m_nodes := float64(sum) / 1000000.0
	fmt.Printf("\n%.4fm nodes searched in %.4fs (%.4fm NPS)\n", m_nodes, seconds_elapsed, m_nodes/seconds_elapsed)

	fmt.Printf("Total score: %d/%d\n", score, len(test))

	// fmt.Printf("Average Branching factor by iteration:\n")
	// var branching float64
	// for d := 2; d <= depth; d++ {
	// 	branching = math.Pow(float64(nodes_per_iteration[d])/float64(nodes_per_iteration[1]), float64(1)/float64(d-1))
	// 	fmt.Printf("%d ABF: %.4f\n", d, branching)
	// }

	fmt.Printf("Overhead: %.4fm\n", float64(load_balancer.Overhead())/1000000.0)
	fmt.Printf("Timeout: %.1fs\n", float64(timeout)/1000.0)
	// fmt.Printf("PV Accuracy: %d/%d (%.2f)\n\n", pv_accuracy[1], pv_accuracy[0]+pv_accuracy[1],
	// 	float64(pv_accuracy[1])/float64(pv_accuracy[0]+pv_accuracy[1]))
}

func correct_move(epd *EPD, move_str string) bool {
	for _, a := range epd.avoid_moves {
		if move_str == a {
			return false
		}
	}
	for _, b := range epd.best_moves {
		if move_str == b {
			return true
		}
	}
	return false
}
