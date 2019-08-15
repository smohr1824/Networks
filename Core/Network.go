// Copyright 2017 - 2019  Stephen T. Mohr
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
	"errors"
	. "fmt"
	"bufio"
	"strconv"
	"sort"
)

type AdjacencyList struct {
	Weights map[uint32]float32
}
type Network struct {
	inEdges map[uint32] map[uint32]float32
	outEdges map[uint32] map[uint32]float32
	directed bool
}

func NewNetwork(directed bool) *Network{
	net := new(Network)
	net.inEdges = make(map[uint32] map[uint32]float32)
	net.outEdges = make(map[uint32] map[uint32]float32)
	net.directed = directed
	return net
}

func NewNetworkFromMatrix(vertices []uint32, weights [][]float32, directed bool) (*Network, error) {
	net := new(Network)
	net.inEdges = make(map[uint32] map[uint32]float32)
	net.outEdges = make(map[uint32] map[uint32]float32)
	net.directed = directed

	if vertices == nil {
		return nil, NewNetworkArgumentNullError("Vertex list must be non-null")
	}

	if weights == nil {
		return nil, NewNetworkArgumentNullError("Adjacency matrix must be non-null")
	}

	vertexCt := len(vertices)
	if vertexCt == 0 || vertexCt != len(weights) || vertexCt != len(weights[0]) {
		return nil, NewNetworkArgumentError(Sprintf("Adjacency matrix must be square, have the same dimensions as the vertex list, and be non-zero; vertices count: %d, weights row count: %d, weights column count: %d", vertexCt, len(weights), len(weights[0])))
	}

	for i:= 0; i < vertexCt; i++ {
		adjacencyList := make(map[uint32]float32)
		for k := 0; k < vertexCt; k++ {
			if k == i {
				continue
			}

			if weights[i][k] != 0 {
				adjacencyList[vertices[k]] = weights[i][k]
				row, contains := net.inEdges[vertices[k]]
				if !contains {
					// vertex not in InEdges
					inList := make(map[uint32]float32)
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

func (network *Network) Vertices(ordered bool) []uint32 {
	keys := make([]uint32, len(network.outEdges))
	i:=0
	for key := range network.outEdges{
		keys[i] = key
		i++
	}
	if !ordered {
		return keys
	} else {
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		return keys
	}
}

func (network *Network) Directed() bool {
	return network.directed
}

/* func (network *Network) Connected() bool {
	retVal := true
	for _, edges  := range network.inEdges {
		if len(edges) == 0 {
			return false
		}
	}

	return retVal
} */

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
			j := network.locOf(vertices, to)
			A[i][j] = wt
		}

	}

	// Need to pick up the in edges to reflect all neighbors in an undirected network
	if (!network.Directed()) {
		for i := range vertices {
			for to, wt := range network.inEdges[vertices[i]] {
				j := network.locOf(vertices, to)
				A[i][j] = wt
			}
		}
	}
	return A
}

func (network *Network) locOf(verts []uint32, vert uint32) int {
	for p, v := range verts {
		if v == vert {
			return p
		}
	}
	return -1
}

func (network *Network) AddVertex(id uint32) {
	_, contained := network.outEdges[id]
	if !contained {
		neighbors := make(map[uint32]float32)
		network.outEdges[id] = neighbors
	}

	_, contained = network.inEdges[id]
	if !contained {
		neighbors := make(map[uint32]float32)
		network.inEdges[id] = neighbors
	}
}

func (network *Network) RemoveVertex(id uint32) {
	tgts, contained := network.outEdges[id]
	if contained {
		for to := range tgts {
			delete(network.inEdges[to], id)
		}
		delete(network.outEdges, id)
	}

	srcs, contained := network.inEdges[id]
	if contained {
		for from := range srcs {
			delete(network.outEdges[from], id)
		}
		delete(network.inEdges,id)
	}
}

func (network *Network) AddEdge(from uint32, to uint32,  weight float32) error {
	if (from == to) {
		return NewNetworkArgumentError(Sprintf("Self-edges are not permitted (vertex %d)", from))
	}

	if network.HasEdge(from, to) {
		return nil // assume the user is happy to have the edge and is not trying for multiple edges
	}

	neighbors, contained := network.outEdges[from]

	if !contained {
		neighbors = make(map[uint32]float32)
		neighbors[to] = weight
		network.outEdges[from] = neighbors
		network.inEdges[from] = make(map[uint32]float32)
	} else {
		neighbors[to] = weight
	}

	// check for the existence of the to vertex and create if needed

	neighbors, contained = network.outEdges[to]
	if !contained {
		neighbors = make(map[uint32]float32)
		network.outEdges[to] = neighbors
		network.inEdges[to] = make(map[uint32]float32)
	}

	_, contained = network.inEdges[to]
	if contained {
		network.inEdges[to][from] = weight
	} else {
		newMap := make(map[uint32]float32)
		newMap[from] = weight
		network.inEdges[to] = newMap
	}

	return nil
}

func (network *Network) RemoveEdge(from uint32 , to uint32) {
	if network.HasEdge(from, to) {
		neighbors := network.outEdges[from]
		delete(neighbors, to)
		delete(network.inEdges[to], from)

	}
}

func (network *Network) GetNeighbors(vertex uint32) map[uint32]float32 {
	neighborsOut, containedOut := network.outEdges[vertex]
	neighborsIn, containedIn := network.inEdges[vertex]

	lenRetVal := 0
	if containedOut {
		lenRetVal += len(neighborsOut)
	}

	if !network.Directed() && containedIn {
		lenRetVal += len(neighborsIn)
	}
	retVal := make(map[uint32]float32, lenRetVal )

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

func (network *Network) GetSources(vertex uint32) map[uint32]float32 {
	ancestors, contained := network.inEdges[vertex]
	if !contained {
		return make(map[uint32]float32)
	} else {
		return ancestors
	}
}

func (network *Network) HasVertex(id uint32) bool{
	_, contained := network.outEdges[id]
	return contained
}

func (network *Network) HasEdge(from uint32, to uint32) bool {
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

func(network *Network) EdgeWeight(from uint32, to uint32) float32{
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

func (network *Network) Degree(vertex uint32) int {
	if !network.HasVertex(vertex) {
		return 0
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
	return retVal
}

func (network *Network) OutDegree(vertex uint32) int {
	if !network.HasVertex(vertex) {
		return 0
	}

	return len(network.outEdges[vertex])
}

func (network *Network) InDegree(vertex uint32) int {
	if !network.HasVertex(vertex) {
		return 0
	}

	return len(network.inEdges[vertex])
}

func (network *Network) InWeights(vertex uint32) float32 {
	if !network.HasVertex(vertex) {
		return 0.0
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
	return retVal
}

func (network *Network) OutWeights(vertex uint32) float32 {
	if !network.HasVertex(vertex) {
		return 0.0
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
	return retVal

}

func (network *Network) Clone() *Network {
	retVal := NewNetwork(network.directed)
	for key, val := range network.outEdges {
		targets := make(map[uint32]float32, len(val))
		for k, v := range val {
			targets[k] = v
		}
		retVal.outEdges[key] = targets

		sources := make(map[uint32]float32, len(network.inEdges[key]))
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
			_, _ = writer.WriteString(strconv.FormatUint(uint64(key), 10) + "\n")
		} else {
			for to, wt := range targets {
				//writer.WriteString(key + delimiter + to + delimiter + strconv.FormatFloat(float64(wt), 'f', -1, 64) + "\n")
				ln := Sprintf("%d%s%d%s%s\n",key,delimiter,to,delimiter,strconv.FormatFloat(float64(wt), 'f', -1, 32))
				_, _ = writer.WriteString(ln)
			}
		}
		writer.Flush()
	}
}

func (network *Network) ListGML(writer *bufio.Writer, level int) error {
	basicIndent := network.indentForLevel(level)
	_, err:= Fprintln(writer, basicIndent + "graph [")
	if err != nil {
		return err
	}
	if network.Directed() {
		_, err = Fprintln(writer, "\tdirected 1")
	} else {
		_, err = Fprintln(writer, "\tdirected 0")
	}
	if err != nil {
		return err
	}
	err = network.listGMLNodes(writer, basicIndent)
	if (err != nil) {
		return err
	}
	err = network.listGMLEdges(writer, basicIndent)
	if err != nil {
		return err
	}
	_, _ = Fprintln(writer, basicIndent+"]")
	return nil
}

func (network *Network) listGMLNodes(writer *bufio.Writer, indent string) error {
	for _, v := range network.Vertices(true) {
		_, _ = Fprintln(writer, indent+"\tnode [")
		_, _ = Fprintln(writer, indent+Sprintf("\t\tid %d", v))
		_, _ = Fprintln(writer, indent + "\t]")
	}
	return nil
}

func (network *Network) listGMLEdges(writer *bufio.Writer, indent string) error {
	for k, v := range network.outEdges {
		for to, wt := range v {
			Fprintln(writer, indent + "\tedge [")
			Fprintln(writer, indent + "\t\tsource " + Sprintf("%d", k))
			Fprintln(writer, indent + "\t\ttarget " + Sprintf("%d", to))
			Fprintln(writer, indent + "\t\tweight " +Sprintf("%f", wt))
			Fprintln(writer, indent + "\t]")
		}
	}
	return nil
}

func (network *Network) indentForLevel(level int) string {
	retVal := ""
	for i:= 0; i < level; i++ {
		retVal += "\t"
	}
	return retVal
}

func (network *Network) StartingVertex(connected bool) (uint32, error) {
	// map keys in Go are always randomized, so grab the first vertex with outgoing edges in the iteration
	for key, val := range network.outEdges {
		if (connected){
			// if we are looking for a vertex with outgoing edges, check the length of the adjacency list
			if len(val) > 0 {
				return key, nil
			}
		} else {
			// ...otherwise, map key iteration is randomized in Go, so return the first one
			return key, nil
		}
	}

	return 0, errors.New("Requested connected starting vertex in a disconnected network")
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

// end utilities
