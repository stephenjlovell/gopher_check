//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"fmt"
	"time"
)

func RunTestSuite(testSuite string, depth, timeout int) {
	test := loadEpdFile(testSuite)
	var moveStr string
	sum, score := 0, 0
	var gt *GameTimer
	var search *Search

	start := time.Now()
	for i, epd := range test {
		gt = NewGameTimer(0, epd.brd.c)
		gt.SetMoveTime(time.Duration(timeout) * time.Millisecond)
		search = NewSearch(SearchParams{depth, false, false, false}, gt, nil, nil)
		search.Start(epd.brd)

		moveStr = ToSAN(epd.brd, search.bestMove)
		if correctMove(epd, moveStr) {
			score += 1
			fmt.Printf("-")
		} else {
			fmt.Printf("%d.", i+1)
		}
		sum += search.nodes
	}
	secondsElapsed := time.Since(start).Seconds()
	mNodes := float64(sum) / 1000000.0
	fmt.Printf("\n%.4fm nodes searched in %.4fs (%.4fm NPS)\n", mNodes, secondsElapsed, mNodes/secondsElapsed)

	fmt.Printf("Total score: %d/%d\n", score, len(test))

	// fmt.Printf("Average Branching factor by iteration:\n")
	// var branching float64
	// for d := 2; d <= depth; d++ {
	// 	branching = math.Pow(float64(nodesPerIteration[d])/float64(nodesPerIteration[1]), float64(1)/float64(d-1))
	// 	fmt.Printf("%d ABF: %.4f\n", d, branching)
	// }

	fmt.Printf("Overhead: %.4fm\n", float64(loadBalancer.Overhead())/1000000.0)
	fmt.Printf("Timeout: %.1fs\n", float64(timeout)/1000.0)
	// fmt.Printf("PV Accuracy: %d/%d (%.2f)\n\n", pvAccuracy[1], pvAccuracy[0]+pvAccuracy[1],
	// 	float64(pvAccuracy[1])/float64(pvAccuracy[0]+pvAccuracy[1]))
}

func correctMove(epd *EPD, moveStr string) bool {
	for _, a := range epd.avoidMoves {
		if moveStr == a {
			return false
		}
	}
	for _, b := range epd.bestMoves {
		if moveStr == b {
			return true
		}
	}
	return false
}
