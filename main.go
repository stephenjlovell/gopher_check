//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/pkg/profile"
)

var version = "0.3.0"

func Max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func Min(a, b int) int {
	if a > b {
		return b
	} else {
		return a
	}
}

func Abs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

func assert(statement bool, failureMessage string) {
	if !statement {
		panic("\nassertion failed: " + failureMessage + "\n")
	}
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	setupChebyshevDistance()
	setupMasks()
	setupMagicMoveGen()
	setupEval()
	setupRand()
	setupZobrist()
	resetMainTt()
	setupLoadBalancer(MaxWorkers())
}

func printName() {
	fmt.Printf("\n---------------------------------------\n")
	fmt.Printf(" \u265B GopherCheck v.%s \u265B\n", version)
	fmt.Printf(" Copyright \u00A9 2014 Stephen J. Lovell\n")
	fmt.Printf("---------------------------------------\n\n")
}

var cpuProfileFlag = flag.Bool("cpuprofile", false,
	"Runs cpu profiler on test suite and outputs to cpu.pprof.")
var memProfileFlag = flag.Bool("memprofile", false,
	"Runs memory profiler on test suite and outputs to mem.pprof.")
var versionFlag = flag.Bool("version", false, "Prints version number and exits.")

func main() {
	flag.Parse()
	if *versionFlag {
		printName()
	} else {
		if *cpuProfileFlag {
			printName()
			defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
			RunTestSuite("test_suites/wac_300.epd", MAX_DEPTH, 5000)
			// run 'go tool pprof -text gopher_check cpu.pprof > cpu_prof.txt' to output profile to text
			// run 'go tool pprof gopher_check cpu.pprof' to enter interactive mode
		} else if *memProfileFlag {
			printName()
			defer profile.Start(profile.MemProfileRate(64), profile.ProfilePath(".")).Stop()
			// run 'go tool pprof -text --alloc_objects gopher_check mem.pprof > mem_profile.txt' to output profile to text
			RunTestSuite("test_suites/wac_150.epd", MAX_DEPTH, 5000)
		} else {
			uci := NewUCIAdapter()
			uci.Read(bufio.NewReader(os.Stdin))
		}
	}
}
