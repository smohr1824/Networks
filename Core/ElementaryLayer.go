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

package Core
import (
	"bufio"
	. "fmt"
	"strconv"
	"strings"
)

type elementaryLayer struct {
	edgeList map[uint32] map[resolvedNodeLayerTuple]float32
	inEdges map[uint32] map[resolvedNodeLayerTuple]float32
	g *Network
	m *MultilayerNetwork
	layerCoordinates string

}

func NewElementaryLayer(M *MultilayerNetwork, G *Network, coordinates string) *elementaryLayer{
	p := new(elementaryLayer)
	p.m = M
	p.g = G
	p.edgeList = make(map[uint32] map[resolvedNodeLayerTuple]float32)
	p.inEdges = make(map[uint32] map[resolvedNodeLayerTuple]float32)
	p.layerCoordinates = coordinates
	return p
}

func (p *elementaryLayer) Vertices(ordered bool) []uint32 {
	return p.g.Vertices(ordered)
}

func (p *elementaryLayer) AspectCoordinates() string {
	return p.m.UnaliasCoordinates(p.layerCoordinates)
}

func (p *elementaryLayer) HasEdge(from resolvedNodeLayerTuple, to resolvedNodeLayerTuple) bool {
	if from.IsSameElementaryLayer(to) {
		return p.g.HasEdge(from.NodeId, to.NodeId)
	} else {
		_, contained := p.edgeList[from.NodeId]
		if contained {
			_, contained = p.edgeList[from.NodeId]
			if contained {
				return true
			} else {
				return false
			}
		} else {
			return false
		}
	}
}

func (p *elementaryLayer) InterlayerAdjacencies(toCoords string) [][]float32 {
	// allocate the matrix
	size := p.g.Order()
	vertices := p.Vertices(true)
	retVal := make([][]float32, size)
	for i:= range retVal {
		retVal[i] = make([]float32, size)
	}
	sCoords := strings.Split(toCoords, ",")
	to := make([]int, len(sCoords))
	for i := 0; i < len(sCoords); i++ {
		to[i], _ = strconv.Atoi(sCoords[i])
	}
	for from, _ := range p.edgeList {
		adjList := p.edgeList[from]
		fromIndex := p.locOf(vertices, from)
		for tgt, _ := range adjList {
			if tgt.AreSameElementaryLayer(to){
				toIndex := p.locOf(vertices, tgt.NodeId)
				retVal[fromIndex][toIndex] = adjList[tgt]
			}
		}
	}
	return retVal
}

func (p *elementaryLayer) LayerAdjacencyMatrix() [][]float32 {
	return p.g.AdjacencyMatrix()
}

func (p *elementaryLayer) CopyNetwork() *Network {
	return p.g.Clone()
}

func (p *elementaryLayer) HasVertex(vertex uint32) bool {
	return p.g.HasVertex(vertex)
}

func (p *elementaryLayer) AddVertex(vertex uint32) {
	p.g.AddVertex(vertex)
}

func (p *elementaryLayer) Order() int {
	return p.g.Order()
}

func (p *elementaryLayer) Degree(vertex uint32) int {
	return p.g.Degree(vertex)
}
func (p *elementaryLayer) InterlayerDegree(vertex uint32) int {
	if !p.HasVertex(vertex) {
		return 0;
	}

	retVal := 0

	edges, ok := p.edgeList[vertex]
	if ok {
		retVal += len(edges)
	}

	edges, ok = p.inEdges[vertex]
	if ok {
		retVal += len(edges)
	}

	return retVal
}

func (p *elementaryLayer) InDegree(vertex uint32) int {
	retVal := p.g.InDegree(vertex)

	ins, ok := p.inEdges[vertex]
	if ok {
		retVal += len(ins)
	}

	return retVal
}

func (p *elementaryLayer) OutDegree(vertex uint32) int {
	retVal := p.g.OutDegree(vertex)
	outs, ok := p.g.outEdges[vertex]
	if ok {
		retVal += len(outs)
	}

	return retVal
}

func (p *elementaryLayer) RemoveVertex(vertex uint32) {
	p.g.RemoveVertex(vertex)

	// remove any interlayer edges
	nodeT := newresolvedNodeLayerTuple(vertex, p.layerCoordinates)
	_, ok := p.edgeList[vertex]
	if ok {
		delete(p.edgeList, vertex)
	}

	edges, ok := p.inEdges[vertex]
	if ok {
		for k, _ := range edges {
			p.m.removeOutEdge(k, *nodeT)
		}
	}
}

func (p *elementaryLayer) AddInEdge(from resolvedNodeLayerTuple, to resolvedNodeLayerTuple, wt float32) {
	if p.HasVertex(from.NodeId) {
		v, ok := p.inEdges[from.NodeId]
		if ok {
			v[to] = wt
		} else {
			m := make(map[resolvedNodeLayerTuple] float32)
			m[to] = wt
			p.inEdges[from.NodeId] = m
		}
	}
}

func (p *elementaryLayer) RemoveOutEdge(from resolvedNodeLayerTuple, to resolvedNodeLayerTuple) {
	edges, ok := p.edgeList[from.NodeId]
	if ok {
		delete(edges, to)
	}
}

func (p *elementaryLayer) RemoveInEdge(tgt resolvedNodeLayerTuple, src resolvedNodeLayerTuple) {
	edges, ok := p.inEdges[tgt.NodeId]
	if ok {
		delete(edges, src)
	}
}

func (p *elementaryLayer) EdgeWeight(from resolvedNodeLayerTuple, to resolvedNodeLayerTuple) float32 {
	if from.IsSameElementaryLayer(to) {
		return p.g.EdgeWeight(from.NodeId, to.NodeId)
	} else {
		edges, ok := p.edgeList[from.NodeId]
		if ok {
			wt, ook := edges[to]
			if ook {
				return wt
			} else {
				return 0.0
			}
		} else {
			return 0.0
		}
	}
}

func (p *elementaryLayer) AddEdge(from resolvedNodeLayerTuple, to resolvedNodeLayerTuple, wt float32) {
	if from.Coordinates != p.layerCoordinates {
		return
	} else {
		if from.IsSameElementaryLayer(to) {
			p.g.AddEdge(from.NodeId, to.NodeId, wt)
		} else {
			edges, ok := p.edgeList[from.NodeId]
			if ok {
				edges[to] = wt
			} else {
				d := make(map[resolvedNodeLayerTuple]float32)
				d[to] = wt
				p.edgeList[from.NodeId] = d
			}
		}
	}
}

func (p *elementaryLayer) RemoveEdge(from resolvedNodeLayerTuple, to resolvedNodeLayerTuple) {
	if from.Coordinates != p.layerCoordinates {
		return
	} else {
		if from.IsSameElementaryLayer(to) {
			p.g.RemoveEdge(from.NodeId, to.NodeId)
		} else {
			edges, ok := p.edgeList[from.NodeId]
			if ok {
				delete(edges, to)
			}
		}
	}
}

func (p *elementaryLayer) GetNeighbors(vertex uint32) map[NodeLayerTuple] float32 {
	retVal := make(map[NodeLayerTuple] float32)
	if !p.HasVertex(vertex) {
		return retVal
	}

	graphNeighbors := p.g.GetNeighbors(vertex)
	layerAspectCoords := p.m.UnaliasCoordinates(p.layerCoordinates)
	for node, _ := range graphNeighbors {
		local := NewNodeLayerTuple(node, layerAspectCoords)
		retVal[*local] = graphNeighbors[node]
	}

	// add interlayer targets
	edges, ok := p.edgeList[vertex]
	if ok {
		for tuple, _ := range edges {
			tgt := NewNodeLayerTuple(tuple.NodeId, p.m.UnaliasCoordinates(tuple.Coordinates))
			retVal[*tgt] = edges[tuple]
		}
	}

	return retVal
}

func (p *elementaryLayer) GetSources(vertex uint32) map[NodeLayerTuple] float32 {
	retVal := make(map[NodeLayerTuple] float32)

	if !p.HasVertex(vertex) {
		return retVal
	}

	graphSources := p.g.GetSources(vertex)
	layerAspectCoordinates := p.m.UnaliasCoordinates(p.layerCoordinates)

	for node, _ := range graphSources {
		local := NewNodeLayerTuple(node, layerAspectCoordinates)
		retVal[*local] = graphSources[node]
	}

	// add interlayer sources
	inedges, ok := p.inEdges[vertex]
	if ok {
		for tuple, _ := range inedges {
			tgt := resolvedNodeLayerTuple {NodeId: tuple.NodeId, Coordinates: tuple.Coordinates}
			inedges[tgt] = inedges[tuple]
		}
	}

	return retVal
}

func (p *elementaryLayer) List(writer *bufio.Writer, delimiter string) {
	p.g.List(writer, delimiter)

	if len(p.edgeList) > 0 {
		Fprintln(writer, "Interlayer edges")
	}

	for from, targets := range p.edgeList {
		for to, wt := range targets {
			Fprintln(writer, strconv.Itoa(int(from)) + ":" + p.m.UnaliasCoordinates(p.layerCoordinates) + delimiter + strconv.Itoa(int(to.NodeId)) + p.m.UnaliasCoordinates(to.Coordinates) + delimiter + Sprintf("%f", wt))
		}
	}
}

func (p *elementaryLayer) ListGML(writer *bufio.Writer, level int) {
	_ = p.g.ListGML(writer, level)
}

func (p *elementaryLayer) ListInterlayerGML(writer *bufio.Writer) {
	indent := "\t"
	for from, targets := range p.edgeList {
		for to, wt :=range targets {
			Fprintln(writer, indent+"edge [")

			Fprintln(writer, indent+"\tsource [")
			Fprintln(writer, indent+"\t\tid "+ strconv.Itoa(int(from)))
			Fprintln(writer, indent + "\t\tcoordinates " + p.m.UnaliasCoordinates(p.layerCoordinates))
			Fprintln(writer, indent + "\t]")

			Fprintln(writer, indent + "\ttarget [")
			Fprintln(writer, indent + "\t\tid " + strconv.Itoa(int(to.NodeId)))
			Fprintln(writer, indent + "\t\tcoordinates " + p.m.UnaliasCoordinates(to.Coordinates))
			Fprintln(writer, indent + "\t]")

			Fprintln(writer, indent + "\tweight " + Sprintf("%f", wt))
			Fprintln(writer, indent + "]")
		}
	}
}

func (p *elementaryLayer) locOf(verts []uint32, vert uint32) int {
	for p, v := range verts {
		if v == vert {
			return p
		}
	}
	return -1
}
