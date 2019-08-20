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

// monolayer fuzzy cognitive map implementation
// state calculations are done exclusively algorithmically as it is more performant than the algebraicmethod
// good bye, 1997

package FuzzyCognitiveMap

import (
	"bufio"
	"errors"
	. "fmt"
	"github.com/smohr1824/Networks/Core"
	"math"
)

type ThresholdFunc func(float32) float32
type FuzzyCognitiveMap struct {
	concepts map[uint32] *CognitiveConcept
	reverseLookup map[string] uint32
	nextNodeId uint32
	// not doing state calculation algebraically allows us to drop the current concepts property from the C# library
	model Core.Network
	modifiedKosko bool
	tfunc ThresholdFunc
	threshold ThresholdType
}

func NewFuzzyCognitiveMapDefault() *FuzzyCognitiveMap {
	retVal := new(FuzzyCognitiveMap)
	retVal.concepts = make(map[uint32] *CognitiveConcept)
	retVal.reverseLookup = make(map[string] uint32)
	retVal.model = *Core.NewNetwork(true)
	retVal.threshold = Bivalent
	retVal.tfunc = bivalent
	retVal.modifiedKosko = false

	return retVal
}

func NewFuzzyCognitiveMap(useModifiedKosko bool, thresholdType ThresholdType) *FuzzyCognitiveMap {
	retVal := new(FuzzyCognitiveMap)
	retVal.concepts = make(map[uint32] *CognitiveConcept)
	retVal.reverseLookup = make(map[string] uint32)
	retVal.model = *Core.NewNetwork(true)
	switch thresholdType {
		case Bivalent:
			retVal.tfunc = bivalent
		case Trivalent:
			retVal.tfunc = trivalent
		case Logistic:
			retVal.tfunc = logistic
		case Custom:
			retVal = nil	// you're going to want to BE SURE to set the threshold function
	}
	retVal.threshold = thresholdType
	retVal.modifiedKosko = useModifiedKosko

	return retVal
}

func (c *FuzzyCognitiveMap) Concepts() map[uint32] *CognitiveConcept {
	return c.concepts
}

func (c *FuzzyCognitiveMap) Threshold() ThresholdType {
	return c.threshold
}

func (c *FuzzyCognitiveMap) AddConcept(conceptName string, initial float32, level float32) bool {
	_, ok := c.reverseLookup[conceptName]
	if !ok {
		c.concepts[c.nextNodeId] = NewCognitiveConcept(conceptName, initial, level)
		c.reverseLookup[conceptName] = c.nextNodeId
		c.model.AddVertex(c.nextNodeId)
		c.nextNodeId++
		return true
	} else {
		// concept exists
		return false
	}
}

func (c *FuzzyCognitiveMap) DeleteConcept(conceptName string) {
	id, ok := c.reverseLookup[conceptName]
	if ok {
		delete(c.reverseLookup, conceptName)
		delete(c.concepts, id)
		c.model.RemoveVertex(id)
	}
}

func (c *FuzzyCognitiveMap) AddInfluence(influences string, influenced string, weight float32) {
	from, okFrom := c.reverseLookup[influences]
	to, okTo := c.reverseLookup[influenced]
	if okFrom && okTo {
		c.model.AddEdge(from, to, weight)
	}
}

func (c *FuzzyCognitiveMap) DeleteInfluence(influences string, influenced string) {
	from, okFrom := c.reverseLookup[influences]
	to, okTo := c.reverseLookup[influenced]
	if okFrom && okTo {
		c.model.RemoveEdge(from, to)
	}
}

func (c *FuzzyCognitiveMap) GetActivationLevel(conceptName string) (float32, error) {
	id, ok := c.reverseLookup[conceptName]
	if ok {
		return c.concepts[id].ActivationLevel, nil
	} else {
		return 0.0, errors.New("Concept " + conceptName + " not found in map")
	}
}

func (c *FuzzyCognitiveMap) Step() {
	// update the concept vector algorithmically
	nextConceptLevels := make(map[uint32] float32)
	for id, concept := range c.concepts {
		sum := float32(0.0)
		influences := c.model.GetSources(id)
		for _, wt := range influences {
			sum += concept.ActivationLevel * wt
		}
		if c.modifiedKosko {
			nextConceptLevels[id] = c.tfunc(concept.ActivationLevel + sum)
		} else {
			nextConceptLevels[id] = c.tfunc(sum)
		}
	}

	// with the new levels calculated, update the FuzzyCognitiveMaps to the new values
	for id, concept := range c.concepts {
		concept.ActivationLevel = nextConceptLevels[id]
	}
}

func (c *FuzzyCognitiveMap) ReportState() map[string] float32 {
	retVal := make(map[string] float32)
	for _, concept := range c.concepts {
		retVal[concept.Name] = concept.ActivationLevel
	}

	return retVal
}

func (c *FuzzyCognitiveMap) Reset() {
	for _, concept := range c.concepts {
		concept.ActivationLevel = concept.initialValue
	}
}

func (c *FuzzyCognitiveMap) SetThresholdFunction(f ThresholdFunc) {
	c.tfunc = f
	c.threshold = Custom
}

func (c *FuzzyCognitiveMap) SwitchThresholdFunction(desiredFunc ThresholdType, f ThresholdFunc) {
	switch desiredFunc {
		case Bivalent:
			c.tfunc = bivalent
			c.threshold = desiredFunc

		case Trivalent:
			c.tfunc = trivalent
			c.threshold = desiredFunc

		case Logistic:
			c.tfunc = logistic
			c.threshold = desiredFunc

		case Custom:
			if f != nil {
				c.tfunc = f
				c.threshold = desiredFunc
			}
	}
}

func (c *FuzzyCognitiveMap) ListGML(writer *bufio.Writer) error {
	_, err := Fprintln(writer, "graph [")
	if err != nil {
		return err
	}
	Fprintln(writer, "\tdirected 1")
	Fprint(writer, "\tthreshold ")
	switch c.threshold {
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

	err = c.listConcepts(writer)
	if err != nil {
		return err
	}
	c.model.ListGMLEdges(writer, "")
	Fprintln(writer, "]")
	return nil
}

func (c *FuzzyCognitiveMap) listConcepts(writer *bufio.Writer) error {
	for id, concept := range c.concepts {
		_, err := Fprintln(writer, "\tnode [")
		_, err = Fprintln(writer, Sprintf("\t\tid %d", id))
		_, err = Fprintln(writer, "\t\tlabel " + concept.Name)
		_, err = Fprintln(writer, Sprintf("\t\tactivation %f.4", concept.ActivationLevel))
		_, err = Fprintln(writer, Sprintf("\t\tinitial %f.4", concept.initialValue))
		_, err = Fprintln(writer, "]")

		if err != nil {
			return err
		}
	}
	return nil
}

// standard threshold functions
func bivalent(f float32) float32 {
	if f > 0.0 {
		return 1.0
	} else {
		return 0.0
	}
}

func trivalent(f float32) float32 {
	if f < 0.5 {
		return -1.0
	}

	if f > 0.5 {
		return 1.0
	}

	return 0.0
}

func logistic(f float32) float32 {
	return float32(1.0/(1 + math.Exp(-5.0 * float64(f))))
}


