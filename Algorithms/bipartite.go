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
	vertex string
	color  uint8
}

// entry point for concurrent bipartite discovery
// The network will only be read from, not written to
// routineCount is the number of concurrent goroutines to use and should be approximately equal to the average order of the network
func ConcurrentBipartite(G *Core.Network, routineCount int) (bool, []string, []string) {
	maxSize := G.Order()
	R := make([]string, 0, maxSize)
	B := make([]string, 0, maxSize)
	colorings := make(map[string]uint8, maxSize)

	// worklist is the queue of vertices on the frontier (BFS)
	worklist := DataStructures.NewQueue()

	isBipartite := true

	start := G.StartingVertex(true)
	if start == "" {
		return false, nil, nil
	}

	// color the starting vertex and load it into the queue
	colorings[start] = red
	R = append(R, start)
	startColor := coloring{start, red}
	worklist.Push(startColor)
	nextChannel := 0
	order := G.Order()

	// coloringChannel receives arrays of colorings from goroutines
	// assignments is an array of the channels for sending a single assignment to a goroutine
	coloringChannel := make(chan []coloring, routineCount)
	assignments := prepareWorkChannels(routineCount)
	defer closeAllChannels(assignments, coloringChannel)

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
			assignments[nextChannel] <- assignment.(coloring)
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
				// closeChannels(assignments)
				return false, nil, nil
			}
		}
	}

	if isBipartite {
		return isBipartite, R, B
	} else {
		return false, nil, nil
	}
}

// create an array of channels for passing work assignments to goroutines
func prepareWorkChannels(count int) []chan coloring {
	assigners := make([]chan coloring, count)
	for i := 0; i < count; i++ {
		assigners[i] = make(chan coloring, 5)
	}

	return assigners
}

func closeChannels(channels []chan coloring) {
	for i := 0; i < len(channels); i++ {
		close(channels[i])
	}
}

func closeAllChannels(channels []chan coloring, mainchan chan []coloring){
	closeChannels(channels)
	close(mainchan)
}

// goroutine enumerates the neighbors, assigns a color, and sends it back to main for review
func serviceAssignments(G *Core.Network, localAssignmentChannel <-chan coloring, coloringChannel chan<- []coloring) {
	assigned, ok := <-localAssignmentChannel
	for ok {
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

		assignments := make([]coloring, len(neighbors))
		i := 0
		for keyddfgd := range neighbors {
			colorassignment := coloring{key, tocolor}
			assignments[i] = colorassignment
			i++
		}
		coloringChannel <- assignments
		assigned, ok = <-localAssignmentChannel
	}
}

// go through an array of proposed colorings from a goroutine
// if not previously seen, add to the map and add it to the work queue
// if seen, make sure there is no conflict, but do not process further
// If a conflict is seen, the graph is not bipartite.
func processColorings(assignedColors []coloring, masterColors map[string]uint8, R *[]string, B *[]string, queue *DataStructures.Queue) bool {
	for _, colored := range assignedColors {
		color, ok := masterColors[colored.vertex]
		if !ok {
			masterColors[colored.vertex] = colored.color
			if colored.color == red {
				*R = append(*R, colored.vertex)
			} else {
				*B = append(*B, colored.vertex)
			}
			queue.Push(colored)
		} else {
			// found, check for conflict, abort all if there is a conflict
			if color != colored.color {
				return true
			}
		}

	}
	return false
}
