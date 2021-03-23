package ast

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Testing Strategy:
//
// Refer to interface_test.go

func TestBooleanNode_ProducedByFactoryMethod(t *testing.T) {
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
			expectedToBytes:   []byte{37, 0},
			expectedString:    "<BOOLEAN[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{false},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{37, 1, 0},
			expectedString:    "<BOOLEAN[1] F>",
		},
		{
			description:       "Size: 3, Variable: 0",
			input:             []interface{}{false, true, true},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{37, 3, 0, 1, 1},
			expectedString:    "<BOOLEAN[3] F T T>",
		},
		{
			description:       "Size: 4, Variable: 1",
			input:             []interface{}{true, false, "var1", false},
			expectedSize:      4,
			expectedVariables: []string{"var1"},
			expectedToBytes:   []byte{},
			expectedString:    "<BOOLEAN[4] T F var1 F>",
		},
		{
			description:       "Size: 3, Variable: 3",
			input:             []interface{}{"foo", "bar", "__var"},
			expectedSize:      3,
			expectedVariables: []string{"foo", "bar", "__var"},
			expectedToBytes:   []byte{},
			expectedString:    "<BOOLEAN[3] foo bar __var>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewBooleanNode(test.input...)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestBooleanNode_ProducedByFillVariables(t *testing.T) {
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
			expectedToBytes:   []byte{37, 0},
			expectedString:    "<BOOLEAN[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{false},
			fillInValues:      map[string]interface{}{},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{37, 1, 0},
			expectedString:    "<BOOLEAN[1] F>",
		},
		{
			description:       "Size: 1, Variable: 1, All variables filled in",
			input:             []interface{}{"var1"},
			fillInValues:      map[string]interface{}{"var1": true},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{37, 1, 1},
			expectedString:    "<BOOLEAN[1] T>",
		},
		{
			description:       "Size: 3, Variable: 2, All variables filled in",
			input:             []interface{}{"var1", false, "var2"},
			fillInValues:      map[string]interface{}{"var1": false, "var2": true},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{37, 3, 0, 0, 1},
			expectedString:    "<BOOLEAN[3] F F T>",
		},
		{
			description:       "Size: 3, Variable: 2, 0 variable filled in",
			input:             []interface{}{"var1", true, "var2"},
			fillInValues:      map[string]interface{}{"var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<BOOLEAN[3] var1 T var2>",
		},
		{
			description:       "Size: 3, Variable: 2, 1 variable filled in",
			input:             []interface{}{"var1", true, "var2"},
			fillInValues:      map[string]interface{}{"var2": false, "var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1"},
			expectedToBytes:   []byte{},
			expectedString:    "<BOOLEAN[3] var1 T F>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewBooleanNode(test.input...).FillVariables(test.fillInValues)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}
