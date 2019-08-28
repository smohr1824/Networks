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
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
	. "fmt"
)

type MultilayerNetwork struct {
	aspects []string  // e.g., {"location","type"}
	indices [][]string  // e.g.,
	directed bool
	elementaryLayers map[string] *elementaryLayer
	nodeIdsAndLayers map[uint32] []*elementaryLayer
}

func NewMultilayerNetwork(aspects []string, indices[][]string, isdirected bool) *MultilayerNetwork {
	p := new(MultilayerNetwork)
	p.directed = isdirected
	if aspects != nil && indices != nil {
		p.aspects = make([]string, len(aspects))
		p.indices = make([][]string, len(aspects))
		for i := 0; i < len(aspects); i++ {
			p.aspects[i] = aspects[i]
			for k, idxs := range indices {
				p.indices[k] = idxs
			}
		}
	}
	p.elementaryLayers = make(map[string] *elementaryLayer)
	p.nodeIdsAndLayers = make(map[uint32] []*elementaryLayer)

	return p
}

func (p *MultilayerNetwork) ElementaryLayers() []string {
	layers := make([]string, len(p.elementaryLayers))
	i := 0
	for _, layer := range p.elementaryLayers {
		layers[i] = p.UnaliasCoordinates(layer.layerCoordinates)
		i++
	}
	sort.Slice(layers, func(i, j int) bool { return layers[i] < layers[j] })
	return layers
}

func (p *MultilayerNetwork) Order() int {
	return len(p.nodeIdsAndLayers)
}

func (p *MultilayerNetwork) Aspects() []string {
	return p.aspects
}

func (p *MultilayerNetwork) Indices(aspect string) []string {
	index := p.locOf(p.aspects, aspect)
	if index == -1 {
		return make([]string, 0)
	} else {
		return p.indices[index]
	}
}

func (p *MultilayerNetwork) UniqueVertices() []uint32 {
	keys := reflect.ValueOf(p.nodeIdsAndLayers).MapKeys()
	verts := make([]uint32, len(keys))
	for i := 0; i < len(keys); i++ {
		verts[i] = uint32(keys[i].Int())
	}
	return verts
}

func (p *MultilayerNetwork) IsNodeAligned() bool {
	vertexGlobalCount := len(p.nodeIdsAndLayers)

	for _, layer := range p.elementaryLayers {
		if layer.Order() != vertexGlobalCount {
			return false
		}
	}
	return true
}

func (p *MultilayerNetwork) Degree(vertex NodeLayerTuple) int {
	coords, err := p.resolveCoordinates(vertex.Coordinates)
	if err == nil {
		layer, ok := p.elementaryLayers[coords]
		if ok {
			retVal := layer.Degree(vertex.NodeId)
			retVal += layer.InterlayerDegree(vertex.NodeId)
			return retVal
		} else {
			return 0
		}
	} else {
		return 0
	}
}

func (p *MultilayerNetwork) InDegree(vertex NodeLayerTuple) int {
	coords, err := p.resolveCoordinates(vertex.Coordinates)
	if err == nil {
		layer, ok := p.elementaryLayers[coords]
		if ok {
			return layer.InDegree(vertex.NodeId)
		} else {
			return 0
		}
	} else {
		return 0
	}
}

func (p *MultilayerNetwork) OutDegree(vertex NodeLayerTuple) int {
	coords, err := p.resolveCoordinates(vertex.Coordinates)
	if err == nil {
		layer, ok := p.elementaryLayers[coords]
		if ok {
			return layer.OutDegree(vertex.NodeId)
		} else {
			return 0
		}
	} else {
		return 0
	}
}

func (p *MultilayerNetwork) HasElementaryLayer(coords string) bool {
	rcoords, err := p.resolveCoordinates(coords)
	if err != nil {
		return false
	} else {
		return p.elementaryLayerExists(rcoords)
	}
}

func (p *MultilayerNetwork) VerticesInLayer(coords string) ([]uint32, error) {
	rcoords, err := p.resolveCoordinates(coords)
	if err == nil {
		layer, ok := p.elementaryLayers[rcoords]
		if ok {
			return layer.Vertices(true), nil
		} else {
			return nil, errors.New(Sprintf("Layer %s not found in network", coords))
		}
	} else {
		return nil, errors.New(Sprintf("Layer %s not found in network", coords))
	}
}

func (p *MultilayerNetwork) HasVertex(vertex NodeLayerTuple) bool {
	rcoords, err := p.resolveCoordinates(vertex.Coordinates)
	if err == nil {
		layer, ok := p.elementaryLayers[rcoords]
		if ok {
			return layer.HasVertex(vertex.NodeId)
		} else {
			return false
		}
	} else {
		return false
	}
}

func (p *MultilayerNetwork) HasEdge(from NodeLayerTuple, to NodeLayerTuple) bool {
	rcoordsFrom, err1 := p.resolveCoordinates(from.Coordinates)
	rcoordsTo, err2 := p.resolveCoordinates(to.Coordinates)
	if err1 == nil && err2 == nil {
		layer, ok := p.elementaryLayers[rcoordsFrom]
		if ok {
			return layer.HasEdge(resolvedNodeLayerTuple{NodeId: from.NodeId, Coordinates: rcoordsFrom}, resolvedNodeLayerTuple{NodeId: to.NodeId, Coordinates: rcoordsTo})
		} else {
			return false
		}
	} else {
		return false
	}
}

func (p *MultilayerNetwork) EdgeWeight(from NodeLayerTuple, to NodeLayerTuple) float32 {
	rcoordsFrom, err1 := p.resolveCoordinates(from.Coordinates)
	rcoordsTo, err2 := p.resolveCoordinates(to.Coordinates)
	if err1 == nil && err2 == nil {
		layer, ok := p.elementaryLayers[rcoordsFrom]
		if ok {
			return layer.EdgeWeight(resolvedNodeLayerTuple{NodeId: from.NodeId, Coordinates: rcoordsFrom}, resolvedNodeLayerTuple{NodeId: to.NodeId, Coordinates: rcoordsTo})
		} else {
			return 0.0
		}
	} else {
		return 0.0
	}
}

func (p *MultilayerNetwork) GetVertexInstances(id uint32) []NodeLayerTuple {
	retVal := make([]NodeLayerTuple, 0)
	layers, ok := p.nodeIdsAndLayers[id]
	if ok {
		for _, layer := range layers {
			retVal = append(retVal, NodeLayerTuple{NodeId: id, Coordinates: layer.AspectCoordinates()})
		}
	}
	return retVal
}

func (p *MultilayerNetwork) GetNeighbors(vertex NodeLayerTuple) map[NodeLayerTuple] float32 {
	retVal := make(map[NodeLayerTuple] float32)
	_, ok := p.nodeIdsAndLayers[vertex.NodeId]
	if !ok {
		return retVal
	}
	resolvedCoordinates, err := p.resolveCoordinates(vertex.Coordinates)
	if err == nil && p.elementaryLayerExists(resolvedCoordinates) {
		// get the explicit neighbors
		neighbors := p.elementaryLayers[resolvedCoordinates].GetNeighbors(vertex.NodeId)
		for node, wt := range neighbors {
			retVal[node] = wt
		}
		// add node-coupled neighbors
		layers, ok := p.nodeIdsAndLayers[vertex.NodeId]
		if ok {
			for _, layer := range layers {
				if layer.layerCoordinates == resolvedCoordinates {
					continue
				} else {
					neighbors := layer.GetNeighbors(vertex.NodeId)
					for nlt, wt := range neighbors {
						retVal[nlt] = wt
					}
				}
			}
		}
	}
	return retVal
}

func (p *MultilayerNetwork) GetSources(vertex NodeLayerTuple, coupled bool) map[NodeLayerTuple] float32 {
	retVal := make(map[NodeLayerTuple] float32)
	_, ok := p.nodeIdsAndLayers[vertex.NodeId]
	if !ok {
		return retVal
	}
	resolvedCoordinates, err := p.resolveCoordinates(vertex.Coordinates)
	if err == nil && p.elementaryLayerExists(resolvedCoordinates) {
		sources := p.elementaryLayers[resolvedCoordinates].GetSources(vertex.NodeId)
		for node, wt := range sources {
			retVal[node] = wt
		}

		if coupled {
			// add node-coupled sources
			for _, layer := range p.nodeIdsAndLayers[vertex.NodeId] {
				if layer.AspectCoordinates() == resolvedCoordinates {
					continue
				} else {
					srcs := layer.GetSources(vertex.NodeId)
					for nlt, wt := range srcs {
						retVal[nlt] = wt
					}
				}
			}
		}
	}
	return retVal
}

func (p *MultilayerNetwork) CategoricalGetNeighbors(vertex NodeLayerTuple, aspectCategory string, ordinal bool) map[NodeLayerTuple] float32 {
	retVal := make(map[NodeLayerTuple] float32)
	_, ok := p.nodeIdsAndLayers[vertex.NodeId]
	if !ok {
		return retVal
	}
	resolvedCoordinates, err := p.resolveCoordinates(vertex.Coordinates)
	if err != nil {
		return retVal
	}
	if !p.elementaryLayerExists(resolvedCoordinates) {
		return retVal
	}

	indexOfAspect := p.locOf(p.aspects, aspectCategory)
	if indexOfAspect == -1 {
		return retVal
	}

	epsilon := 1
	if !ordinal {
		epsilon = len(p.indices[indexOfAspect])
	}

	for _, layer := range p.nodeIdsAndLayers[vertex.NodeId] {
		if layer.layerCoordinates == resolvedCoordinates {
			continue
		}
		outOfAspect := false

		lResolved := strings.Split(layer.layerCoordinates, ",")
		resolved := strings.Split(resolvedCoordinates, ",")
		ilResolved := make([]int, len(resolved))
		iResolved := make([]int, len(resolved))

		for i:= 0; i < len(resolved); i++ {
			il, err1 := strconv.Atoi(lResolved[i])
			ir, err2 := strconv.Atoi(resolved[i])
			if err1 != nil || err2 != nil {
				return retVal
			}
			ilResolved[i] = il
			iResolved[i] = ir
		}

		for k := 0; k < len(iResolved); k++ {
			if k != indexOfAspect {
				continue
			}
			if ilResolved[k] != iResolved[k] {
				outOfAspect = true
				break
			}
		}
		if outOfAspect {
			continue
		}

		// any layer that survives to this point is in aspect, see if it is ordinal if required
		if abs(iResolved[indexOfAspect] - ilResolved[indexOfAspect]) > epsilon {
			continue
		}

		nghrs := layer.GetNeighbors(vertex.NodeId)
		for k, v := range nghrs {
			retVal[k] = v
		}
	}
	return retVal
}

func (p *MultilayerNetwork) CategoricalGetSources(vertex NodeLayerTuple, aspectCategory string, ordinal bool) map[NodeLayerTuple] float32 {
	retVal := make(map[NodeLayerTuple] float32)
	_, ok := p.nodeIdsAndLayers[vertex.NodeId]
	if !ok {
		return retVal
	}
	resolvedCoordinates, err := p.resolveCoordinates(vertex.Coordinates)
	if err != nil {
		return retVal
	}
	if !p.elementaryLayerExists(resolvedCoordinates) {
		return retVal
	}

	indexOfAspect := p.locOf(p.aspects, aspectCategory)
	if indexOfAspect == -1 {
		return retVal
	}

	epsilon := 1
	if !ordinal {
		epsilon = len(p.indices[indexOfAspect])
	}

	for _, layer := range p.nodeIdsAndLayers[vertex.NodeId] {
		if layer.layerCoordinates == resolvedCoordinates {
			continue
		}
		outOfAspect := false

		lResolved := strings.Split(layer.layerCoordinates, ",")
		resolved := strings.Split(resolvedCoordinates, ",")
		ilResolved := make([]int, len(resolved))
		iResolved := make([]int, len(resolved))

		for i:= 0; i < len(resolved); i++ {
			il, err1 := strconv.Atoi(lResolved[i])
			ir, err2 := strconv.Atoi(resolved[i])
			if err1 != nil || err2 != nil {
				return retVal
			}
			ilResolved[i] = il
			iResolved[i] = ir
		}

		for k := 0; k < len(iResolved); k++ {
			if k != indexOfAspect {
				continue
			}
			if ilResolved[k] != iResolved[k] {
				outOfAspect = true
				break
			}
		}
		if outOfAspect {
			continue
		}

		// any layer that survives to this point is in aspect, see if it is ordinal if required
		if abs(iResolved[indexOfAspect] - ilResolved[indexOfAspect]) > epsilon {
			continue
		}

		nghrs := layer.GetSources(vertex.NodeId)
		for k, v := range nghrs {
			retVal[k] = v
		}
	}
	return retVal
}

func (p *MultilayerNetwork) UnaliasCoordinates(rcoords string) string {
	retVal := ""
	coords := strings.Split(rcoords, ",")
	for i := 0; i < len(p.aspects); i++ {
		icoord, err := strconv.Atoi(coords[i])
		if err != nil {
			return ""
		}
		if icoord > len(p.indices[i]) - 1 {
			return ""
		} else {
			retVal += p.indices[i][icoord]
		}
		if i < len(p.aspects) - 1 {
			retVal += ","
		}
	}
	return retVal
}
func (p *MultilayerNetwork) AddElementaryLayer(coordinates string, G *Network) (bool, error) {
	if coordinates == "" {
		return false, NewNetworkArgumentError("Coordinates cannot be null")
	}

	if G.Directed() != p.directed {
		return false, NewNetworkArgumentError("Both the multilayer network and the elementary layer must have the same value of directed")
	}

	resolved, err := p.resolveCoordinates(coordinates)
	if err == nil {
		if p.addElementaryNetworkToMultilayerNetwork(resolved, G) {
			return true, nil
		} else {
			return false, nil
		}
	} else {
		return false, err
	}
}

func (p *MultilayerNetwork) RemoveElementaryLayer(coords string) (bool, error) {
	if coords == "" {
		return false, NewNetworkArgumentError("Coordinates cannot be null")
	}

	resolved, err := p.resolveCoordinates(coords)
	if err == nil {
		return p.removeElementaryLayerFromMultilayerNetwork(resolved), nil
	} else {
		return false, err
	}
}

func (p *MultilayerNetwork) ListGML(writer *bufio.Writer) {
	Fprintln(writer, "multilayer_network [")
	if p.directed {
		Fprintln(writer, "\tdirected 1")
	} else {
		Fprintln(writer, "\tdirected 0")
	}

	Fprintln(writer, "\taspects")

	for i := 0; i < len(p.aspects); i++ {
		Fprint(writer, "\t\t" + p.aspects[i] + " ")
		Fprintln(writer, strings.Join(p.indices[i], ","))
	}
	Fprintln(writer, "\t]")

	// serialize the layer coordinates and its constituent network, but defer the interlayer edges
	for coords, layer := range p.elementaryLayers {
		Fprintln(writer, "\tlayer [")
		aspectCoords := p.UnaliasCoordinates(coords)
		Fprintln(writer, "\t\tcoordinates " + aspectCoords)
		layer.ListGML(writer, 2)
		Fprintln(writer, "\t]")
	}

	// now write all the interlayer edges
	for _, layer := range p.elementaryLayers {
		layer.ListInterlayerGML(writer)
	}
	Fprintln(writer, "]")
}

func (p *MultilayerNetwork) ListAllLayersGML(writer *bufio.Writer, level int) {
	for coords, layer := range p.elementaryLayers {
		Fprintln(writer, "\tlayer [")
		aspectCoords := p.UnaliasCoordinates(coords)
		Fprintln(writer, "\t\tcoordinates " + aspectCoords)
		layer.ListGML(writer, level)
		Fprintln(writer, "\t]")
	}
}

func (p *MultilayerNetwork) ListAllInterlayerEdges(writer *bufio.Writer) {
	for _, layer := range p.elementaryLayers {
		layer.ListInterlayerGML(writer)
	}
}

func (p *MultilayerNetwork) AddVertex(vertex NodeLayerTuple) (bool, error) {
	rVertex, err := p.resolveNodeLayerTuple(vertex)
	if err != nil {
		return false, err
	}

	if !p.elementaryLayerExists(rVertex.Coordinates) {
		return false, NewNetworkArgumentError("Elementary layer does not exist at " + rVertex.Coordinates)
	}

	layer, _ := p.elementaryLayers[rVertex.Coordinates]		// already know the layer exists from the previous test
	if !layer.HasVertex(rVertex.NodeId) {
		layer.AddVertex(rVertex.NodeId)

		layers, ok := p.nodeIdsAndLayers[rVertex.NodeId]
		if !ok {
			layerList := make([] *elementaryLayer, 1)
			layerList[0] = layer
			p.nodeIdsAndLayers[rVertex.NodeId] = layerList
		} else {
			if p.layerLoc(layers, *layer)  != -1 {
				layers = append(layers, layer)
			}
		}
		return true, nil
	} else {
		return false, NewNetworkArgumentError(Sprintf("%s already exists", vertex.ToString()))
	}
}

func (p *MultilayerNetwork) RemoveVertex(vertex NodeLayerTuple) (bool, error) {
	rVertex, err := p.resolveNodeLayerTuple(vertex)
	if err != nil {
		return false, err
	}

	if !p.elementaryLayerExists(rVertex.Coordinates) {
		return false, NewNetworkArgumentError("Elementary layer does not exist at " + rVertex.Coordinates)
	}

	layer, _ := p.elementaryLayers[rVertex.Coordinates]
	if layer.HasVertex(rVertex.NodeId) {
		layer.RemoveVertex(rVertex.NodeId)
		layers, _ := p.nodeIdsAndLayers[rVertex.NodeId]
		loc := p.layerLoc(layers, *layer)
		if loc != -1 {
			layers[loc] = layers[len(layers) - 1]
			layers[len(layers) - 1] = nil		// turn last into nil
			layers = layers[:len(layers) - 1]
		}
		if len(layers) == 0 {
			delete(p.nodeIdsAndLayers, rVertex.NodeId)
		}
		return true, nil
	} else {
		return false, NewNetworkArgumentError(Sprintf("Vertex %d does not exist in layer %s", vertex.NodeId, vertex.Coordinates))
	}
}

func (p *MultilayerNetwork) AddEdge(from NodeLayerTuple, to NodeLayerTuple, wt float32) (bool, error) {
	if from.NodeId == to.NodeId && from.Coordinates == to.Coordinates {
		return false, NewNetworkArgumentError(Sprintf("Self-edges are not permitted (vertex %s, %s)", from.ToString(), to.ToString()))
	}

	if from.NodeId == to.NodeId {
		return false, NewNetworkArgumentError(Sprintf("Categorical edges are implicit and have weight zero (vertices %s, %s)", from.ToString(), to.ToString()))
	}

	rFrom, err1 := p.resolveNodeLayerTuple(from)
	rTo, err2 := p.resolveNodeLayerTuple(to)

	if err1 != nil || err2 != nil || !p.elementaryLayerExists(rFrom.Coordinates) || !p.elementaryLayerExists(rTo.Coordinates) {
		return false, NewNetworkArgumentError(Sprintf("The elementary layer for one or more vertices does not exist (vertices passed are %s, %s)", from.ToString(), to.ToString()))
	}

	fromLayer, _ := p.elementaryLayers[rFrom.Coordinates]
	toLayer, _ := p.elementaryLayers[rTo.Coordinates]

	_, ok := p.nodeIdsAndLayers[from.NodeId]
	if !ok {
		// add an entry for the node
		layers := make([] *elementaryLayer, 1)
		p.nodeIdsAndLayers[from.NodeId] = layers
	}

	_, ok = p.nodeIdsAndLayers[to.NodeId]
	if !ok {
		layers := make([] *elementaryLayer, 1)
		p.nodeIdsAndLayers[to.NodeId] = layers
	}

	if fromLayer.layerCoordinates == toLayer.layerCoordinates {
		// special case, intralayer add
		if !fromLayer.HasVertex(rFrom.NodeId) {
			_, ok := p.nodeIdsAndLayers[rFrom.NodeId]
			if ok {
				p.nodeIdsAndLayers[rFrom.NodeId] = append(p.nodeIdsAndLayers[rFrom.NodeId], fromLayer)
			} else {
				layers := make([] *elementaryLayer, 1)
				layers = append(layers, fromLayer)
				p.nodeIdsAndLayers[rFrom.NodeId] = layers
			}
		}

		if !fromLayer.HasVertex(rTo.NodeId) {
			_, ok := p.nodeIdsAndLayers[rTo.NodeId]
			if ok {
				p.nodeIdsAndLayers[rTo.NodeId] = append(p.nodeIdsAndLayers[rTo.NodeId], fromLayer)
			} else {
				layers := make([] *elementaryLayer, 1)
				layers = append(layers, fromLayer)
				p.nodeIdsAndLayers[rTo.NodeId] = layers
			}
		}
		fromLayer.AddEdge(rFrom, rTo, wt)
		return true, nil
	} else {
		if !fromLayer.HasVertex(rFrom.NodeId) {
			fromLayer.AddVertex(rFrom.NodeId)
			p.nodeIdsAndLayers[rFrom.NodeId] = append(p.nodeIdsAndLayers[rFrom.NodeId], fromLayer)
		}

		if !toLayer.HasVertex(rTo.NodeId) {
			toLayer.AddVertex(rTo.NodeId)
			p.nodeIdsAndLayers[rTo.NodeId] = append(p.nodeIdsAndLayers[rTo.NodeId], fromLayer)
		}
		// vertices definitely exist, add the edge
		fromLayer.AddEdge(rFrom, rTo, wt)
		toLayer.AddInEdge(rTo, rFrom, wt)
		return true, nil
	}
}

func (p *MultilayerNetwork) RemoveEdge(from NodeLayerTuple, to NodeLayerTuple) (bool, error) {
	rFrom, err1 := p.resolveNodeLayerTuple(from)
	rTo, err2 := p.resolveNodeLayerTuple(to)

	if err1 != nil || err2 != nil || !p.elementaryLayerExists(rFrom.Coordinates) || !p.elementaryLayerExists(rTo.Coordinates) {
		return false, NewNetworkArgumentError(Sprintf("The elementary layer for one or more vertices does not exist (vertices passed are %s, %s)", from.ToString(), to.ToString()))
	}

	fromLayer, _ := p.elementaryLayers[rFrom.Coordinates]
	toLayer, _ := p.elementaryLayers[rTo.Coordinates]

	if fromLayer.HasEdge(rFrom, rTo) {
		fromLayer.RemoveEdge(rFrom, rTo)
		toLayer.RemoveInEdge(rTo, rFrom)
	}

	return true, nil
}

func (p *MultilayerNetwork) MakeSupraAdjacencyMatrix() [][]float32 {
	if !p.IsNodeAligned() {
		return nil
	}

	dimension := p.getDimension()
	// make a zero'd out matrix
	retVal := make([][]float32, dimension)
	for i := 0; i < dimension; i++ {
		retVal[i] = make([]float32, dimension)
	}

	layerList := make([][]string, 0)
	aspect := p.aspects[0]

	// build an ordered, hierarchical list of all elementary layer indices
	coords := make([]string,0)
	for _, mark := range p.indices[0] {
		coords = append(coords, mark)
		p.recurseAdjacencyMatrix(&layerList, aspect, 0, 0, &coords)
		coords = append(coords[:0], coords[1:]...)
	}

	// The list dictates how the supra-adjacency matrix is organized.
	// Begin building the matrix block by block, with each block representing an elementary layer adjacency matrix (on the diagonal)
	// or interlayer adjacencies (off the diagonal)
	blockCt := len(layerList)
	for row := 0; row < blockCt; row++ {
		for column := 0; column < blockCt; column++ {
			rowCoord := layerList[row]
			colCoord := layerList[column]
			rowResolved, _ := p.resolveCoordinates(strings.Join(rowCoord, ","))
			colResolved, _ := p.resolveCoordinates(strings.Join(colCoord, ","))
			layer := p.elementaryLayers[rowResolved]
			if row == column {
				p.insertLayerAdjacencies(retVal, layer.LayerAdjacencyMatrix(), row, column)
			} else {
				p.insertLayerAdjacencies(retVal, layer.InterlayerAdjacencies(colResolved), row, column)
			}
		}
	}
	return retVal
}

func (p *MultilayerNetwork) GetLayer(layerCoordinates string) *Network {
	resolved, err := p.resolveCoordinates(layerCoordinates)

	if err != nil || !p.elementaryLayerExists(resolved) {
		return nil
	}
	return p.elementaryLayers[resolved].CopyNetwork()
}


func (p *MultilayerNetwork) indentForLevel(level int) string {
	retVal := ""
	for i := 0; i < level; i++ {
		retVal += "\t"
	}
	return retVal
}

func (p *MultilayerNetwork) elementaryLayerExists(rcoordinates string) bool {
	_, ok :=  p.elementaryLayers[rcoordinates]
	return ok
}

func (p *MultilayerNetwork) addElementaryNetworkToMultilayerNetwork(coords string, G *Network) bool {
	inVertices := G.Vertices(true)
	layer := NewElementaryLayer(p, G, coords)

	for _, vertex := range inVertices {
		layers, ok := p.nodeIdsAndLayers[vertex]
		if ok {
			p.nodeIdsAndLayers[vertex] = append(layers, layer)
		} else {
			lyrs := make([]*elementaryLayer, 0)
			lyrs = append(lyrs, layer)
			p.nodeIdsAndLayers[vertex] = lyrs
		}
	}
	p.elementaryLayers[coords] = layer
	return true
}

func (p *MultilayerNetwork) removeElementaryLayerFromMultilayerNetwork(resolved string) bool {
	if p.elementaryLayerExists(resolved) {
		layer := p.elementaryLayers[resolved]
		vertices := layer.Vertices(true)

		for _, vertex := range vertices {
			layers, ok := p.nodeIdsAndLayers[vertex]
			i := 0
			if ok {
				for ; i < len(layers); i++ {
					if layers[i] == layer {
						break
					}
				}
				if i < len(layers) {
					// remove the ith layer by copying down the rest of the slice
					layers[i] = layers[len(layers) - 1]
					layers[len(layers) - 1] = nil		// turn last into nil
					layers = layers[:len(layers) - 1]	// truncate
				}
			}
		}
		delete(p.elementaryLayers, resolved)
		return true
	} else {
		return false
	}
}

func (p *MultilayerNetwork) resolveNodeLayerTuple(tuple NodeLayerTuple) (resolvedNodeLayerTuple, error) {
	coordinates, err := p.resolveCoordinates(tuple.Coordinates)
	if err == nil {
		return resolvedNodeLayerTuple{NodeId: tuple.NodeId, Coordinates: coordinates}, nil
	} else {
		return resolvedNodeLayerTuple{NodeId:0, Coordinates:""}, NewNetworkArgumentError(tuple.Coordinates + " cannot be found in the network")
	}
}

// takes a comma delimited list of aspect index names and converts it to a comma delimited list of integer indices
// Each index is the index into indices for a particular aspect. For example, if the aspect location has the indices PHL, ATL, NYC, and
// the aspect type has the indices tech, transport, finance, then "PHL,transport" is resolved to "0,1"
func (p *MultilayerNetwork) resolveCoordinates(coordinateString string) (string, error) {
	indices := strings.Split(coordinateString, ",")
	if len(indices) != len(p.aspects) {
		return "", nil
	}
	s := ""
	index := 0
	for i:= 0; i < len(indices); i++ {
		index = p.locOf(p.indices[i], indices[i])
		if index == -1 {
			return "", errors.New(Sprintf("Index %s not found for aspect %s", indices[i], p.aspects[i]))
		}
		s += strconv.Itoa(index)
		if i < len(indices) -1 {
			s += ","
		}
	}

	return s, nil
}

// find the index of a string in an array of strings
func (p *MultilayerNetwork) locOf(col []string, item string) int {
	for i, val := range col {
		if val == item {
			return i
		}
	}
	return -1
}

func (p *MultilayerNetwork) layerLoc(layers [] *elementaryLayer, layer elementaryLayer) int {
	for i := 0; i < len(layers); i++ {
		if layers[i].layerCoordinates == layer.layerCoordinates {
			return i
		}
	}
	return -1
}

func (p *MultilayerNetwork) getDimension() int {
	elemSize := len(p.nodeIdsAndLayers)
	ct := 1
	for _, aspect := range p.aspects {
		ct *= len(p.indices[p.locOf(p.aspects, aspect)])
	}

	return ct * elemSize
}

// construct a flattening of the aspects by using recursion such that the last aspect repeats its indices for each index of the next to last aspect, then the next to last, and so on until the first aspect
// enumerates its indices once
func (p *MultilayerNetwork) recurseAdjacencyMatrix(allCoords *[][]string, aspect string, blockRow int, blockColumn int, layerCoords *[]string) {
	index := p.locOf(p.aspects, aspect) + 1
	if index == len(p.aspects) - 1 {
		// innermost aspect
		for _, stop := range p.indices[index] {
			//elemLayerCoords := make([]string, len(*layerCoords))
			elemLayerCoords :=make([]string, 0)
			// concatenate layerCoords to the end of elemLayerCoords
			elemLayerCoords = append(elemLayerCoords, *layerCoords...)
			elemLayerCoords = append(elemLayerCoords, stop)
			*allCoords = append(*allCoords, elemLayerCoords)
		}
	} else {
		curAspect := p.aspects[index]
		for _, stop := range p.indices[index] {
			*layerCoords = append(*layerCoords, stop)
			p.recurseAdjacencyMatrix(allCoords, curAspect, blockRow, blockColumn, layerCoords)
			// delete the first entry of layerCoords by copying the remaining slice down one element
			*layerCoords = (*layerCoords)[:len(*layerCoords) - 1]

		}
	}
}

func (p *MultilayerNetwork) insertLayerAdjacencies(supra [][]float32, layerMatrix [][]float32, rowBlock int, colBlock int) {
	size := len(layerMatrix[0])
	rowOffset := size * rowBlock
	colOffset := size * colBlock

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			supra[rowOffset + i][colOffset + j] = layerMatrix[i][j]
		}
	}
}

func (p *MultilayerNetwork) removeOutEdge(from resolvedNodeLayerTuple, to resolvedNodeLayerTuple) {
	if p.elementaryLayerExists(from.Coordinates) {
		p.elementaryLayers[from.Coordinates].RemoveOutEdge(from, to)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}