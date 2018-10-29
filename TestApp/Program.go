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
	"bufio"
	"flag"
	"fmt"
	"github.com/smohr1824/Networks/Algorithms"
	"github.com/smohr1824/Networks/Core"
	"os"
	"runtime/pprof"
	"time"
)

// uncomment next two to profile
//var cpuprofile = flag.String("cpuprofile", "cpu_test.prof", "write cpu profile to `file`")
//var memprofile = flag.String("memprofile", "mem_test.prof", "write memory profile to `file`")

// uncomment next two to not profile
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")


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

	/* ser := Core.NewNetworkSerializer("|")
	G, _ := ser.ReadNetworkFromFile("hasedgestest.dat", false)
	order1 := G.Order()
	size1 := G.Size()
	fmt.Println(fmt.Sprintf("Vertices %d, edge %d", order1, size1))
	if !G.HasEdge("C", "A"){

 	} else {
 		wt := G.EdgeWeight("C", "A")
 		wt = G.EdgeWeight("A", "B")
 		wt++
 		test := G.HasEdge("A", "B")
 		if test {}
	}

	G, _ = ser.ReadNetworkFromFile("newadjtest.dat", false)
	neighbors := G.GetNeighbors("D")
	k := len(neighbors)
	k++
	matrix := G.AdjacencyMatrix()
	matrix[0][0] = 0 */

	//ser := Core.NewNetworkSerializer("|")
	//G, _ := ser.ReadNetworkFromFile("bipartitetest.dat", true)
	G := Core.NewNetwork(true)
	for i := 0; i < 200; i++ {
		place := fmt.Sprintf("P%d", i)
		transition := fmt.Sprintf("T%d", i)

		G.AddVertex(place)
		G.AddVertex(transition)
	}
	G.AddVertex("P200")

	for j:= 0; j < 200; j++ {
		place := fmt.Sprintf("P%d", j)
		transition := fmt.Sprintf("T%d", j)
		G.AddEdge(place, transition, 1)
	}
	G.AddEdge("T199", "P200", 1)

	for k:= 0; k < 199; k++ {
		place := fmt.Sprintf("P%d", k)
		trans := fmt.Sprintf("T%d", k + 1)
		G.AddEdge(place, trans, 1)
	}

	for l:= 0; l < 199; l++ {
		place := fmt.Sprintf("P%d", l + 2)
		trans := fmt.Sprintf("T%d", l)
		G.AddEdge(trans, place, 1)
	}

	start := time.Now()
	numgos := 2
	isIt, R, B := Algorithms.ConcurrentBipartite(G, numgos)
	elapsed := time.Since(start)
	if isIt {
		fmt.Println(fmt.Sprintf("Is bipartite, R has %d, B has %d memebers", len(R), len(B)))
		fmt.Println(fmt.Sprintf("Elapsed time %s using %d goroutines", elapsed, numgos))
	} else {
		fmt.Println("Not bipartite")
	}

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	fmt.Println(text)

	// command line args are the edge list filename, delimiter (e.g., "," for a csv file, and number of concurrent routines (partitions) to use
	// edge list is from node delimiter to node on one line
	/*csvsr := Core.NewNetworkSerializer(os.Args[2])
	pythonnetwork, err := csvsr.ReadNetworkFromFile(os.Args[1], true)
	size := pythonnetwork.Size()
	fmt.Println(fmt.Sprintf("Number of edges is %d", size))
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
	communities := Algorithms.ConcurrentSLPA(pythonnetwork, 20, .4, 123456, procs, 2)
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
	}*/

}



