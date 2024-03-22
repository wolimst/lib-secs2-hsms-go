package hsms

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/GunsonJack/lib-secs2-hsms-go/pkg/ast"
)

// Tests HSMS parser
//
// Testing Strategy:
//
// Parse input bytes of a HSMS message, and tests the result message using its
// public observer methods. Also test that the parsed message's HSMS byte
// representation is equal to the input bytes.
//
// Partitions:
//
// - data message:
//   - stream code: [0, 128)
//   - function code: [0, 256)
//   - wait bit: true, false
//   - data item:
//     - size: 0, 1, ...
//     - type: list, ascii, binary, boolean, I1, I2, I4, I8, F4, F8, U1, U2, U4, U8
//     - value:
//       - ascii: ascii characters
//       - binary: [0, 256)
//       - boolean: 0, 1
//       - I1, I2, I4, I8, U1, U2, U4, U8, F4, F8: [min, max] for each type
//
// - control message: select.req, select.rsp,
//                    deselect.req, deselect.rsp,
//                    linktest.req, linktest.rsp,
//                    reject.req, separate.req,
//                    undefined control message
//
// - session id: [0, 65536)
// - system bytes: [0x00000000, 0x000000FF]

func TestParser_DataMessage(t *testing.T) {
	var tests = []struct {
		description          string // test case description
		input                []byte // input to the parser
		expectedType         string // expected message type
		expectedStreamCode   int
		expectedFunctionCode int
		expectedWaitBit      string
		expectedSessionID    int
		expectedSystemBytes  []byte
		expectedString       string
	}{
		{
			description:          "S0F0 empty data item",
			input:                []byte{0, 0, 0, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedType:         "data message",
			expectedStreamCode:   0,
			expectedFunctionCode: 0,
			expectedWaitBit:      "false",
			expectedSessionID:    0,
			expectedSystemBytes:  []byte{0, 0, 0, 0},
			expectedString:       "S0F0 H<->E\n.",
		},
		{
			description: `S1F1 W <A "lorem ipsum">`,
			input: []byte{
				0, 0, 0, 23, 0, 1, 129, 1, 0, 0, 0, 0, 0, 1,
				0x41, 11, 0x6C, 0x6F, 0x72, 0x65, 0x6D, 0x20, 0x69, 0x70, 0x73, 0x75, 0x6D,
			},
			expectedType:         "data message",
			expectedStreamCode:   1,
			expectedFunctionCode: 1,
			expectedWaitBit:      "true",
			expectedSessionID:    1,
			expectedSystemBytes:  []byte{0, 0, 0, 1},
			expectedString:       "S1F1 W H<->E\n<A \"lorem ipsum\">\n.",
		},
		{
			description: `S50F50 <B[0]>`,
			input: []byte{
				0, 0, 0, 12, 0, 2, 50, 50, 0, 0, 0, 0, 0, 2,
				33, 0,
			},
			expectedType:         "data message",
			expectedStreamCode:   50,
			expectedFunctionCode: 50,
			expectedWaitBit:      "false",
			expectedSessionID:    2,
			expectedSystemBytes:  []byte{0, 0, 0, 2},
			expectedString:       "S50F50 H<->E\n<B[0]>\n.",
		},
		{
			description: `S126F254 <BOOLEAN[2] T F>`,
			input: []byte{
				0, 0, 0, 14, 0xFE, 0xFE, 126, 254, 0, 0, 0xFE, 0xFE, 0xFE, 0xFE,
				37, 2, 1, 0,
			},
			expectedType:         "data message",
			expectedStreamCode:   126,
			expectedFunctionCode: 254,
			expectedWaitBit:      "false",
			expectedSessionID:    65278,
			expectedSystemBytes:  []byte{0xFE, 0xFE, 0xFE, 0xFE},
			expectedString:       "S126F254 H<->E\n<BOOLEAN[2] T F>\n.",
		},
		{
			description: `S127F255 W <F4[3] -1.0 0.0 3.141592>`,
			input: []byte{
				0, 0, 0, 24, 0xFF, 0xFE, 255, 255, 0, 0, 0xFF, 0xFF, 0xFF, 0xFE,
				0x91, 12,
				0xBF, 0x80, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x40, 0x49, 0x0F, 0xD8,
			},
			expectedType:         "data message",
			expectedStreamCode:   127,
			expectedFunctionCode: 255,
			expectedWaitBit:      "true",
			expectedSessionID:    65534,
			expectedSystemBytes:  []byte{0xFF, 0xFF, 0xFF, 0xFE},
			expectedString:       "S127F255 W H<->E\n<F4[3] -1 0 3.141592>\n.",
		},
		{
			description: `S0F0 <F8[3] -1 0 1>`,
			input: []byte{
				0, 0, 0, 36, 0xFF, 0xFF, 0, 0, 0, 0, 0xFF, 0xFF, 0xFF, 0xFF,
				0x81, 24,
				0xBF, 0xF0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x3F, 0xF0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			expectedType:         "data message",
			expectedStreamCode:   0,
			expectedFunctionCode: 0,
			expectedWaitBit:      "false",
			expectedSessionID:    65535,
			expectedSystemBytes:  []byte{0xFF, 0xFF, 0xFF, 0xFF},
			expectedString:       "S0F0 H<->E\n<F8[3] -1 0 1>\n.",
		},
		{
			description: `S0F0, nested list`,
			input: []byte{
				0, 0, 0, 88, 0xFF, 0xFF, 0, 0, 0, 0, 0xFF, 0xFF, 0xFF, 0xFF,
				0x01, 3, // L[3]
				0x01, 0, //   L[0]
				0x01, 4, //   L[4]
				0x65, 0,
				0x69, 2, 0x80, 0x00,
				0x71, 8,
				0xFF, 0xFF, 0xFF, 0xFF,
				0, 0, 0, 0,
				0x61, 32,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0x2A,
				0x01, 4, //   L[4]
				0xA5, 4, 0, 1, 0xFE, 0xFF,
				0xA9, 4, 0, 0, 0xFF, 0xFF,
				0xB1, 4, 0, 0, 0, 0x2A,
				0xA1, 0,
			},
			expectedType:         "data message",
			expectedStreamCode:   0,
			expectedFunctionCode: 0,
			expectedWaitBit:      "false",
			expectedSessionID:    65535,
			expectedSystemBytes:  []byte{0xFF, 0xFF, 0xFF, 0xFF},
			expectedString: `S0F0 H<->E
<L[3]
  <L[0]>
  <L[4]
    <I1[0]>
    <I2[1] -32768>
    <I4[2] -1 0>
    <I8[4] -2 -1 0 42>
  >
  <L[4]
    <U1[4] 0 1 254 255>
    <U2[2] 0 65535>
    <U4[1] 42>
    <U8[0]>
  >
>
.`,
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		msg, _ := Parse(test.input)
		assert.Equal(t, test.expectedType, msg.Type())
		assert.Equal(t, test.input, msg.ToBytes())
		assert.Equal(t, test.expectedStreamCode, msg.(*ast.DataMessage).StreamCode())
		assert.Equal(t, test.expectedFunctionCode, msg.(*ast.DataMessage).FunctionCode())
		assert.Equal(t, test.expectedWaitBit, msg.(*ast.DataMessage).WaitBit())
		assert.Equal(t, test.expectedSessionID, msg.(*ast.DataMessage).SessionID())
		assert.Equal(t, test.expectedSystemBytes, msg.(*ast.DataMessage).SystemBytes())
		assert.Equal(t, test.expectedString, fmt.Sprint(msg))
		assert.Len(t, msg.(*ast.DataMessage).Variables(), 0)
	}
}

func TestParser_ControlMessage(t *testing.T) {
	var tests = []struct {
		input        []byte // input to the parser
		expectedType string // expected message type
	}{
		{
			input:        []byte{0, 0, 0, 10, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0},
			expectedType: "select.req",
		},
		{
			input:        []byte{0, 0, 0, 10, 0, 1, 0, 1, 0, 2, 0, 0, 0, 1},
			expectedType: "select.rsp",
		},
		{
			input:        []byte{0, 0, 0, 10, 1, 0, 0, 0, 0, 3, 3, 2, 1, 0},
			expectedType: "deselect.req",
		},
		{
			input:        []byte{0, 0, 0, 10, 0xAA, 0xBB, 0, 2, 0, 4, 0xFC, 0xFD, 0xFE, 0xFF},
			expectedType: "deselect.rsp",
		},
		{
			input:        []byte{0, 0, 0, 10, 0xFF, 0xFE, 0, 0, 0, 5, 0xFF, 0xFF, 0xFF, 0xFE},
			expectedType: "linktest.req",
		},
		{
			input:        []byte{0, 0, 0, 10, 0xFF, 0xFF, 0, 0, 0, 6, 0xFF, 0xFF, 0xFF, 0xFF},
			expectedType: "linktest.rsp",
		},
		{
			input:        []byte{0, 0, 0, 10, 0x12, 0x34, 9, 3, 0, 7, 0xFC, 0xFD, 0xFE, 0xFF},
			expectedType: "reject.req",
		},
		{
			input:        []byte{0, 0, 0, 10, 0xFE, 0xFE, 0, 0, 0, 9, 0xFE, 0xFE, 0xFE, 0xFE},
			expectedType: "separate.req",
		},
	}
	for _, test := range tests[len(tests)-1:] {
		msg, _ := Parse(test.input)
		assert.Equal(t, test.expectedType, msg.Type())
		assert.Equal(t, test.input, msg.ToBytes())
	}
}
