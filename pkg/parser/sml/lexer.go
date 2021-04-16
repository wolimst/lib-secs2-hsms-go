package sml

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// The design of this lexer is based on "Lexical Scanning in Go" by Rob Pike
// https://talks.golang.org/2011/lex.slide

// token represents a tokenized text string that a lexer identified.
type token struct {
	typ  tokenType // token type
	val  string    // tokenized text
	line int       // tokenized text's line number
	col  int       // tokenized text's column number
}

type tokenType int

const (
	tokenTypeEOF        tokenType = iota // EOF
	tokenTypeError                       // lexing error
	tokenTypeComment                     // starting with '//', ending with newline or EOF
	tokenTypeMessageEnd                  // '.'

	// Message header
	tokenTypeStreamFunction // 'S' [0-9]+ 'F' [0-9]+, case insensitive
	tokenTypeWaitBit        // 'W', '[W]', case insensitive
	tokenTypeDirection      // 'H->E', 'H<-E', 'H<->E', case insensitive
	tokenTypeMessageName    // Series of characters except whitespaces and comment delimiter

	// Message text
	tokenTypeLeftAngleBracket  // '<'
	tokenTypeRightAngleBracket // '>'
	tokenTypeDataItemType      // 'L', 'B', 'BOOLEAN', 'A', 'F4', 'F8', 'I1', 'I2', 'I4', 'I8', 'U1', 'U2', 'U4', 'U8', case insensitive
	tokenTypeDataItemSize      // '[' [0-9]+ ('..' [0-9]+)? ']'
	tokenTypeNumber            // decimal, hexadecimal, octal, binary, floating-point number including scientific notation, case insensitive
	tokenTypeBool              // 'T', 'F', case insensitive
	tokenTypeVariable          // [A-Za-z_] [A-Za-z0-9_]* ('[' [0-9]+ ']')?
	tokenTypeQuotedString      // string enclosed with double quotes, e.g. "quoted string"
	tokenTypeEllipsis          // '...'
)

// lexer represents the state of the lexical scanner.
type lexer struct {
	input     string     // input string being lexed
	lastState stateFn    // last lexing state function
	state     stateFn    // next lexing state function to enter
	pos       int        // current position in the input
	start     int        // start position of a token being lexed in input string
	width     int        // width of last rune read from input
	tokens    chan token // the channel to report scanned tokens
}

const eof rune = -1

// lex creates a new scanner for the input string.
func lex(input string) *lexer {
	l := &lexer{
		input:  input,
		state:  lexMessageHeader,
		tokens: make(chan token, 2),
	}
	return l
}

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune. Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// emit passes a token to the client.
func (l *lexer) emit(t tokenType) {
	line, col := l.lineColumn()
	l.tokens <- token{typ: t, val: l.input[l.start:l.pos], line: line, col: col}
	l.start = l.pos
}

// emitUppercase passes a token with a uppercase token value to the client.
func (l *lexer) emitUppercase(t tokenType) {
	line, col := l.lineColumn()
	l.tokens <- token{typ: t, val: strings.ToUpper(l.input[l.start:l.pos]), line: line, col: col}
	l.start = l.pos
}

// emitSpaceRemoved passes a token to the client, with all spaces in token value removed.
func (l *lexer) emitSpaceRemoved(t tokenType) {
	line, col := l.lineColumn()
	val := make([]rune, 0, l.pos-l.start)
	for _, r := range l.input[l.start:l.pos] {
		if !unicode.IsSpace(r) {
			val = append(val, r)
		}
	}
	l.tokens <- token{typ: t, val: string(val), line: line, col: col}
	l.start = l.pos
}

// emitEOF passes a EOF token to the client.
func (l *lexer) emitEOF() {
	line, col := l.lineColumn()
	l.tokens <- token{typ: tokenTypeEOF, val: "EOF", line: line, col: col}
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// lineColumn returns line and column number of current start position.
func (l *lexer) lineColumn() (line, column int) {
	// Doing it this way means we don't have to worry about peek double counting
	line = 1 + strings.Count(l.input[:l.start], "\n")
	lineStart := 1 + strings.LastIndex(l.input[:l.start], "\n")
	column = 1 + utf8.RuneCountInString(l.input[lineStart:l.start])
	return line, column
}

// errorf returns an error token and terminates the running lexer.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	line, col := l.lineColumn()
	l.tokens <- token{tokenTypeError, fmt.Sprintf(format, args...), line, col}
	return l.terminate()
}

// nextToken returns the next token from the input.
// If lexer.tokens channel is closed, it will return EOF token.
func (l *lexer) nextToken() token {
	for {
		select {
		case tok, ok := <-l.tokens:
			if !ok {
				return token{typ: tokenTypeEOF}
			}
			return tok
		default:
			l.lastState, l.state = l.state, l.state(l)
		}
	}
	// should not reach here
}

// stateFn represents the state of the lexer as a function that returns the next state
type stateFn func(*lexer) stateFn

// terminate closes the l.tokens channel and terminates the scan by
// passing back a nil pointer as the next state function.
func (l *lexer) terminate() stateFn {
	close(l.tokens)
	return nil
}

// lexMessageHeader scans the elements that can appear in the message header.
func lexMessageHeader(l *lexer) stateFn {
	for {
		// Handle a line comment
		if strings.HasPrefix(l.input[l.pos:], "//") {
			return lexComment
		}

		// Handle stream function code
		re := regexp.MustCompile(`^[Ss]\d+[Ff]\d+`)
		if loc := re.FindStringIndex(l.input[l.pos:]); loc != nil {
			l.pos += loc[1]
			l.emitUppercase(tokenTypeStreamFunction)
			return lexMessageHeader
		}

		// Handle wait bit
		re = regexp.MustCompile(`^([Ww]|\[[Ww]\])`)
		if loc := re.FindStringIndex(l.input[l.pos:]); loc != nil {
			l.pos += loc[1]
			l.emitUppercase(tokenTypeWaitBit)
			return lexMessageHeader
		}

		// Handle message direction
		re = regexp.MustCompile(`^[Hh](->|<->|<-)[Ee]`)
		if loc := re.FindStringIndex(l.input[l.pos:]); loc != nil {
			l.pos += loc[1]
			l.emitUppercase(tokenTypeDirection)
			return lexMessageHeader
		}

		switch r := l.next(); r {
		case eof:
			return lexEOF
		case ' ', '\t', '\r', '\n':
			l.ignore()
		case '.':
			l.emit(tokenTypeMessageEnd)
			return lexMessageHeader
		case '<':
			l.emit(tokenTypeLeftAngleBracket)
			return lexMessageText
		default:
			for {
				r := l.next()
				if r == eof || unicode.IsSpace(r) || strings.HasPrefix(l.input[l.pos-1:], "//") {
					l.backup()
					break
				}
			}
			l.emit(tokenTypeMessageName)
			return lexMessageHeader
		}
	}
	// should not reach here
}

// lexMessageText scans the elements inside the message text.
func lexMessageText(l *lexer) stateFn {
	for {
		// Handle a line comment
		if strings.HasPrefix(l.input[l.pos:], "//") {
			return lexComment
		}

		re := regexp.MustCompile(`^\.\.\.(\[\d+\])?`)
		if loc := re.FindStringIndex(l.input[l.pos:]); loc != nil {
			l.pos += loc[1]
			l.emit(tokenTypeEllipsis)
			return lexMessageText
		}

		// Handle data types or variables
		re = regexp.MustCompile(`^[A-Za-z_]\w*`)
		if loc := re.FindStringIndex(l.input[l.pos:]); loc != nil {
			switch strings.ToUpper(l.input[l.pos : l.pos+loc[1]]) {
			case "L", "A", "B", "BOOLEAN", "F4", "F8",
				"I1", "I2", "I4", "I8", "U1", "U2", "U4", "U8":
				l.pos += loc[1]
				l.emitUppercase(tokenTypeDataItemType)
				return lexMessageText
			case "T", "F":
				l.pos += loc[1]
				l.emitUppercase(tokenTypeBool)
				return lexMessageText
			default:
				l.pos += loc[1]
				// Handle optional array-like notation
				re = regexp.MustCompile(`^(\[\d+\])+`)
				if loc = re.FindStringIndex(l.input[l.pos:]); loc != nil {
					l.pos += loc[1]
				}
				l.emit(tokenTypeVariable)
				return lexMessageText
			}
		}

		r := l.next()
		// Handle number
		if r == '+' || r == '-' || isDigit(r) || (r == '.' && isDigit(l.peek())) {
			l.backup()
			return lexNumber
		}

		switch r {
		case eof:
			return lexEOF
		case '<':
			l.emit(tokenTypeLeftAngleBracket)
			return lexMessageText
		case '>':
			l.emit(tokenTypeRightAngleBracket)
			return lexMessageText
		case '.':
			l.emit(tokenTypeMessageEnd)
			return lexMessageHeader
		case '[':
			l.backup()
			return lexDataItemSize
		case '"':
			l.backup()
			return lexQuotedString
		case ' ', '\t', '\r', '\n':
			l.ignore()
		default:
			return l.errorf("unexpected character in data item: %#U", r)
		}
	}
	// should not reach here
}

// lexEOF scans a EOF which is known to be present, and terminates the running lexer.
func lexEOF(l *lexer) stateFn {
	l.emitEOF()
	return l.terminate()
}

// lexComment scans a line comment.
// The line comment delimiter "//" is known to be present.
// Returns the previous state function which called lexComment.
func lexComment(l *lexer) stateFn {
	i := strings.Index(l.input[l.pos:], "\n")
	if i < 0 {
		l.pos = len(l.input)
		l.emit(tokenTypeComment)
		return lexEOF
	}

	for unicode.IsSpace(rune(l.input[l.pos+i-1])) {
		i -= 1
	}
	l.pos += i
	l.emit(tokenTypeComment)
	return l.lastState
}

// lexDataItemSize scans a data item's size, e.g. [2] or [2..7].
// The left square bracket is known to be present.
func lexDataItemSize(l *lexer) stateFn {
	numberFound := false
	l.accept("[")
	l.acceptRun(" \t\r\n")
	if l.accept("0123456789") {
		numberFound = true
		l.acceptRun("0123456789")
		l.acceptRun(" \t\r\n")
	}
	if strings.HasPrefix(l.input[l.pos:], "..") {
		l.pos += 2
		l.acceptRun(" \t\r\n")
		if l.accept("0123456789") {
			numberFound = true
			l.acceptRun("0123456789")
			l.acceptRun(" \t\r\n")
		}
	}
	if !(l.accept("]") && numberFound) {
		return l.errorf("invalid data item size")
	}
	l.emitSpaceRemoved(tokenTypeDataItemSize)
	return lexMessageText
}

// lexQuotedString scans a string inside double quotes.
// The left double quote is known to be present.
func lexQuotedString(l *lexer) stateFn {
	l.accept(`"`)
	i := strings.Index(l.input[l.pos:], `"`)
	j := strings.IndexAny(l.input[l.pos:], "\r\n")
	if i < 0 || (j > 0 && j < i) {
		return l.errorf("unclosed quoted string")
	}
	l.pos += i + 1 // Include the double quote
	l.emit(tokenTypeQuotedString)
	return lexMessageText
}

// lexNumber scans a number, which is known to be present.
func lexNumber(l *lexer) stateFn {
	// Optional number sign
	l.accept("+-")

	// Handle decimal, hexadecimal, binary number
	digits := "0123456789" // default is decimal
	if l.accept("0") {
		if l.accept("xX") {
			digits = "0123456789abcdefABCDEF"
		} else if l.accept("bB") {
			digits = "01"
		} else if l.accept("oO") {
			digits = "01234567"
		}
	}
	l.acceptRun(digits)

	// Handle floating-point number
	if l.accept(".") {
		l.acceptRun(digits)
	}

	// Handle scientific notation
	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}

	// Next thing must not be alphanumeric
	if isAlphaNumeric(l.peek()) {
		l.next()
		return l.errorf("invalid number syntax: %q", l.input[l.start:l.pos])
	}

	l.emit(tokenTypeNumber)
	return lexMessageText
}

// Helper functions

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isDigit reports whether r is a digit.
func isDigit(r rune) bool {
	return ('0' <= r && r <= '9')
}
