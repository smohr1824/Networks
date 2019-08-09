// Copyright 2017 - 2018  Stephen T. Mohr
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
	"strings"
	"strconv"
)

type GMLTokenizer struct {
	deadchars string;
}

func NewGMLTokenizer() *GMLTokenizer{
	tokenizer := new (GMLTokenizer)
	tokenizer.deadchars = "\t \r\n"
	return tokenizer
}

func (tokenizer *GMLTokenizer) EatWhitespace(reader *bufio.Reader){
	ch, _, err := reader.ReadRune()
	if err == nil {
		for ; err == nil && strings.ContainsRune(tokenizer.deadchars, ch); {
			if string(ch) == "#" {
				reader.UnreadRune()
				reader.ReadLine()
			}
			ch, _, err = reader.ReadRune()
		}
		if err == nil {
			reader.UnreadRune()
		}
	}
}

func (tokenizer *GMLTokenizer) ReadNextToken (reader *bufio.Reader) string {
	token := ""
	ch, _, err := reader.ReadRune()
	if err == nil {
		for ; err == nil && !strings.ContainsRune(tokenizer.deadchars, ch); {
			token = token + string(ch)
			if string(ch) == "[" || string(ch) == "]" {
				break
			}
			ch, _, err = reader.ReadRune()
		}
	}
	return token
}

func (tokenizer *GMLTokenizer) ReadListRecord (reader *bufio.Reader) map[string] string {
	props := make(map[string]string)

	tokenizer.EatWhitespace(reader)

	key := tokenizer.ReadNextToken(reader)
	if key == "[" {
		tokenizer.EatWhitespace(reader)
		key = tokenizer.ReadNextToken(reader)
	}

	for ; key != "]"; {
		tokenizer.EatWhitespace(reader)
		value := tokenizer.ReadNextValue(reader)
		if value == "[" {
			nestLevel := 1
			ch, _, err := reader.ReadRune()

			// process possibly nested records
			for ; err == nil && nestLevel > 0; {
				if string(ch) == "[" {
					nestLevel++
				}

				if string(ch) == "]" {
					nestLevel--
				}

				value = value + string(ch)
				ch, _, err = reader.ReadRune()
			}
		}
		(props)[key] = value
		tokenizer.EatWhitespace(reader)
		key = tokenizer.ReadNextToken(reader)
	}

	return props
}

func (tokenizer *GMLTokenizer) ReadNextValue(reader *bufio.Reader) string {
	value := ""

	ch, _, err := reader.ReadRune()
	if err == nil {
		if string(ch) != "'" && string(ch) != "\"" {
			for ; err == nil; {
				if !strings.ContainsRune(tokenizer.deadchars, ch) {
					value += string(ch)
					ch, _, err = reader.ReadRune()
				} else {
					break
				}
			}
		} else {
			// ch indicates a quoted string literal
			ch, _, err = reader.ReadRune()
			for ;err == nil && (string(ch) != "'") && (string(ch) != "\""); {
				value += string(ch)
				ch, _, err = reader.ReadRune()
			}
		}
	}
	return value
}

func (tokenizer *GMLTokenizer) PositionStartOfRecordOrArray(reader *bufio.Reader) int {
	retVal := 0

	tokenizer.EatWhitespace(reader)
	key := tokenizer.ReadNextToken(reader)
	if key == "[" {
		tokenizer.EatWhitespace(reader)
	} else {
		retVal = -1
	}

	return retVal
}

func (tokenizer *GMLTokenizer) ProcessFloatProp(prop string) (float32, error) {
	val, err := strconv.ParseFloat(prop, 32)
	return float32(val), err
}

func (tokenizer *GMLTokenizer) ConsumeUnknownValue(reader *bufio.Reader) {
	tokenizer.EatWhitespace(reader)
	value := tokenizer.ReadNextValue(reader)
	if value == "[" {
		tokenizer.ReadListRecord(reader)
	}
}


