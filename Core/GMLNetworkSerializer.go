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

// top level control of GML serialization/deserialization
package Core

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

func WriteNetworkToFile(net *Network, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return NewIoCreateError(fmt.Sprintf("Error creating %s for output: %s", filename, err.Error()))
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	writeNetwork(net, w)
	w.Flush()
	return nil
}

func writeNetwork(net *Network, writer *bufio.Writer) {
	net.ListGML(writer, 0)
}

func ReadNetworkFromFile(filename string) (*Network, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	f.Seek(0, io.SeekStart)
	reader := bufio.NewReader(f)

	return readNetwork(reader)
}

func readNetwork(reader *bufio.Reader) (*Network, error) {
	gmlTokenizer := NewGMLTokenizer()
	gmlTokenizer.EatWhitespace(reader)
	top := gmlTokenizer.ReadNextToken(reader)
	if top == "graph" {
		gmlTokenizer.EatWhitespace(reader)
		start := gmlTokenizer.ReadNextToken(reader)
		if start == "[" {
			gmlTokenizer.EatWhitespace(reader)
			net, err := processGraph(reader)
			return net, err
		}
	}
	return nil, errors.New("Top level structure wrong, could not read GML network")
}

func processGraph(reader *bufio.Reader) (*Network, error) {
	globalState := 1
	created := false
	gmlTokenizer := NewGMLTokenizer()
	unfinished := true
	net := NewNetwork(true)

	for ; unfinished; {
		// EOF check
		_, _, err := reader.ReadRune()
		if err != nil {
			break
		} else {
			reader.UnreadRune()
		}
		gmlTokenizer.EatWhitespace(reader)
		token := gmlTokenizer.ReadNextToken(reader)
		switch(token) {
			case "directed":
				gmlTokenizer.EatWhitespace(reader)
				if gmlTokenizer.ReadNextValue(reader) == "1" {
					net = NewNetwork(true)
					created = true
				} else {
					net = NewNetwork(false)
					created = true
				}

			case "node":
				if globalState < 3 {
					globalState = 2
					nodeProps := gmlTokenizer.ReadListRecord(reader)
					sId, ok := nodeProps["id"]
					if ok {
						id, err := processNodeId(sId)
						if err == nil {
							net.AddVertex(id)
						}
					} else {
						return nil, errors.New("Missing node id")
					}
				} else {
					return nil, errors.New("Node record found out of place in file")
				}

			case "edge":
				if globalState <= 3 {
					globalState = 3
					edgeProps := gmlTokenizer.ReadListRecord(reader)
					src, ok1 := edgeProps["source"]
					tgt, ok2 := edgeProps["target"]
					wt, ok3 := edgeProps["weight"]
					if ok1 && ok2 && ok3 {
						srcId, err1 := processNodeId(src)
						tgtId, err2 := processNodeId(tgt)
						wtVal, err3 := gmlTokenizer.ProcessFloatProp(wt)
						if err1 == nil && err2 == nil && err3 == nil {
							err := net.AddEdge(srcId, tgtId, wtVal)
							if err != nil {
								return nil, errors.New("Error adding edge")
							}
						}
					}

				} else {
					return nil, errors.New("Edge record found out of order in file")
				}

			case "]":
				unfinished = false

			default:
				gmlTokenizer.ConsumeUnknownValue(reader)

		}
	}

	if !created {
		return nil, errors.New("Network not created, directed property not found")
	}
	return net, nil
}

func processNodeId(sId string) (uint32, error) {
	id, err := strconv.ParseUint(sId, 10, 32)
	if err != nil {
		return 0, errors.New("Error reading " + sId + "; " + err.Error())
	}

	return uint32(id), nil
}
