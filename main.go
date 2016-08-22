//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------
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

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/pkg/profile"
)

var setup_once sync.Once

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
func min(a, b int) int {
	if a > b {
		return b
	} else {
		return a
	}
}
func abs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

func assert(statement bool, failure_message string) {
	if !statement {
		panic("\nassertion failed: " + failure_message + "\n")
	}
}

func setup() {
	setup_once.Do(func() {
		num_cpu := runtime.NumCPU()
		runtime.GOMAXPROCS(num_cpu)
		setup_chebyshev_distance()
		setup_masks()
		setup_magic_move_gen()
		setup_eval()
		setup_rand()
		setup_zobrist()
		reset_main_tt()
		setup_load_balancer(num_cpu)
	})
}

var version = "0.1.1"

func print_name() {
	fmt.Printf("\n---------------------------------------\n")
	fmt.Printf(" \u265B GopherCheck v.%s \u265B\n", version)
	fmt.Printf(" Copyright \u00A9 2014 Stephen J. Lovell\n")
	fmt.Printf("---------------------------------------\n\n")
}

var cpu_profile_flag = flag.Bool("cpuprofile", false, "Runs cpu profiler on test suite.")
var mem_profile_flag = flag.Bool("memprofile", false, "Runs memory profiler on test suite.")
var version_flag = flag.Bool("version", false, "Prints version number and exits.")

func main() {
	flag.Parse()
	if *version_flag {
		print_name()
	} else {
		setup()
		if *cpu_profile_flag {
			print_name()
			defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
			RunTestSuite("test_suites/wac_75.epd", MAX_DEPTH, 5000)
			// run 'go tool pprof -text gopher_check cpu.pprof > cpu_prof.txt' to output profile to text
		} else if *mem_profile_flag {
			print_name()
			defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
			RunTestSuite("test_suites/wac_150.epd", MAX_DEPTH, 5000)
		} else {
			uci := NewUCIAdapter()
			uci.Read(bufio.NewReader(os.Stdin))
		}
	}
}
