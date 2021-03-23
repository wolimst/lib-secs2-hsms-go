package ast

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Testing Strategy:
//
// Create a new instance using the factory methods or FillVariables(),
// and test the result of public observer methods Size(), FillInStringLength(),
// Variables(), ToBytes(), and String().
//
// Partitions:
//
// When the node contains string literal
// - Length of the string: 0, 1, ...
// - Non-printable characters (LF, TAB, etc.) in string literal: true, false
// - Position of the non-printable characters: head, middle, tail
//
// When the node contains variable
// - Fill-in string min length: 0, 1, ...
// - Fill-in string max length: -1, 0, 1, ...
// - Fill-in string length: 0, 1, ...
// - Non-printable characters in fill-in string: true, false
// - Position of the non-printable chracters: head, middle, tail

func TestASCIINode_NoVariable_ProducedByFactoryMethod(t *testing.T) {
	var tests = []struct {
		description             string   // Test case description
		input                   string   // Input to the factory method
		expectedSize            int      // expected result from Size()
		expectedFillInStrLenMin int      // expected result of min from FillInStringLength()
		expectedFillInStrLenMax int      // expected result of max from FillInStringLength()
		expectedVariables       []string // expected result from Variables()
		expectedToBytes         []byte   // expected result from ToBytes()
		expectedString          string   // expected result from String()
	}{
		{
			description:             "Length: 0, Empty string literal",
			input:                   "",
			expectedSize:            0,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 0},
			expectedString:          `<A[0]>`,
		},
		{
			description:             "Length: 1",
			input:                   "A",
			expectedSize:            1,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 1, 65},
			expectedString:          `<A "A">`,
		},
		{
			description:             "Length: 2",
			input:                   ".*",
			expectedSize:            2,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 2, 0x2E, 0x2A},
			expectedString:          `<A ".*">`,
		},
		{
			description:             "Length: 11",
			input:                   "lorem ipsum",
			expectedSize:            11,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 11, 0x6C, 0x6F, 0x72, 0x65, 0x6D, 0x20, 0x69, 0x70, 0x73, 0x75, 0x6D},
			expectedString:          `<A "lorem ipsum">`,
		},
		{
			description:             "Length: 1, Non-printable char only",
			input:                   "\n",
			expectedSize:            1,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 1, 0x0A},
			expectedString:          `<A 0x0A>`,
		},
		{
			description:             "Length: 6, Non-printable chars at text head",
			input:                   "\r\ntext",
			expectedSize:            6,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 6, 0x0D, 0x0A, 0x74, 0x65, 0x78, 0x74},
			expectedString:          `<A 0x0D 0x0A "text">`,
		},
		{
			description:             "Length: 6, Non-printable chars at text tail",
			input:                   "text\n\x00",
			expectedSize:            6,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 6, 0x74, 0x65, 0x78, 0x74, 0x0A, 0x00},
			expectedString:          `<A "text" 0x0A 0x00>`,
		},
		{
			description:             "Length: 6, Non-printable chars in between texts",
			input:                   "te\x09\x7Fxt",
			expectedSize:            6,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 6, 0x74, 0x65, 0x09, 0x7F, 0x78, 0x74},
			expectedString:          `<A "te" 0x09 0x7F "xt">`,
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewASCIINode(test.input)
		min, max := node.(*ASCIINode).FillInStringLength()
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedFillInStrLenMin, min)
		assert.Equal(t, test.expectedFillInStrLenMax, max)
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestASCIINode_Variable_ProducedByFactoryMethod(t *testing.T) {
	var tests = []struct {
		description             string        // Test case description
		input                   []interface{} // Inputs to the factory method (name, minLength, maxLength)
		expectedSize            int           // expected result from Size()
		expectedFillInStrLenMin int           // expected result of min from FillInStringLength()
		expectedFillInStrLenMax int           // expected result of max from FillInStringLength()
		expectedVariables       []string      // expected result from Variables()
		expectedToBytes         []byte        // expected result from ToBytes()
		expectedString          string        // expected result from String()
	}{
		{
			description:             "Fill-in string length limit: 0, -1",
			input:                   []interface{}{"a", 0, -1},
			expectedSize:            -1,
			expectedFillInStrLenMin: 0,
			expectedFillInStrLenMax: -1,
			expectedVariables:       []string{"a"},
			expectedToBytes:         []byte{},
			expectedString:          `<A a>`,
		},
		{
			description:             "Fill-in string length limit: 0, 0",
			input:                   []interface{}{"var", 0, 0},
			expectedSize:            -1,
			expectedFillInStrLenMin: 0,
			expectedFillInStrLenMax: 0,
			expectedVariables:       []string{"var"},
			expectedToBytes:         []byte{},
			expectedString:          `<A[0] var>`,
		},
		{
			description:             "Fill-in string length limit: 0, 1",
			input:                   []interface{}{"__var", 0, 1},
			expectedSize:            -1,
			expectedFillInStrLenMin: 0,
			expectedFillInStrLenMax: 1,
			expectedVariables:       []string{"__var"},
			expectedToBytes:         []byte{},
			expectedString:          `<A[0..1] __var>`,
		},
		{
			description:             "Fill-in string length limit: 1, 1",
			input:                   []interface{}{"var1", 1, 1},
			expectedSize:            -1,
			expectedFillInStrLenMin: 1,
			expectedFillInStrLenMax: 1,
			expectedVariables:       []string{"var1"},
			expectedToBytes:         []byte{},
			expectedString:          `<A[1] var1>`,
		},
		{
			description:             "Fill-in string length limit: 2, -1",
			input:                   []interface{}{"var", 2, -1},
			expectedSize:            -1,
			expectedFillInStrLenMin: 2,
			expectedFillInStrLenMax: -1,
			expectedVariables:       []string{"var"},
			expectedToBytes:         []byte{},
			expectedString:          `<A[2..] var>`,
		},
		{
			description:             "Fill-in string length limit: 2, 10",
			input:                   []interface{}{"var", 2, 10},
			expectedSize:            -1,
			expectedFillInStrLenMin: 2,
			expectedFillInStrLenMax: 10,
			expectedVariables:       []string{"var"},
			expectedToBytes:         []byte{},
			expectedString:          `<A[2..10] var>`,
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewASCIINodeVariable(test.input[0].(string), test.input[1].(int), test.input[2].(int))
		min, max := node.(*ASCIINode).FillInStringLength()
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedFillInStrLenMin, min)
		assert.Equal(t, test.expectedFillInStrLenMax, max)
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestASCIINode_Variable_ProducedByFillVariables(t *testing.T) {
	var tests = []struct {
		description             string                 // Test case description
		input                   []interface{}          // Inputs to the factory method (name, minLength, maxLength)
		inputFillInValues       map[string]interface{} // Input to the FillVariables()
		expectedSize            int                    // expected result from Size()
		expectedFillInStrLenMin int                    // expected result of min from FillInStringLength()
		expectedFillInStrLenMax int                    // expected result of max from FillInStringLength()
		expectedVariables       []string               // expected result from Variables()
		expectedToBytes         []byte                 // expected result from ToBytes()
		expectedString          string                 // expected result from String()
	}{
		{
			description:             "No fill-in value in input",
			input:                   []interface{}{"a", 0, -1},
			inputFillInValues:       map[string]interface{}{},
			expectedSize:            -1,
			expectedFillInStrLenMin: 0,
			expectedFillInStrLenMax: -1,
			expectedVariables:       []string{"a"},
			expectedToBytes:         []byte{},
			expectedString:          `<A a>`,
		},
		{
			description:             "Fill empty string literal",
			input:                   []interface{}{"var", 0, 0},
			inputFillInValues:       map[string]interface{}{"var": ""},
			expectedSize:            0,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 0},
			expectedString:          `<A[0]>`,
		},
		{
			description:             "Fill-in string length: 1, length in range",
			input:                   []interface{}{"var", 0, 1},
			inputFillInValues:       map[string]interface{}{"var": "A"},
			expectedSize:            1,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 1, 65},
			expectedString:          `<A "A">`,
		},
		{
			description:             "Fill-in string length: 1, length exact match",
			input:                   []interface{}{"var", 1, 1},
			inputFillInValues:       map[string]interface{}{"var": "B"},
			expectedSize:            1,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 1, 66},
			expectedString:          `<A "B">`,
		},
		{
			description:             "Fill-in string length: 11, length no upper bound",
			input:                   []interface{}{"var", 2, -1},
			inputFillInValues:       map[string]interface{}{"var": "lorem ipsum"},
			expectedSize:            11,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 11, 0x6C, 0x6F, 0x72, 0x65, 0x6D, 0x20, 0x69, 0x70, 0x73, 0x75, 0x6D},
			expectedString:          `<A "lorem ipsum">`,
		},
		{
			description:             "Fill-in string length: 5, length equal to lower bound",
			input:                   []interface{}{"var", 5, 10},
			inputFillInValues:       map[string]interface{}{"var": "hello", "foo": "bar"},
			expectedSize:            5,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 5, 0x68, 0x65, 0x6C, 0x6C, 0x6F},
			expectedString:          `<A "hello">`,
		},
		{
			description:             "Fill-in string length: 2, Non-printable chars only",
			input:                   []interface{}{"var", 0, -1},
			inputFillInValues:       map[string]interface{}{"var": "\r\n", "foo": "bar"},
			expectedSize:            2,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 2, 0x0D, 0x0A},
			expectedString:          `<A 0x0D 0x0A>`,
		},
		{
			description:             "Fill-in string length: 6, Non-printable chars at text head",
			input:                   []interface{}{"var", 0, -1},
			inputFillInValues:       map[string]interface{}{"var": "\r\ntext", "foo": "bar"},
			expectedSize:            6,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 6, 0x0D, 0x0A, 0x74, 0x65, 0x78, 0x74},
			expectedString:          `<A 0x0D 0x0A "text">`,
		},
		{
			description:             "Fill-in string length: 6, Non-printable chars at text tail",
			input:                   []interface{}{"var", 0, -1},
			inputFillInValues:       map[string]interface{}{"var": "text\n\x00", "foo": "bar"},
			expectedSize:            6,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 6, 0x74, 0x65, 0x78, 0x74, 0x0A, 0x00},
			expectedString:          `<A "text" 0x0A 0x00>`,
		},
		{
			description:             "Fill-in string length: 6, Non-printable chars in between texts",
			input:                   []interface{}{"var", 0, -1},
			inputFillInValues:       map[string]interface{}{"var": "te\x09\x7Fxt", "foo": "bar"},
			expectedSize:            6,
			expectedFillInStrLenMin: -2,
			expectedFillInStrLenMax: -2,
			expectedVariables:       []string{},
			expectedToBytes:         []byte{0x41, 6, 0x74, 0x65, 0x09, 0x7F, 0x78, 0x74},
			expectedString:          `<A "te" 0x09 0x7F "xt">`,
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewASCIINodeVariable(test.input[0].(string), test.input[1].(int), test.input[2].(int))
		node = node.FillVariables(test.inputFillInValues)
		min, max := node.(*ASCIINode).FillInStringLength()

		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedFillInStrLenMin, min)
		assert.Equal(t, test.expectedFillInStrLenMax, max)
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}
