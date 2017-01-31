//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"fmt"
	"testing"
)

func TestEPDParsing(t *testing.T) {
	test, err := loadEpdFile("test_suites/wac_300.epd")
	if err != nil {
		fmt.Print(err)
		return
	}
	for _, epd := range test {
		epd.Print()
	}
}
