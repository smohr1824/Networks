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
	"math"
	"testing"
)

func TestBasicFCM(t *testing.T) {
	fcm := makeBasicFCM()

	for i := 0; i < 5; i++ {
		fcm.Step()
	}

	if len(fcm.Concepts()) != 5 {
		t.Errorf("Should have 5 concepts, actually have %d", len(fcm.Concepts()))
	}

	state := fcm.ReportState()

	if state["A"] != 1.0 {
		t.Errorf("Value of A should be 1.0, is %.2f", state["A"])
	}
	if state["B"] != 0.0 {
		t.Errorf("Value of B should be 0.0, is %.2f", state["B"])
	}

	if state["C"] != 1.0 {
		t.Errorf("Value of C should be 1.0, is %.2f", state["C"])
	}

	if state["D"] != 0.0 {
		t.Errorf("Value of D should be 0.0, is %.2f", state["D"])
	}

	if state["E"] != 0.0 {
		t.Errorf("Value of E should be 0.0, is %.2f", state["E"])
	}

	fcm.Reset()
	fcm.SwitchThresholdFunction(Logistic, nil)

	for i := 0; i < 5; i++ {
		fcm.Step()
	}

	state = fcm.ReportState()
	if math.Abs(float64(state["A"]) - 1.0) > 0.05 {
		t.Errorf("A should be 1.0, is %.2f", state["A"])
	}

	if math.Abs(float64(state["B"]) - 1.0) > 0.05 {
		t.Errorf("B should be 0.0, is %.2f", state["B"])
	}

	if math.Abs(float64(state["C"]) - 0.9) > 0.05 {
		t.Errorf("C should be 0.9, is %.2f", state["C"])
	}

	if math.Abs(float64(state["D"]) - 0.5) > 0.05 {
		t.Errorf("D should be 0.5, is %.2f", state["D"])
	}

	if math.Abs(float64(state["E"]) - 0.0) > 0.05 {
		t.Errorf("E should be 0.0, is %.2f", state["E"])
	}
}

func TestReadWriteFCM(t *testing.T) {
	fcm := makeBasicFCM()

	err := WriteFCMToFile(fcm, "..\\Work\\basic.fcm")
	if err != nil {
		t.Error(err.Error())
	}

	fcm2, err := ReadFCMFromFile("..\\Work\\basic.fcm")
	if err != nil {
		t.Error("Error reading FCM file: " + err.Error())
	}

	if len(fcm.Concepts()) != len(fcm2.Concepts()) {
		t.Error("Unequal number of concepts")
		t.Logf("In memory has %d concepts", len(fcm.Concepts()))
		t.Logf("Read has %d concepts", len(fcm2.Concepts()))
	}
	for _, concept := range fcm2.Concepts() {
		conceptStd, err := fcm.GetConcept(concept.Name)
		if err != nil {
			t.Error("Couldn't find concept " + concept.Name)
		}

		// check on Go quality -- delete after test
		if conceptStd.Name != concept.Name || conceptStd.ActivationLevel != concept.ActivationLevel || conceptStd.GetInitialValue() != concept.GetInitialValue() {
			t.Error("Guess I don't understand equality in Go")
		}

	}
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

