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

// Serialization/deserialization of a simple graph using text-format adjacency lists
package Core

import (
	"os"
	"fmt"
	"strings"
	"bufio"
	"strconv"
)

type NetworkSerializer struct {
	delimiter string
}

func NewNetworkSerializer(Delimiter string) *NetworkSerializer {
	serializer := new(NetworkSerializer)
	serializer.delimiter = Delimiter
	return serializer
}

func NewDefaultNetworkSerializer() *NetworkSerializer {
	serializer := new(NetworkSerializer)
	serializer.delimiter = "|"
	return serializer
}

func (serializer *NetworkSerializer) ReadNetworkFromFile(filename string, directed bool) (*Network, error) {
	f, err := os.Open(filename)
	if err != nil {
		return NewNetwork(directed), err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	retVal := serializer.readNetwork(scanner, directed)
	return retVal, err

}


func (serializer *NetworkSerializer) WriteNetworkToFile(net *Network, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return NewIoCreateError(fmt.Sprintf("Error creating %s for output: %s", filename, err.Error()))
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	serializer.writeNetwork(net, w)
	w.Flush()
	return nil
}

// read a network in edge list format
func (serializer *NetworkSerializer) readNetwork(scanner *bufio.Scanner, directed bool) *Network {
	network := NewNetwork(directed)

	for scanner.Scan() {
		fields := splitAndClean(scanner.Text(), serializer.delimiter)
		ct := len(fields)
		if ct == 1 {
			// vertex only, just add
			network.AddVertex(fields[0])
			continue;
		}

		if ct > 3 {
			continue
		}

		var wt float32 = 1.0
		if ct == 3 {
			wtWide, err := strconv.ParseFloat(fields[2], 32)
			// if there is a parse error, go with the default of 1
			if err == nil {
				wt = float32(wtWide)
			}
			network.AddEdge(fields[0], fields[1], wt)
		}

		if ct == 2 && fields[1] == "" {
			network.AddVertex(fields[0])
			continue
		}

		if ct == 2 && fields[0] == "" {
			continue
		}

		network.AddEdge(fields[0], fields[1], 1)


	}
	return network
}

func (serializer *NetworkSerializer) writeNetwork(net *Network, writer *bufio.Writer) {
	net.List(writer, serializer.delimiter)
}

func splitAndClean(line string, delimiter string) []string {
	fields := strings.Split(line, delimiter)
	retVal := make([]string, 0, len(fields))
	leng := len(fields)
	for i:=0; i < leng; i++ {
		//if len(strings.TrimSpace(fields[i])) > 0 {
			retVal = append(retVal, strings.TrimSpace(fields[i]))
		//}
	}
	return retVal
}