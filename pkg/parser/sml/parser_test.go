package sml

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests SECS Message Language (SML) parser
//
// Testing Strategy:
//
// Parse input string, and tests the result messages using their public observer methods,
// errors and warnings by their numbers, positions and texts.
// When no error is found, use the parsed messages' string representation as input string,
// and test that the parser create same result from the input string.
//
// Dependency:
//
// SECS-II abstract syntax tree in /pkg/ast
//
// Partitions:
//
// - Number of messages: 0, 1, ...
// - Message header:
//   - Stream code: [0, 128), error otherwise
//   - Function code: [0, 256), error otherwise
//   - Wait bit: false (empty string), true (W), optional ([W])
//               error when wait bit true and function code is even
//   - Direction: H->E, H<-E, H<->E
//                warning when not specified, direction will be H<->E
//   - Message name: empty string, ASCII characters, unicode characters, ...
// - Message text:
//   - '<': error when unexpected token found
//   - '>': error when unexpected token found
//   - Data item type: L, B, BOOLEAN, A, F4, F8, I1, I2, I4, I8, U1, U2, U4, U8
//                     error when unexpected token found
//   - Data item size: not specified, fixed size, ranged size, e.g. [1..10], [1..], [..10]
//                     error when number of values overflows the size
//   - Data item value:
//     - Common: variable,
//               error when unexpected token is found
//     - List: data item, ellipsis
//     - Binary: decimal, binary, octal, hexadecimal in [0, 256)
//               error when range overflow, error when number cannot be parsed
//     - Boolean: T, F
//     - ASCII: quoted string, ASCII number code
//              error when non-ASCII character is found
//     - F4, F8: decimal, binary, octal, hexadecimal number, possibly with scientific notation
//               error when range overflow, error when number cannot be parsed
//     - I1, I2, I4, I8: decimal, binary, octal, hexadecimal integer
//                       error when range overflow, error when number cannot be parsed
//     - U1, U2, U4, U8: decimal, binary, octal, hexadecimal unsigned integer
//                       error when range overflow, error when number cannot be parsed
//
// Input space is huge; Some important cases to check:
// - Nested data items and ellipsis
// - Case insensitivity for tokenTypeStreamFunction, tokenTypeWaitBit, tokenTypeDataItemType, tokenTypeNumber
// - Optional strings in input, like direction in message header
// - Comments; it can be anywhere except inside quoted string
// - Errors and warnings

func TestParser_NoErrorCases(t *testing.T) {
	var tests = []struct {
		description              string   // Test case description
		input                    string   // Input to the parser
		expectedNumberOfMessages int      // expected number of parsed messages
		expectedNumberOfErrors   int      // expected number of parsing errors
		expectedNumberOfWarnings int      // expected number of parsing warnings
		expectedString           []string // expected string representation of messages
	}{
		{
			description:              "empty input",
			input:                    "",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   0,
			expectedNumberOfWarnings: 0,
			expectedString:           []string{},
		},
		{
			description:              "0 message",
			input:                    "// comment 코멘트注釈ဍီကာ\n",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   0,
			expectedNumberOfWarnings: 0,
			expectedString:           []string{},
		},
		{
			description:              "1 message, no data item",
			input:                    "S0F0 H->E .",
			expectedNumberOfMessages: 1,
			expectedNumberOfErrors:   0,
			expectedNumberOfWarnings: 0,
			expectedString:           []string{"S0F0 H->E\n."},
		},
		{
			description:              "1 message, ASCII node",
			input:                    `S1F1 W H<-E <A "text">.`,
			expectedNumberOfMessages: 1,
			expectedNumberOfErrors:   0,
			expectedNumberOfWarnings: 0,
			expectedString:           []string{"S1F1 W H<-E\n<A \"text\">\n."},
		},
		{
			description:              "1 message, Binary node",
			input:                    `S63F127 [W] H<->E <B[4] 0b0 0xFE 255 var>.`,
			expectedNumberOfMessages: 1,
			expectedNumberOfErrors:   0,
			expectedNumberOfWarnings: 0,
			expectedString:           []string{"S63F127 [W] H<->E\n<B[4] 0b0 0b11111110 0b11111111 var>\n."},
		},
		{
			description:              "1 message, Boolean node",
			input:                    `S126F254 H->E TestMessage <BOOLEAN T F var>.`,
			expectedNumberOfMessages: 1,
			expectedNumberOfErrors:   0,
			expectedNumberOfWarnings: 0,
			expectedString:           []string{"S126F254 H->E TestMessage\n<BOOLEAN[3] T F var>\n."},
		},
		{
			description: "2 messages, F4, F8 node",
			input: `S126F254 H->E TestMessage1 <F4 +0.1 var -0.1>. 
			        S127F255 H->E TestMessage2 <F8 1e3 1E-3 .5e-1>.`,
			expectedNumberOfMessages: 2,
			expectedNumberOfErrors:   0,
			expectedNumberOfWarnings: 0,
			expectedString: []string{
				"S126F254 H->E TestMessage1\n<F4[3] 0.1 var -0.1>\n.",
				"S127F255 H->E TestMessage2\n<F8[3] 1000 0.001 0.05>\n.",
			},
		},
		{
			description: "4 messages, I1, I2, I4, I8 node",
			input: `S0F0 H->E TestMessage1 <I1 -128 -64 -1 0 1 64 127>. 
			        S0F0 H->E TestMessage2 <I2 -32768 __var 32767>.
			        S0F0 H->E TestMessage3 <I4 -2147483648 __var 2147483647>.
			        S0F0 H->E TestMessage4 <I8 -9223372036854775808 __var 9223372036854775807>.`,
			expectedNumberOfMessages: 4,
			expectedNumberOfErrors:   0,
			expectedNumberOfWarnings: 0,
			expectedString: []string{
				"S0F0 H->E TestMessage1\n<I1[7] -128 -64 -1 0 1 64 127>\n.",
				"S0F0 H->E TestMessage2\n<I2[3] -32768 __var 32767>\n.",
				"S0F0 H->E TestMessage3\n<I4[3] -2147483648 __var 2147483647>\n.",
				"S0F0 H->E TestMessage4\n<I8[3] -9223372036854775808 __var 9223372036854775807>\n.",
			},
		},
		{
			description: "4 messages, U1, U2, U4, U8 node",
			input: `S0F0 H->E TestMessage1 <U1[0..4] 0 1 128 255>. 
			        S0F0 H->E TestMessage2 <U2[4..4] 0 1 var 65535>.
			        S0F0 H->E TestMessage3 <U4[..4] 0 1 var 4294967295>.
			        S0F0 H->E TestMessage4 <U8[0..] 0 1 var1 var2 18446744073709551615>.`,
			expectedNumberOfMessages: 4,
			expectedNumberOfErrors:   0,
			expectedNumberOfWarnings: 0,
			expectedString: []string{
				"S0F0 H->E TestMessage1\n<U1[4] 0 1 128 255>\n.",
				"S0F0 H->E TestMessage2\n<U2[4] 0 1 var 65535>\n.",
				"S0F0 H->E TestMessage3\n<U4[4] 0 1 var 4294967295>\n.",
				"S0F0 H->E TestMessage4\n<U8[5] 0 1 var1 var2 18446744073709551615>\n.",
			},
		},
		{
			description: "2 messages, Nested list node with ellipsis and comment",
			input: `S0F0 H->E TestMessage1 // message header comment
<L          // comment
  <L[0]>    // comment
  <L[2]     // comment
    <A[0]>  // comment
    <B[0]>  // comment
  >         // comment
  ...       // comment
>           // comment
.           // comment
            // comment
S0F0 H->E TestMessage2
<L
//comment  <L
  <L
    <I1 foo>
    <L
      <I2 bar>
      var
      ...
      // comment
    >
    // comment
    ...
  >
  ...
  <I1 0>
>
// comment
.`,
			expectedNumberOfMessages: 2,
			expectedNumberOfErrors:   0,
			expectedNumberOfWarnings: 0,
			expectedString: []string{
				`S0F0 H->E TestMessage1
<L
  <L[0]>
  <L[2]
    <A[0]>
    <B[0]>
  >
  ...
>
.`,
				`S0F0 H->E TestMessage2
<L
  <L
    <I1[1] foo>
    <L
      <I2[1] bar>
      var
      ...
    >
    ...
  >
  ...
  <I1[1] 0>
>
.`,
			},
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		msgs, errs, warnings := Parse(test.input)
		assert.Len(t, msgs, test.expectedNumberOfMessages)
		assert.Len(t, errs, test.expectedNumberOfErrors)
		assert.Len(t, warnings, test.expectedNumberOfWarnings)
		for j, msg := range msgs {
			str := fmt.Sprint(msg)
			assert.Equal(t, test.expectedString[j], str)
			reparsedMsgs, reparsedErrs, reparsedWarnings := Parse(str)
			assert.Len(t, reparsedMsgs, 1)
			assert.Len(t, reparsedErrs, 0)
			assert.Len(t, reparsedWarnings, 0)
			assert.Equal(t, msg, reparsedMsgs[0])
		}
	}
}

func TestParser_CommonErrorCases(t *testing.T) {
	var tests = []struct {
		description              string   // Test case description
		input                    string   // Input to the parser
		expectedNumberOfMessages int      // expected number of parsed messages
		expectedNumberOfErrors   int      // expected number of parsing errors
		expectedNumberOfWarnings int      // expected number of parsing warnings
		expectedErrorString      []string // expected error strings in form of "line:col:subset of error text"
		expectedWarningString    []string // expected warning strings, same form as expected error string
	}{
		{
			description:              "stream function: unexpected token",
			input:                    "SxFy",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"1:1:expected stream function"},
			expectedWarningString:    []string{},
		},
		{
			description:              "stream function: stream overflow",
			input:                    "S128F255 H->E .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"1:1:overflow"},
			expectedWarningString:    []string{},
		},
		{
			description:              "stream function: function overflow",
			input:                    "S127F256 H->E .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"1:1:overflow"},
			expectedWarningString:    []string{},
		},
		{
			description:              "wait bit: true on reply message, direction: not specified",
			input:                    "S1F2 W .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 1,
			expectedErrorString:      []string{"1:6:wait bit"},
			expectedWarningString:    []string{"1:8:direction"},
		},
		{
			description:              "message text: unexpected token",
			input:                    "S0F0 H->E TestMessage\n//comment\n*",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"3:1:expected"},
			expectedWarningString:    []string{},
		},
		{
			description:              "data item: invalid data item type",
			input:                    "S0F0 H->E TestMessage\n<BOOL[1] T>",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:2:data item type"},
			expectedWarningString:    []string{},
		},
		{
			description:              "data item size: syntax error",
			input:                    "S0F0 H->E TestMessage\n<B[-3] 0> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:3:syntax error"},
			expectedWarningString:    []string{},
		},
		{
			description:              "data item size: underflow",
			input:                    "S0F0 H->E TestMessage\n<B[3] 0> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:3:overflow"},
			expectedWarningString:    []string{},
		},
		{
			description:              "data item size: underflow",
			input:                    "S0F0 H->E TestMessage\n<B[3..] 0> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:3:overflow"},
			expectedWarningString:    []string{},
		},
		{
			description:              "data item size: overflow",
			input:                    "S0F0 H->E TestMessage\n<B[..2] 0 1 2> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:3:overflow"},
			expectedWarningString:    []string{},
		},
		{
			description:              "missing token: message end",
			input:                    "S0F0 H->E TestMessage\n<B[0]>\n",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"3:1:expected message end"},
			expectedWarningString:    []string{},
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		msgs, errs, warnings := Parse(test.input)
		assert.Len(t, msgs, test.expectedNumberOfMessages)
		assert.Len(t, errs, test.expectedNumberOfErrors)
		assert.Len(t, warnings, test.expectedNumberOfWarnings)
		for j, err := range errs {
			s := strings.Split(test.expectedErrorString[j], ":")
			lineCol := fmt.Sprintf("Ln %s, Col %s", s[0], s[1])
			errTextSubset := s[2]
			assert.Truef(
				t, strings.HasPrefix(err, lineCol),
				"Wrong error position, expected %s, got %s",
				strings.Split(err, ":")[0], lineCol,
			)
			assert.Contains(t, err, errTextSubset)
		}
		for j, warning := range warnings {
			s := strings.Split(test.expectedWarningString[j], ":")
			lineCol := fmt.Sprintf("Ln %s, Col %s", s[0], s[1])
			warningTextSubset := s[2]
			assert.Truef(
				t, strings.HasPrefix(warning, lineCol),
				"Wrong warning position, expected %s, got %s",
				strings.Split(warning, ":")[0], lineCol,
			)
			assert.Contains(t, warning, warningTextSubset)
		}
	}
}

func TestParser_List_ErrorCases(t *testing.T) {
	var tests = []struct {
		description              string   // Test case description
		input                    string   // Input to the parser
		expectedNumberOfMessages int      // expected number of parsed messages
		expectedNumberOfErrors   int      // expected number of parsing errors
		expectedNumberOfWarnings int      // expected number of parsing warnings
		expectedErrorString      []string // expected error strings in form of "line:col:subset of error text"
		expectedWarningString    []string // expected warning strings, same form as expected error string
	}{
		{
			description: "duplicated variable name",
			input: `S0F0 H->E TestMessage
<L
<A[1] foo>
<A[1] foo>
<B[1] foo>
<BOOLEAN[1] foo>
<F4[1] foo>
<I1[1] foo>
<U1[1] foo>
foo
>.`,
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   7,
			expectedNumberOfWarnings: 0,
			expectedErrorString: []string{
				"4:7:duplicated var",
				"5:7:duplicated var",
				"6:13:duplicated var",
				"7:8:duplicated var",
				"8:8:duplicated var",
				"9:8:duplicated var",
				"10:1:duplicated var",
			},
			expectedWarningString: []string{},
		},
		{
			description:              "ellipsis is the first item",
			input:                    "S0F0 H->E TestMessage\n<L\n...\n>.",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"3:1:ellipsis cannot be the first"},
			expectedWarningString:    []string{},
		},
		{
			description: "ellipsis naming warning",
			input: `S0F0 H->E TestMessage
<L
<L
foo
...[1]
>
...
>.`,
			expectedNumberOfMessages: 1,
			expectedNumberOfErrors:   0,
			expectedNumberOfWarnings: 1,
			expectedErrorString:      []string{},
			expectedWarningString:    []string{"5:1:ellipsis count"},
		},
		{
			description:              "unexpected token",
			input:                    "S0F0 H->E TestMessage\n<L[1] T>\n.",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:7:expected child data item"},
			expectedWarningString:    []string{},
		},
		{
			description:              "unexpected token",
			input:                    "S0F0 H->E TestMessage\n<L[1] !@#>\n.",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:7:syntax error"},
			expectedWarningString:    []string{},
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		msgs, errs, warnings := Parse(test.input)
		assert.Len(t, msgs, test.expectedNumberOfMessages)
		assert.Len(t, errs, test.expectedNumberOfErrors)
		assert.Len(t, warnings, test.expectedNumberOfWarnings)
		for j, err := range errs {
			s := strings.Split(test.expectedErrorString[j], ":")
			lineCol := fmt.Sprintf("Ln %s, Col %s", s[0], s[1])
			errTextSubset := s[2]
			assert.Truef(
				t, strings.HasPrefix(err, lineCol),
				"Wrong error position, expected %s, got %s",
				strings.Split(err, ":")[0], lineCol,
			)
			assert.Contains(t, err, errTextSubset)
		}
		for j, warning := range warnings {
			s := strings.Split(test.expectedWarningString[j], ":")
			lineCol := fmt.Sprintf("Ln %s, Col %s", s[0], s[1])
			warningTextSubset := s[2]
			assert.Truef(
				t, strings.HasPrefix(warning, lineCol),
				"Wrong warning position, expected %s, got %s",
				strings.Split(warning, ":")[0], lineCol,
			)
			assert.Contains(t, warning, warningTextSubset)
		}
	}
}

func TestParser_ASCII_ErrorCases(t *testing.T) {
	var tests = []struct {
		description              string   // Test case description
		input                    string   // Input to the parser
		expectedNumberOfMessages int      // expected number of parsed messages
		expectedNumberOfErrors   int      // expected number of parsing errors
		expectedNumberOfWarnings int      // expected number of parsing warnings
		expectedErrorString      []string // expected error strings in form of "line:col:subset of error text"
		expectedWarningString    []string // expected warning strings, same form as expected error string
	}{
		{
			description:              "non-ascii characters",
			input:                    "S0F0 H->E TestMessage\n<A \"စာသား\"> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:4:expected ASCII"},
			expectedWarningString:    []string{},
		},
		{
			description:              "invalid character number code",
			input:                    "S0F0 H->E TestMessage\n<A 0.01> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:4:number code"},
			expectedWarningString:    []string{},
		},
		{
			description:              "non-ascii number code",
			input:                    "S0F0 H->E TestMessage\n<A 128> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:4:overflow"},
			expectedWarningString:    []string{},
		},
		{
			description:              "variable with literal",
			input:                    "S0F0 H->E TestMessage\n<A \"text\" 65 66 var> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:17:variable"},
			expectedWarningString:    []string{},
		},
		{
			description:              "unexpected token",
			input:                    "S0F0 H->E TestMessage\n<A BOOLEAN> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:4:expected quoted string"},
			expectedWarningString:    []string{},
		},
		{
			description:              "unexpected token (error token)",
			input:                    "S0F0 H->E TestMessage\n<A[..10] !@#> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:10:syntax error"},
			expectedWarningString:    []string{},
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		msgs, errs, warnings := Parse(test.input)
		assert.Len(t, msgs, test.expectedNumberOfMessages)
		assert.Len(t, errs, test.expectedNumberOfErrors)
		assert.Len(t, warnings, test.expectedNumberOfWarnings)
		for j, err := range errs {
			s := strings.Split(test.expectedErrorString[j], ":")
			lineCol := fmt.Sprintf("Ln %s, Col %s", s[0], s[1])
			errTextSubset := s[2]
			assert.Truef(
				t, strings.HasPrefix(err, lineCol),
				"Wrong error position, expected %s, got %s",
				strings.Split(err, ":")[0], lineCol,
			)
			assert.Contains(t, err, errTextSubset)
		}
		for j, warning := range warnings {
			s := strings.Split(test.expectedWarningString[j], ":")
			lineCol := fmt.Sprintf("Ln %s, Col %s", s[0], s[1])
			warningTextSubset := s[2]
			assert.Truef(
				t, strings.HasPrefix(warning, lineCol),
				"Wrong warning position, expected %s, got %s",
				strings.Split(warning, ":")[0], lineCol,
			)
			assert.Contains(t, warning, warningTextSubset)
		}
	}
}

func TestParser_Binary_ErrorCases(t *testing.T) {
	var tests = []struct {
		description              string   // Test case description
		input                    string   // Input to the parser
		expectedNumberOfMessages int      // expected number of parsed messages
		expectedNumberOfErrors   int      // expected number of parsing errors
		expectedNumberOfWarnings int      // expected number of parsing warnings
		expectedErrorString      []string // expected error strings in form of "line:col:subset of error text"
		expectedWarningString    []string // expected warning strings, same form as expected error string
	}{
		{
			description:              "underflow",
			input:                    "S0F0 H->E TestMessage\n<B -1> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:4:overflow"},
			expectedWarningString:    []string{},
		},
		{
			description:              "overflow",
			input:                    "S0F0 H->E TestMessage\n<B 256> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:4:overflow"},
			expectedWarningString:    []string{},
		},
		{
			description:              "unexpected token",
			input:                    "S0F0 H->E TestMessage\n<B[1] T> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:7:expected number or variable"},
			expectedWarningString:    []string{},
		},
		{
			description:              "unexpected token (error token)",
			input:                    "S0F0 H->E TestMessage\n<B[2] !@#> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:7:syntax error"},
			expectedWarningString:    []string{},
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		msgs, errs, warnings := Parse(test.input)
		assert.Len(t, msgs, test.expectedNumberOfMessages)
		assert.Len(t, errs, test.expectedNumberOfErrors)
		assert.Len(t, warnings, test.expectedNumberOfWarnings)
		for j, err := range errs {
			s := strings.Split(test.expectedErrorString[j], ":")
			lineCol := fmt.Sprintf("Ln %s, Col %s", s[0], s[1])
			errTextSubset := s[2]
			assert.Truef(
				t, strings.HasPrefix(err, lineCol),
				"Wrong error position, expected %s, got %s",
				strings.Split(err, ":")[0], lineCol,
			)
			assert.Contains(t, err, errTextSubset)
		}
		for j, warning := range warnings {
			s := strings.Split(test.expectedWarningString[j], ":")
			lineCol := fmt.Sprintf("Ln %s, Col %s", s[0], s[1])
			warningTextSubset := s[2]
			assert.Truef(
				t, strings.HasPrefix(warning, lineCol),
				"Wrong warning position, expected %s, got %s",
				strings.Split(warning, ":")[0], lineCol,
			)
			assert.Contains(t, warning, warningTextSubset)
		}
	}
}

func TestParser_Boolean_ErrorCases(t *testing.T) {
	var tests = []struct {
		description              string   // Test case description
		input                    string   // Input to the parser
		expectedNumberOfMessages int      // expected number of parsed messages
		expectedNumberOfErrors   int      // expected number of parsing errors
		expectedNumberOfWarnings int      // expected number of parsing warnings
		expectedErrorString      []string // expected error strings in form of "line:col:subset of error text"
		expectedWarningString    []string // expected warning strings, same form as expected error string
	}{
		{
			description:              "unexpected token",
			input:                    "S0F0 H->E TestMessage\n<BOOLEAN[1] 10> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:13:expected boolean"},
			expectedWarningString:    []string{},
		},
		{
			description:              "unexpected token (error token)",
			input:                    "S0F0 H->E TestMessage\n<BOOLEAN[1] !@#> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:13:syntax error"},
			expectedWarningString:    []string{},
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		msgs, errs, warnings := Parse(test.input)
		assert.Len(t, msgs, test.expectedNumberOfMessages)
		assert.Len(t, errs, test.expectedNumberOfErrors)
		assert.Len(t, warnings, test.expectedNumberOfWarnings)
		for j, err := range errs {
			s := strings.Split(test.expectedErrorString[j], ":")
			lineCol := fmt.Sprintf("Ln %s, Col %s", s[0], s[1])
			errTextSubset := s[2]
			assert.Truef(
				t, strings.HasPrefix(err, lineCol),
				"Wrong error position, expected %s, got %s",
				strings.Split(err, ":")[0], lineCol,
			)
			assert.Contains(t, err, errTextSubset)
		}
		for j, warning := range warnings {
			s := strings.Split(test.expectedWarningString[j], ":")
			lineCol := fmt.Sprintf("Ln %s, Col %s", s[0], s[1])
			warningTextSubset := s[2]
			assert.Truef(
				t, strings.HasPrefix(warning, lineCol),
				"Wrong warning position, expected %s, got %s",
				strings.Split(warning, ":")[0], lineCol,
			)
			assert.Contains(t, warning, warningTextSubset)
		}
	}
}

func TestParser_Float_ErrorCases(t *testing.T) {
	var tests = []struct {
		description              string   // Test case description
		input                    string   // Input to the parser
		expectedNumberOfMessages int      // expected number of parsed messages
		expectedNumberOfErrors   int      // expected number of parsing errors
		expectedNumberOfWarnings int      // expected number of parsing warnings
		expectedErrorString      []string // expected error strings in form of "line:col:subset of error text"
		expectedWarningString    []string // expected warning strings, same form as expected error string
	}{
		{
			description:              "F4 overflow",
			input:                    "S0F0 H->E TestMessage\n<F4 1e99999> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:5:overflow"},
			expectedWarningString:    []string{},
		},
		{
			description:              "F8 overflow",
			input:                    "S0F0 H->E TestMessage\n<F8 1e99999> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:5:overflow"},
			expectedWarningString:    []string{},
		},
		{
			description:              "unexpected token",
			input:                    "S0F0 H->E TestMessage\n<F4[1] T> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:8:expected float"},
			expectedWarningString:    []string{},
		},
		{
			description:              "unexpected token (error token)",
			input:                    "S0F0 H->E TestMessage\n<F4[1] !@#> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:8:syntax error"},
			expectedWarningString:    []string{},
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		msgs, errs, warnings := Parse(test.input)
		assert.Len(t, msgs, test.expectedNumberOfMessages)
		assert.Len(t, errs, test.expectedNumberOfErrors)
		assert.Len(t, warnings, test.expectedNumberOfWarnings)
		for j, err := range errs {
			s := strings.Split(test.expectedErrorString[j], ":")
			lineCol := fmt.Sprintf("Ln %s, Col %s", s[0], s[1])
			errTextSubset := s[2]
			assert.Truef(
				t, strings.HasPrefix(err, lineCol),
				"Wrong error position, expected %s, got %s",
				strings.Split(err, ":")[0], lineCol,
			)
			assert.Contains(t, err, errTextSubset)
		}
		for j, warning := range warnings {
			s := strings.Split(test.expectedWarningString[j], ":")
			lineCol := fmt.Sprintf("Ln %s, Col %s", s[0], s[1])
			warningTextSubset := s[2]
			assert.Truef(
				t, strings.HasPrefix(warning, lineCol),
				"Wrong warning position, expected %s, got %s",
				strings.Split(warning, ":")[0], lineCol,
			)
			assert.Contains(t, warning, warningTextSubset)
		}
	}
}

func TestParser_Int_ErrorCases(t *testing.T) {
	var tests = []struct {
		description              string   // Test case description
		input                    string   // Input to the parser
		expectedNumberOfMessages int      // expected number of parsed messages
		expectedNumberOfErrors   int      // expected number of parsing errors
		expectedNumberOfWarnings int      // expected number of parsing warnings
		expectedErrorString      []string // expected error strings in form of "line:col:subset of error text"
		expectedWarningString    []string // expected warning strings, same form as expected error string
	}{
		{
			description: "underflow",
			input: `S0F0 H->E TestMessage
<L[4]
<I1 -129>
<I2 -32769>
<I4 -2147483649>
<I8 -9223372036854775809>
>.`,
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   4,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"3:5:overflow", "4:5:overflow", "5:5:overflow", "6:5:overflow"},
			expectedWarningString:    []string{},
		},
		{
			description: "overflow",
			input: `S0F0 H->E TestMessage
<L[4]
<I1 128>
<I2 32768>
<I4 2147483648>
<I8 9223372036854775808>
>.`,
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   4,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"3:5:overflow", "4:5:overflow", "5:5:overflow", "6:5:overflow"},
			expectedWarningString:    []string{},
		},
		{
			description:              "unexpected token",
			input:                    "S0F0 H->E TestMessage\n<I1[2] 0.12 T> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   2,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:8:expected integer", "2:13:expected integer"},
			expectedWarningString:    []string{},
		},
		{
			description:              "unexpected token (error token)",
			input:                    "S0F0 H->E TestMessage\n<I1[1] !@#> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:8:syntax error"},
			expectedWarningString:    []string{},
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		msgs, errs, warnings := Parse(test.input)
		assert.Len(t, msgs, test.expectedNumberOfMessages)
		assert.Len(t, errs, test.expectedNumberOfErrors)
		assert.Len(t, warnings, test.expectedNumberOfWarnings)
		for j, err := range errs {
			s := strings.Split(test.expectedErrorString[j], ":")
			lineCol := fmt.Sprintf("Ln %s, Col %s", s[0], s[1])
			errTextSubset := s[2]
			assert.Truef(
				t, strings.HasPrefix(err, lineCol),
				"Wrong error position, expected %s, got %s",
				strings.Split(err, ":")[0], lineCol,
			)
			assert.Contains(t, err, errTextSubset)
		}
		for j, warning := range warnings {
			s := strings.Split(test.expectedWarningString[j], ":")
			lineCol := fmt.Sprintf("Ln %s, Col %s", s[0], s[1])
			warningTextSubset := s[2]
			assert.Truef(
				t, strings.HasPrefix(warning, lineCol),
				"Wrong warning position, expected %s, got %s",
				strings.Split(warning, ":")[0], lineCol,
			)
			assert.Contains(t, warning, warningTextSubset)
		}
	}
}

func TestParser_Uint_ErrorCases(t *testing.T) {
	var tests = []struct {
		description              string   // Test case description
		input                    string   // Input to the parser
		expectedNumberOfMessages int      // expected number of parsed messages
		expectedNumberOfErrors   int      // expected number of parsing errors
		expectedNumberOfWarnings int      // expected number of parsing warnings
		expectedErrorString      []string // expected error strings in form of "line:col:subset of error text"
		expectedWarningString    []string // expected warning strings, same form as expected error string
	}{
		{
			description: "overflow",
			input: `S0F0 H->E TestMessage
<L[4]
<U1 256>
<U2 65536>
<U4 4294967296>
<U8 18446744073709551616>
>.`,
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   4,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"3:5:overflow", "4:5:overflow", "5:5:overflow", "6:5:overflow"},
			expectedWarningString:    []string{},
		},
		{
			description:              "unexpected token",
			input:                    "S0F0 H->E TestMessage\n<U1[1] -1 T> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   2,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:8:expected unsigned integer", "2:11:expected unsigned integer"},
			expectedWarningString:    []string{},
		},
		{
			description:              "unexpected token (error token)",
			input:                    "S0F0 H->E TestMessage\n<U1[1] !@#> .",
			expectedNumberOfMessages: 0,
			expectedNumberOfErrors:   1,
			expectedNumberOfWarnings: 0,
			expectedErrorString:      []string{"2:8:syntax error"},
			expectedWarningString:    []string{},
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		msgs, errs, warnings := Parse(test.input)
		assert.Len(t, msgs, test.expectedNumberOfMessages)
		assert.Len(t, errs, test.expectedNumberOfErrors)
		assert.Len(t, warnings, test.expectedNumberOfWarnings)
		for j, err := range errs {
			s := strings.Split(test.expectedErrorString[j], ":")
			lineCol := fmt.Sprintf("Ln %s, Col %s", s[0], s[1])
			errTextSubset := s[2]
			assert.Truef(
				t, strings.HasPrefix(err, lineCol),
				"Wrong error position, expected %s, got %s",
				strings.Split(err, ":")[0], lineCol,
			)
			assert.Contains(t, err, errTextSubset)
		}
		for j, warning := range warnings {
			s := strings.Split(test.expectedWarningString[j], ":")
			lineCol := fmt.Sprintf("Ln %s, Col %s", s[0], s[1])
			warningTextSubset := s[2]
			assert.Truef(
				t, strings.HasPrefix(warning, lineCol),
				"Wrong warning position, expected %s, got %s",
				strings.Split(warning, ":")[0], lineCol,
			)
			assert.Contains(t, warning, warningTextSubset)
		}
	}
}
