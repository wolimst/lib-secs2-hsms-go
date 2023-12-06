package ast

import (
	"fmt"
	"strings"
	"unicode"
)

// ASCIINode is a immutable data type that represents a ASCII string in a SECS-II message.
// Implements ItemNode.
//
// It contains either a string that consists of ASCII characters,
// or a variable which can be used to fill the string value later.
// ASCII data type is one of the special cases in the SECS-II data types;
// the size of ASCII data type is the length of the string, and therefore,
// there could be only one variable if exist.
type ASCIINode struct {
	value    string            // a string literal that consists of ASCII characters
	variable asciiNodeVariable // a struct that contains information on the variable
	isValue  bool              // a flag that represents which data is set; value or variable

	// Rep invariants
	// - If isValue == true, variable shouldn't be used and it should have zero-value
	//   else, value shouldn't be used and it should have zero-value
	// - value should consist of ASCII characters
	// - variable.name should adhere to the variable naming rule; refer to interface.go
	// - variable.minLength >= 0, variable.maxLength >= -1
	// - variable.minLength <= variable.maxLength, when variable.maxLength != -1
}

type asciiNodeVariable struct {
	name      string // variable name
	minLength int    // minimum length of the string value to be filled; -1 means no limit
	maxLength int    // maximum length of the string value to be filled; -1 means no limit
}

// Factory methods

// NewASCIINode creates a new ASCIINode that contains the input string.
//
// The input string should consist of ASCII chracters.
func NewASCIINode(str string) ItemNode {
	if getDataByteLength("ascii", len(str)) > MAX_BYTE_SIZE {
		panic("string length limit exceeded")
	}

	node := &ASCIINode{value: str, isValue: true}
	node.checkRep()
	return node
}

// NewASCIINodeVariable creates a new ASCIINode that contains a variable.
//
// name should be a valid variable name as specified in the interface documentation.
// minLength and maxLength represents the length range of the string value to be filled.
//
// minLength and maxLength should meet following conditions.
// minLength >= 0, maxLength >= -1, where -1 means no limit.
// minLength <= maxLength, when maxLength != -1.
func NewASCIINodeVariable(name string, minLength, maxLength int) ItemNode {
	node := &ASCIINode{
		variable: asciiNodeVariable{name, minLength, maxLength},
		isValue:  false,
	}
	node.checkRep()
	return node
}

// Public methods

// Size implements DataItemNode.Size().
//
// If the node have a variable, returns -1.
func (node *ASCIINode) Size() int {
	if !node.isValue {
		return -1
	}
	return len(node.value)
}

func (node *ASCIINode) Type() string{
	return "ascii"
}

func (node *ASCIINode) Value() string {
	return node.value
}

// FillInStringLength returns the minimun and the maximum string length that can be
// filled into the variable of this ASCIINode.
//
// Return value of -1 means no limit.
// If the node doesn't have variable, it will return (-2, -2).
func (node *ASCIINode) FillInStringLength() (min int, max int) {
	if node.isValue {
		return -2, -2
	}
	return node.variable.minLength, node.variable.maxLength
}

// Variables implements DataItemNode.Variables().
func (node *ASCIINode) Variables() []string {
	if node.isValue {
		return []string{}
	}
	return []string{node.variable.name}
}

// FillVariables implements ItemNode.FillVariables().
//
// The fill-in value must be acceptable by the NewASCIINode factory method, and
// it should be in range of the fill-in string length.
func (node *ASCIINode) FillVariables(values map[string]interface{}) ItemNode {
	if node.isValue {
		return node
	}

	if _, ok := values[node.variable.name]; !ok {
		return node
	}

	value, ok := values[node.variable.name].(string)
	if !ok {
		panic("fill-in value has invalid type for ASCIINode")
	}

	if len(value) < node.variable.minLength {
		panic("fill-in string length overflow")
	}

	if node.variable.maxLength != -1 && node.variable.maxLength < len(value) {
		panic("fill-in string length overflow")
	}

	return NewASCIINode(value)
}

// ToBytes implements ItemNode.ToBytes()
func (node *ASCIINode) ToBytes() []byte {
	if !node.isValue {
		return []byte{}
	}

	result, err := getHeaderBytes("ascii", node.Size())
	if err != nil {
		return []byte{}
	}

	for _, ch := range node.value {
		result = append(result, byte(ch))
	}

	return result
}

// String returns the string representation of the node.
func (node *ASCIINode) String() string {
	if !node.isValue {
		var lengthStr string
		min, max := node.variable.minLength, node.variable.maxLength

		if min == 0 && max == -1 {
			// empty lengthStr
		} else if min == max {
			lengthStr = fmt.Sprintf("[%d]", max)
		} else if max == -1 {
			lengthStr = fmt.Sprintf("[%d..]", min)
		} else {
			lengthStr = fmt.Sprintf("[%d..%d]", min, max)
		}
		return fmt.Sprintf("<A%s %s>", lengthStr, node.variable.name)
	}

	if node.value == "" {
		return "<A[0]>"
	}

	var sb strings.Builder
	printableState := false
	for _, ch := range node.value {
		if ch < 32 || ch == 127 {
			// ch is a non-printable control character
			// 32: space, which is the first printable character, 127: del
			if printableState {
				printableState = false
				sb.WriteString(`"`) // Close double quote
			}
			fmt.Fprintf(&sb, " 0x%02X", ch) // 0xNN format
		} else {
			// c is a printable character
			if !printableState {
				printableState = true
				sb.WriteString(` "`) // Open double quote
			}
			sb.WriteRune(ch)
		}
	}
	// Close the double quote if in printable state
	if printableState {
		sb.WriteString(`"`)
	}

	return fmt.Sprintf(`<A%s>`, sb.String())
}

// Private methods

func (node *ASCIINode) checkRep() {
	if node.isValue {
		if node.variable.name != "" || node.variable.minLength != 0 || node.variable.maxLength != 0 {
			panic("value and variable should not be used at the same time")
		}

		for _, ch := range node.value {
			if ch > unicode.MaxASCII {
				panic("encountered non-ASCII character")
			}
		}
	} else {
		if node.value != "" {
			panic("value and variable should not be used at the same time")
		}

		if !isValidVarName(node.variable.name) {
			panic("invalid variable name")
		}

		if node.variable.minLength < 0 || node.variable.maxLength < -1 {
			panic("invalid fill-in string length")
		}

		if node.variable.maxLength != -1 {
			if node.variable.minLength > node.variable.maxLength {
				panic("invalid fill-in string length")
			}
		}
	}
}
