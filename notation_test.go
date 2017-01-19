//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

"fmt"
"testing"

func TestEPDParsing(t *testing.T) {
  test := load_epd_file("test_suites/wac_300.epd")

  for _, epd := range test {
    epd.Print()
  }

}
