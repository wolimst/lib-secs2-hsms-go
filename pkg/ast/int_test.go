package ast

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Testing Strategy:
//
// Refer to interface_test.go

// I1 type

func TestI1Node_ProducedByFactoryMethod(t *testing.T) {
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
			expectedToBytes:   []byte{0x65, 0},
			expectedString:    "<I1[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x65, 1, 0},
			expectedString:    "<I1[1] 0>",
		},
		{
			description:       "Size: 3, Variable: 0",
			input:             []interface{}{-1, 0, 1},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x65, 3, 0xFF, 0, 1},
			expectedString:    "<I1[3] -1 0 1>",
		},
		{
			description:       "Size: 6, Variable: 2",
			input:             []interface{}{"var1", -128, -64, 64, 127, "var2"},
			expectedSize:      6,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<I1[6] var1 -128 -64 64 127 var2>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewIntNode(1, test.input...)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestI1Node_ProducedByFillVariables(t *testing.T) {
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
			expectedToBytes:   []byte{0x65, 0},
			expectedString:    "<I1[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			fillInValues:      map[string]interface{}{},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x65, 1, 0},
			expectedString:    "<I1[1] 0>",
		},
		{
			description:       "Size: 1, Variable: 1, All variables filled in",
			input:             []interface{}{"var1"},
			fillInValues:      map[string]interface{}{"var1": -128},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x65, 1, 0x80},
			expectedString:    "<I1[1] -128>",
		},
		{
			description:       "Size: 3, Variable: 2, All variables filled in",
			input:             []interface{}{"var1", 126, "var2"},
			fillInValues:      map[string]interface{}{"var1": -48, "var2": 127},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x65, 3, 0xD0, 126, 127},
			expectedString:    "<I1[3] -48 126 127>",
		},
		{
			description:       "Size: 3, Variable: 2, 0 variable filled in",
			input:             []interface{}{"var1", 126, "var2"},
			fillInValues:      map[string]interface{}{"var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<I1[3] var1 126 var2>",
		},
		{
			description:       "Size: 3, Variable: 2, 1 variable filled in",
			input:             []interface{}{"var1", 126, "var2"},
			fillInValues:      map[string]interface{}{"var1": -127, "var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<I1[3] -127 126 var2>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewIntNode(1, test.input...).FillVariables(test.fillInValues)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

// I2 type

func TestI2Node_ProducedByFactoryMethod(t *testing.T) {
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
			expectedToBytes:   []byte{0x69, 0},
			expectedString:    "<I2[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x69, 2, 0, 0},
			expectedString:    "<I2[1] 0>",
		},
		{
			description:       "Size: 3, Variable: 0",
			input:             []interface{}{-1, 0, 1},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x69, 6, 0xFF, 0xFF, 0, 0, 0, 1},
			expectedString:    "<I2[3] -1 0 1>",
		},
		{
			description:       "Size: 6, Variable: 2",
			input:             []interface{}{"var1", -32768, -32767, 32766, 32767, "var2"},
			expectedSize:      6,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<I2[6] var1 -32768 -32767 32766 32767 var2>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewIntNode(2, test.input...)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestI2Node_ProducedByFillVariables(t *testing.T) {
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
			expectedToBytes:   []byte{0x69, 0},
			expectedString:    "<I2[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			fillInValues:      map[string]interface{}{},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x69, 2, 0, 0},
			expectedString:    "<I2[1] 0>",
		},
		{
			description:       "Size: 1, Variable: 1, All variables filled in",
			input:             []interface{}{"var1"},
			fillInValues:      map[string]interface{}{"var1": 1},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x69, 2, 0, 1},
			expectedString:    "<I2[1] 1>",
		},
		{
			description:       "Size: 3, Variable: 2, All variables filled in",
			input:             []interface{}{"var1", 256, "var2"},
			fillInValues:      map[string]interface{}{"var1": -32768, "var2": 32767},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x69, 6, 0x80, 0x00, 0x01, 0x00, 0x7F, 0xFF},
			expectedString:    "<I2[3] -32768 256 32767>",
		},
		{
			description:       "Size: 3, Variable: 2, 0 variable filled in",
			input:             []interface{}{"var1", 256, "var2"},
			fillInValues:      map[string]interface{}{"var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<I2[3] var1 256 var2>",
		},
		{
			description:       "Size: 3, Variable: 2, 1 variable filled in",
			input:             []interface{}{"var1", 256, "var2"},
			fillInValues:      map[string]interface{}{"var1": -333, "var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<I2[3] -333 256 var2>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewIntNode(2, test.input...).FillVariables(test.fillInValues)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

// I4 type

func TestI4Node_ProducedByFactoryMethod(t *testing.T) {
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
			expectedToBytes:   []byte{0x71, 0},
			expectedString:    "<I4[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x71, 4, 0, 0, 0, 0},
			expectedString:    "<I4[1] 0>",
		},
		{
			description:       "Size: 3, Variable: 0",
			input:             []interface{}{-1, 0, 1},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x71, 12, 0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0, 0, 0, 0, 1},
			expectedString:    "<I4[3] -1 0 1>",
		},
		{
			description:       "Size: 6, Variable: 2",
			input:             []interface{}{"var1", -2147483648, -2147483647, 2147483646, 2147483647, "var2"},
			expectedSize:      6,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<I4[6] var1 -2147483648 -2147483647 2147483646 2147483647 var2>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewIntNode(4, test.input...)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestI4Node_ProducedByFillVariables(t *testing.T) {
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
			expectedToBytes:   []byte{0x71, 0},
			expectedString:    "<I4[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			fillInValues:      map[string]interface{}{},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x71, 4, 0, 0, 0, 0},
			expectedString:    "<I4[1] 0>",
		},
		{
			description:       "Size: 1, Variable: 1, All variables filled in",
			input:             []interface{}{"var1"},
			fillInValues:      map[string]interface{}{"var1": -2147483648},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x71, 4, 0x80, 0x00, 0x00, 0x00},
			expectedString:    "<I4[1] -2147483648>",
		},
		{
			description:       "Size: 3, Variable: 2, All variables filled in",
			input:             []interface{}{"var1", -2, "var2"},
			fillInValues:      map[string]interface{}{"var1": -2147483647, "var2": 2147483647},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x71, 12, 0x80, 0x00, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFE, 0x7F, 0xFF, 0xFF, 0xFF},
			expectedString:    "<I4[3] -2147483647 -2 2147483647>",
		},
		{
			description:       "Size: 3, Variable: 2, 0 variable filled in",
			input:             []interface{}{"var1", 254, "var2"},
			fillInValues:      map[string]interface{}{"var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<I4[3] var1 254 var2>",
		},
		{
			description:       "Size: 3, Variable: 2, 1 variable filled in",
			input:             []interface{}{"var1", 55555, "var2"},
			fillInValues:      map[string]interface{}{"var2": 7777777, "var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1"},
			expectedToBytes:   []byte{},
			expectedString:    "<I4[3] var1 55555 7777777>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewIntNode(4, test.input...).FillVariables(test.fillInValues)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

// I8 type

func TestI8Node_ProducedByFactoryMethod(t *testing.T) {
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
			expectedToBytes:   []byte{0x61, 0},
			expectedString:    "<I8[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x61, 8, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedString:    "<I8[1] 0>",
		},
		{
			description:       "Size: 3, Variable: 0",
			input:             []interface{}{-1, 0, 1},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes: []byte{
				0x61, 24,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 1,
			},
			expectedString: "<I8[3] -1 0 1>",
		},
		{
			description:       "Size: 6, Variable: 2",
			input:             []interface{}{"var1", math.MinInt64, math.MinInt64 + 1, math.MaxInt64 - 1, math.MaxInt64, "var2"},
			expectedSize:      6,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<I8[6] var1 -9223372036854775808 -9223372036854775807 9223372036854775806 9223372036854775807 var2>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewIntNode(8, test.input...)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestI8Node_ProducedByFillVariables(t *testing.T) {
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
			expectedToBytes:   []byte{0x61, 0},
			expectedString:    "<I8[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			fillInValues:      map[string]interface{}{},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x61, 8, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedString:    "<I8[1] 0>",
		},
		{
			description:       "Size: 1, Variable: 1, All variables filled in",
			input:             []interface{}{"var1"},
			fillInValues:      map[string]interface{}{"var1": math.MaxInt64},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x61, 8, 0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			expectedString:    "<I8[1] 9223372036854775807>",
		},
		{
			description:       "Size: 3, Variable: 2, All variables filled in",
			input:             []interface{}{"var1", -7777777, "var2"},
			fillInValues:      map[string]interface{}{"var1": math.MinInt64, "var2": math.MaxInt64 - 1},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes: []byte{
				0x61, 24,
				0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x89, 0x52, 0x0F,
				0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE,
			},
			expectedString: "<I8[3] -9223372036854775808 -7777777 9223372036854775806>",
		},
		{
			description:       "Size: 3, Variable: 2, 0 variable filled in",
			input:             []interface{}{"var1", 254, "var2"},
			fillInValues:      map[string]interface{}{"var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<I8[3] var1 254 var2>",
		},
		{
			description:       "Size: 3, Variable: 2, 1 variable filled in",
			input:             []interface{}{"var1", 55555, "var2"},
			fillInValues:      map[string]interface{}{"var2": 7777777, "var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1"},
			expectedToBytes:   []byte{},
			expectedString:    "<I8[3] var1 55555 7777777>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewIntNode(8, test.input...).FillVariables(test.fillInValues)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestI8Node_FactoryMethodInputTypes(t *testing.T) {
	node := NewIntNode(
		8,
		int(-16), int8(-8), int16(-4), int32(-2), int64(-1),
		uint(0), uint8(1), uint16(2), uint32(4), uint64(8),
	)

	assert.Equal(t, 10, node.Size())
	assert.Equal(t, "<I8[10] -16 -8 -4 -2 -1 0 1 2 4 8>", fmt.Sprint(node))
}
