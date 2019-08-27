// Copyright 2017 - 2019 Stephen T. Mohr
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
	"flag"
	"fmt"
	"github.com/smohr1824/Networks/Core"
	. "github.com/smohr1824/Networks/FuzzyCognitiveMaps"
	"os"
	"runtime/pprof"
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

	/*N := Core.NewNetwork(true)
	N.AddEdge(1, 2, 1)
	N.AddEdge(1, 3, 1)
	N.AddEdge(2, 5, 2)
	N.AddEdge(3, 4, 2)
	N.AddEdge(5, 6, 3)
	N.AddEdge(4, 6, 3)
	Core.WriteNetworkToFile(N, "simple.gml")

	net, err := Core.ReadNetworkFromFile("simple.gml")
	if err != nil {
		fmt.Println("Error reading network")
	} else {
		fmt.Println(fmt.Sprintf("Read network with %d nodes and size %d", net.Order(), net.Size()))
	}*/

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

	/*ser := Core.NewNetworkSerializer("|")
	G, _ := ser.ReadNetworkFromFile("bipartitetest.dat", true)
	G := Core.NewNetwork(true)
	for i := 0; i < 2000; i++ {
		place := fmt.Sprintf("%d", i)
		transition := fmt.Sprintf("%d", i)

		G.AddVertex(place)
		G.AddVertex(transition)
	}
	G.AddVertex("P2000")

	for j:= 0; j < 2000; j++ {
		place := fmt.Sprintf("P%d", j)
		transition := fmt.Sprintf("T%d", j)
		G.AddEdge(place, transition, 1)
	}
	G.AddEdge("T1999", "P2000", 1)

	for k:= 0; k < 1999; k++ {
		place := fmt.Sprintf("P%d", k)
		trans := fmt.Sprintf("T%d", k + 1)
		G.AddEdge(place, trans, 1)
	}

	for l:= 0; l < 1999; l++ {
		place := fmt.Sprintf("P%d", l + 2)
		trans := fmt.Sprintf("T%d", l)
		G.AddEdge(trans, place, 1)
	}

	start := time.Now()
	numgos := 4
	isIt, R, B := Algorithms.ConcurrentBipartite(G, numgos)
	elapsed := time.Since(start)
	if isIt {
		fmt.Println(fmt.Sprintf("Is bipartite, R has %d, B has %d members", len(R), len(B)))
		fmt.Println(fmt.Sprintf("Elapsed time %s using %d goroutines", elapsed, numgos))
	} else {
		fmt.Println("Not bipartite")
	}

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	fmt.Println(text) */

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
	ser := Core.NewNetworkSerializer("|")
	G, err := ser.ReadNetworkFromFile("displays2.dat", false)
	if err != nil {
		t.Errorf("Error reading test file")
	}

	communities := Algorithms.ConcurrentSLPA(G, 20, 0.3, 3000, 1, 2)
	t.Errorf(fmt.Sprintf("%d communities found", len(communities)))
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
// test SLPA
	/*ser := Core.NewNetworkSerializer("|")
	G, err := ser.ReadNetworkFromFile("displays2.dat", false)
	if err != nil {
		fmt.Println("Error reading test file")
	}

	communities := Algorithms.ConcurrentSLPA(G, 20, 0.3, 3000, 2, 2)
	fmt.Println(fmt.Sprintf("%d communities found", len(communities)))
	fmt.Println(communities)*/

	// BasicMLFCM()
	/*fcm := BuildMLBasic()
	s := NewMLFCMSerializer()
	s.WriteMLFCMToFile(fcm, "..\\Work\\mlbasic.fcm")
	fcm2,_ := s.ReadMLFCMFromFile("..\\Work\\MLBasic.fcm")
	layers := fcm2.ListLayers()
	for _,layer := range layers {
		fmt.Println(layer)
	}
	fmt.Println("Initial State")
	writeState(fcm)
	fmt.Println("")

	for i := 1; i < 4; i++ {
		fcm.Step()
		fmt.Println(fmt.Sprintf("Iteration %d", i))
		writeState(fcm)
		fmt.Println("")
	}
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan() */

	Q, err := Core.ReadMultilayerNetworkFromFile("multilayer_three_aspects.gml")
	if err != nil {
		fmt.Println("Error reading test file multilayer_three_aspects.gml")
		return
	}
	supra := Q.MakeSupraAdjacencyMatrix()
	for _, row := range supra {
		for _, elt := range row {
			fmt.Printf("%.0f ", elt)
		}
		fmt.Println("")
	}
	// test basic fcm
	/*fcm := NewFuzzyCognitiveMapDefault()
	fcm.AddConcept("A", 1.0, 1.0)
	fcm.AddConcept("B", 0.0, 0.0)
	fcm.AddConcept("C", 1.0, 1.0)
	fcm.AddConcept("D", 0.0, 0.0)
	fcm.AddConcept("E", 0.0, 0.0)

	fcm.AddInfluence("B", "A", 1.0)
	fcm.AddInfluence("A", "C", 1.0)
	fcm.AddInfluence("C", "E", 1.0)
	fcm.AddInfluence("E", "D", 1.0)
	fcm.AddInfluence("D", "C", -1.0)
	fcm.AddInfluence("B", "E", -1.0)
	fcm.AddInfluence("E", "A", -1.0)
	fcm.AddInfluence("D", "B", 1.0)
	fcm.AddInfluence("E", "F", -1.0)

	fcm.Step() */
	// test bipartite
	/*G := Core.NewNetwork(false)

	for i:= 0; i < 1001; i += 2 {
		G.AddVertex(uint32(i))
		G.AddVertex(uint32(i + 1))
	}

	for i := 0; i < 1001; i += 2 {
		G.AddEdge(uint32(i), uint32(i + 1), 1.0)
	}
	for k := 1001; k > 1; k -= 2 {
		G.AddEdge(uint32(k), uint32(k - 3), 1.0)
	}

	isIt, R, B := Algorithms.ConcurrentBipartite(G, 4)
	if isIt {
		fmt.Println("Bipartite")
		fmt.Println(fmt.Sprintf("R is %d items long", len(R)))
		fmt.Println(fmt.Sprintf("B is %d items long", len(B)))
	} */
}

func BasicMLFCM() {
	fcm := BuildMLBasic()

	concepts := fcm.ListConcepts()
	s := ""
	for _, conName := range concepts {
		s += conName + " "
	}
	fmt.Println(s)
}

func writeState(fcm *MultilayerFuzzyCognitiveMap) {
	concepts := fcm.ListConcepts()
	sMain := ""
	for _, concept := range concepts {
		aggLevel, _ := fcm.GetActivationLevel(concept)
		sMain += fmt.Sprintf("%s: %.4f ", concept, aggLevel)
	}
	fmt.Println(sMain)
	layers := fcm.ListLayers()
	for _, layer := range layers {
		levels, _ := fcm.GetLayerActivationLevels(layer)
		fmt.Println(layer)
		sLevel := ""
		for k, v := range levels {
			sLevel += fmt.Sprintf("%s: %.4f ", k, v)
		}
		fmt.Println(sLevel)
	}
}

func BuildMLBasic() *MultilayerFuzzyCognitiveMap {
	indices := []string {"I", "II"}
	dimensions := []string{"levels"}
	allindices := make([][]string,1)
	allindices[0] = indices

	fcm:= NewMultilayerFuzzyCognitiveMap(dimensions, allindices,false, Logistic)

	fcm.AddConceptToLayer("A", "I", float32(1.0), float32(1.0), false)
	fcm.AddConceptToLayer("B", "I", float32(0.0), float32(0.0), false)
	fcm.AddConceptToLayer("C", "I", float32(0.0), float32(0.0), false)

	fcm.AddConceptToLayer("A", "II", float32(1.0), float32(1.0), false)
	fcm.AddConceptToLayer("D", "II", float32(0.0), float32(0.0), false);
	fcm.AddConceptToLayer("E", "II", float32(0.0), float32(0.0), false)

	fcm.AddInfluence("A", "I", "B", "I", float32(1.0))
	fcm.AddInfluence("A", "I", "C", "I", float32(1.0))
	fcm.AddInfluence("A", "II", "D", "II", float32(1.0))
	fcm.AddInfluence("D", "II", "E", "II", float32(1.0))
	fcm.AddInfluence("E", "II", "A", "I", float32(1.0))
	return fcm
}



