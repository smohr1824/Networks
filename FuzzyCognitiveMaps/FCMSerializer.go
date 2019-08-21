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
	"fmt"
	. "github.com/smohr1824/Networks/Core"
	"io"
	"os"
	"strconv"
	"strings"
)

func WriteFCMToFile(fcm *FuzzyCognitiveMap, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return NewIoCreateError(fmt.Sprintf("Error creating %s for output: %s", filename, err.Error()))
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	writeNetwork(fcm, w)
	w.Flush()
	return nil
}

func writeNetwork(fcm *FuzzyCognitiveMap, writer *bufio.Writer) error {
	return fcm.ListGML(writer)
}

func ReadFCMFromFile(filename string) (*FuzzyCognitiveMap, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	f.Seek(0, io.SeekStart)
	reader := bufio.NewReader(f)

	return readFCM(reader)
}

func readFCM(reader *bufio.Reader) (*FuzzyCognitiveMap, error) {
	gmlTokenizer := NewGMLTokenizer()
	gmlTokenizer.EatWhitespace(reader)
	top := gmlTokenizer.ReadNextToken(reader)
	if top == "graph" {
		gmlTokenizer.EatWhitespace(reader)
		start := gmlTokenizer.ReadNextToken(reader)
		if start == "[" {
			gmlTokenizer.EatWhitespace(reader)
			net, err := processFCM(reader)
			return net, err
		}
	}
	return nil, errors.New("Top level structure wrong, could not read GML network")
}

func processFCM(reader *bufio.Reader) (*FuzzyCognitiveMap, error) {
	globalState := 1
	unfinished := true
	ttype := Bivalent
	conceptLookup := make(map[uint32] string)
	modified := false
	var graph *FuzzyCognitiveMap = nil

	for ; unfinished; {
		// EOF check
		_, _, err := reader.ReadRune()
		if err != nil {
			break
		} else {
			reader.UnreadRune()
		}

		gmlTokenizer := NewGMLTokenizer()
		gmlTokenizer.EatWhitespace(reader)
		token := gmlTokenizer.ReadNextToken(reader)
		switch(strings.ToLower(token)) {
		case "directed":
			if globalState == 1 {
				gmlTokenizer.EatWhitespace(reader)
				gmlTokenizer.ReadNextValue(reader)
			} else {
				return nil, errors.New(fmt.Sprintf("Property %s found out of order", token))
			}

		case "threshold":
			if globalState == 1 {
				gmlTokenizer.EatWhitespace(reader)
				threshname := gmlTokenizer.ReadNextValue(reader)
				switch(strings.ToLower(threshname)) {
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
				return nil, errors.New("Threshold property found ouf of order")
			}
		case "rule":
			if globalState == 1 {
				gmlTokenizer.EatWhitespace(reader)
				rulename := gmlTokenizer.ReadNextValue(reader)
				if strings.ToLower(rulename) == "modified" {
					modified = true
				} else {
					modified = false
				}

			} else {
				return nil, errors.New("Rule record found out of order")
			}

		case "node":
			if globalState <= 2 {
				globalState = 2
				if graph == nil {
					graph = NewFuzzyCognitiveMap(modified, ttype)
				}
				nodeDictionary := gmlTokenizer.ReadListRecord(reader)
				err := processConcept(nodeDictionary, graph, conceptLookup)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("Concept (node) record found out of order")
			}

		case "edge":
			if globalState > 1 && globalState <= 3 {
				globalState = 3
				edgeProps := gmlTokenizer.ReadListRecord(reader)
				err := processEdge(edgeProps, graph, conceptLookup)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("Influence (edge) record found out of order")
			}

		case "]":
			unfinished = false
		default:
			gmlTokenizer.ConsumeUnknownValue(reader)
		}
	}
	return graph, nil
}

func processConcept(nodeProps map[string] string, graph *FuzzyCognitiveMap, lookup map[uint32] string) error {
	id, okId := nodeProps["id"]
	label, okLabel := nodeProps["label"]
	initial, okInitial := nodeProps["initial"]
	if okId && okLabel && okInitial {
		cId, err := processNodeId(id)
		if err != nil {
			return errors.New("Error converting a concept's node id")
		}
		gmlTokenizer := NewGMLTokenizer()
		finitial, err := gmlTokenizer.ProcessFloatProp(initial)
		if err != nil {
			return errors.New("Error converting initial activation value for concept")
		}
		activation, ok := nodeProps["activation"]
		factivation := float32(0.0)
		if ok {
			factivation, err = gmlTokenizer.ProcessFloatProp(activation)
			if err != nil {
				return errors.New("Error converting activation value")
			}
		} else {
			factivation = finitial
		}
		lookup[cId] = label
		graph.AddConcept(label, finitial, factivation)
	} else {
		if okLabel {
			return errors.New("Concept " + label + " missing one or more required properties")
		} else {
			return errors.New("Unnamed concept missing one or more required properties")
		}
	}
	return nil
}

func processEdge(edgeProps map[string] string, graph *FuzzyCognitiveMap, lookup map[uint32] string) error {
	src, okSrc := edgeProps["source"]
	tgt, okTgt := edgeProps["target"]
	wt, okWt := edgeProps["weight"]

	if okSrc && okTgt && okWt {
		gmlTokenizer := NewGMLTokenizer()
		srcId, errSrc := processNodeId(src)
		tgtId, errTgt := processNodeId(tgt)
		fwt, errWt := gmlTokenizer.ProcessFloatProp(wt)

		if errSrc == nil && errTgt == nil && errWt == nil {
			graph.AddInfluence(lookup[srcId], lookup[tgtId], fwt)
		} else {
			return errors.New("Error converting source, target, or weight proeprty values.")
		}
	} else {
		return errors.New("Influence (edge) missing one or more required properties")
	}
	return nil
}

func processNodeId(sId string) (uint32, error) {
	id, err := strconv.ParseUint(sId, 10, 32)
	if err != nil {
		return 0, errors.New("Error reading " + sId + "; " + err.Error())
	}

	return uint32(id), nil
}

