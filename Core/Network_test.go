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

// core Network tests
package Core

import (
	"math"
	"testing"
)

func TestStartingVertex(t *testing.T) {
	N := NewNetwork(true)

	N.AddVertex(0) // A
	N.AddVertex(1) // B
	N.AddVertex(2) // C
	N.AddVertex(3) // D
	N.AddVertex(4) // E
	N.AddVertex(5) // F

	s, err := N.StartingVertex(true)
	if err == nil {
		t.Errorf("Expected err, saw %d", s)
	}

	err = N.AddEdge(4, 5, 1.0)
	if err != nil {
		panic(err)
	}
	s, err = N.StartingVertex(true)
	if s != 4 {
		t.Errorf("Expected starting vertex of E, got %d", s)
	}
	err = N.AddEdge(0, 1, 1.0)
	err = N.AddEdge(1, 2, 1.0)
	if err != nil {
		panic(err)
	}

	s, err = N.StartingVertex(true)
	if !(s == 0 || s == 1 || s == 4){
		t.Errorf("Expected one of 0, 1, 4 but saw %d", s)
	}
}

func TestVertexBasic(t *testing.T) {
	N := NewNetwork(true)

	err := N.AddEdge(1, 2, 1)
	err = N.AddEdge(1, 3, 1)
	err = N.AddEdge(2, 3, 2)
	err = N.AddEdge(1, 4, 3)

	if err != nil {
		panic(err)
	}
	if N.Order() != 4 {
		t.Errorf("Wrong number of vertices, failed Order()")
	}

	if len(N.GetNeighbors(1)) != 3 {
		t.Errorf("Vertex 1 should have three neighbors, found %d", len(N.GetNeighbors(1)))
	}

	if !N.HasEdge(2, 3) {
		t.Error("Did not find edge from 2 to 3")
	}

	if N.HasEdge(2, 1) {
		t.Error("Found unexpected edge between 2 and 1")
	}

	n, err := N.OutDegree(1)
	if err != nil {
		t.Errorf(err.Error())
	} else {
		if n != 3 {
			t.Error("Wrong number of out edges from vertex 1")
		}
	}

	N.RemoveVertex(3)
	if N.HasVertex(3) {
		t.Error("Vertex 3 not removed")
	}
	if N.HasEdge(1, 3) {
		t.Error("Found edge from 1 to 3 after deleting vertex 3")
	}
}

func TestSize(t *testing.T) {
	G := makeSimple(true)
	size := G.Size()
	if size != 9 {
		t.Errorf("Wrong size computed for directed network. Expected 9, got %d.", size)
	}

	G = makeSimple(false)
	size = G.Size()
	if size != 9 {
		t.Errorf("Wrong size computed for undirected network. Expected 9, got %d.", size)
	}
}

func TestDensity(t *testing.T) {
	G := NewNetwork(true)
	err := G.AddEdge(1, 2, 1);
	err = G.AddEdge(1, 3, 1);
	err = G.AddEdge(2, 3, 2);
	err = G.AddEdge(1, 4, 3);

	if err != nil {
		panic(err)
	}
	density := G.Density()
	if math.Abs((0.33 - density)) > 0.01 {
		t.Errorf("Wrong density computed for directed graph.  Expected 0.33, got %f.", density)
	}

	G = NewNetwork(false);
	err = G.AddEdge(1, 2, 1);
	err = G.AddEdge(1, 3, 1);
	err = G.AddEdge(2, 3, 2);
	err = G.AddEdge(1, 4, 3);

	if err != nil {
		panic(err)
	}
	density = G.Density()
	if math.Abs((0.66 - density)) > 0.01 {
		t.Errorf("Wrong density computed for undirected graph.  Expected 0.66, got %f.", density)
	}
}

func makeSimple(directed bool) *Network {
	G := NewNetwork(directed)
	err := G.AddEdge(1, 2, 1.0)
	err = G.AddEdge(1, 3, 1.0)
	err = G.AddEdge(1, 6, 1.0);
	err = G.AddEdge(2, 4, 1.0);
	err = G.AddEdge(4, 6, 1.0);
	err = G.AddEdge(3, 5, 1.0);
	err = G.AddEdge(5, 6, 1.0);
	err = G.AddEdge(2, 5, 1.0);
	err = G.AddEdge(3, 4, 1.0);if err != nil {
		panic(err)
	}
	return G
}
