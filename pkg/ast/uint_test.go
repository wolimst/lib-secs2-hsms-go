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

// U1 type

func TestU1Node_ProducedByFactoryMethod(t *testing.T) {
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
			expectedToBytes:   []byte{0xA5, 0},
			expectedString:    "<U1[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xA5, 1, 0},
			expectedString:    "<U1[1] 0>",
		},
		{
			description:       "Size: 3, Variable: 0",
			input:             []interface{}{0, 1, 2},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xA5, 3, 0, 1, 2},
			expectedString:    "<U1[3] 0 1 2>",
		},
		{
			description:       "Size: 5, Variable: 2",
			input:             []interface{}{"var1", 128, 254, 255, "var2"},
			expectedSize:      5,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<U1[5] var1 128 254 255 var2>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewUintNode(1, test.input...)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestU1Node_ProducedByFillVariables(t *testing.T) {
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
			expectedToBytes:   []byte{0xA5, 0},
			expectedString:    "<U1[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			fillInValues:      map[string]interface{}{},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xA5, 1, 0},
			expectedString:    "<U1[1] 0>",
		},
		{
			description:       "Size: 1, Variable: 1, All variables filled in",
			input:             []interface{}{"var1"},
			fillInValues:      map[string]interface{}{"var1": 1},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xA5, 1, 1},
			expectedString:    "<U1[1] 1>",
		},
		{
			description:       "Size: 3, Variable: 2, All variables filled in",
			input:             []interface{}{"var1", 254, "var2"},
			fillInValues:      map[string]interface{}{"var1": 2, "var2": 255},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xA5, 3, 2, 254, 255},
			expectedString:    "<U1[3] 2 254 255>",
		},
		{
			description:       "Size: 3, Variable: 2, 0 variable filled in",
			input:             []interface{}{"var1", 254, "var2"},
			fillInValues:      map[string]interface{}{"var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<U1[3] var1 254 var2>",
		},
		{
			description:       "Size: 3, Variable: 2, 1 variable filled in",
			input:             []interface{}{"var1", 254, "var2"},
			fillInValues:      map[string]interface{}{"var2": 255, "var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1"},
			expectedToBytes:   []byte{},
			expectedString:    "<U1[3] var1 254 255>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewUintNode(1, test.input...).FillVariables(test.fillInValues)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

// U2 type

func TestU2Node_ProducedByFactoryMethod(t *testing.T) {
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
			expectedToBytes:   []byte{0xA9, 0},
			expectedString:    "<U2[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xA9, 2, 0, 0},
			expectedString:    "<U2[1] 0>",
		},
		{
			description:       "Size: 3, Variable: 0",
			input:             []interface{}{0, 1, 2},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xA9, 6, 0, 0, 0, 1, 0, 2},
			expectedString:    "<U2[3] 0 1 2>",
		},
		{
			description:       "Size: 5, Variable: 2",
			input:             []interface{}{"var1", 1024, 65534, 65535, "var2"},
			expectedSize:      5,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<U2[5] var1 1024 65534 65535 var2>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewUintNode(2, test.input...)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestU2Node_ProducedByFillVariables(t *testing.T) {
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
			expectedToBytes:   []byte{0xA9, 0},
			expectedString:    "<U2[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			fillInValues:      map[string]interface{}{},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xA9, 2, 0, 0},
			expectedString:    "<U2[1] 0>",
		},
		{
			description:       "Size: 1, Variable: 1, All variables filled in",
			input:             []interface{}{"var1"},
			fillInValues:      map[string]interface{}{"var1": 1},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xA9, 2, 0, 1},
			expectedString:    "<U2[1] 1>",
		},
		{
			description:       "Size: 3, Variable: 2, All variables filled in",
			input:             []interface{}{"var1", 254, "var2"},
			fillInValues:      map[string]interface{}{"var1": 2, "var2": 65535},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xA9, 6, 0, 2, 0, 254, 0xFF, 0xFF},
			expectedString:    "<U2[3] 2 254 65535>",
		},
		{
			description:       "Size: 3, Variable: 2, 0 variable filled in",
			input:             []interface{}{"var1", 254, "var2"},
			fillInValues:      map[string]interface{}{"var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<U2[3] var1 254 var2>",
		},
		{
			description:       "Size: 3, Variable: 2, 1 variable filled in",
			input:             []interface{}{"var1", 254, "var2"},
			fillInValues:      map[string]interface{}{"var2": 65534, "var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1"},
			expectedToBytes:   []byte{},
			expectedString:    "<U2[3] var1 254 65534>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewUintNode(2, test.input...).FillVariables(test.fillInValues)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

// U4 type

func TestU4Node_ProducedByFactoryMethod(t *testing.T) {
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
			expectedToBytes:   []byte{0xB1, 0},
			expectedString:    "<U4[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xB1, 4, 0, 0, 0, 0},
			expectedString:    "<U4[1] 0>",
		},
		{
			description:       "Size: 3, Variable: 0",
			input:             []interface{}{0, 1, 2},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xB1, 12, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 2},
			expectedString:    "<U4[3] 0 1 2>",
		},
		{
			description:       "Size: 5, Variable: 2",
			input:             []interface{}{"var1", 65536, 1<<32 - 2, 1<<32 - 1, "var2"},
			expectedSize:      5,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<U4[5] var1 65536 4294967294 4294967295 var2>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewUintNode(4, test.input...)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestU4Node_ProducedByFillVariables(t *testing.T) {
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
			expectedToBytes:   []byte{0xB1, 0},
			expectedString:    "<U4[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			fillInValues:      map[string]interface{}{},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xB1, 4, 0, 0, 0, 0},
			expectedString:    "<U4[1] 0>",
		},
		{
			description:       "Size: 1, Variable: 1, All variables filled in",
			input:             []interface{}{"var1"},
			fillInValues:      map[string]interface{}{"var1": 1},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xB1, 4, 0, 0, 0, 1},
			expectedString:    "<U4[1] 1>",
		},
		{
			description:       "Size: 3, Variable: 2, All variables filled in",
			input:             []interface{}{"var1", 4294967294, "var2"},
			fillInValues:      map[string]interface{}{"var1": 65536, "var2": 0xFFFFFFFF},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xB1, 12, 0, 1, 0, 0, 0xFF, 0xFF, 0xFF, 0xFE, 0xFF, 0xFF, 0xFF, 0xFF},
			expectedString:    "<U4[3] 65536 4294967294 4294967295>",
		},
		{
			description:       "Size: 3, Variable: 2, 0 variable filled in",
			input:             []interface{}{"var1", 254, "var2"},
			fillInValues:      map[string]interface{}{"var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<U4[3] var1 254 var2>",
		},
		{
			description:       "Size: 3, Variable: 2, 1 variable filled in",
			input:             []interface{}{"var1", 55555, "var2"},
			fillInValues:      map[string]interface{}{"var2": 7777777, "var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1"},
			expectedToBytes:   []byte{},
			expectedString:    "<U4[3] var1 55555 7777777>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewUintNode(4, test.input...).FillVariables(test.fillInValues)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

// U8 type

func TestU8Node_ProducedByFactoryMethod(t *testing.T) {
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
			expectedToBytes:   []byte{0xA1, 0},
			expectedString:    "<U8[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xA1, 8, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedString:    "<U8[1] 0>",
		},
		{
			description:       "Size: 3, Variable: 0",
			input:             []interface{}{0, 1, 2},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes: []byte{
				0xA1, 24,
				0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 1,
				0, 0, 0, 0, 0, 0, 0, 2,
			},
			expectedString: "<U8[3] 0 1 2>",
		},
		{
			description:       "Size: 5, Variable: 2",
			input:             []interface{}{"var1", math.MaxUint32 + 1, uint64(math.MaxUint64 - 1), uint64(math.MaxUint64), "var2"},
			expectedSize:      5,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<U8[5] var1 4294967296 18446744073709551614 18446744073709551615 var2>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewUintNode(8, test.input...)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestU8Node_ProducedByFillVariables(t *testing.T) {
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
			expectedToBytes:   []byte{0xA1, 0},
			expectedString:    "<U8[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			fillInValues:      map[string]interface{}{},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xA1, 8, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedString:    "<U8[1] 0>",
		},
		{
			description:       "Size: 1, Variable: 1, All variables filled in",
			input:             []interface{}{"var1"},
			fillInValues:      map[string]interface{}{"var1": 1},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0xA1, 8, 0, 0, 0, 0, 0, 0, 0, 1},
			expectedString:    "<U8[1] 1>",
		},
		{
			description:       "Size: 3, Variable: 2, All variables filled in",
			input:             []interface{}{"var1", uint64(math.MaxUint64 - 1), "var2"},
			fillInValues:      map[string]interface{}{"var1": math.MaxUint32 + 1, "var2": uint64(math.MaxUint64)},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes: []byte{
				0xA1, 24,
				0, 0, 0, 1, 0, 0, 0, 0,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			},
			expectedString: "<U8[3] 4294967296 18446744073709551614 18446744073709551615>",
		},
		{
			description:       "Size: 3, Variable: 2, 0 variable filled in",
			input:             []interface{}{"var1", 254, "var2"},
			fillInValues:      map[string]interface{}{"var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<U8[3] var1 254 var2>",
		},
		{
			description:       "Size: 3, Variable: 2, 1 variable filled in",
			input:             []interface{}{"var1", 55555, "var2"},
			fillInValues:      map[string]interface{}{"var2": 7777777, "var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1"},
			expectedToBytes:   []byte{},
			expectedString:    "<U8[3] var1 55555 7777777>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewUintNode(8, test.input...).FillVariables(test.fillInValues)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestU8Node_FactoryMethodInputTypes(t *testing.T) {
	node := NewUintNode(
		8,
		int(0), int8(1), int16(2), int32(4), int64(8),
		uint(16), uint8(32), uint16(64), uint32(128), uint64(256),
	)

	assert.Equal(t, 10, node.Size())
	assert.Equal(t, "<U8[10] 0 1 2 4 8 16 32 64 128 256>", fmt.Sprint(node))
}
