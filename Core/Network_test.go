// Copyright 2017 - 2018 Stephen T. Mohr
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
	"fmt"
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

	N.AddEdge(4, 5, 1.0)
	s, err = N.StartingVertex(true)
	if s != 4 {
		t.Errorf("Expected starting vertex of E, got %d", s)
	}
	N.AddEdge(0, 1, 1.0)
	N.AddEdge(1, 2, 1.0)

	s, err = N.StartingVertex(true)
	if !(s == 0 || s == 1 || s == 4){
		t.Errorf("Expected one of 0, 1, 4 but saw %d", s)
	}
}

func TestVertextBasic(t *testing.T) {
	N := NewNetwork(true)

	N.AddEdge(1, 2, 1)
	N.AddEdge(1, 3, 1)
	N.AddEdge(2, 3, 2)
	N.AddEdge(1, 4, 3)

	if N.Order() != 4 {
		t.Errorf("Wrong number of vertices, failed Order()")
	}

	if len(N.GetNeighbors(1)) != 3 {
		t.Errorf(fmt.Sprintf("Vertex 1 should have three neighbors, found %d", N.GetNeighbors((1))))
	}

	if !N.HasEdge(2, 3) {
		t.Errorf("Did not find edge from 2 to 3")
	}

	if N.HasEdge(2, 1) {
		t.Errorf("Found unexpected edge between 2 and 1")
	}

	n, err := N.OutDegree(1)
	if err != nil {
		t.Errorf(err.Error())
	} else {
		if n != 3 {
			t.Errorf("Wrong number of out edges from vertex 1")
		}
	}

	N.RemoveVertex(3)
	if N.HasVertex(3) {
		t.Errorf("Vertex 3 not removed")
	}
	if N.HasEdge(1, 3) {
		t.Errorf("Found edge from 1 to 3 after deleting vertex 3")
	}
}
