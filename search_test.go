//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import "testing"

func TestPlayingStrength(t *testing.T) {
	printName()
	timeout := 2000
	RunTestSuite("test_suites/wac_300.epd", MAX_DEPTH, timeout)
}
