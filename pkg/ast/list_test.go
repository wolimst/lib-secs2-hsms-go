package ast

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Testing Strategy:
//
// Refer to interface.go
// Variables(), FillValues(), ToBytes(), String() should run recursively.
//
// Partitions:
//
// - Size of ListNode: 0, 1, ...
// - Data value type in ListNode: ListNode, other ItemNode, variable, variable with array-like notation,
//                                ellipsis, nested ellipsis with array-like notation
// - The number of nested ListNode: 0, 1, ...

func TestListNode_ProducedByFactoryMethod(t *testing.T) {
	var tests = []struct {
		description       string        // Test case description
		input             []interface{} // Input to the factory method
		expectedSize      int           // expected result from Size()
		expectedVariables []string      // expected result from Variables()
		expectedToBytes   []byte        // expected result from ToBytes()
		expectedString    string        // expected result from String()
	}{
		{
			description:       "Size: 0",
			input:             []interface{}{},
			expectedSize:      0,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x01, 0},
			expectedString:    `<L[0]>`,
		},
		{
			description:       "Size: 1, Contains ordinary ItemNode",
			input:             []interface{}{NewASCIINode("text")},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x01, 1, 0x41, 4, 0x74, 0x65, 0x78, 0x74},
			expectedString: `<L[1]
  <A "text">
>`,
		},
		{
			description:       "Size: 2, Contains ordinary ItemNodes",
			input:             []interface{}{NewASCIINode("text"), NewIntNode(1, 11, 22)},
			expectedSize:      2,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x01, 2, 0x41, 4, 0x74, 0x65, 0x78, 0x74, 0x65, 2, 11, 22},
			expectedString: `<L[2]
  <A "text">
  <I1[2] 11 22>
>`,
		},
		{
			description:       "Size: 3, Contains ordinary ItemNodes with variable",
			input:             []interface{}{NewASCIINodeVariable("varC", 0, -1), NewIntNode(1, "varB", "varA"), NewBinaryNode()},
			expectedSize:      3,
			expectedVariables: []string{"varC", "varB", "varA"},
			expectedToBytes:   []byte{},
			expectedString: `<L[3]
  <A varC>
  <I1[2] varB varA>
  <B[0]>
>`,
		},
		{
			description:       "Size: 4, Contains ordinary ItemNodes, a variable and a ellipsis",
			input:             []interface{}{NewASCIINodeVariable("varC", 0, -1), NewIntNode(1, "varB", "varA"), "varNode", "..."},
			expectedSize:      4,
			expectedVariables: []string{"varC", "varB", "varA", "varNode", "..."},
			expectedToBytes:   []byte{},
			expectedString: `<L
  <A varC>
  <I1[2] varB varA>
  varNode
  ...
>`,
		},
		{
			description: "Size: 2, Nested list level: 1, No variable",
			input: []interface{}{
				NewListNode(),
				NewListNode(NewIntNode(1, 33, 55)),
			},
			expectedSize:      2,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x01, 2, 0x01, 0, 0x01, 1, 0x65, 2, 33, 55},
			expectedString: `<L[2]
  <L[0]>
  <L[1]
    <I1[2] 33 55>
  >
>`,
		},
		{
			description: "Size: 2, Nested list level: 2, No variable",
			input: []interface{}{
				NewListNode(
					NewIntNode(1, 33, 55),
					NewListNode(NewASCIINode("text")),
				),
				NewIntNode(2, 77, 99),
			},
			expectedSize:      2,
			expectedVariables: []string{},
			expectedToBytes: []byte{
				0x01, 2,
				0x01, 2, 0x65, 2, 33, 55, 0x01, 1, 0x41, 4, 0x74, 0x65, 0x78, 0x74,
				0x69, 4, 0, 77, 0, 99,
			},
			expectedString: `<L[2]
  <L[2]
    <I1[2] 33 55>
    <L[1]
      <A "text">
    >
  >
  <I2[2] 77 99>
>`,
		},
		{
			description: "Size: 3, Nested list level: 2, Contains variables with array-like notation",
			input: []interface{}{
				NewListNode(
					NewIntNode(1, "foo[0]"),
					NewListNode("varNode[0][0]", "varNode[0][1]"),
					NewIntNode(1, "foo[1]"),
					NewListNode("varNode[1][0]", "varNode[1][1]"),
				),
				"...",
				"varNode2",
			},
			expectedSize: 3,
			expectedVariables: []string{
				"foo[0]", "varNode[0][0]", "varNode[0][1]",
				"foo[1]", "varNode[1][0]", "varNode[1][1]", "...", "varNode2",
			},
			expectedToBytes: []byte{},
			expectedString: `<L
  <L[4]
    <I1[1] foo[0]>
    <L
      varNode[0][0]
      varNode[0][1]
    >
    <I1[1] foo[1]>
    <L
      varNode[1][0]
      varNode[1][1]
    >
  >
  ...
  varNode2
>`,
		},
		{
			description: "Size: 3, Nested list level: 2, Contains nested ellipsis",
			input: []interface{}{
				NewIntNode(1, "var"),
				NewListNode(
					NewIntNode(1, "foo"),
					NewListNode(
						NewIntNode(1, "bar"),
						"varNode",
						"...[0]",
					),
					"...[1]",
				),
				"...[2]",
			},
			expectedSize:      3,
			expectedVariables: []string{"var", "foo", "bar", "varNode", "...[0]", "...[1]", "...[2]"},
			expectedToBytes:   []byte{},
			expectedString: `<L
  <I1[1] var>
  <L
    <I1[1] foo>
    <L
      <I1[1] bar>
      varNode
      ...
    >
    ...
  >
  ...
>`,
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewListNode(test.input...)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}

func TestListNode_ProducedByFillValues(t *testing.T) {
	var tests = []struct {
		description       string                 // Test case description
		input             []interface{}          // Input to the factory method
		inputFillInValues map[string]interface{} // Input to FillValues()
		expectedSize      int                    // expected result from Size()
		expectedVariables []string               // expected result from Variables()
		expectedToBytes   []byte                 // expected result from ToBytes()
		expectedString    string                 // expected result from String()
	}{
		{
			description:       "Size: 0",
			input:             []interface{}{},
			inputFillInValues: map[string]interface{}{},
			expectedSize:      0,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x01, 0},
			expectedString:    `<L[0]>`,
		},
		{
			description:       "Size: 1, No variable",
			input:             []interface{}{NewASCIINode("text")},
			inputFillInValues: map[string]interface{}{},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x01, 1, 0x41, 4, 0x74, 0x65, 0x78, 0x74},
			expectedString: `<L[1]
  <A "text">
>`,
		},
		{
			description:       "Size: 1, Variable: 1, All variables filled in",
			input:             []interface{}{NewASCIINodeVariable("var", 0, -1)},
			inputFillInValues: map[string]interface{}{"var": "text"},
			expectedSize:      1,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x01, 1, 0x41, 4, 0x74, 0x65, 0x78, 0x74},
			expectedString: `<L[1]
  <A "text">
>`,
		},
		{
			description:       "Size: 3, Variable: 3, All variables filled in",
			input:             []interface{}{NewASCIINodeVariable("varC", 0, -1), NewIntNode(1, "varB", "varA"), NewBinaryNode()},
			inputFillInValues: map[string]interface{}{"varC": "text", "varB": 0, "varA": 1, "foo": "bar"},
			expectedSize:      3,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x01, 3, 0x41, 4, 0x74, 0x65, 0x78, 0x74, 0x65, 2, 0, 1, 0x21, 0},
			expectedString: `<L[3]
  <A "text">
  <I1[2] 0 1>
  <B[0]>
>`,
		},
		{
			description:       "Size: 3, Variable: 3, 0 variables filled in",
			input:             []interface{}{NewASCIINodeVariable("varC", 0, -1), NewIntNode(1, "varB", "varA"), NewBinaryNode()},
			inputFillInValues: map[string]interface{}{"foo": "bar"},
			expectedSize:      3,
			expectedVariables: []string{"varC", "varB", "varA"},
			expectedToBytes:   []byte{},
			expectedString: `<L[3]
  <A varC>
  <I1[2] varB varA>
  <B[0]>
>`,
		},
		{
			description:       "Size: 3, Variable: 3, 2 variables filled in",
			input:             []interface{}{NewASCIINodeVariable("varC", 0, -1), NewIntNode(1, "varB", "varA"), NewBinaryNode()},
			inputFillInValues: map[string]interface{}{"varC": "text", "varB": 0},
			expectedSize:      3,
			expectedVariables: []string{"varA"},
			expectedToBytes:   []byte{},
			expectedString: `<L[3]
  <A "text">
  <I1[2] 0 varA>
  <B[0]>
>`,
		},
		{
			description:       "Fill except ellipsis",
			input:             []interface{}{NewASCIINodeVariable("var", 0, -1), "varNode", "..."},
			inputFillInValues: map[string]interface{}{"var": "text", "varNode": NewIntNode(1, 0, 1)},
			expectedSize:      3,
			expectedVariables: []string{"..."},
			expectedToBytes:   []byte{},
			expectedString: `<L
  <A "text">
  <I1[2] 0 1>
  ...
>`,
		},
		{
			description:       "Fill except ellipsis, ItemNode variable contains a variable",
			input:             []interface{}{NewASCIINodeVariable("var", 0, -1), "varNode", "..."},
			inputFillInValues: map[string]interface{}{"var": "text", "varNode": NewIntNode(1, 0, "var2"), "...[10]": 0},
			expectedSize:      3,
			expectedVariables: []string{"var2", "..."},
			expectedToBytes:   []byte{},
			expectedString: `<L
  <A "text">
  <I1[2] 0 var2>
  ...
>`,
		},
		{
			description:       "Fill ellipsis = 0",
			input:             []interface{}{NewASCIINodeVariable("var", 0, -1), "varNode", "..."},
			inputFillInValues: map[string]interface{}{"...": 0},
			expectedSize:      2,
			expectedVariables: []string{"var", "varNode"},
			expectedToBytes:   []byte{},
			expectedString: `<L
  <A var>
  varNode
>`,
		},
		{
			description:       "Fill ellipsis = 1",
			input:             []interface{}{NewASCIINodeVariable("var", 0, -1), "varNode", "..."},
			inputFillInValues: map[string]interface{}{"...": 1},
			expectedSize:      4,
			expectedVariables: []string{"var[0]", "varNode[0]", "var[1]", "varNode[1]"},
			expectedToBytes:   []byte{},
			expectedString: `<L
  <A var[0]>
  varNode[0]
  <A var[1]>
  varNode[1]
>`,
		},
		{
			description:       "Fill ellipsis = 2, some variables with array-like notation",
			input:             []interface{}{NewASCIINodeVariable("var", 0, -1), "varNode", "..."},
			inputFillInValues: map[string]interface{}{"...": 2, "var[2]": "text", "varNode[2]": NewIntNode(1, 0, 1)},
			expectedSize:      6,
			expectedVariables: []string{"var[0]", "varNode[0]", "var[1]", "varNode[1]"},
			expectedToBytes:   []byte{},
			expectedString: `<L
  <A var[0]>
  varNode[0]
  <A var[1]>
  varNode[1]
  <A "text">
  <I1[2] 0 1>
>`,
		},
		{
			description: "Nest level: 1, No variable",
			input: []interface{}{
				NewListNode(),
				NewListNode(NewIntNode(1, 33, 55)),
			},
			inputFillInValues: map[string]interface{}{},
			expectedSize:      2,
			expectedVariables: []string{},
			expectedToBytes:   []byte{0x01, 2, 0x01, 0, 0x01, 1, 0x65, 2, 33, 55},
			expectedString: `<L[2]
  <L[0]>
  <L[1]
    <I1[2] 33 55>
  >
>`,
		},
		{
			description: "Nest level: 2, Fill some variables with array-like notation",
			input: []interface{}{
				NewListNode(
					NewIntNode(1, "foo[0]"),
					NewListNode("varNode[0][0]", "varNode[0][1]"),
					NewIntNode(1, "foo[1]"),
					NewListNode("varNode[1][0]", "varNode[1][1]"),
				),
			},
			inputFillInValues: map[string]interface{}{
				"foo[0]":        0,
				"foo[1]":        1,
				"varNode[0][0]": NewIntNode(1, 0, 1),
			},
			expectedSize:      1,
			expectedVariables: []string{"varNode[0][1]", "varNode[1][0]", "varNode[1][1]"},
			expectedToBytes:   []byte{},
			expectedString: `<L[1]
  <L[4]
    <I1[1] 0>
    <L
      <I1[2] 0 1>
      varNode[0][1]
    >
    <I1[1] 1>
    <L
      varNode[1][0]
      varNode[1][1]
    >
  >
>`,
		},
		{
			description: "Nest level: 2, Items exist after ellipsis",
			input: []interface{}{
				NewListNode(
					NewIntNode(1, "foo"),
					"...[0]",
					NewASCIINode(""),
					NewBinaryNode(),
				),
				NewListNode(
					"varNode1",
					"...[1]",
					"varNode2",
				),
			},
			inputFillInValues: map[string]interface{}{"...[0]": 2, "...[1]": 1},
			expectedSize:      2,
			expectedVariables: []string{"foo[0]", "foo[1]", "foo[2]", "varNode1[0]", "varNode1[1]", "varNode2"},
			expectedToBytes:   []byte{},
			expectedString: `<L[2]
  <L[5]
    <I1[1] foo[0]>
    <I1[1] foo[1]>
    <I1[1] foo[2]>
    <A[0]>
    <B[0]>
  >
  <L
    varNode1[0]
    varNode1[1]
    varNode2
  >
>`,
		},
		{
			description: "Nest level: 1, Fill only outer ellipsis",
			input: []interface{}{
				NewListNode(
					NewIntNode(1, 0),
					"...[0]",
				),
				"...[1]",
			},
			inputFillInValues: map[string]interface{}{"...[1]": 2},
			expectedSize:      3,
			expectedVariables: []string{"...[0]", "...[1]", "...[2]"},
			expectedToBytes:   []byte{},
			expectedString: `<L[3]
  <L
    <I1[1] 0>
    ...
  >
  <L
    <I1[1] 0>
    ...
  >
  <L
    <I1[1] 0>
    ...
  >
>`,
		},
		{
			description: "Nest level: 1, Fill only inner ellipsis",
			input: []interface{}{
				NewListNode(
					NewIntNode(1, 0),
					"...[0]",
				),
				"...[1]",
			},
			inputFillInValues: map[string]interface{}{"...[0]": 4},
			expectedSize:      2,
			expectedVariables: []string{"..."},
			expectedToBytes:   []byte{},
			expectedString: `<L
  <L[5]
    <I1[1] 0>
    <I1[1] 0>
    <I1[1] 0>
    <I1[1] 0>
    <I1[1] 0>
  >
  ...
>`,
		},
		{
			description: "Nest level: 2, Contains nested ellipsis, Fill all variables",
			input: []interface{}{
				NewListNode(
					NewIntNode(1, "foo"),
					NewListNode(
						NewIntNode(1, "bar"),
						"varNode",
						"...[0]",
					),
					"...[1]",
				),
				"...[2]",
				NewIntNode(1, "var"),
			},
			inputFillInValues: map[string]interface{}{
				"...[0]": 2,
				"...[1]": 1,
				"...[2]": 0,
				// fillEllipsis() will create following item node
				// <L[1]
				//   <L[4]
				//     <I1[1] foo[0]>
				//     <L[6]
				//       <I1[1] bar[0][0]>
				//       varNode[0][0]
				//       <I1[1] bar[0][1]>    | repeated
				//       varNode[0][1]        | 2 times
				//       <I1[1] bar[0][2]>    |
				//       varNode[0][2]        |
				//     >
				//     <I1[1] foo[1]>                       | repeated
				//     <L[6]                                | 1 time
				//       <I1[1] bar[1][0]>                  |
				//       varNode[1][0]                      |
				//       <I1[1] bar[1][1]>    | repeated    |
				//       varNode[1][1]        | 2 times     |
				//       <I1[1] bar[1][2]>    |             |
				//       varNode[1][2]        |             |
				//     >                                    |
				//   >
				// >
				"foo[0]":        0,
				"bar[0][0]":     1,
				"bar[0][1]":     2,
				"bar[0][2]":     3,
				"foo[1]":        4,
				"bar[1][0]":     5,
				"bar[1][1]":     6,
				"bar[1][2]":     7,
				"varNode[0][0]": NewListNode(),
				"varNode[0][1]": NewBooleanNode(),
				"varNode[0][2]": NewBooleanNode(true),
				"varNode[1][0]": NewListNode(NewBooleanNode(false)),
				"varNode[1][1]": NewBooleanNode(true),
				"varNode[1][2]": NewBooleanNode(true, false),
				"var":           8,
			},
			expectedSize:      2,
			expectedVariables: []string{},
			expectedToBytes: []byte{
				0x01, 2,
				0x01, 4, // <L[4]
				0x65, 1, 0,
				0x01, 6, // <L[6]
				0x65, 1, 1,
				0x01, 0, // <L[0]
				0x65, 1, 2,
				0x25, 0,
				0x65, 1, 3,
				0x25, 1, 1,
				0x65, 1, 4,
				0x01, 6, // <L[6]
				0x65, 1, 5,
				0x01, 1, 0x25, 1, 0, // <L[1]
				0x65, 1, 6,
				0x25, 1, 1,
				0x65, 1, 7,
				0x25, 2, 1, 0,
				0x65, 1, 8,
			},
			expectedString: `<L[2]
  <L[4]
    <I1[1] 0>
    <L[6]
      <I1[1] 1>
      <L[0]>
      <I1[1] 2>
      <BOOLEAN[0]>
      <I1[1] 3>
      <BOOLEAN[1] T>
    >
    <I1[1] 4>
    <L[6]
      <I1[1] 5>
      <L[1]
        <BOOLEAN[1] F>
      >
      <I1[1] 6>
      <BOOLEAN[1] T>
      <I1[1] 7>
      <BOOLEAN[2] T F>
    >
  >
  <I1[1] 8>
>`,
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		node := NewListNode(test.input...).FillValues(test.inputFillInValues)
		assert.Equal(t, test.expectedSize, node.Size())
		assert.Equal(t, test.expectedVariables, node.Variables())
		assert.Equal(t, test.expectedToBytes, node.ToBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(node))
	}
}
