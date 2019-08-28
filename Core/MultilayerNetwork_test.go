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

func TestMultilayerNeighbors(t *testing.T) {
	Q, err := ReadMultilayerNetworkFromFile("multilayer_three_aspects.gml")
	if err != nil {
		t.Error("Error reading test file multilayer_three_aspects.gml")
		return
	}

	nlt := NewNodeLayerTuple(uint32(2), "I,A,1")
	n := Q.GetNeighbors(*nlt)
	if len(n) != 14 {
		t.Errorf("Expected 14 neighbors, found %d", len(n))
	}

	nlt1 := NewNodeLayerTuple(3, "I,B,2")
	nlt2 := NewNodeLayerTuple(1, "I,A,1")
	nlt3 := NewNodeLayerTuple(3, "I,A,1")

	_, ok1 := n[*nlt1]
	_, ok2 := n[*nlt2]
	_, ok3 := n[*nlt3]
	if !ok1 || !ok2 || !ok3 {
		t.Errorf("One or more explicit neighbors missing")
		for ngbr, _ := range n {
			t.Logf("Neighbor: %d in layer %s", ngbr.NodeId, ngbr.Coordinates)
		}
	}

}

func TestSupraadjacency(t *testing.T) {
	Q, err := ReadMultilayerNetworkFromFile("multilayer_three_aspects.gml")
	if err != nil {
		t.Error("Error reading test file multilayer_three_aspects.gml")
		return
	}
	supra := Q.MakeSupraAdjacencyMatrix()

	// check a block on the diagonal, i.e., an elementary layer's adjacency matrix
	if !(supra[21][22] == 1 && supra[22][21] == 1 && supra[23][21] == 1) {
		t.Error("Failure to find adjacencies within II,B,2")
	}

	// check all the explicit interlayer adjacencies
	if !(supra[0][7] == 1 && supra[1][11] == 1 && supra[20][3] == 1 && supra[35][0] == 1) {
		t.Error("Failure to find interlayer adjacencies")
	}

	// test three elements that MUST BE zero
	if (supra[3][14] != 0 || supra[11][4] != 0 || supra[32][8] != 0) {
		t.Error("Found elements that should have been zero but were nozero")
	}

	// granted, we only tested 0.7% of the supra-adjacency matrix
}

func listSupraAdjacencyMatrix(t *testing.T) {
	Q, err := ReadMultilayerNetworkFromFile("multilayer_three_aspects.gml")
	if err != nil {
		t.Error("Error reading test file multilayer_three_aspects.gml")
		return
	}
	supra := Q.MakeSupraAdjacencyMatrix()
	for _, row := range supra {
		for _, elt := range row {
			t.Logf("%.2f ", elt)
		}
	}
}