// Copyright 2017 - 2018 Stephen T. Mohr, OSIsoft, LLC
// MIT License

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"github.com/smohr1824/Networks/Core"
	"github.com/smohr1824/Networks/Algorithms"
	"fmt"
	"os"
	"runtime"
	"time"
	"strconv"
	"runtime/pprof"
	"flag"
)

var cpuprofile = flag.String("cpuprofile", "cpu_test.prof", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "mem_test.prof", "write memory profile to `file`")

func main() {
	//cpuSetting := runtime.GOMAXPROCS(0)
	//cpuAvailable := runtime.NumCPU()
	//fmt.Println(fmt.Sprintf("Max number of CPUs/threads to use: %d", cpuSetting))
	//fmt.Println(fmt.Sprintf("Available CPUs/threads: %d", cpuAvailable))


	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Println("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Println("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// command line args are the edge list filename, delimiter (e.g., "," for a csv file, and number of concurrent routines (partitions) to use
	// edge list is from node delimiter to node on one line
	csvsr := Core.NewNetworkSerializer(os.Args[2])
	pythonnetwork, err := csvsr.ReadNetworkFromFile(os.Args[1], true)
	if err != nil {
		fmt.Println("Error on read")
		return
	}

	fmt.Println(fmt.Sprintf("Read %d nodes", pythonnetwork.Order()))
	start := time.Now()
	procs, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println("Number of processors must be an integer")
	}
	communities := Algorithms.ConcurrentSLPA(pythonnetwork, 20, .4, time.Now().Unix(), procs)
	end := time.Now()

	dur := end.Sub(start)
	fmt.Println(fmt.Sprintf("Duration: %f", dur.Seconds()))
	fmt.Println(fmt.Sprintf("%d communities found", len(communities)))
	//fmt.Println(communities)

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			fmt.Println("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Println("could not write memory profile: ", err)
		}
		f.Close()
	}

}



