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
	. "fmt"
	"testing"
)

func TestBasicFCM(t *testing.T) {
	fcm := makeBasicFCM()

	for i := 0; i < 5; i++ {
		fcm.Step()
	}
	t.Logf("Number of concepts: %d", len(fcm.Concepts()))
	state := fcm.ReportState()
	for name, level := range state {
		t.Log(name + ": " + Sprintf("%.2f", level))
	}
	t.Logf("")
}

func makeBasicFCM() *FuzzyCognitiveMap{
	fcm := NewFuzzyCognitiveMapDefault()
	fcm.AddConcept("A", 1.0, 1.0)
	fcm.AddConcept("B", 0.0, 0.0)
	fcm.AddConcept("C", 1.0, 1.0)
	fcm.AddConcept("D", 0.0, 0.0)
	fcm.AddConcept("E", 0.0, 0.0)

	fcm.AddInfluence("B", "A", 1.0)
	fcm.AddInfluence("A", "C", 1.0)
	fcm.AddInfluence("C", "E", 1.0)
	fcm.AddInfluence("E", "D", 1.0)
	fcm.AddInfluence("D", "C", -1.0)
	fcm.AddInfluence("B", "E", -1.0)
	fcm.AddInfluence("E", "A", -1.0)
	fcm.AddInfluence("D", "B", 1.0)
	fcm.AddInfluence("E", "F", -1.0)

	return fcm
}

