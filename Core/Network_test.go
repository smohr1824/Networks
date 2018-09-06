// Copyright 2017 - 2018 Stephen T. Mohr, OSIsoft, LLC
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
	"testing"
	"strings"
)

func TestStartingVertex(t *testing.T) {
	N := NewNetwork(true)

	N.AddVertex("A")
	N.AddVertex("B")
	N.AddVertex("C")
	N.AddVertex("D")
	N.AddVertex("E")
	N.AddVertex("F")

	s := N.StartingVertex(true)
	if s != "" {
		t.Errorf("Expected empty string, saw %s", s)
	}

	N.AddEdge("E", "F", 1.0)
	s = N.StartingVertex(true)
	if s != "E" {
		t.Errorf("Expected starting vertex of E, got %s", s)
	}
	N.AddEdge("A", "B", 1.0)
	N.AddEdge("B", "C", 1.0)

	s = N.StartingVertex(true)
	if !strings.Contains("ABE", s){
		t.Errorf("Expected one of A, B, E but saw %s", s)
	}
}
