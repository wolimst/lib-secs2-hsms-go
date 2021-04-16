package sml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Testing Strategy:
//
// Create a lexer using lex() and call nextToken() repeatedly
// to test that the lexer tokenize the input correctly.
//
// Partitions:
//
// The input space is quite huge; Some important cases to check:
// - Check all token types
// - Nested data items
// - Case insensitivity for tokenTypeStreamFunction, tokenTypeWaitBit, tokenTypeDataItemType, tokenTypeNumber
// - Optional strings in input, like dot in the tokenTypeNumber
// - Comments; it can be anywhere except inside quoted string
// - Errors

// Helper functions and variables

var (
	// Tokens with fixed value
	// The error messages, line and column numbers are ignored to simplify testing

	tokenError      = token{tokenTypeError, "error", -1, -1}
	tokenMessageEnd = token{tokenTypeMessageEnd, ".", -1, -1}
	tokenLAB        = token{tokenTypeLeftAngleBracket, "<", -1, -1}
	tokenRAB        = token{tokenTypeRightAngleBracket, ">", -1, -1}
)

// doLex run the lexer and returns the identified tokens, except EOF.
func doLex(input string, initStateFn stateFn) (tokens []token) {
	lexer := lex(input)
	lexer.state = initStateFn
	tokens = []token{}

	tok := lexer.nextToken()
	for tok.typ != tokenTypeEOF {
		// ignore error message and line number to ease testing
		switch tok.typ {
		case tokenTypeError:
			tok.val = "error"
			fallthrough
		case tokenTypeMessageEnd, tokenTypeLeftAngleBracket, tokenTypeRightAngleBracket:
			tok.line = -1
			tok.col = -1
		}
		tokens = append(tokens, tok)
		tok = lexer.nextToken()
	}
	return tokens
}

// Tests

func TestLexer_EOF(t *testing.T) {
	lexer := lex("")
	tok := lexer.nextToken()

	assert.Equal(t, tokenTypeEOF, tok.typ)
	assert.Equal(t, 1, tok.line)
	assert.Equal(t, 1, tok.col)
}

func TestLexer_MessageHeader(t *testing.T) {
	var tests = []struct {
		description string
		input       string
		expected    []token
	}{
		{
			description: "Stream function code, Uppercase",
			input:       "S1F1",
			expected:    []token{{tokenTypeStreamFunction, "S1F1", 1, 1}},
		},
		{
			description: "Stream function code, Direction, Uppercase",
			input:       "S1F1 H->E",
			expected:    []token{{tokenTypeStreamFunction, "S1F1", 1, 1}, {tokenTypeDirection, "H->E", 1, 6}},
		},
		{
			description: "Stream function code, Wait bit, Mixed case",
			input:       "S9f99 W",
			expected:    []token{{tokenTypeStreamFunction, "S9F99", 1, 1}, {tokenTypeWaitBit, "W", 1, 7}},
		},
		{
			description: "Stream function code, Wait bit, Direction, Mixed case",
			input:       "s42F9\tw H<->e",
			expected:    []token{{tokenTypeStreamFunction, "S42F9", 1, 1}, {tokenTypeWaitBit, "W", 1, 7}, {tokenTypeDirection, "H<->E", 1, 9}},
		},
		{
			description: "Stream function code, Wait bit, Direction, Message name, Mixed case",
			input:       "s42F9\tW h<->E message_name",
			expected:    []token{{tokenTypeStreamFunction, "S42F9", 1, 1}, {tokenTypeWaitBit, "W", 1, 7}, {tokenTypeDirection, "H<->E", 1, 9}, {tokenTypeMessageName, "message_name", 1, 15}},
		},
		{
			description: "Stream function code, Wait bit, lowercase, trailing spaces",
			input:       "s42f128  [w] h<-e   ",
			expected:    []token{{tokenTypeStreamFunction, "S42F128", 1, 1}, {tokenTypeWaitBit, "[W]", 1, 10}, {tokenTypeDirection, "H<-E", 1, 14}},
		},
		{
			description: "Whitespace types, Unicode message name",
			input:       "\t \t메시지メッセージသတင်းစကား✉ \t\n S999F999 \t\r\n [w]\r\n",
			expected:    []token{{tokenTypeMessageName, "메시지メッセージသတင်းစကား✉", 1, 4}, {tokenTypeStreamFunction, "S999F999", 2, 2}, {tokenTypeWaitBit, "[W]", 3, 2}},
		},
		{
			description: "Without spaces",
			input:       "S1F3wH->EMSGNAME//Comment",
			expected:    []token{{tokenTypeStreamFunction, "S1F3", 1, 1}, {tokenTypeWaitBit, "W", 1, 5}, {tokenTypeDirection, "H->E", 1, 6}, {tokenTypeMessageName, "MSGNAME", 1, 10}, {tokenTypeComment, "//Comment", 1, 17}},
		},
	}
	for i, test := range tests {
		t.Logf("Test #%d: %s", i, test.description)
		tokens := doLex(test.input, lexMessageHeader)
		assert.Equal(t, test.expected, tokens)
	}
}

func TestLexer_DataType(t *testing.T) {
	var tests = []struct {
		input    string
		expected []token
	}{
		{
			input:    "L",
			expected: []token{{tokenTypeDataItemType, "L", 1, 1}},
		},
		{
			input:    " boolean",
			expected: []token{{tokenTypeDataItemType, "BOOLEAN", 1, 2}},
		},
		{
			input:    "\tBOOLEAN",
			expected: []token{{tokenTypeDataItemType, "BOOLEAN", 1, 2}},
		},
		{
			input:    "\nBoolean",
			expected: []token{{tokenTypeDataItemType, "BOOLEAN", 2, 1}},
		},
		{
			input:    "\r\n\r\n\tb",
			expected: []token{{tokenTypeDataItemType, "B", 3, 2}},
		},
		{
			input:    "\r\n\r\n\ta",
			expected: []token{{tokenTypeDataItemType, "A", 3, 2}},
		},
		{
			input:    "I1",
			expected: []token{{tokenTypeDataItemType, "I1", 1, 1}},
		},
		{
			input:    "U4",
			expected: []token{{tokenTypeDataItemType, "U4", 1, 1}},
		},
		{
			input:    "f8",
			expected: []token{{tokenTypeDataItemType, "F8", 1, 1}},
		},
	}
	for _, test := range tests {
		tokens := doLex(test.input, lexMessageText)
		assert.Equal(t, test.expected, tokens)
	}
}

func TestLexer_DataItemSize(t *testing.T) {
	var tests = []struct {
		input    string
		expected []token
	}{
		{
			input:    "[0]",
			expected: []token{{tokenTypeDataItemSize, "[0]", 1, 1}},
		},
		{
			input:    "[ 1 ]",
			expected: []token{{tokenTypeDataItemSize, "[1]", 1, 1}},
		},
		{
			input:    "[  42\t\t]",
			expected: []token{{tokenTypeDataItemSize, "[42]", 1, 1}},
		},
		{
			input:    "[0..42]",
			expected: []token{{tokenTypeDataItemSize, "[0..42]", 1, 1}},
		},
		{
			input:    "[\n0 ..\n42\n]",
			expected: []token{{tokenTypeDataItemSize, "[0..42]", 1, 1}},
		},
		{
			input:    "[0..]",
			expected: []token{{tokenTypeDataItemSize, "[0..]", 1, 1}},
		},
		{
			input:    "[..42]",
			expected: []token{{tokenTypeDataItemSize, "[..42]", 1, 1}},
		},
		{ // Wrong syntax
			input:    "[0 ... 42]",
			expected: []token{tokenError},
		},
	}
	for _, test := range tests {
		tokens := doLex(test.input, lexMessageText)
		assert.Equal(t, test.expected, tokens)
	}
}

func TestLexer_Number(t *testing.T) {
	var tests = []struct {
		input    string
		expected []token
	}{
		{
			input:    "0",
			expected: []token{{tokenTypeNumber, "0", 1, 1}},
		},
		{
			input:    "1",
			expected: []token{{tokenTypeNumber, "1", 1, 1}},
		},
		{
			input:    "142",
			expected: []token{{tokenTypeNumber, "142", 1, 1}},
		},
		{
			input:    "+0",
			expected: []token{{tokenTypeNumber, "+0", 1, 1}},
		},
		{
			input:    "+1",
			expected: []token{{tokenTypeNumber, "+1", 1, 1}},
		},
		{
			input:    "+42",
			expected: []token{{tokenTypeNumber, "+42", 1, 1}},
		},
		{
			input:    "-0",
			expected: []token{{tokenTypeNumber, "-0", 1, 1}},
		},
		{
			input:    "-1",
			expected: []token{{tokenTypeNumber, "-1", 1, 1}},
		},
		{
			input:    "-42",
			expected: []token{{tokenTypeNumber, "-42", 1, 1}},
		},
		{
			input:    "1.042",
			expected: []token{{tokenTypeNumber, "1.042", 1, 1}},
		},
		{
			input:    ".042",
			expected: []token{{tokenTypeNumber, ".042", 1, 1}},
		},
		{
			input:    "-.042",
			expected: []token{{tokenTypeNumber, "-.042", 1, 1}},
		},
		{
			input:    "1E0",
			expected: []token{{tokenTypeNumber, "1E0", 1, 1}},
		},
		{
			input:    "1.496e+8",
			expected: []token{{tokenTypeNumber, "1.496e+8", 1, 1}},
		},
		{
			input:    "6.626e-34",
			expected: []token{{tokenTypeNumber, "6.626e-34", 1, 1}},
		},
		{
			input:    "0x042EFF 0XABC",
			expected: []token{{tokenTypeNumber, "0x042EFF", 1, 1}, {tokenTypeNumber, "0XABC", 1, 10}},
		},
		{
			input:    "0b0101 0B100",
			expected: []token{{tokenTypeNumber, "0b0101", 1, 1}, {tokenTypeNumber, "0B100", 1, 8}},
		},
		{
			input:    "0o777 0O076",
			expected: []token{{tokenTypeNumber, "0o777", 1, 1}, {tokenTypeNumber, "0O076", 1, 7}},
		},
		{ // Wrong syntax
			input:    "42BF",
			expected: []token{tokenError},
		},
	}
	for _, test := range tests {
		tokens := doLex(test.input, lexMessageText)
		assert.Equal(t, test.expected, tokens)
	}
}

func TestLexer_Variable(t *testing.T) {
	var tests = []struct {
		input    string
		expected []token
	}{
		{
			input:    "ack",
			expected: []token{{tokenTypeVariable, "ack", 1, 1}},
		},
		{
			input:    "ECID",
			expected: []token{{tokenTypeVariable, "ECID", 1, 1}},
		},
		{
			input:    "var_42[0]",
			expected: []token{{tokenTypeVariable, "var_42[0]", 1, 1}},
		},
		{
			input:    "__var42[1][2][42]",
			expected: []token{{tokenTypeVariable, "__var42[1][2][42]", 1, 1}},
		},
		// variable names that are similar to reserved names
		{
			input:    "List",
			expected: []token{{tokenTypeVariable, "List", 1, 1}},
		},
		{
			input:    "binary",
			expected: []token{{tokenTypeVariable, "binary", 1, 1}},
		},
		{
			input:    "booleanT",
			expected: []token{{tokenTypeVariable, "booleanT", 1, 1}},
		},
		{
			input:    "ascii",
			expected: []token{{tokenTypeVariable, "ascii", 1, 1}},
		},
		{
			input:    "float",
			expected: []token{{tokenTypeVariable, "float", 1, 1}},
		},
		{
			input:    "true",
			expected: []token{{tokenTypeVariable, "true", 1, 1}},
		},
		{
			input:    "False",
			expected: []token{{tokenTypeVariable, "False", 1, 1}},
		},
	}
	for _, test := range tests {
		tokens := doLex(test.input, lexMessageText)
		assert.Equal(t, test.expected, tokens)
	}
}

func TestLexer_QuotedString(t *testing.T) {
	var tests = []struct {
		input    string
		expected []token
	}{
		{
			input:    `""`,
			expected: []token{{tokenTypeQuotedString, `""`, 1, 1}},
		},
		{
			input:    `"a"`,
			expected: []token{{tokenTypeQuotedString, `"a"`, 1, 1}},
		},
		{
			input:    `"QUOTED"`,
			expected: []token{{tokenTypeQuotedString, `"QUOTED"`, 1, 1}},
		},
		{
			input:    `"123!@#'()[]-+\//"`,
			expected: []token{{tokenTypeQuotedString, `"123!@#'()[]-+\//"`, 1, 1}},
		},
		{
			input:    `" with  spaces "`,
			expected: []token{{tokenTypeQuotedString, `" with  spaces "`, 1, 1}},
		},
		{
			input:    "\"\twith\t\ttabs\t\"",
			expected: []token{{tokenTypeQuotedString, "\"\twith\t\ttabs\t\"", 1, 1}},
		},
		// Wrong syntax; quoted strings must finished in one line
		{
			input:    "\"line feed\n\"",
			expected: []token{tokenError},
		},
		{
			input:    "\"carriage return\r\"",
			expected: []token{tokenError},
		},
	}
	for _, test := range tests {
		tokens := doLex(test.input, lexMessageText)
		assert.Equal(t, test.expected, tokens)
	}
}

func TestLexer_Ellipsis_Bool_Error(t *testing.T) {
	var tests = []struct {
		input    string
		expected []token
	}{
		{ // Ellipsis
			input:    "...",
			expected: []token{{tokenTypeEllipsis, "...", 1, 1}},
		},
		{ // Bool
			input:    "T",
			expected: []token{{tokenTypeBool, "T", 1, 1}},
		},
		{
			input:    "F",
			expected: []token{{tokenTypeBool, "F", 1, 1}},
		},
		{
			input:    "t f",
			expected: []token{{tokenTypeBool, "T", 1, 1}, {tokenTypeBool, "F", 1, 3}},
		},
		{ // Error in lexMessageText
			input:    "!@#",
			expected: []token{tokenError},
		},
	}
	for _, test := range tests {
		tokens := doLex(test.input, lexMessageText)
		assert.Equal(t, test.expected, tokens)
	}
}

func TestLexer_FullMessage_NestedList_Comments(t *testing.T) {
	var tests = []struct {
		input    string
		expected []token
	}{
		{
			input: "// Input with comments only\r\n" +
				"   // 2nd line comment and EOF",
			expected: []token{{tokenTypeComment, "// Input with comments only", 1, 1}, {tokenTypeComment, "// 2nd line comment and EOF", 2, 4}},
		},
		{
			input: `S99F99 [W] H<->E TestMessage
<L [4]
  <B 0b1>
  <U4 42>
  <A[..10] ALTX>
  <L[2]
    <BOOLEAN[1] F>
    <F4[0]>
  >
>
.`,
			expected: []token{
				// Message header
				{tokenTypeStreamFunction, "S99F99", 1, 1},
				{tokenTypeWaitBit, "[W]", 1, 8},
				{tokenTypeDirection, "H<->E", 1, 12},
				{tokenTypeMessageName, "TestMessage", 1, 18},
				// <L [4]
				tokenLAB,
				{tokenTypeDataItemType, "L", 2, 2},
				{tokenTypeDataItemSize, "[4]", 2, 4},
				// <B 0b1>
				tokenLAB,
				{tokenTypeDataItemType, "B", 3, 4},
				{tokenTypeNumber, "0b1", 3, 6},
				tokenRAB,
				// <U4 42>
				tokenLAB,
				{tokenTypeDataItemType, "U4", 4, 4},
				{tokenTypeNumber, "42", 4, 7},
				tokenRAB,
				// <A[..10] ALTX>
				tokenLAB,
				{tokenTypeDataItemType, "A", 5, 4},
				{tokenTypeDataItemSize, "[..10]", 5, 5},
				{tokenTypeVariable, "ALTX", 5, 12},
				tokenRAB,
				// <L[2]
				tokenLAB,
				{tokenTypeDataItemType, "L", 6, 4},
				{tokenTypeDataItemSize, "[2]", 6, 5},
				// <BOOLEAN[1] F>
				tokenLAB,
				{tokenTypeDataItemType, "BOOLEAN", 7, 6},
				{tokenTypeDataItemSize, "[1]", 7, 13},
				{tokenTypeBool, "F", 7, 17},
				tokenRAB,
				// <F4[0]>
				tokenLAB,
				{tokenTypeDataItemType, "F4", 8, 6},
				{tokenTypeDataItemSize, "[0]", 8, 8},
				tokenRAB,
				// >
				tokenRAB,
				// >.
				tokenRAB,
				tokenMessageEnd,
			},
		},
		{
			input: `// Test Message
S99F99// [W] wait bit is commented out
// Message text start
<L [4] // Comment inside message text
  // Unicode chracters in comment 한글がなカナ漢字
  <B 0// b1
  >
  <U// 4 42
  >
  <A AL// TX
  >
  <L[2]
    // <BOOLEAN[1] F>
    < // >
  <A "string // Error, missing closing double quote
  >
>
.`,
			expected: []token{
				{tokenTypeComment, "// Test Message", 1, 1},
				{tokenTypeStreamFunction, "S99F99", 2, 1},
				{tokenTypeComment, "// [W] wait bit is commented out", 2, 7},
				{tokenTypeComment, "// Message text start", 3, 1},
				tokenLAB, {tokenTypeDataItemType, "L", 4, 2}, {tokenTypeDataItemSize, "[4]", 4, 4}, {tokenTypeComment, "// Comment inside message text", 4, 8},
				{tokenTypeComment, "// Unicode chracters in comment 한글がなカナ漢字", 5, 3},
				tokenLAB, {tokenTypeDataItemType, "B", 6, 4}, {tokenTypeNumber, "0", 6, 6}, {tokenTypeComment, "// b1", 6, 7},
				tokenRAB,
				tokenLAB, {tokenTypeVariable, "U", 8, 4}, {tokenTypeComment, "// 4 42", 8, 5},
				tokenRAB,
				tokenLAB, {tokenTypeDataItemType, "A", 10, 4}, {tokenTypeVariable, "AL", 10, 6}, {tokenTypeComment, "// TX", 10, 8},
				tokenRAB,
				tokenLAB, {tokenTypeDataItemType, "L", 12, 4}, {tokenTypeDataItemSize, "[2]", 12, 5},
				{tokenTypeComment, "// <BOOLEAN[1] F>", 13, 5},
				tokenLAB, {tokenTypeComment, "// >", 14, 7},
				tokenLAB, {tokenTypeDataItemType, "A", 15, 4}, tokenError,
				// Error occurred; following input is ignored
			},
		},
	}
	for _, test := range tests {
		tokens := doLex(test.input, lexMessageHeader)
		assert.Equal(t, test.expected, tokens)
	}
}
