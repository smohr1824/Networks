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
	"bufio"
	"errors"
	. "fmt"
	"github.com/smohr1824/Networks/Core"
	"io"
	"os"
	"strings"
)

type MLFCMSerializer struct {
	tokenizer *Core.GMLTokenizer
}

func NewMLFCMSerializer() *MLFCMSerializer {
	s := new(MLFCMSerializer)
	s.tokenizer = Core.NewGMLTokenizer()
	return s
}

func (s *MLFCMSerializer) WriteMLFCMToFile(fcm *MultilayerFuzzyCognitiveMap, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return Core.NewIoCreateError(Sprintf("Error creating %s for output: %s", filename, err.Error()))
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	s.WriteMLFCM(fcm, w)
	w.Flush()
	return nil
}

func (s *MLFCMSerializer)WriteMLFCM(fcm *MultilayerFuzzyCognitiveMap, writer *bufio.Writer) error {
	return fcm.ListGML(writer)
}

func (s *MLFCMSerializer) ReadMLFCMFromFile(filename string) (*MultilayerFuzzyCognitiveMap, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	f.Seek(0, io.SeekStart)
	reader := bufio.NewReader(f)

	return s.ReadMLFCM(reader)
}

func (s *MLFCMSerializer) ReadMLFCM(reader *bufio.Reader) (*MultilayerFuzzyCognitiveMap, error) {
	s.tokenizer.EatWhitespace(reader)
	top := s.tokenizer.ReadNextToken(reader)
	if top == "multilayer_network" {
		s.tokenizer.EatWhitespace(reader)
		start := s.tokenizer.ReadNextToken(reader)
		if start == "[" {
			s.tokenizer.EatWhitespace(reader)
			Q, err := s.processMLFCM(reader)
			return Q, err
		}
	} else {
		return nil, errors.New("Incorrect top level record")
	}
	return nil, errors.New("Unknown error")
}

func (s *MLFCMSerializer) processMLFCM(reader *bufio.Reader) (*MultilayerFuzzyCognitiveMap, error) {
	globalState := 1
	unfinished := true
	ttype := Bivalent
	modified := false
	var graph *MultilayerFuzzyCognitiveMap = nil

	for ; unfinished; {
		// EOF check
		_, _, err := reader.ReadRune()
		if err != nil {
			break
		} else {
			_ = reader.UnreadRune()
		}

		s.tokenizer.EatWhitespace(reader)
		token := s.tokenizer.ReadNextToken(reader)
		switch strings.ToLower(token) {
		case "directed":
			if globalState == 1 {
				s.tokenizer.EatWhitespace(reader)
				s.tokenizer.ReadNextToken(reader)
			} else {
				return nil, errors.New("Property directed found out of order")
			}
		case "threshold":
			if globalState == 1 {
				s.tokenizer.EatWhitespace(reader)
				threshname := s.tokenizer.ReadNextValue(reader)
				switch strings.ToLower(threshname) {
				case "bivalent":
					ttype = Bivalent
				case "trivalent":
					ttype = Trivalent
				case "logistic":
					ttype = Logistic
				case "custom":
					ttype = Custom
				}
			} else {
				return nil, errors.New("Property threshold found out of order")
			}
		case "rule":
			if globalState == 1 {
				s.tokenizer.EatWhitespace(reader)
				rulename := s.tokenizer.ReadNextValue(reader)
				if strings.ToLower(rulename) == "modified" {
					modified = true
				} else {
					modified = false
				}
			} else {
				return nil, errors.New("Property rule found out of order")
			}
		case "aspects":
			if globalState == 1 {
				globalState = 2
				aspects, indices := Core.ReadAspects(reader)
				graph = NewMultilayerFuzzyCognitiveMap(aspects, indices, modified, ttype)
			} else {
				return nil, errors.New("Property aspects found out of order")
			}
		case "concept":
			if globalState >= 2 && globalState <= 3 {
				globalState = 3
				nodeProps := s.tokenizer.ReadListRecord(reader)
				id, ok := nodeProps["id"]
				if ok {
					concept, err := s.processConcept(nodeProps)
					if err != nil {
						return nil, err
					}
					uId, err := processNodeId(id)
					if err == nil {
						if! graph.AddConcept(concept, uId) {
							return nil, errors.New(Sprintf("Concept %s, id = %d already exists in the network or cannot be added", concept.Name, uId))
						}
					}
				} else {
					return nil, errors.New("Id missing for concept record")
				}
			} else {
				return nil, errors.New("Concept record found out of order")
			}
		case "]":
			unfinished = false
		case "layer":
			if globalState >= 3 && globalState <= 4 {
				globalState = 4
				err := s.readLayer(reader, graph)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("Layer record found out of order")
			}
		case "edge":
			if globalState >= 4 && globalState <= 5 {
				globalState = 5
				err := s.readInterlayerEdge(reader, graph)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("Interlayer edge record found out of order")
			}
		default:
			s.tokenizer.ConsumeUnknownValue(reader)
		}
	}
	for _, concept := range graph.concepts {
		graph.recomputeAggregateActivationLevel(concept.Name)
	}
	return graph, nil
}

func (s *MLFCMSerializer) readLayer(reader *bufio.Reader, fcm *MultilayerFuzzyCognitiveMap) error {
	if s.tokenizer.PositionStartOfRecordOrArray(reader) != -1 {
		key := s.tokenizer.ReadNextToken(reader)
		if strings.ToLower(key) == "coordinates" {
			s.tokenizer.EatWhitespace(reader)
			coords := s.tokenizer.ReadNextValue(reader)
			s.tokenizer.EatWhitespace(reader)
			network, err := Core.ReadNetwork(reader)
			if err == nil {
				fcm.AddElementaryLayer(coords, network)
				s.tokenizer.EatWhitespace(reader)
				s.tokenizer.ReadNextToken(reader)
			} else {
				return errors.New("Error deserializing network for elementary layer " + coords)
			}
		} else {
			return errors.New("Missing coordinates on layer record")
		}
	} else {
		return errors.New("Malformed elementary layer record")
	}
	return nil
}

func (s *MLFCMSerializer) readInterlayerEdge(reader *bufio.Reader, fcm *MultilayerFuzzyCognitiveMap) error {
	edgeWt := float32(1.0)
	edgeProps := s.tokenizer.ReadListRecord(reader)
	_, okSource := edgeProps["source"]
	_, okTarget := edgeProps["target"]
	src := Core.NodeLayerTuple{ NodeId: 0, Coordinates: ""}
	tgt := Core.NodeLayerTuple{ NodeId: 0, Coordinates:"" }
	var err error
	if okSource && okTarget {
		for key, value := range edgeProps {
			switch strings.ToLower(key) {
			case "source":
				sreader := bufio.NewReader(strings.NewReader(value))
				src, err = Core.ProcessQualifiedNode(sreader)
				if err != nil {
					return err
				}
			case "target":
				sreader := bufio.NewReader(strings.NewReader(value))
				tgt, err = Core.ProcessQualifiedNode(sreader)
				if err != nil {
					return err
				}
			case "weight":
				edgeWt, err = s.tokenizer.ProcessFloatProp(value)
				if err != nil {
					return err
				}
			}
		}

	} else {
		return errors.New("missing source or target for influence (edge) record")
	}
	fcm.AddInfluenceTuples(src, tgt, edgeWt)
	return nil
}

func (s *MLFCMSerializer) processConcept(props map[string] string) (MultilayerCognitiveConcept, error) {
	layerLevels := make(map[string] string)
	concept := MultilayerCognitiveConcept{
		Name:                  "",
		initialValue:          0,
		ActivationLevel:       0,
		layerActivationLevels: nil,
	}
	aggregate := float32(0.0)
	label, labelOk := props["label"]
	initial, initialOk := props["initial"]
	if labelOk && initialOk {
		finitial, err := s.tokenizer.ProcessFloatProp(initial)
		if err != nil {
			return MultilayerCognitiveConcept{
				Name:                  "",
				initialValue:          0,
				ActivationLevel:       0,
				layerActivationLevels: nil,
			}, err
		}
		for key, value := range props {
			switch strings.ToLower(key){
			case "aggregate":
				aggregate, err = s.tokenizer.ProcessFloatProp(value)
				if err != nil {
					return MultilayerCognitiveConcept{
						Name:                  "",
						initialValue:          0,
						ActivationLevel:       0,
						layerActivationLevels: nil,
					}, err
				}
			case "levels":
				sreader := bufio.NewReader(strings.NewReader(value))
				layerLevels = s.tokenizer.ReadListRecord(sreader)
			}
		}
		_, ok := props["aggregate"]
		if !ok {
			aggregate = finitial
		}
		concept = *NewMultilayerCognitiveConcept(label, finitial, aggregate)
		for key, value := range layerLevels {
			levelValue, err := s.tokenizer.ProcessFloatProp(value)
			if err == nil {
				concept.setLayerLevel(key, levelValue)
			}
		}
	}
	return concept, nil
}

