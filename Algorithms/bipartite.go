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

package Algorithms

import (
	"github.com/smohr1824/DataStructures"
	"github.com/smohr1824/Networks/Core"
)

const (
	red = iota
	blue
)

type coloring struct {
	vertex uint32
	color  uint8
}

// entry point for concurrent bipartite discovery
// The network will only be read from, not written to
// routineCount is the number of concurrent goroutines to use and should be approximately equal to the average degree of the network
func ConcurrentBipartite(G *Core.Network, routineCount int) (bool, []uint32, []uint32) {
	maxSize := G.Order()
	R := make([]uint32, 0, maxSize)
	B := make([]uint32, 0, maxSize)
	colorings := make(map[uint32]uint8, maxSize)

	// worklist is the queue of vertices on the frontier (BFS)
	worklist := DataStructures.NewQueue()
	isBipartite := true

	start, err := G.StartingVertex(true)
	if err != nil {
		return false, nil, nil
	}


	// color the starting vertex and load it into the queue
	colorings[start] = red
	R = append(R, start)
	startColor := coloring{start, red}
	initial := make([]coloring, 1)
	initial[0] = startColor
	worklist.Push(initial)
	nextChannel := 0
	order := G.Order()

	// coloringChannel receives arrays of colorings from goroutines
	// assignments is an array of the channels for sending a single assignment to a goroutine
	coloringChannel := make(chan []coloring, routineCount*5)

	assignments := prepareWorkChannels(routineCount)
	defer func() {
		close(coloringChannel)
	}()

	isBipartite = true

	// start the goroutines
	for i := 0; i < routineCount; i++ {
		go serviceAssignments(G, assignments[i], coloringChannel)
	}

	// start processing
	for len(colorings) < order {

		// assign any available work items
		qlen := worklist.Length()
		if qlen > 0 {
			assignment := worklist.Pop()
			assignments[nextChannel] <- assignment.([]coloring)
			nextChannel++
			if nextChannel >= routineCount {
				nextChannel = 0
			}
		}

		select {
		// process colorings from goroutines
		case coloringMsg := <-coloringChannel:
			trigger := processColorings(coloringMsg, colorings, &R, &B, worklist)
			if trigger {
				closeChannels(assignments)
				return false, nil, nil
			}
		}
	}

	// close the work assignment channels to cause the goroutines to terminate
	closeChannels(assignments)
	if isBipartite {
		return isBipartite, R, B
	} else {
		return false, nil, nil
	}
}

// create an array of channels for passing work assignments to goroutines
func prepareWorkChannels(count int) []chan []coloring {
	assigners := make([]chan []coloring, count)
	for i := 0; i < count; i++ {
		assigners[i] = make(chan []coloring, 5)
	}

	return assigners
}

func closeChannels(channels []chan []coloring) {
	for i := 0; i < len(channels); i++ {
		close(channels[i])
	}
}

// goroutine enumerates the neighbors, assigns a color, and sends it back to main for review
func serviceAssignments(G *Core.Network, localAssignmentChannel <-chan []coloring, coloringChannel chan<- []coloring) {
	// depending on the timing, a goroutine may send on the main coloring channel after it has been closed, hence this call
	defer func() { recover() } ()
	assignments, ok := <-localAssignmentChannel
	for ok {
		for _, assigned := range assignments {
			parentVertex := assigned.vertex
			parentColor := assigned.color

			// get the neighbors, turn them into colorings, and send the array back to main
			var tocolor uint8
			if parentColor == red {
				tocolor = blue
			} else {
				tocolor = red
			}

			neighbors := G.GetNeighbors(parentVertex)

			// pick up sources to ensure reachability in directed graphs
			if G.Directed() {
				predecessors := G.GetSources(parentVertex)
				for key, value := range predecessors {
					neighbors[key] = value
				}
			}

			newassignments := make([]coloring, len(neighbors))
			i := 0
			for key := range neighbors {
				colorassignment := coloring{key, tocolor}
				newassignments[i] = colorassignment
				i++
			}

			coloringChannel <- newassignments
		}
		assignments, ok = <-localAssignmentChannel
	}
}

// go through an array of proposed colorings from a goroutine
// if not previously seen, add to the map and add it to the work queue
// if seen, make sure there is no conflict, but do not process further
// If a conflict is seen, the graph is not bipartite.
func processColorings(assignedColors []coloring, masterColors map[uint32]uint8, R *[]uint32, B *[]uint32, queue *DataStructures.Queue) bool {
	filteredColorings := make([]coloring, 0, len(assignedColors))
	for _, colored := range assignedColors {
		color, ok := masterColors[colored.vertex]
		if !ok {
			masterColors[colored.vertex] = colored.color
			if colored.color == red {
				*R = append(*R, colored.vertex)
			} else {
				*B = append(*B, colored.vertex)
			}
			//queue.Push(colored)
			filteredColorings = append(filteredColorings, colored)
		} else {
			// found, check for conflict, abort all if there is a conflict
			if color != colored.color {
				return true
			}
		}

	}
	queue.Push(filteredColorings)
	return false
}
