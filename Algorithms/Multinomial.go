// Copyright 2017 - 2018 -- Stephen T. Mohr, OSIsoft, LLC
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

// Basic multinomial distribution stuff

package Algorithms

import (
	"math/rand"
	//"sort"
	"sync"
)
type MultinomialLabels struct {
	slots []float64
	rand rand.Rand
	labels []int
}

func NewMultinomialLabels(labelsObserved *sync.Map, seed int64) *MultinomialLabels {
	multi := new(MultinomialLabels)


	labels := make([]int, 0)
	values := make([]int, 0)

	labelsObserved.Range(func(k, v interface{}) bool {
		labels = append(labels, k.(int))
		values = append(values, v.(int))
		return true
	})
	/*for obs := range labelsObserved {
		//labels = append(labels, obs.Key)
		labels = append(labels, obs.Key.(int))
		//values = append(values, obs.Value)
		values = append(values, obs.Value.(int))
	}*/

	sum:= sumObservations(values)
	probs := calcProbabilities(values, sum)
	bounds := make([]float64, len(probs))
	multi.rand = *rand.New(rand.NewSource(seed))
	var top = 0.0



	for i, v := range probs {
		bounds[i] = top + v
	}
	multi.slots = bounds

	//sort.Ints(labels)
	multi.labels = labels
	return multi
}

func (dist *MultinomialLabels) NextSample() int {
	roll := dist.rand.Float64()
	for i := 0; i < len(dist.slots); i++ {
		if roll < dist.slots[i] {
			return dist.labels[i]
		}
	}
	if len(dist.slots )!= len(dist.labels) || len(dist.labels) == 0 {
		return 0
	}
	return dist.labels[len(dist.labels) - 1]
}

func sumObservations(counts []int) int {
	retVal := 0
	for _, count := range counts {
		retVal += count
	}
	return retVal

}

// calculate the observed probability for each observed label and populate an array
func calcProbabilities(counts []int, sum int) []float64 {
	probs := make([]float64, len(counts))
	idx := 0
	for v := range counts {
		probs[idx] = float64(v)/float64(sum)
		idx++
	}
	return probs
}