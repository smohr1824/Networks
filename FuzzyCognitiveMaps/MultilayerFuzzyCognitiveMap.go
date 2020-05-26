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

package FuzzyCognitiveMap

import (
	"bufio"
	"errors"
	. "fmt"
	"github.com/smohr1824/Networks/Core"
	"sort"
)

type MultilayerFuzzyCognitiveMap struct {
	concepts      map[uint32]*MultilayerCognitiveConcept
	reverseLookup map[string]uint32
	nextNodeId    uint32
	// not doing state calculation algebraically allows us to drop the current concepts property from the C# library
	model         Core.MultilayerNetwork
	modifiedKosko bool
	tfunc         ThresholdFunc
	threshold     ThresholdType
}

func NewMultilayerFuzzyCognitiveMapDefault(aspects []string, indices [][]string) *MultilayerFuzzyCognitiveMap {
	retVal := new(MultilayerFuzzyCognitiveMap)
	retVal.concepts = make(map[uint32]*MultilayerCognitiveConcept)
	retVal.reverseLookup = make(map[string]uint32)
	retVal.model = *Core.NewMultilayerNetwork(aspects, indices, true)
	retVal.threshold = Bivalent
	retVal.tfunc = bivalent
	retVal.modifiedKosko = false

	return retVal
}

func NewMultilayerFuzzyCognitiveMap(aspects []string, indices [][]string, useModifiedKosko bool, thresholdType ThresholdType) *MultilayerFuzzyCognitiveMap {
	retVal := new(MultilayerFuzzyCognitiveMap)
	retVal.concepts = make(map[uint32]*MultilayerCognitiveConcept)
	retVal.reverseLookup = make(map[string]uint32)
	retVal.model = *Core.NewMultilayerNetwork(aspects, indices, true)
	switch thresholdType {
	case Bivalent:
		retVal.tfunc = bivalent
	case Trivalent:
		retVal.tfunc = trivalent
	case Logistic:
		retVal.tfunc = logistic
	case Custom:
		retVal.tfunc = nil // you're going to want to BE SURE to set the threshold function
	default:
		retVal.tfunc = nil
	}
	retVal.threshold = thresholdType
	retVal.modifiedKosko = useModifiedKosko

	return retVal
}

func (c *MultilayerFuzzyCognitiveMap) Concepts() map[uint32]*MultilayerCognitiveConcept {
	return c.concepts
}

func (c *MultilayerFuzzyCognitiveMap) ListLayers() []string {
	return c.model.ElementaryLayers()
}

// returns concept names in a consistent order (typically, the order in which concepts are added/deserialized
func (c *MultilayerFuzzyCognitiveMap) ListConcepts() []string {
	keys := make([]uint32, len(c.concepts))
	i := 0
	for id := range c.concepts {
		keys[i] = id
		i++
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	cons := make([]string, len(keys))
	for k := 0; k < len(keys); k++ {
		cons[k] = c.concepts[keys[k]].Name
	}
	return cons
}

func (c *MultilayerFuzzyCognitiveMap) Threshold() ThresholdType {
	return c.threshold
}

func (c *MultilayerFuzzyCognitiveMap) ReportAggregateLevel(concept string) (float32, error) {
	conceptId, ok := c.reverseLookup[concept]
	if ok {
		return c.concepts[conceptId].ActivationLevel, nil
	} else {
		return 0.0, errors.New("concept " + concept + " not found in map")
	}
}

func (c *MultilayerFuzzyCognitiveMap) ReportLayerLevels(concept string) (map[string]float32, error) {
	conceptId, ok := c.reverseLookup[concept]
	// make a copy of the layer activation levels
	levels := make(map[string]float32)
	if ok {
		for k, v := range c.concepts[conceptId].layerActivationLevels {
			levels[k] = v
		}
		return levels, nil
	} else {
		return levels, errors.New("Concept " + concept + " not found in map.")
	}
}

// support for deserialization-- node id management is relaxed
func (c *MultilayerFuzzyCognitiveMap) AddConcept(concept MultilayerCognitiveConcept, id uint32) bool {
	_, okReverse := c.reverseLookup[concept.Name]
	_, okConcept := c.concepts[id]
	if !okReverse && !okConcept {
		c.concepts[id] = &concept
		c.reverseLookup[concept.Name] = id
		if id >= c.nextNodeId {
			c.nextNodeId = id + 1
		}
		return true
	} else {
		return false
	}
}

func (c *MultilayerFuzzyCognitiveMap) AddConceptToLayer(name string, coords string, level float32, initial float32, fast bool) bool {
	existingKey, ok := c.reverseLookup[name]
	if !ok {
		concept := NewMultilayerCognitiveConcept(name, initial, level)
		concept.layerActivationLevels[coords] = initial
		c.concepts[c.nextNodeId] = concept
		c.reverseLookup[name] = c.nextNodeId
		tuple := Core.NewNodeLayerTuple(c.nextNodeId, coords)
		if !c.model.HasElementaryLayer(coords) {
			_, _ = c.model.AddElementaryLayer(coords, Core.NewNetwork(true))
		}
		_, _ = c.model.AddVertex(*tuple)
		concept.setLayerLevel(coords, level)
		c.nextNodeId++
		return true
	} else {
		tuple := Core.NewNodeLayerTuple(existingKey, coords)
		if !c.model.HasVertex(*tuple) {
			if !c.model.HasElementaryLayer(coords) {
				_, _ = c.model.AddElementaryLayer(coords, Core.NewNetwork(true))
			}
			concept, _ := c.concepts[existingKey]
			concept.layerActivationLevels[coords] = initial
			_, _ = c.model.AddVertex(*tuple)
			concept.setLayerLevel(coords, level)
			if !fast {
				c.recomputeAggregateActivationLevel(concept.Name)
			}
			return true
		} else {
			return false
		}
	}
}

func (c *MultilayerFuzzyCognitiveMap) DeleteConcept(conceptName string, coords string) {
	id, ok := c.reverseLookup[conceptName]
	if ok {
		delete(c.reverseLookup, conceptName)
		layers := c.concepts[id].GetLayers()
		for _, layer := range layers {
			tuple := Core.NewNodeLayerTuple(id, layer)
			_, _ = c.model.RemoveVertex(*tuple)
		}
		delete(c.concepts, id)
	}
}

func (c *MultilayerFuzzyCognitiveMap) AddInfluence(influences string, influencesCoords string, influenced string, influencedCoords string, wt float32) {
	from, okFrom := c.reverseLookup[influences]
	to, okTo := c.reverseLookup[influenced]
	if okFrom && okTo {
		influencesTuple := Core.NewNodeLayerTuple(from, influencesCoords)
		influencedTuple := Core.NewNodeLayerTuple(to, influencedCoords)

		_, _ = c.model.AddEdge(*influencesTuple, *influencedTuple, wt)

	}
}

func (c *MultilayerFuzzyCognitiveMap) AddInfluenceTuples(influencesTuple Core.NodeLayerTuple, influencedTuple Core.NodeLayerTuple, wt float32) {
	_, okFrom := c.concepts[influencesTuple.NodeId]
	_, okTo := c.concepts[influencedTuple.NodeId]
	if okFrom && okTo {
		_, _ = c.model.AddEdge(influencesTuple, influencedTuple, wt)
	}
}

func (c *MultilayerFuzzyCognitiveMap) DeleteInfluence(influences string, influencesCoords string, influenced string, influencedCoords string) {
	from, fromOk := c.reverseLookup[influences]
	to, toOk := c.reverseLookup[influenced]
	if fromOk && toOk {
		influencesTuple := Core.NewNodeLayerTuple(from, influencesCoords)
		influencedTuple := Core.NewNodeLayerTuple(to, influencedCoords)

		_, _ = c.model.RemoveEdge(*influencesTuple, *influencedTuple)
	}
}

func (c *MultilayerFuzzyCognitiveMap) GetActivationLevel(conceptName string) (float32, error) {
	id, ok := c.reverseLookup[conceptName]
	if !ok {
		return 0.0, errors.New("Concept " + conceptName + " not found in map")
	} else {
		concept, _ := c.concepts[id]
		return concept.GetAggregateActivationLevel(), nil
	}
}

func (c *MultilayerFuzzyCognitiveMap) GetConceptLayerActivationLevel(conceptName string, conceptCoords string) (float32, error) {
	id, ok := c.reverseLookup[conceptName]
	if !ok {
		return 0.0, errors.New("Concept " + conceptName + " not found in map.")
	} else {
		concept, _ := c.concepts[id]
		return concept.GetLayerActivationLevel(conceptCoords)
	}
}

func (c *MultilayerFuzzyCognitiveMap) GetLayerActivationLevels(coords string) (map[string]float32, error) {
	if c.model.HasElementaryLayer(coords) {
		iverts, _ := c.model.VerticesInLayer(coords)
		levels := make(map[string]float32)
		for _, vertId := range iverts {
			concept, _ := c.concepts[vertId]
			layerLevel, _ := c.GetConceptLayerActivationLevel(concept.Name, coords)
			levels[concept.Name] = layerLevel
		}
		return levels, nil
	} else {
		return nil, errors.New(Sprintf("Layer %s not found in network", coords))
	}
}

func (c *MultilayerFuzzyCognitiveMap) Step() {
	nextConceptLevels := make(map[uint32]*MultilayerCognitiveConcept)

	for conceptId, concept := range c.concepts {
		next := NewMultilayerCognitiveConcept(concept.Name, 0.0, 0.0)
		layers := concept.GetLayers()
		for _, layer := range layers {
			instance := Core.NewNodeLayerTuple(conceptId, layer)
			sources := c.model.GetSources(*instance, false)
			sum := float32(0.0)
			for tuple, value := range sources {
				level, _ := c.concepts[tuple.NodeId].GetLayerActivationLevel(tuple.Coordinates)
				sum += level * value
			}
			if c.modifiedKosko {
				val, _ := c.concepts[conceptId].GetLayerActivationLevel(layer)
				next.setLayerLevel(layer, c.tfunc(val+sum))
			} else {
				next.setLayerLevel(layer, c.tfunc(sum))
			}
		}

		// finished all layer instances of a concept, so compute the aggregate and add the next gen concept to the map
		total := float32(0.0)
		for _, layer := range next.GetLayers() {
			value, _ := next.GetLayerActivationLevel(layer)
			total += value
		}

		if c.modifiedKosko {
			next.ActivationLevel = c.tfunc(next.ActivationLevel + total)
		} else {
			next.ActivationLevel = c.tfunc(total)
		}
		nextConceptLevels[conceptId] = next
	}

	for id, concept := range nextConceptLevels {
		c.concepts[id].ActivationLevel = concept.ActivationLevel
		for _, layer := range concept.GetLayers() {
			newVal, _ := concept.GetLayerActivationLevel(layer)
			c.concepts[id].setLayerLevel(layer, newVal)
		}
	}
}

func (c *MultilayerFuzzyCognitiveMap) AddElementaryLayer(coords string, G *Core.Network) {
	_, _ = c.model.AddElementaryLayer(coords, G)
}

func (c *MultilayerFuzzyCognitiveMap) ListGML(writer *bufio.Writer) error {
	_, err := Fprintln(writer, "multilayer_network [")
	// error handling -- check for first, return early if there is a problem; after that, write optimistically and return if there is
	// a problem (an error will probably persist
	if err != nil {
		return err
	}
	_, err = Fprintln(writer, "\tdirected 1")
	_, err = Fprint(writer, "\tthreshold ")
	switch c.Threshold() {
	case Bivalent:
		Fprintln(writer, "\"bivalent\"")
	case Trivalent:
		Fprintln(writer, "\"trivalent\"")
	case Logistic:
		Fprintln(writer, "\"logistic\"")
	case Custom:
		Fprintln(writer, "\"custom\"")
	}
	if c.modifiedKosko {
		Fprintln(writer, "\trule \"modified\"")
	} else {
		Fprintln(writer, "\trule \"kosko\"")
	}

	Fprintln(writer, "\taspects [")
	aspects := c.model.Aspects()
	for _, aspect := range aspects {
		indices := c.model.Indices(aspect)
		_, err = Fprint(writer, "\t\t"+aspect+" \"")
		sindices := ""
		for k, index := range indices {
			sindices += index
			if k < len(indices)-1 {
				sindices += ","
			}
		}
		_, err = Fprintln(writer, sindices+"\"")
	}
	Fprintln(writer, "\t]")

	concepts := c.ListConcepts()
	//for id, concept := range c.concepts {
	for _, name := range concepts {
		conceptId := c.reverseLookup[name]
		concept := c.concepts[conceptId]
		Fprintln(writer, "\t concept [")
		Fprintln(writer, Sprintf("\t\tid %d", conceptId))
		Fprintln(writer, "\t\tlabel \""+concept.Name+"\"")
		Fprintln(writer, Sprintf("\t\tinitial %.4f", concept.initialValue))
		Fprintln(writer, Sprintf("\t\taggregate %.4f", concept.ActivationLevel))
		Fprintln(writer, "\t\tlevels [")
		layers := concept.GetLayers()
		for _, layer := range layers {
			level, _ := concept.GetLayerActivationLevel(layer)
			Fprintln(writer, Sprintf("\t\t\t%s %.4f", layer, level))
		}
		Fprintln(writer, "\t\t]")
		Fprintln(writer, "\t]")
	}
	c.model.ListAllLayersGML(writer, 2)
	c.model.ListAllInterlayerEdges(writer)
	Fprintln(writer, "]")
	return err
}

func (c *MultilayerFuzzyCognitiveMap) recomputeAggregateActivationLevel(conceptName string) {
	id, ok := c.reverseLookup[conceptName]
	if ok {
		total := float32(0.0)
		concept := c.concepts[id]
		for _, layer := range concept.GetLayers() {
			f, _ := concept.GetLayerActivationLevel(layer)
			total += f
		}

		if c.modifiedKosko {
			concept.ActivationLevel = c.tfunc(concept.ActivationLevel + total)
		} else {
			concept.ActivationLevel = c.tfunc(total)
		}
	}
}
