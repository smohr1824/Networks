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

import "errors"

type MultilayerCognitiveConcept struct {
	Name string
	initialValue float32
	ActivationLevel float32
	layerActivationLevels map[string] float32
}

func NewMultilayerCognitiveConcept(name string, initialValue float32, level float32) *MultilayerCognitiveConcept {
	retVal := new (MultilayerCognitiveConcept)
	retVal.Name = name
	retVal.initialValue = initialValue
	retVal.ActivationLevel = level
	retVal.layerActivationLevels = make(map[string] float32)
	return retVal
}

func (c *MultilayerCognitiveConcept) SetInitialLevel(level float32) {
	c.initialValue = level
}

func (c *MultilayerCognitiveConcept) GetInitialValue() float32 {
	return c.initialValue
}

func (c *MultilayerCognitiveConcept) LayerCount() int {
	return len(c.layerActivationLevels)
}

func (c *MultilayerCognitiveConcept) GetLayers() []string {
	levels := make([]string, 0)
	for k, _ := range c.layerActivationLevels {
		levels = append(levels, k)
	}
	return levels
}

func (c *MultilayerCognitiveConcept) GetAggregateActivationLevel() float32 {
	return c.ActivationLevel
}

func (c *MultilayerCognitiveConcept) GetLayerActivationLevel(coords string) (float32, error) {
	level, ok := c.layerActivationLevels[coords]
	if ok {
		return level, nil
	} else {
		return 0.0, errors.New("Concept does not participate in layer " + coords)
	}
}

func (c *MultilayerCognitiveConcept) setLayerLevel(coords string, level float32) {
		c.layerActivationLevels[coords] = level
}