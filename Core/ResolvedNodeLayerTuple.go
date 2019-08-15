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

package Core

import (
	"bufio"
)

type resolvedNodeLayerTuple struct {
	NodeId uint32
	Coordinates string // map type cannot use []string as a key, so this is a comma delimited list of integer aspect indices
}

func newresolvedNodeLayerTuple(id uint32, coordinates string) *resolvedNodeLayerTuple {
	p := new (resolvedNodeLayerTuple)
	p.NodeId = id
	p.Coordinates = coordinates
	return p
}

func (p *resolvedNodeLayerTuple) IsSameElementaryLayer(b resolvedNodeLayerTuple) bool {
	if p.ToString() == b.ToString() {
		return true
	} else {
		return false
	}
}

func (p *resolvedNodeLayerTuple) AreSameElementaryLayer(b []int) bool {
	s := ""
	for i, c := range b {
		s += string(c)
		if i < len(b) - 1 {
			s += ","
		}
	}
	if p.ToString() == s {
		return true
	} else {
		return false
	}
}

func (p *resolvedNodeLayerTuple) List(writer bufio.Writer) {
	writer.WriteString(p.ToString())
}

func (p *resolvedNodeLayerTuple) ToString() string {
	retVal := ""
	for i, coord := range p.Coordinates {
		retVal += string(coord)
		if i < len(p.Coordinates) - 1 {
			retVal += ","
		}
	}
	return retVal
}