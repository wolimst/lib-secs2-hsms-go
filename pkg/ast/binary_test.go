package ast

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Testing Strategy:
//
// Refer to interface_test.go

func TestBinaryNode_ProducedByFactoryMethod(t *testing.T) {
	var tests = []struct {
		description       string        // Test case description
		input             []interface{} // Input to the factory method
		expectedSize      int           // expected result from Size()
		expectedVariables []string      // expected result from Variables()
		expectedToBytes   []byte        // expected result from ToBytes()
		expectedString    string        // expected result from String()
	}{
		{
			description:       "Size: 0, Variable: 0",
			input:             []interface{}{},
			expectedSize:      0,
			expectedVariables: []string{},
			expectedToBytes:   []byte{33, 0},
			expectedString:    "<B[0]>",
		},
		{
			description:       "Size: 1, Integer input, Variable: 0",
			input:             []interface{}{0},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{33, 1, 0},
			expectedString:    "<B[1] 0b0>",
		},
		{
			description:       "Size: 3, Integer input, Variable: 0",
			input:             []interface{}{1, 2, 255},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{33, 3, 1, 2, 255},
			expectedString:    "<B[3] 0b1 0b10 0b11111111>",
		},
		{
			description:       "Size: 4, Binary string input, Variable: 1",
			input:             []interface{}{"_var1", "0b00", "0b01", "0b11111111"},
			expectedSize:      4,
			expectedVariables: []string{"_var1"},
			expectedToBytes:   []byte{},
			expectedString:    "<B[4] _var1 0b0 0b1 0b11111111>",
		},
		{
			description:       "Size: 7, Integer and binary string input, Variable: 3",
			input:             []interface{}{"0b1", "foo", 2, "0b1111", "bar", 42, "__var"},
			expectedSize:      7,
			expectedVariables: []string{"foo", "bar", "__var"},
			expectedToBytes:   []byte{},
			expectedString:    "<B[7] 0b1 foo 0b10 0b1111 bar 0b101010 __var>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewBinaryNode(test.input...)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestBinaryNode_ProducedByFillVariables(t *testing.T) {
	var tests = []struct {
		description       string                 // Test case description
		input             []interface{}          // Input to the factory method
		fillInValues      map[string]interface{} // Input to the FillVariables()
		expectedSize      int                    // expected result from Size()
		expectedVariables []string               // expected result from Variables()
		expectedToBytes   []byte                 // expected result from ToBytes()
		expectedString    string                 // expected result from String()
	}{
		{
			description:       "Size: 0, Variable: 0",
			input:             []interface{}{},
			fillInValues:      map[string]interface{}{},
			expectedSize:      0,
			expectedVariables: []string{},
			expectedToBytes:   []byte{33, 0},
			expectedString:    "<B[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			fillInValues:      map[string]interface{}{},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{33, 1, 0},
			expectedString:    "<B[1] 0b0>",
		},
		{
			description:       "Size: 1, Variable: 1, All variables filled in",
			input:             []interface{}{"var1"},
			fillInValues:      map[string]interface{}{"var1": 1},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{33, 1, 1},
			expectedString:    "<B[1] 0b1>",
		},
		{
			description:       "Size: 3, Variable: 2, All variables filled in",
			input:             []interface{}{"var1", "0b11", "var2"},
			fillInValues:      map[string]interface{}{"var1": "0b001", "var2": 5},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{33, 3, 1, 3, 5},
			expectedString:    "<B[3] 0b1 0b11 0b101>",
		},
		{
			description:       "Size: 3, Variable: 2, 0 variable filled in",
			input:             []interface{}{"var1", "0b11", "var2"},
			fillInValues:      map[string]interface{}{"var3": 0},
			expectedSize:      3,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<B[3] var1 0b11 var2>",
		},
		{
			description:       "Size: 3, Variable: 2, 1 variable filled in",
			input:             []interface{}{"var1", "0b11", "var2"},
			fillInValues:      map[string]interface{}{"var2": "0b101", "var3": 0},
			expectedSize:      3,
			expectedVariables: []string{"var1"},
			expectedToBytes:   []byte{},
			expectedString:    "<B[3] var1 0b11 0b101>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewBinaryNode(test.input...).FillVariables(test.fillInValues)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}
