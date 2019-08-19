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

// parser for Multilayer network GML -- uses lexer GMLTokenizer

package Core

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func WriteMultilayerNetworkToFile(M *MultilayerNetwork, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return NewIoCreateError(fmt.Sprintf("Error creating %s for output: %s", filename, err.Error()))
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	WriteMultilayerNetwork(M, w)
	w.Flush()
	return nil
}

func WriteMultilayerNetwork(M *MultilayerNetwork, writer *bufio.Writer) {
	M.ListGML(writer)
}

func ReadMultilayerNetworkFromFile(filename string) (*MultilayerNetwork, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	f.Seek(0, io.SeekStart)
	reader := bufio.NewReader(f)

	return ReadMultilayerNetwork(reader)
}

func ReadMultilayerNetwork(reader *bufio.Reader) (*MultilayerNetwork, error) {
	tokenizer := NewGMLTokenizer()
	tokenizer.EatWhitespace(reader)

	top := tokenizer.ReadNextToken(reader)
	if top == "multilayer_network" {
		tokenizer.EatWhitespace(reader)
		start := tokenizer.ReadNextToken(reader)
		if start == "[" {
			tokenizer.EatWhitespace(reader)
			Q, err := processMLNetwork(reader)
			return Q, err
		} else {
			return nil, errors.New("malformed top-level record (missing opening bracket)")
		}
	} else {
		return nil, errors.New("incorrect top level record")
	}
}

func processMLNetwork(reader *bufio.Reader) (*MultilayerNetwork, error) {
	globalState := 1
	created := false
	directed := false
	gmlTokenizer := NewGMLTokenizer()
	unfinished := true
	 Q := NewMultilayerNetwork(nil, nil, false)

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
		switch (token) {
			case "directed":
			gmlTokenizer.EatWhitespace(reader)
			if gmlTokenizer.ReadNextValue(reader) == "1" {
				directed = true
			} else {
				directed = false
			}

			case "aspects":
				if globalState == 1 {
					globalState = 2
					aspects, indices := readAspects(reader)
					if len(aspects) == 0 {
						return nil, errors.New("No aspects read")
					}
					Q = NewMultilayerNetwork(aspects, indices, directed)
				} else {
					return nil, errors.New("aspects record found out of place")
				}
				created = true

			case "layer":
				if globalState > 1 && globalState <= 3 {
					globalState = 3
					err := readLayer(reader, Q)
					if err != nil {
						return nil, err
					}
				} else {
					return nil, errors.New("layer record found out of place")
				}

		case "edge":
				if globalState >=3 && globalState <= 4 {
					globalState = 4
					err := readInterlayerEdge(reader, Q)
					if err != nil {
						return nil, err
					}
				} else {
					return nil, errors.New("interlayer edge record found out of place")
				}

			case "]":
				unfinished = false

			default:
				gmlTokenizer.ConsumeUnknownValue(reader)

		}
	}

	if !created {
		return nil, errors.New("multilayer network not created, aspects record not found")
	} else {
		return Q, nil
	}
}

func readAspects(reader *bufio.Reader) ([]string, [][]string) {
	gmlTokenizer := NewGMLTokenizer()
	aspectDictionary := gmlTokenizer.ReadListRecord(reader)
	aspects := make([]string, 0)
	indices := make([][]string, 0)
	for aspect, indexValues := range aspectDictionary {
		aspects = append(aspects, aspect)
		indices = append(indices, strings.Split(indexValues, ","))
	}
	return aspects, indices
}

func readLayer(reader *bufio.Reader, graph *MultilayerNetwork) error {
	gmlTokenizer := NewGMLTokenizer()
	if gmlTokenizer.PositionStartOfRecordOrArray(reader) != -1 {
		key := gmlTokenizer.ReadNextToken(reader)
		if key == "coordinates" {
			gmlTokenizer.EatWhitespace(reader)
			coords := gmlTokenizer.ReadNextValue(reader)
			gmlTokenizer.EatWhitespace(reader)
			net, err := readNetwork(reader)
			if err != nil {
				return errors.New("error deserializing network for elementary layer " + coords)
			} else {
				graph.AddElementaryLayer(coords, net)
				gmlTokenizer.EatWhitespace(reader)
				gmlTokenizer.ReadNextToken(reader)
				return nil
			}
		} else {
			return errors.New("missing coordinates on layer record")
		}
	} else {
		return errors.New("malformed elementary layer record")
	}
}

func readInterlayerEdge(reader *bufio.Reader, graph *MultilayerNetwork) error {
	gmlTokenizer := NewGMLTokenizer()
	edgeProps := gmlTokenizer.ReadListRecord(reader)
	_, okSrc := edgeProps["source"]
	_, okTgt := edgeProps["target"]
	src := NodeLayerTuple{NodeId:0, Coordinates:""}
	tgt := NodeLayerTuple{NodeId:0, Coordinates:""}
	err := errors.New("")
	var wt float32
	wt = 1.0

	if okSrc && okTgt {
		for k, v := range edgeProps {
			switch strings.ToLower(k) {
				case "source":
					sreader := bufio.NewReader(strings.NewReader(v))
					src, err = processQualifiedNode(sreader)
					if err != nil {
						return err
					}

				case "target":
					sreader := bufio.NewReader(strings.NewReader(v))
					tgt, err = processQualifiedNode(sreader)
					if err != nil {
						return err
					}

				case "weight":
					f, err := gmlTokenizer.ProcessFloatProp(v)
					if err != nil {
						return errors.New("unable to convert value of weight of an interlayer edge to a floating point value")
					} else {
						wt = f
					}
			}
		}
		if src.Coordinates != "" && tgt.Coordinates != "" {
			graph.AddEdge(src, tgt, float32(wt))
		} else {
			return errors.New("Missing source and/or target for interlayer edge")
		}
	}
	return nil
}

func processQualifiedNode(reader *bufio.Reader) (NodeLayerTuple, error){
	gmlTokenizer := NewGMLTokenizer()
	nodeProps := gmlTokenizer.ReadListRecord(reader)
	id, okId := nodeProps["id"]
	coords, okCoords := nodeProps["coordinates"]
	if !okId {
		return NodeLayerTuple{NodeId:0, Coordinates:""}, interlayerEdgeError("id","<missing>")
	}
	if !okCoords {
		return NodeLayerTuple{NodeId:0, Coordinates:""}, interlayerEdgeError("coordinates", "<missing>")
	}
	nodeId, err := strconv.Atoi(id)
	if err != nil {
		return NodeLayerTuple{NodeId:0, Coordinates:""}, interlayerEdgeError("id", err.Error())
	}
	return *NewNodeLayerTuple(uint32(nodeId), coords), nil
}

func interlayerEdgeError(keyword string, msg string) error {
	return errors.New(fmt.Sprintf("error formatting or converting interlayer edge %s = %s", keyword, msg))
}

