// Copyright 2017 - 2019 Stephen T. Mohr
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

// core MultilayerNetwork tests

package Core

import (
	"testing"
)

func TestBasicMultilayer(t *testing.T) {
	/*ser := NewNetworkSerializer("|")
	G, err := ser.ReadNetworkFromFile("electrical.dat", true)
	if err != nil {
		t.Errorf(err.Error())
	}
	H, err := ser.ReadNetworkFromFile("flow.dat", true)
	if err != nil {
		t.Errorf(err.Error())
	}
	I, err := ser.ReadNetworkFromFile("control.dat", true)
	if err != nil {
		t.Errorf(err.Error())
	}

	J, err := ser.ReadNetworkFromFile("electrical.dat", true)
	if err != nil {
		t.Errorf(err.Error())
	}
	K, err := ser.ReadNetworkFromFile("flow.dat", true)
	if err != nil {
		t.Errorf(err.Error())
	}
	L, err := ser.ReadNetworkFromFile("control.dat", true)
	if err != nil {
		t.Errorf(err.Error())
	}*/

	procIndices := []string {"electrical", "flow", "control"}
	locIndices := []string {"PHL", "SLTC"}
	allIndices := make([][]string, 2)
	allIndices[0] = procIndices
	allIndices[1] = locIndices
	//aspects := []string {"process", "site"}

	/*Q := NewMultilayerNetwork(aspects, allIndices, true)

	// add elementary layers
	added, err := Q.AddElementaryLayer("electrical,PHL", G)
	if added == false {
		t.Error(err.Error())
	}
	added, err = Q.AddElementaryLayer("flow,PHL", H)
	if added == false {
		t.Error(err.Error())
	}
	added, err = Q.AddElementaryLayer("control,PHL", I)
	if added == false {
		t.Error(err.Error())
	}

	added, err = Q.AddElementaryLayer("electrical,SLTC", J)
	if added == false {
		t.Error(err.Error())
	}
	added, err = Q.AddElementaryLayer("flow,SLTC", K)
	if added == false {
		t.Error(err.Error())
	}
	added, err = Q.AddElementaryLayer("control,SLTC", L)
	if added == false {
		t.Error(err.Error())
	}

	added, err = Q.AddEdge(*NewNodeLayerTuple(1, "electrical,SLTC"), *NewNodeLayerTuple(2,"control,SLTC"), 2)
	if added == false {
		t.Error(err.Error())
	}
	added, err = Q.AddEdge(*NewNodeLayerTuple(2, "electrical,SLTC"), *NewNodeLayerTuple(3, "control,SLTC"), 2)
	if added == false {
		t.Error(err.Error())
	}
	added, err = Q.AddEdge(*NewNodeLayerTuple(3, "control,PHL"), *NewNodeLayerTuple(1, "control,SLTC"), 4)
	if added == false {
		t.Error(err.Error())
	}

	added, err = Q.AddEdge(*NewNodeLayerTuple(4, "flow,SLTC"), *NewNodeLayerTuple(5, "flow,SLTC"), 2)
	if added == false {
		t.Error(err.Error())
	}

	added, err = Q.AddEdge(*NewNodeLayerTuple(7,"fusion,SLTC"), *NewNodeLayerTuple(8, "fusion,SLTC"), 1)
	if added {
		t.Error("Added edge between non-existent vertices in non-existent elementary layers")
	} */

	Q, err := ReadMultilayerNetworkFromFile("multilayer_test.gml")
	if err != nil {
		t.Error(err.Error())
	}

	//WriteMultilayerNetworkToFile(Q, "multilayer_test.gml")
	ct := Q.Order()
	if ct != 5 {
		t.Errorf("Expected order of 5, got %d", ct)
	}

	nlt1 := NewNodeLayerTuple(2,"electrical,SLTC")
	dg := Q.Degree(*nlt1)
	if dg != 3 {
		t.Errorf("Degree for node %s: expected 3, found %d", nlt1.ToString(), dg)
	}

	nlt2 := NewNodeLayerTuple(3,"control,PHL")
	dg = Q.Degree(*nlt2)
	if dg != 3 {
		t.Errorf("Degree for node %s: expected 3, found %d", nlt2.ToString(), dg)
	}

	m := Q.GetNeighbors(*nlt2)
	if len(m) != 3 {
		t.Errorf("GetNeighbors for %s found %d, expected 3", nlt2.ToString(), len(m))
	}

	t.Logf("Expect 5:control,PHL, 1:control,SLTC, and 5: control,SLTC in some order")
	t.Logf("Neighbors for %s:", nlt2.ToString())
	for k,v := range m {
		t.Logf("%s wt = %f", k.ToString(), v)
	}

	deleted, err := Q.RemoveEdge(*nlt2, *NewNodeLayerTuple(1, "control,SLTC"))
	if !deleted {
		t.Error(err.Error())
	}
	if deleted {
		dg = Q.Degree(*nlt2)
		if dg !=2 {
			t.Errorf("Deleted edge from %s, but got degree %d instead of 2", nlt2.ToString(), dg)
		}
	}

	arr := Q.Indices("process")
	aBreak := false
	bBreak := false
	if len(arr) != 3 {
		aBreak = true
		t.Error("Indices for aspect 'process' wrong length")
	}
	for i:= 0; i < len(arr); i++ {
		if arr[i] != procIndices[i] {
			bBreak = true
		}
	}
	if bBreak {
		t.Error("One or more indices for aspect 'process' is/are incorrect")
	}
	if aBreak || bBreak {
		for j := 0; j < len(arr); j++ {
			t.Log(arr[j])
		}

		t.Errorf("Indices for aspect 'process' erroneously reported")

		for i := 0; i < len(arr); i++ {
			t.Logf("%s ", arr[i])
		}
	}

	if Q.IsNodeAligned() {
		t.Error("ML Network incorrectly reported as node-aligned")
	}

	in := Q.InDegree(*nlt1)
	out := Q.OutDegree(*nlt1)
	if in != 1 {
		t.Errorf("In degree for %s: expected 1, saw %d", nlt1.ToString(), in)
	}
	if out != 2 {
		t.Errorf("Out degree for %s: expected 2, saw %d", nlt2.ToString(), out)
	}

}