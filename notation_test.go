//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"fmt"
	"testing"
)

// Verify that required FEN fields are parsed correctly.
func TestEPDParsing(t *testing.T) {
	test, err := LoadEpdFile("test_suites/wac_300.epd")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, epd := range test {
		fen := BoardToFEN(epd.brd)
		if fen != epd.fen {
			fmt.Print("\n")
			fmt.Println(epd.fen)
			fmt.Println(fen)
			panic("FEN string parsing not symmetric.")
		}
	}
}
