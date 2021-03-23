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

// F4 type

func TestF4Node_ProducedByFactoryMethod(t *testing.T) {
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
			expectedToBytes:   []byte{0x91, 0},
			expectedString:    "<F4[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x91, 4, 0, 0, 0, 0},
			expectedString:    "<F4[1] 0>",
		},
		{
			description:       "Size: 3, Variable: 0",
			input:             []interface{}{-1.0, 0.0, 1.0},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes: []byte{
				0x91, 12,
				0xBF, 0x80, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x3F, 0x80, 0x00, 0x00,
			},
			expectedString: "<F4[3] -1 0 1>",
		},
		{
			description: "Size: 6, Variable: 2",
			input: []interface{}{
				"var1",
				-math.MaxFloat32,
				-math.SmallestNonzeroFloat32,
				math.SmallestNonzeroFloat32,
				math.MaxFloat32,
				"var2",
			},
			expectedSize:      6,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString: fmt.Sprintf(
				"<F4[6] var1 %g %g %g %g var2>",
				float32(-math.MaxFloat32),
				float32(-math.SmallestNonzeroFloat32),
				float32(math.SmallestNonzeroFloat32),
				float32(math.MaxFloat32),
			),
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewFloatNode(4, test.input...)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestF4Node_ProducedByFillVariables(t *testing.T) {
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
			expectedToBytes:   []byte{0x91, 0},
			expectedString:    "<F4[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{math.SmallestNonzeroFloat32},
			fillInValues:      map[string]interface{}{},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x91, 4, 0x00, 0x00, 0x00, 0x01},
			expectedString:    fmt.Sprintf("<F4[1] %g>", float32(math.SmallestNonzeroFloat32)),
		},
		{
			description:       "Size: 1, Variable: 1, All variables filled in",
			input:             []interface{}{"var1"},
			fillInValues:      map[string]interface{}{"var1": math.MaxFloat32},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x91, 4, 0x7F, 0x7F, 0xFF, 0xFF},
			expectedString:    fmt.Sprintf("<F4[1] %g>", float32(math.MaxFloat32)),
		},
		{
			description:       "Size: 3, Variable: 2, All variables filled in",
			input:             []interface{}{"var1", 3.141592, "var2"},
			fillInValues:      map[string]interface{}{"var1": -math.SmallestNonzeroFloat32, "var2": -math.MaxFloat32},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes: []byte{
				0x91, 12,
				0x80, 0x00, 0x00, 0x01,
				0x40, 0x49, 0x0F, 0xD8,
				0xFF, 0x7F, 0xFF, 0xFF},
			expectedString: fmt.Sprintf("<F4[3] %g 3.141592 %g>", float32(-math.SmallestNonzeroFloat32), float32(-math.MaxFloat32)),
		},
		{
			description:       "Size: 3, Variable: 2, 0 variable filled in",
			input:             []interface{}{"var1", 3.141592, "var2"},
			fillInValues:      map[string]interface{}{"var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<F4[3] var1 3.141592 var2>",
		},
		{
			description:       "Size: 3, Variable: 2, 1 variable filled in",
			input:             []interface{}{"var1", 3.141592, "var2"},
			fillInValues:      map[string]interface{}{"var2": 2.71828, "var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1"},
			expectedToBytes:   []byte{},
			expectedString:    "<F4[3] var1 3.141592 2.71828>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewFloatNode(4, test.input...).FillVariables(test.fillInValues)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

// F8 type

func TestF8Node_ProducedByFactoryMethod(t *testing.T) {
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
			expectedToBytes:   []byte{0x81, 0},
			expectedString:    "<F8[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{0},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x81, 8, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedString:    "<F8[1] 0>",
		},
		{
			description:       "Size: 3, Variable: 0",
			input:             []interface{}{-1.0, 0.0, 1.0},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes: []byte{
				0x81, 24,
				0xBF, 0xF0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x3F, 0xF0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			expectedString: "<F8[3] -1 0 1>",
		},
		{
			description: "Size: 6, Variable: 2",
			input: []interface{}{
				"var1",
				-math.MaxFloat64,
				-math.SmallestNonzeroFloat64,
				math.SmallestNonzeroFloat64,
				math.MaxFloat64,
				"var2",
			},
			expectedSize:      6,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString: fmt.Sprintf(
				"<F8[6] var1 %g %g %g %g var2>",
				-math.MaxFloat64,
				-math.SmallestNonzeroFloat64,
				math.SmallestNonzeroFloat64,
				math.MaxFloat64,
			),
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewFloatNode(8, test.input...)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestF8Node_ProducedByFillVariables(t *testing.T) {
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
			expectedToBytes:   []byte{0x81, 0},
			expectedString:    "<F8[0]>",
		},
		{
			description:       "Size: 1, Variable: 0",
			input:             []interface{}{math.SmallestNonzeroFloat64},
			fillInValues:      map[string]interface{}{},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x81, 8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			expectedString:    fmt.Sprintf("<F8[1] %g>", math.SmallestNonzeroFloat64),
		},
		{
			description:       "Size: 1, Variable: 1, All variables filled in",
			input:             []interface{}{"var1"},
			fillInValues:      map[string]interface{}{"var1": math.MaxFloat64},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x81, 8, 0x7F, 0xEF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			expectedString:    fmt.Sprintf("<F8[1] %g>", math.MaxFloat64),
		},
		{
			description:       "Size: 3, Variable: 2, All variables filled in",
			input:             []interface{}{"var1", -2.0, "var2"},
			fillInValues:      map[string]interface{}{"var1": -math.SmallestNonzeroFloat64, "var2": -math.MaxFloat64},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes: []byte{
				0x81, 24,
				0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
				0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0xFF, 0xEF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			},
			expectedString: fmt.Sprintf("<F8[3] %g -2 %g>", -math.SmallestNonzeroFloat64, -math.MaxFloat64),
		},
		{
			description:       "Size: 3, Variable: 2, 0 variable filled in",
			input:             []interface{}{"var1", 3.141592, "var2"},
			fillInValues:      map[string]interface{}{"var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1", "var2"},
			expectedToBytes:   []byte{},
			expectedString:    "<F8[3] var1 3.141592 var2>",
		},
		{
			description:       "Size: 3, Variable: 2, 1 variable filled in",
			input:             []interface{}{"var1", 3.141592, "var2"},
			fillInValues:      map[string]interface{}{"var2": 2.71828, "var3": "ASCII"},
			expectedSize:      3,
			expectedVariables: []string{"var1"},
			expectedToBytes:   []byte{},
			expectedString:    "<F8[3] var1 3.141592 2.71828>",
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewFloatNode(8, test.input...).FillVariables(test.fillInValues)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestF8Node_FactoryMethodInputTypes(t *testing.T) {
	node := NewFloatNode(
		8,
		int(0), int8(1), int16(2), int32(4), int64(8),
		uint(16), uint8(32), uint16(64), uint32(128), uint64(256),
		float32(math.SmallestNonzeroFloat32), float64(math.SmallestNonzeroFloat64),
	)

	assert.Equal(t, 12, node.Size())
	assert.Equal(
		t,
		fmt.Sprintf("<F8[12] 0 1 2 4 8 16 32 64 128 256 %g %g>",
			float64(math.SmallestNonzeroFloat32),
			float64(math.SmallestNonzeroFloat64),
		),
		fmt.Sprint(node),
	)
}
