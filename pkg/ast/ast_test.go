package ast

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests the abstract syntax tree of a SECS-II message.
//
// The data item nodes that a message contains, are tested separately;
// refer to interface_test.go, and test files of each interface implementations.
//
// Testing Strategy:
//
// Create a instance of MessageNode using the factory methods or FillVariables(),
// and test the result of public observer methods.
//
// Partitions:
//
// - message name length: 0, 1, ...
// - stream code: 0, 1, ..., 126, 127
// - function code: 0, 1, ..., 254, 255
// - wait bit: 0 (false), 1 (true), 2 (optional)
// - direction: H->E, H<-E, H<->E
// - data item: empty, ASCII, list, nested list, and other nodes
// - Input to ToHSMS() observer method:
//   - deviceID: 0, 1, ..., max (=1<<16-1)
//   - systemBytes: 0, 1, ..., max (=1<<32-1) represented as []byte

func TestMessageNode_ProducedByFactoryMethod_EmptyItem(t *testing.T) {
	msg := NewEmptyMessageNode("empty_message", 0, 0, 2, "H->E")

	assert.Equal(t, "empty_message", msg.Name())
	assert.Equal(t, 0, msg.StreamCode())
	assert.Equal(t, 0, msg.FunctionCode())
	assert.Equal(t, "optional", msg.WaitBit())
	assert.Equal(t, "H->E", msg.Direction())
	assert.Equal(t, "S0F0 [W] H->E empty_message", msg.Header())
	assert.Equal(t, []string{}, msg.Variables())
	assert.Equal(t, []byte{}, msg.ToHSMS(0, []byte{0, 0, 0, 0}))
	assert.Equal(t, "S0F0 [W] H->E empty_message\n.", fmt.Sprint(msg))
}

func TestMessageNode_ProducedByFactoryMethod(t *testing.T) {
	var tests = []struct {
		description       string   // Test case description
		inputMessageName  string   // Input to the factory method
		inputStreamCode   int      // Input to the factory method
		inputFunctionCode int      // Input to the factory method
		inputWaitBit      int      // Input to the factory method
		inputDirection    string   // Input to the factory method
		inputItemNode     ItemNode // Input to the factory method
		inputDeviceID     uint16   // Input to ToHSMS()
		inputSystemBytes  []byte   // Input to ToHSMS()
		expectedHeader    string   // expected result from Header()
		expectedVariables []string // expected result from Variables()
		expectedToHSMS    []byte   // expected result from ToHSMS()
		expectedString    string   // expected result from String()
	}{
		{
			description:       "S0F0 H->E, lower boundary, empty node",
			inputMessageName:  "",
			inputStreamCode:   0,
			inputFunctionCode: 0,
			inputWaitBit:      0,
			inputDirection:    "H->E",
			inputItemNode:     NewEmptyItemNode(),
			inputDeviceID:     0,
			inputSystemBytes:  []byte{0, 0, 0, 0},
			expectedVariables: []string{},
			expectedToHSMS:    []byte{0, 0, 0, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedString:    "S0F0 H->E\n.",
		},
		{
			description:       "S1F1 W H<-E A, lower boundary + 1, ASCII node",
			inputMessageName:  "A",
			inputStreamCode:   1,
			inputFunctionCode: 1,
			inputWaitBit:      1,
			inputDirection:    "H<-E",
			inputItemNode:     NewASCIINode("text"),
			inputDeviceID:     1,
			inputSystemBytes:  []byte{0, 0, 0, 1},
			expectedVariables: []string{},
			expectedToHSMS: []byte{
				0, 0, 0, 16, 0, 1, 0x81, 1, 0, 0, 0, 0, 0, 1,
				0x41, 4, 0x74, 0x65, 0x78, 0x74,
			},
			expectedString: "S1F1 W H<-E A\n<A \"text\">\n.",
		},
		{
			description:       "S64F128 H<->E message_name, intermediate values, boolean node",
			inputMessageName:  "message_name",
			inputStreamCode:   64,
			inputFunctionCode: 128,
			inputWaitBit:      0,
			inputDirection:    "H<->E",
			inputItemNode:     NewBooleanNode(true, false),
			inputDeviceID:     256,
			inputSystemBytes:  []byte{0x12, 0x34, 0x56, 0x78},
			expectedVariables: []string{},
			expectedToHSMS: []byte{
				0, 0, 0, 14, 0x01, 0x00, 0x40, 0x80, 0, 0, 0x12, 0x34, 0x56, 0x78,
				37, 2, 1, 0,
			},
			expectedString: "S64F128 H<->E message_name\n<BOOLEAN[2] T F>\n.",
		},
		{
			description:       "S126F254 [W] H->E メッセージ名, upper boundary - 1, I1 node with variable",
			inputMessageName:  "メッセージ",
			inputStreamCode:   126,
			inputFunctionCode: 254,
			inputWaitBit:      2,
			inputDirection:    "H->E",
			inputItemNode:     NewListNode(NewIntNode(1, "var")),
			inputDeviceID:     0xFFFE,
			inputSystemBytes:  []byte{0xFF, 0xFF, 0xFF, 0xFE},
			expectedVariables: []string{"var"},
			expectedToHSMS:    []byte{},
			expectedString:    "S126F254 [W] H->E メッセージ\n<L[1]\n  <I1[1] var>\n>\n.",
		},
		{
			description:       "S127F255 W H<->E 메시지_이름, upper boundary, nested list node",
			inputMessageName:  "메시지_이름",
			inputStreamCode:   127,
			inputFunctionCode: 255,
			inputWaitBit:      1,
			inputDirection:    "H<->E",
			inputItemNode:     NewListNode(NewListNode(), NewListNode(NewIntNode(1, 33, 55))),
			inputDeviceID:     0xFFFF,
			inputSystemBytes:  []byte{0xFF, 0xFF, 0xFF, 0xFF},
			expectedVariables: []string{},
			expectedToHSMS: []byte{
				0, 0, 0, 20, 0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0xFF, 0xFF, 0xFF, 0xFF,
				0x01, 2, 0x01, 0, 0x01, 1, 0x65, 2, 33, 55,
			},
			expectedString: `S127F255 W H<->E 메시지_이름
<L[2]
  <L[0]>
  <L[1]
    <I1[2] 33 55>
  >
>
.`,
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		msg := NewMessageNode(
			test.inputMessageName,
			test.inputStreamCode,
			test.inputFunctionCode,
			test.inputWaitBit,
			test.inputDirection,
			test.inputItemNode,
		)
		assert.Equal(t, test.expectedVariables, msg.Variables())
		assert.Equal(t, test.expectedToHSMS, msg.ToHSMS(test.inputDeviceID, test.inputSystemBytes))
		assert.Equal(t, test.expectedString, fmt.Sprint(msg))
	}
}

func TestMessageNode_ProducedByFillVariables(t *testing.T) {
	var tests = []struct {
		description       string                 // Test case description
		inputMessageName  string                 // Input to the factory method
		inputStreamCode   int                    // Input to the factory method
		inputFunctionCode int                    // Input to the factory method
		inputWaitBit      int                    // Input to the factory method
		inputDirection    string                 // Input to the factory method
		inputItemNode     ItemNode               // Input to the factory method
		inputFillInValues map[string]interface{} // input to FillVariables()
		inputDeviceID     uint16                 // Input to ToHSMS()
		inputSystemBytes  []byte                 // Input to ToHSMS()
		expectedHeader    string                 // expected result from Header()
		expectedVariables []string               // expected result from Variables()
		expectedToHSMS    []byte                 // expected result from ToHSMS()
		expectedString    string                 // expected result from String()
	}{
		{
			description:       "S99F99 W H<->E, Fill in all variables",
			inputMessageName:  "",
			inputStreamCode:   99,
			inputFunctionCode: 99,
			inputWaitBit:      1,
			inputDirection:    "H<->E",
			inputItemNode:     NewListNode(NewIntNode(1, "var1", "var2"), NewBooleanNode("var3")),
			inputFillInValues: map[string]interface{}{"var1": 0, "var2": -1, "var3": true},
			inputDeviceID:     0x1234,
			inputSystemBytes:  []byte{0x56, 0x78, 0x9A, 0xBC},
			expectedVariables: []string{},
			expectedToHSMS: []byte{
				0, 0, 0, 19, 0x12, 0x34, 0xE3, 0x63, 0, 0, 0x56, 0x78, 0x9A, 0xBC,
				0x01, 2, 0x65, 2, 0x00, 0xFF, 37, 1, 1,
			},
			expectedString: `S99F99 W H<->E
<L[2]
  <I1[2] 0 -1>
  <BOOLEAN[1] T>
>
.`,
		},
		{
			description:       "S99F99 W H<->E, 0 variable filled in",
			inputMessageName:  "",
			inputStreamCode:   99,
			inputFunctionCode: 99,
			inputWaitBit:      1,
			inputDirection:    "H<->E",
			inputItemNode:     NewListNode(NewIntNode(1, "var1", "var2"), NewBooleanNode("var3")),
			inputFillInValues: map[string]interface{}{"foo": "bar"},
			inputDeviceID:     0x1234,
			inputSystemBytes:  []byte{0x56, 0x78, 0x9A, 0xBC},
			expectedVariables: []string{"var1", "var2", "var3"},
			expectedToHSMS:    []byte{},
			expectedString: `S99F99 W H<->E
<L[2]
  <I1[2] var1 var2>
  <BOOLEAN[1] var3>
>
.`,
		},
		{
			description:       "S99F99 W H<->E, 1 out of 3 variable filled in",
			inputMessageName:  "",
			inputStreamCode:   99,
			inputFunctionCode: 99,
			inputWaitBit:      1,
			inputDirection:    "H<->E",
			inputItemNode:     NewListNode(NewIntNode(1, "var1", "var2"), NewBooleanNode("var3")),
			inputFillInValues: map[string]interface{}{"var3": false, "foo": "bar"},
			inputDeviceID:     0x1234,
			inputSystemBytes:  []byte{0x56, 0x78, 0x9A, 0xBC},
			expectedVariables: []string{"var1", "var2"},
			expectedToHSMS:    []byte{},
			expectedString: `S99F99 W H<->E
<L[2]
  <I1[2] var1 var2>
  <BOOLEAN[1] F>
>
.`,
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		msg := NewMessageNode(
			test.inputMessageName,
			test.inputStreamCode,
			test.inputFunctionCode,
			test.inputWaitBit,
			test.inputDirection,
			test.inputItemNode,
		)
		msg = msg.FillVariables(test.inputFillInValues)
		assert.Equal(t, test.expectedVariables, msg.Variables())
		assert.Equal(t, test.expectedToHSMS, msg.ToHSMS(test.inputDeviceID, test.inputSystemBytes))
		assert.Equal(t, test.expectedString, fmt.Sprint(msg))
	}
}
