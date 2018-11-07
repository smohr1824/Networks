// Copyright 2017 - 2018  Stephen T. Mohr
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

// Basic graph/network capabilities

package Core

import (
	"fmt"
	"bufio"
	"strconv"
	"sort"
)

type AdjacencyList struct {
	Weights map[string]float32
}
type Network struct {
	inEdges map[string] map[string]float32
	outEdges map[string] map[string]float32
	directed bool
}

func NewNetwork(directed bool) *Network{
	net := new(Network)
	net.inEdges = make(map[string] map[string]float32)
	net.outEdges = make(map[string] map[string]float32)
	net.directed = directed
	return net
}

func NewNetworkFromMatrix(vertices []string, weights [][]float32, directed bool) (*Network, error) {
	net := new(Network)
	net.inEdges = make(map[string] map[string]float32)
	net.outEdges = make(map[string] map[string]float32)
	net.directed = directed

	if vertices == nil {
		return nil, NewNetworkArgumentNullError("Vertex list must be non-null")
	}

	if weights == nil {
		return nil, NewNetworkArgumentNullError("Adjacency matrix must be non-null")
	}

	vertexCt := len(vertices)
	if vertexCt == 0 || vertexCt != len(weights) || vertexCt != len(weights[0]) {
		return nil, NewNetworkArgumentError(fmt.Sprintf("Adjacency matrix must be square, have the same dimensions as the vertex list, and be non-zero; vertices count: %d, weights row count: %d, weights column count: %d", vertexCt, len(weights), len(weights[0])))
	}

	for i:= 0; i < vertexCt; i++ {
		adjacencyList := make(map[string]float32)
		for k := 0; k < vertexCt; k++ {
			if k == i {
				continue
			}

			if weights[i][k] != 0 {
				adjacencyList[vertices[k]] = weights[i][k]
				row, contains := net.inEdges[vertices[k]]
				if !contains {
					// vertex not in InEdges
					inList := make(map[string]float32)
					inList[vertices[k]] = weights[i][k]
					net.inEdges[vertices[k]] = inList
				} else {
					row[vertices[i]] = weights[i][k]
				}
			}


		}
		net.outEdges[vertices[i]] = adjacencyList
	}
	return net, nil
}
// public

func (network *Network) Vertices(ordered bool) []string {
	keys := make([]string, len(network.outEdges))
	i:=0
	for key := range network.outEdges{
		keys[i] = key
		i++
	}
	if !ordered {
		return keys
	} else {
		sort.Strings(keys)
		return keys
	}
}

func (network *Network) Directed() bool {
	return network.directed
}

func (network *Network) Connected() bool {
	retVal := true
	for edges := range network.outEdges {
		if len(edges) == 0 {
			retVal = false
		}
	}

	return retVal
}

func (network *Network) Order() int {
	return len(network.outEdges)
}

func (network *Network) Density() float64 {
	edgeCt := network.countEdges()
	order := len(network.outEdges)
	retVal := float64(edgeCt)/float64(order * (order - 1))
	if (!network.Directed()) {
		retVal = 2 * retVal
	}

	return retVal
}

func (network *Network) Size() int {
	return network.countEdges()
}

func (network *Network) AdjacencyMatrix() [][]float32 {
	order := network.Order()
	A:=[][]float32{}
	for i := 0; i < order; i++ {
		A = append(A, make([]float32, order))
	}

	vertices := network.Vertices(true)

	for i := range vertices { // row = i
		for to, wt := range network.outEdges[vertices[i]] {
			j := indexOf(vertices, to)
			A[i][j] = wt
		}

	}

	// Need to pick up the in edges to reflect all neighbors in an undirected network
	if (!network.Directed()) {
		for i := range vertices {
			for to, wt := range network.inEdges[vertices[i]] {
				j := indexOf(vertices, to)
				A[i][j] = wt
			}
		}
	}
	return A
}

func (network *Network) AddVertex(id string) {
	_, contained := network.outEdges[id]
	if !contained {
		neighbors := make(map[string]float32)
		network.outEdges[id] = neighbors
	}

	_, contained = network.inEdges[id]
	if !contained {
		neighbors := make(map[string]float32)
		network.inEdges[id] = neighbors
	}
}

func (network *Network) RemoveVertex(id string) {
	tgts, contained := network.outEdges[id]
	if contained {
		for to, _ := range tgts {
			delete(network.inEdges[to], id)
		}
		delete(network.outEdges, id)
	}

	srcs, contained := network.inEdges[id]
	if contained {
		for from, _ := range srcs {
			delete(network.outEdges[from], id)
		}
		delete(network.inEdges,id)
	}
}

func (network *Network) AddEdge(from string, to string,  weight float32) error {
	if (from == to) {
		return NewNetworkArgumentError(fmt.Sprintf("Self-edges are not permitted (vertex %s)", from))
	}

	if network.HasEdge(from, to) {
		return nil // assume the user is happy to have the edge and is not trying for multiple edges
	}

	neighbors, contained := network.outEdges[from]

	if !contained {
		neighbors = make(map[string]float32)
		neighbors[to] = weight
		network.outEdges[from] = neighbors
		network.inEdges[from] = make(map[string]float32)
	} else {
		neighbors[to] = weight
	}

	// check for the existence of the to vertex and create if needed

	neighbors, contained = network.outEdges[to]
	if !contained {
		neighbors = make(map[string]float32)
		network.outEdges[to] = neighbors
		network.inEdges[to] = make(map[string]float32)
	}

	_, contained = network.inEdges[to]
	if contained {
		network.inEdges[to][from] = weight
	} else {
		newMap := make(map[string]float32)
		newMap[from] = weight
		network.inEdges[to] = newMap
	}

	return nil
}

func (network *Network) RemoveEdge(from string , to string) {
	if network.HasEdge(from, to) {
		neighbors := network.outEdges[from]
		delete(neighbors, to)
		delete(network.inEdges[to], from)

	}
}

func (network *Network) GetNeighbors(vertex string) map[string]float32 {
	neighborsOut, containedOut := network.outEdges[vertex]
	neighborsIn, containedIn := network.inEdges[vertex]

	lenRetVal := 0
	if containedOut {
		lenRetVal += len(neighborsOut)
	}

	if !network.Directed() && containedIn {
		lenRetVal += len(neighborsIn)
	}
	retVal := make(map[string]float32, lenRetVal )

	for to, wt := range neighborsOut {
		retVal[to] = wt
	}

	if (!network.Directed() && containedIn) {
		for to, wt := range neighborsIn {
			retVal[to] = wt
		}
	}

	return retVal
}

func (network *Network) GetSources(vertex string) map[string]float32 {
	ancestors, contained := network.inEdges[vertex]
	if !contained {
		return make(map[string]float32)
	} else {
		return ancestors
	}
}

func (network *Network) HasVertex(id string) bool{
	_, contained := network.outEdges[id]
	return contained
}

func (network *Network) HasEdge(from string, to string) bool {
	adjList, contained := network.outEdges[from]
	if contained {
		_, contained2 := adjList[to]
		if (contained2) {
			return true
		} else {
			if !network.Directed() {
				adjList, contained := network.inEdges[from]
				if contained {
					_, contained2 := adjList[to]
					if contained2 {
						return true
					} else {
						return false
					}
				}
			} else {
				return false
			}
		}
	} else {
		return false
	}

	return false
}

func(network *Network) EdgeWeight(from string, to string) float32{
	if network.HasEdge(from, to){
		wt, contained := network.outEdges[from][to]
		if contained {
			return wt
		} else {
			if !network.Directed() {
				return network.inEdges[from][to]
			} else {
				return 0.0
			}
		}
	} else {
		return 0.0
	}
}

func (network *Network) Degree(vertex string) (int, error) {
	if !network.HasVertex(vertex) {
		return 0, NewNetworkArgumentError(fmt.Sprintf("Vertex %s is not a member of this network", vertex))
	}

	// return the sum of the in and out edges
	retVal := 0
	_, contained := network.outEdges[vertex]
	if contained {
		retVal = len(network.outEdges[vertex])
	}
	_, contained = network.inEdges[vertex]
	if contained {
		retVal += len(network.inEdges[vertex])
	}
	return retVal, nil
}

func (network *Network) OutDegree(vertex string) (int, error) {
	if !network.HasVertex(vertex) {
		return 0, NewNetworkArgumentError(fmt.Sprintf("Vertex %s is not a member of the network", vertex))
	}

	return len(network.outEdges[vertex]), nil
}

func (network *Network) InDegree(vertex string) (int, error) {
	if !network.HasVertex(vertex) {
		return 0, NewNetworkArgumentError(fmt.Sprintf("Vertex %s is not a member of the network", vertex))
	}

	return len(network.inEdges[vertex]), nil
}

func (network *Network) InWeights(vertex string) (float32, error) {
	if !network.HasVertex(vertex) {
		return 0.0, NewNetworkArgumentError(fmt.Sprintf("Vertex %s is not part of the network", vertex))
	}

	var retVal float32 = 0.0


	for _, val := range network.inEdges[vertex] {
		retVal += val
	}
	if !network.directed {
		for _, val := range network.outEdges[vertex] {
			retVal += val
		}
	}
	return retVal, nil
}

func (network *Network) OutWeights(vertex string) (float32, error) {
	if !network.HasVertex(vertex) {
		return 0.0, NewNetworkArgumentError(fmt.Sprintf("Vertex %s is not part of the network", vertex))
	}

	var retVal float32 = 0.0

	for _, val := range network.outEdges[vertex] {
		retVal += val
	}

	if !network.Directed() {
		for _, val := range network.inEdges[vertex] {
			retVal += val
		}
	}
	return retVal, nil

}

func (network *Network) Clone() *Network {
	retVal := NewNetwork(network.directed)
	for key, val := range network.outEdges {
		targets := make(map[string]float32, len(val))
		for k, v := range val {
			targets[k] = v
		}
		retVal.outEdges[key] = targets

		sources := make(map[string]float32, len(network.inEdges[key]))
		for k, v := range network.inEdges[key] {
			sources[k] = v
		}
		retVal.inEdges[key] = targets
	}

	return retVal

}

func (network *Network) List(writer *bufio.Writer, delimiter string) {
	for key, targets := range network.outEdges {
		if len(targets) == 0 {
			writer.WriteString(key + "\n")
		} else {
			for to, wt := range targets {
				//writer.WriteString(key + delimiter + to + delimiter + strconv.FormatFloat(float64(wt), 'f', -1, 64) + "\n")
				ln := fmt.Sprintf("%s%s%s%s%s\n",key,delimiter,to,delimiter,strconv.FormatFloat(float64(wt), 'f', -1, 32))
				writer.WriteString(ln)
			}
		}
		writer.Flush()
	}
}

func (network *Network) StartingVertex(connected bool) string {
	retVal := ""

	// map keys in Go are always randomized, so grab the first vertex with outgoing edges in the iteration
	for key, val := range network.outEdges {
		if (connected){
			// if we are looking for a vertex with outgoing edges, check the length of the adjacency list
			if len(val) > 0 {
				retVal = key
				break
			}
		} else {
			// ...otherwise, map key iteration is randomized in Go, so return the first one
			retVal = key
			break
		}
	}

	return retVal
}

// end public

// utilities

func(network *Network) countEdges()	int {
	edgeCt := 0
	for _, neighbors := range network.outEdges {
		edgeCt += len(neighbors)
	}
	return edgeCt
}

func indexOf(vertexList []string, vertex string) int {
	for k, v := range vertexList {
		if v == vertex {
			return k
		}
	}
	return -1
}

	// end utilities
