package sml

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/GunsonJack/lib-secs2-hsms-go/pkg/ast"
)

// Parse parses the input string, and return parsed message nodes and parsing errors/warnings.
//
// input should have UTF-8 encoding.
//
// No messages is returned if error exist in the input.
// errors and warnings have format of "Ln x, Col y: error text".
func Parse(input string) (messages []*ast.DataMessage, errors, warnings []string) {
	p := &parser{
		input:      input,
		lexer:      lex(input),
		tokenQueue: []token{},
		messages:   []*ast.DataMessage{},
		errors:     []parseError{},
		warnings:   []parseError{},
	}

	for p.peek().typ != tokenTypeEOF {
		if ok := p.parseMessage(); !ok {
			break
		}
	}

	errors = make([]string, 0, len(p.errors))
	warnings = make([]string, 0, len(p.warnings))
	for _, err := range p.errors {
		errors = append(errors, err.string())
	}
	for _, warning := range p.warnings {
		warnings = append(warnings, warning.string())
	}

	if len(errors) > 0 {
		return []*ast.DataMessage{}, errors, warnings
	}
	return p.messages, errors, warnings
}

type parser struct {
	input         string             // input string to parse
	lexer         *lexer             // lexer to tokenize the input string
	tokenQueue    []token            // token queue that the lexer tokenized
	variableNames map[string]bool    // variable names in a message to check duplicates
	ellipsisCount int                // ellipsis count in a message
	messages      []*ast.DataMessage // parsed messages
	errors        []parseError       // parsing errors
	warnings      []parseError       // parsing warnings
}

type parseError struct {
	line int
	col  int
	text string
}

func (pe *parseError) string() string {
	return fmt.Sprintf("Ln %d, Col %d: %s", pe.line, pe.col, pe.text)
}

// peek returns the next token.
func (p *parser) peek() token {
	if len(p.tokenQueue) == 0 {
		var t token
		for {
			// ignore comment token
			if t = p.lexer.nextToken(); t.typ != tokenTypeComment {
				break
			}
		}
		p.tokenQueue = append(p.tokenQueue, t)
	}
	return p.tokenQueue[0]
}

// accentAny returns the next token, and removes it from the token queue.
func (p *parser) acceptAny() token {
	t := p.peek()
	p.tokenQueue = p.tokenQueue[1:]
	return t
}

// accept returns the next token, and if the token type matches, removes the
// token from the token queue. The second return value ok is true if and only
// if the token type matches.
func (p *parser) accept(typ tokenType) (t token, ok bool) {
	t = p.peek()
	if t.typ == typ {
		return p.acceptAny(), true
	}
	return t, false
}

// errorf create parse error and append it to parser.errors slice.
func (p *parser) errorf(t token, format string, args ...interface{}) {
	p.errors = append(p.errors, parseError{t.line, t.col, fmt.Sprintf(format, args...)})
}

func (p *parser) warningf(t token, format string, args ...interface{}) {
	p.warnings = append(p.warnings, parseError{t.line, t.col, fmt.Sprintf(format, args...)})
}

// parseMessage parses a SECS-II message.
// Returns ok == false when parsing failed to stop the parser.
// When some non-critical errors occurred, parsed values might be changed to
// correct the error and continue parsing. The non-critical error will be
// handled at the end of the parsing operation.
func (p *parser) parseMessage() (ok bool) {
	p.variableNames = map[string]bool{}
	p.ellipsisCount = 0

	var (
		stream    int
		function  int
		waitBit   int
		direction string
		msgName   string
		dataItem  ast.ItemNode
	)
	stream, function, ok = p.parseStreamFunctionCode()
	if !ok {
		return false
	}

	if t, ok := p.accept(tokenTypeWaitBit); ok {
		if t.val == "W" {
			waitBit = 1
			if function%2 == 0 {
				waitBit = 0
				p.errorf(t, "wait bit cannot be true on reply message (function code is even)")
			}
		} else if t.val == "[W]" {
			waitBit = 2
		}
	}

	if t, ok := p.accept(tokenTypeDirection); ok {
		direction = t.val
	} else {
		p.warningf(t, `missing message direction, "H<->E" will be used`)
		direction = "H<->E"
	}

	if t, ok := p.accept(tokenTypeMessageName); ok {
		msgName = t.val
	}

	dataItem, ok = p.parseMessageText()
	if !ok {
		return false
	}

	if t, ok := p.accept(tokenTypeMessageEnd); !ok {
		p.errorf(t, "expected message end character '.', found %q", t.val)
		return false
	}

	message := ast.NewDataMessage(msgName, stream, function, waitBit, direction, dataItem)
	p.messages = append(p.messages, message)
	return true
}

// parseStreamFunctionCode parses the stream function token.
// Returns ok == false when stream function token isn't found, to stop parsing the message.
// When some non-critical errors occurred, parsed values might be changed to
// correct the error and continue parsing. The non-critical error will be
// handled at the end of the parsing operation.
func (p *parser) parseStreamFunctionCode() (stream, function int, ok bool) {
	t, ok := p.accept(tokenTypeStreamFunction)
	if !ok {
		p.errorf(t, "expected stream function, found %q", t.val)
		return -1, -1, false
	}

	i := strings.Index(t.val, "F")
	stream, _ = strconv.Atoi(t.val[1:i])
	function, _ = strconv.Atoi(t.val[i+1:])
	if !(0 <= stream && stream < 128) {
		p.errorf(t, "stream code range overflow, should be in range of [0, 128)")
		stream = 0
	}
	if !(0 <= function && function < 256) {
		p.errorf(t, "function code range overflow, should be in range of [0, 256)")
		function = 0
	}
	return stream, function, true
}

// parseMessageText parses the message text.
// Returns ok == false when unexpected token is found, to stop parsing the message.
func (p *parser) parseMessageText() (item ast.ItemNode, ok bool) {
	switch t := p.peek(); t.typ {
	case tokenTypeMessageEnd:
		return ast.NewEmptyItemNode(), true
	case tokenTypeLeftAngleBracket:
		return p.parseDataItem()
	default:
		p.errorf(t, "expected '<' or '.', found %q", t.val)
		return ast.NewEmptyItemNode(), false
	}
	// should not reach here
}

// parseDataItem parses a data item.
// Returns ok == false when unexpected token is found, to stop parsing the message.
// When some non-critical errors occurred, parsed values might be changed to
// correct the error and continue parsing. The non-critical error will be
// handled at the end of the parsing operation.
func (p *parser) parseDataItem() (item ast.ItemNode, ok bool) {
	tokenLAB, ok := p.accept(tokenTypeLeftAngleBracket)
	if !ok {
		p.errorf(tokenLAB, "expected '<', found %q", tokenLAB.val)
		return ast.NewEmptyItemNode(), false
	}

	defer func() {
		if r := recover(); r != nil {
			p.errorf(tokenLAB, "%v", r)
			p.warningf(tokenLAB, "Recovered from panic %q. Please submit an issue to handle this error in parser.", r)
			// override return value
			item, ok = ast.NewEmptyItemNode(), false
		}
	}()

	var dataItemType string
	if t, ok := p.accept(tokenTypeDataItemType); ok {
		dataItemType = t.val
	} else {
		p.errorf(t, "invalid data item type: %q", t.val)
		return ast.NewEmptyItemNode(), false
	}

	var tokenDataItemSize token
	var sizeStart, sizeEnd int = 0, -1
	if t := p.peek(); t.typ == tokenTypeDataItemSize {
		tokenDataItemSize, sizeStart, sizeEnd = p.parseDataItemSize()
	} else if t.typ == tokenTypeError {
		p.errorf(t, "syntax error: %s", t.val)
		return ast.NewEmptyItemNode(), false
	}

	switch dataItemType {
	case "L":
		item, ok = p.parseList()
	case "A":
		item, ok = p.parseASCII(sizeStart, sizeEnd)
	case "B":
		item, ok = p.parseBinary()
	case "BOOLEAN":
		item, ok = p.parseBoolean()
	case "F4":
		item, ok = p.parseFloat4()
	case "F8":
		item, ok = p.parseFloat8()
	case "I1":
		item, ok = p.parseInt1()
	case "I2":
		item, ok = p.parseInt2()
	case "I4":
		item, ok = p.parseInt4()
	case "I8":
		item, ok = p.parseInt8()
	case "U1":
		item, ok = p.parseUint1()
	case "U2":
		item, ok = p.parseUint2()
	case "U4":
		item, ok = p.parseUint4()
	case "U8":
		item, ok = p.parseUint8()
	}
	if !ok {
		return ast.NewEmptyItemNode(), false
	}

	if item.Size() >= 0 {
		// (ASCIINode with variable).Size() == -1
		p.checkDataItemSizeError(item.Size(), sizeStart, sizeEnd, tokenDataItemSize)
	}

	if t, ok := p.accept(tokenTypeRightAngleBracket); !ok {
		p.errorf(t, "expected '>', found %q", t.val)
		return ast.NewEmptyItemNode(), false
	}

	return item, ok
}

// parseDataItemSize parses data item size token. Returns lower and upper bound
// of data item size.
func (p *parser) parseDataItemSize() (token, int, int) {
	sizeStart, sizeEnd := 0, -1
	t, ok := p.accept(tokenTypeDataItemSize)
	if ok {
		// possible token value: [x], [x..], [..y], [x..y], where x and y are integers
		i := strings.Index(t.val, "..")
		if i == -1 {
			// '..' is not found
			sizeStart, _ = strconv.Atoi(t.val[1 : len(t.val)-1])
			sizeEnd = sizeStart
		} else {
			sizeStart, _ = strconv.Atoi(t.val[1:i]) // Atoi return 0 when syntax error
			end, err := strconv.Atoi(t.val[i+2 : len(t.val)-1])
			if err == nil || err.(*strconv.NumError).Err == strconv.ErrRange {
				sizeEnd = end
			}
		}
	}
	return t, sizeStart, sizeEnd
}

// checkDataItemSizeError submit error on the token, if the size is not in range.
// upperLimit = -1 means no upper limit.
func (p *parser) checkDataItemSizeError(size, lowerLimit, upperLimit int, t token) {
	if upperLimit == -1 {
		if lowerLimit > size {
			p.errorf(t, "data item size overflow, got size of %d", size)
		}
	} else if !(lowerLimit <= size && size <= upperLimit) {
		p.errorf(t, "data item size overflow, got size of %d", size)
	}
}

// parseList parses a list data item.
// Returns ok == false when unexpected token is found, to stop parsing the message.
// When some non-critical errors occurred, parsed values might be changed to
// correct the error and continue parsing. The non-critical error will be
// handled at the end of the parsing operation.
func (p *parser) parseList() (item ast.ItemNode, ok bool) {
	values := []interface{}{}

	count := 0
	for {
		switch t := p.peek(); t.typ {
		case tokenTypeLeftAngleBracket:
			childItem, ok := p.parseDataItem()
			if !ok {
				return ast.NewEmptyItemNode(), false
			}
			values = append(values, childItem)

		case tokenTypeVariable:
			t = p.acceptAny()
			if _, ok := p.variableNames[t.val]; ok {
				p.errorf(t, "duplicated variable name %q", t.val)
				values = append(values, ast.NewEmptyItemNode())
			} else {
				p.variableNames[t.val] = true
				values = append(values, t.val)
			}

		case tokenTypeEllipsis:
			t = p.acceptAny()
			if count == 0 {
				p.errorf(t, "ellipsis cannot be the first item in list")
				return ast.NewEmptyItemNode(), false
			}
			val := fmt.Sprintf("...[%d]", p.ellipsisCount)
			p.ellipsisCount += 1
			if t.val != "..." && t.val != val {
				p.warningf(t, "wrong ellipsis count, %q will be used", val)
			}
			values = append(values, val)

		case tokenTypeRightAngleBracket:
			return ast.NewListNode(values...), true

		case tokenTypeError:
			p.errorf(t, "syntax error: %s", t.val)
			return ast.NewEmptyItemNode(), false

		default:
			p.errorf(t, "expected child data item, variable, ellipsis, or '>', found %q", t.val)
			return ast.NewEmptyItemNode(), false
		}

		count += 1
	}
	// should not reach here
}

func (p *parser) getDataItemValueTokens() []token {
	tokens := []token{}
	for {
		switch p.peek().typ {
		case tokenTypeNumber, tokenTypeBool, tokenTypeQuotedString, tokenTypeVariable:
			tokens = append(tokens, p.acceptAny())
		case tokenTypeRightAngleBracket:
			return tokens
		default:
			tokens = append(tokens, p.acceptAny())
			return tokens
		}
	}
	// should not reach here
}

// parseASCII parses a ASCII data item.
// Returns ok == false when unexpected token is found, to stop parsing the message.
// When some non-critical errors occurred, parsed values might be changed to
// correct the error and continue parsing. The non-critical error will be
// handled at the end of the parsing operation.
func (p *parser) parseASCII(minLength, maxLength int) (item ast.ItemNode, ok bool) {
	var literal string

	tokens := p.getDataItemValueTokens()
	for _, t := range tokens {
		switch t.typ {
		case tokenTypeQuotedString:
			val, _ := strconv.Unquote(t.val)
			for _, r := range val {
				if r > unicode.MaxASCII {
					val = ""
					p.errorf(t, "expected ASCII characters, found %q", r)
					break
				}
			}
			literal += val

		case tokenTypeNumber:
			val, err := strconv.ParseUint(t.val, 0, 0)
			if err != nil {
				if err.(*strconv.NumError).Err == strconv.ErrSyntax {
					p.errorf(t, "expected ASCII number code, found %q", t.val)
				}
			}
			if val > unicode.MaxASCII {
				val = 0
				p.errorf(t, "overflows ASCII range, found %q", t.val)
			}
			literal += string(byte(val))

		case tokenTypeVariable:
			if len(tokens) != 1 {
				p.errorf(t, "variable cannot co-exist with other literals in ASCII data item")
				return ast.NewEmptyItemNode(), false
			}

			if _, ok := p.variableNames[t.val]; ok {
				p.errorf(t, "duplicated variable name %q", t.val)
				return ast.NewASCIINode(strings.Repeat("*", minLength)), true
			} else {
				p.variableNames[t.val] = true
				return ast.NewASCIINodeVariable(t.val, minLength, maxLength), true
			}

		case tokenTypeError:
			p.errorf(t, "syntax error: %s", t.val)
			return ast.NewEmptyItemNode(), false

		default:
			p.errorf(t, "expected quoted string, ASCII number code or variable, found %q", t.val)
			return ast.NewEmptyItemNode(), false
		}
	}

	return ast.NewASCIINode(literal), true
}

// parseBinary parses a binary data item.
// Returns ok == false when unexpected token is found, to stop parsing the message.
// When some non-critical errors occurred, parsed values might be changed to
// correct the error and continue parsing. The non-critical error will be
// handled at the end of the parsing operation.
func (p *parser) parseBinary() (item ast.ItemNode, ok bool) {
	values := []interface{}{}

	for _, t := range p.getDataItemValueTokens() {
		switch t.typ {
		case tokenTypeNumber:
			val, _ := strconv.ParseInt(t.val, 0, 0)
			if !(0 <= val && val < 256) {
				val = 0
				p.errorf(t, "binary value overflow, should be in range of [0, 256)")
			}
			values = append(values, int(val))

		case tokenTypeVariable:
			if _, ok := p.variableNames[t.val]; ok {
				p.errorf(t, "duplicated variable name %q", t.val)
				values = append(values, 0)
			} else {
				p.variableNames[t.val] = true
				values = append(values, t.val)
			}

		case tokenTypeError:
			p.errorf(t, "syntax error: %s", t.val)
			return ast.NewEmptyItemNode(), false

		default:
			p.errorf(t, "expected number or variable, found %q", t.val)
			return ast.NewEmptyItemNode(), false
		}
	}

	return ast.NewBinaryNode(values...), true
}

// parseBoolean parses a boolean data item.
// Returns ok == false when unexpected token is found, to stop parsing the message.
// When some non-critical errors occurred, parsed values might be changed to
// correct the error and continue parsing. The non-critical error will be
// handled at the end of the parsing operation.
func (p *parser) parseBoolean() (item ast.ItemNode, ok bool) {
	values := []interface{}{}

	for _, t := range p.getDataItemValueTokens() {
		switch t.typ {
		case tokenTypeBool:
			if t.val == "T" {
				values = append(values, true)
			} else {
				values = append(values, false)
			}

		case tokenTypeVariable:
			if _, ok := p.variableNames[t.val]; ok {
				p.errorf(t, "duplicated variable name %q", t.val)
				values = append(values, false)
			} else {
				p.variableNames[t.val] = true
				values = append(values, t.val)
			}

		case tokenTypeError:
			p.errorf(t, "syntax error: %s", t.val)
			return ast.NewEmptyItemNode(), false

		default:
			p.errorf(t, "expected boolean value or variable, found %q", t.val)
			return ast.NewEmptyItemNode(), false
		}
	}

	return ast.NewBooleanNode(values...), true
}

// parseFloat4 parses a F4 data item.
// Returns ok == false when unexpected token is found, to stop parsing the message.
func (p *parser) parseFloat4() (item ast.ItemNode, ok bool) {
	return p.parseFloat(4)
}

// parseFloat8 parses a F8 data item.
// Returns ok == false when unexpected token is found, to stop parsing the message.
func (p *parser) parseFloat8() (item ast.ItemNode, ok bool) {
	return p.parseFloat(8)
}

// parseFloat parses F4 and F8 data items.
// Returns ok == false when unexpected token is found, to stop parsing the message.
// When some non-critical errors occurred, parsed values might be changed to
// correct the error and continue parsing. The non-critical error will be
// handled at the end of the parsing operation.
func (p *parser) parseFloat(byteSize int) (item ast.ItemNode, ok bool) {
	values := []interface{}{}

	for _, t := range p.getDataItemValueTokens() {
		switch t.typ {
		case tokenTypeNumber:
			val, err := strconv.ParseFloat(t.val, byteSize*8)
			if err != nil {
				val = 0
				if err.(*strconv.NumError).Err == strconv.ErrRange {
					p.errorf(t, "F%d range overflow", byteSize)
				} else {
					p.errorf(t, "expected float, found %q", t.val)
				}
			}
			values = append(values, val)

		case tokenTypeVariable:
			if _, ok := p.variableNames[t.val]; ok {
				p.errorf(t, "duplicated variable name %q", t.val)
				values = append(values, 0)
			} else {
				p.variableNames[t.val] = true
				values = append(values, t.val)
			}

		case tokenTypeError:
			p.errorf(t, "syntax error: %s", t.val)
			return ast.NewEmptyItemNode(), false

		default:
			p.errorf(t, "expected float or variable, found %q", t.val)
			return ast.NewEmptyItemNode(), false
		}
	}

	return ast.NewFloatNode(byteSize, values...), true
}

// parseInt1 parses a I1 data item.
// Returns ok == false when unexpected token is found, to stop parsing the message.
func (p *parser) parseInt1() (item ast.ItemNode, ok bool) {
	return p.parseInt(1)
}

// parseInt2 parses a I2 data item.
// Returns ok == false when unexpected token is found, to stop parsing the message.
func (p *parser) parseInt2() (item ast.ItemNode, ok bool) {
	return p.parseInt(2)
}

// parseInt4 parses a I4 data item.
// Returns ok == false when unexpected token is found, to stop parsing the message.
func (p *parser) parseInt4() (item ast.ItemNode, ok bool) {
	return p.parseInt(4)
}

// parseInt8 parses a I8 data item.
// Returns ok == false when unexpected token is found, to stop parsing the message.
func (p *parser) parseInt8() (item ast.ItemNode, ok bool) {
	return p.parseInt(8)
}

// parseInt parses integer data items.
// Returns ok == false when unexpected token is found, to stop parsing the message.
// When some non-critical errors occurred, parsed values might be changed to
// correct the error and continue parsing. The non-critical error will be
// handled at the end of the parsing operation.
func (p *parser) parseInt(byteSize int) (item ast.ItemNode, ok bool) {
	values := []interface{}{}

	for _, t := range p.getDataItemValueTokens() {
		switch t.typ {
		case tokenTypeNumber:
			val, err := strconv.ParseInt(t.val, 0, byteSize*8)
			if err != nil {
				if err.(*strconv.NumError).Err == strconv.ErrRange {
					p.errorf(t, "I%d range overflow", byteSize)
				} else {
					p.errorf(t, "expected integer, found %q", t.val)
				}
			}
			values = append(values, val)

		case tokenTypeVariable:
			if _, ok := p.variableNames[t.val]; ok {
				p.errorf(t, "duplicated variable name %q", t.val)
				values = append(values, 0)
			} else {
				p.variableNames[t.val] = true
				values = append(values, t.val)
			}

		case tokenTypeError:
			p.errorf(t, "syntax error: %s", t.val)
			return ast.NewEmptyItemNode(), false

		default:
			p.errorf(t, "expected integer or variable, found %q", t.val)
			return ast.NewEmptyItemNode(), false
		}
	}

	return ast.NewIntNode(byteSize, values...), true
}

// parseUint1 parses a U1 data item.
// Returns ok == false when unexpected token is found, to stop parsing the message.
func (p *parser) parseUint1() (item ast.ItemNode, ok bool) {
	return p.parseUint(1)
}

// parseUint2 parses a U2 data item.
// Returns ok == false when unexpected token is found, to stop parsing the message.
func (p *parser) parseUint2() (item ast.ItemNode, ok bool) {
	return p.parseUint(2)
}

// parseUint4 parses a U4 data item.
// Returns ok == false when unexpected token is found, to stop parsing the message.
func (p *parser) parseUint4() (item ast.ItemNode, ok bool) {
	return p.parseUint(4)
}

// parseUint8 parses a U8 data item.
// Returns ok == false when unexpected token is found, to stop parsing the message.
func (p *parser) parseUint8() (item ast.ItemNode, ok bool) {
	return p.parseUint(8)
}

// parseInt parses unsigned integer data items.
// Returns ok == false when unexpected token is found, to stop parsing the message.
// When some non-critical errors occurred, parsed values might be changed to
// correct the error and continue parsing. The non-critical error will be
// handled at the end of the parsing operation.
func (p *parser) parseUint(byteSize int) (item ast.ItemNode, ok bool) {
	values := []interface{}{}

	for _, t := range p.getDataItemValueTokens() {
		switch t.typ {
		case tokenTypeNumber:
			val, err := strconv.ParseUint(t.val, 0, byteSize*8)
			if err != nil {
				if err.(*strconv.NumError).Err == strconv.ErrRange {
					p.errorf(t, "U%d range overflow", byteSize)
				} else {
					p.errorf(t, "expected unsigned integer, found %q", t.val)
				}
			}
			values = append(values, val)

		case tokenTypeVariable:
			if _, ok := p.variableNames[t.val]; ok {
				p.errorf(t, "duplicated variable name %q", t.val)
				values = append(values, 0)
			} else {
				p.variableNames[t.val] = true
				values = append(values, t.val)
			}

		case tokenTypeError:
			p.errorf(t, "syntax error: %s", t.val)
			return ast.NewEmptyItemNode(), false

		default:
			p.errorf(t, "expected unsigned integer or variable, found %q", t.val)
			return ast.NewEmptyItemNode(), false
		}
	}

	return ast.NewUintNode(byteSize, values...), true
}
