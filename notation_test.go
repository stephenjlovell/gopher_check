//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import "testing"

func TestEPDParsing(t *testing.T) {
	test := loadEpdFile("test_suites/wac300.epd")

	for _, epd := range test {
		epd.Print()
	}

}
