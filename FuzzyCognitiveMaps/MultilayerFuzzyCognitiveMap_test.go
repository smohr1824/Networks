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
	"fmt"
	"testing"
)

func TestBasicMLFCM(t *testing.T) {
	fcm := BuildMLBasic()

	for i := 1; i < 4; i++ {
		fcm.Step()
	}

	layerIValue, _ := fcm.GetConceptLayerActivationLevel("A", "I")
	if layerIValue != float32(1.0) {
		t.Error("Incorrect activation of concept A after three iterations")
	}
}

func TestBasicMLFCMSerialization(t *testing.T) {
	fcm := BuildMLBasic()
	s := NewMLFCMSerializer()
	s.WriteMLFCMToFile(fcm, "..\\Work\\mlbasic.fcm")
	fcm2, err := s.ReadMLFCMFromFile("..\\Work\\mlbasic.fcm")
	if err != nil {
		t.Errorf("Error deserializing ML FCM: %s", err.Error())
	}
	if len(fcm2.concepts) != len(fcm.concepts) {
		t.Error("Incorrect number of concepts read")
	}

	for k,v := range fcm2.concepts {
		if len(v.layerActivationLevels) != len(fcm.concepts[k].layerActivationLevels) {
			t.Errorf("Concept id %d has a different number of layer activation levels, expected %d saw %d", k, len(fcm.concepts[k].layerActivationLevels), len(v.layerActivationLevels))
		}
	}

	for i := 1; i < 4; i++ {
		fcm.Step()
		fcm2.Step()
	}
	valOrig, err := fcm.GetConceptLayerActivationLevel("A", "I")
	if err != nil {
		t.Errorf("Error trying to get original activation level for layer I: %s", err.Error())
	}
	valRead, err := fcm2.GetConceptLayerActivationLevel("A", "I")
	if err != nil {
		t.Errorf("Error trying to get original activation level for layer I: %s", err.Error())
	}

	if valOrig != valRead {
		t.Errorf("Expected a layer activation level of %.4f, read ML FCM has a level of %.4f", valOrig, valRead)
	}
}

func writeState(t *testing.T, fcm *MultilayerFuzzyCognitiveMap, concepts []string) {
	for _, conName:= range concepts {
		agg,_ := fcm.GetActivationLevel(conName)
		t.Logf("%s aggregate: %.4f", conName, agg)
		layer2Level, _ := fcm.ReportLayerLevels(conName)
		s := ""
		for layer, val := range layer2Level {
			s += fmt.Sprintf("%s: %.4f", layer, val)
		}
	}
}

func BuildMLBasic() *MultilayerFuzzyCognitiveMap {
	indices := []string {"I", "II"}
	dimensions := []string{"levels"}
	allindices := make([][]string,1)
	allindices[0] = indices

	fcm:= NewMultilayerFuzzyCognitiveMap(dimensions, allindices, false, Bivalent)

	fcm.AddConceptToLayer("A", "I", float32(1.0), float32(1.0), false)
	fcm.AddConceptToLayer("B", "I", float32(0.0), float32(0.0), false)
	fcm.AddConceptToLayer("C", "I", float32(0.0), float32(0.0), false)

	fcm.AddConceptToLayer("A", "II", float32(1.0), float32(1.0), false)
	fcm.AddConceptToLayer("D", "II", float32(0.0), float32(0.0), false);
	fcm.AddConceptToLayer("E", "II", float32(0.0), float32(0.0), false)

	fcm.AddInfluence("A", "I", "B", "I", float32(1.0))
	fcm.AddInfluence("A", "I", "C", "I", float32(1.0))
	fcm.AddInfluence("A", "II", "D", "II", float32(1.0))
	fcm.AddInfluence("D", "II", "E", "II", float32(1.0))
	fcm.AddInfluence("E", "II", "A", "I", float32(1.0))
	return fcm
}
