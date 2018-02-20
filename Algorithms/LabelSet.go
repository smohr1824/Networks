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


// Attempt to make thread-safe map of node ids to label observations

package Algorithms

import (
	"sync"
)
type LabelSet struct  {
	observedLabels sync.Map //[int] int
	sync.RWMutex
}

type LabelObservation struct {
	Key   interface{}
	Value interface{}
}

func MakeLabelSet() *LabelSet {
	set := new(LabelSet)
	//set.observedLabels = make(map[int]int)
	return set
}

func (ls *LabelSet) Iterate() <-chan LabelObservation {
	c := make(chan LabelObservation)
	//ls.RLock()
	//defer ls.RUnlock()
	f := func() {
		//ls.Lock()

		//for k, v := range ls.observedLabels {
		ls.observedLabels.Range(func(k, v interface{}) bool {
			c <- LabelObservation{k, v}
			return true
			//}
		})
		close(c)
		//ls.Unlock()
	}
	go f()
	return c
}

func (ls *LabelSet) SetLabel(label int, value int) {
	//ls.Lock()
	//ls.observedLabels[label] = value
	ls.observedLabels.Store(label, value)
	//ls.Unlock()
}

func (ls *LabelSet) DeleteLabel(label int) {
	ls.Lock()
	//delete(ls.observedLabels, label)
	ls.observedLabels.Delete(label)
	ls.Unlock()
}

func (ls *LabelSet) IncrementLabel(label int) {
	//ls.Lock()
	//count, ok := ls.observedLabels[label]
	count, ok := ls.observedLabels.Load(label)
	if ok {
		//ls.observedLabels[label] = count + 1
		ls.observedLabels.Store(label, count.(int)+1)
	} else {
		//ls.observedLabels[label] = 1
		ls.observedLabels.Store(label, 1)
	}
	//ls.Unlock()
}