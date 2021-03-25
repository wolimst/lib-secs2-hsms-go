package ast

import (
	"fmt"
	"regexp"
	"sort"
)

const MAX_BYTE_SIZE = 1<<24 - 1

// ItemNode is a interface of immutable data types, that represents a data item in a SECS-II message.
// It contains an array consists of data values or variables which can be used to fill the data values later.
// E.g., A boolean node should be able to represent a SECS-II data item of <BOOLEAN[3] T F varName>.
//
// A variable name could contain a string with alphanumerics and the underbar.
// Also, ellipsis literal "..." is a special variable name that can be only used in ListNode,
// which means repetition of elements in the ListNode before it.
// A number cannot be the first letter of the variable name.
// Each variable name in a data item node should be unique.
//
// A variable name also could contain characters that is like array index accessor
// in most programming languages, at the end of its name, e.g. varName[0], varName[1], etc.
// However, it doesn't mean that there's underlying array; in case of the example above,
// varName[0] it self is the name of the variable, and varName would not exist when not specified.
// It might be a source of some confusion, therefore, it is recommended to use
// variable names with this array-like notation, only in ListNodes, when there are repeating elements.
// ListNode containing variables with array-like notation would be created,
// when filling values into a ListNode containing variables and ellipsis,
// e.g. <L <A varName> ...> can be <L[2] <A varName[0]> <A varName[1]> >,
// when the ellipsis is filled in with 1 (1 repetition).
// For more detailed information, refer to the documentation of ListNode.
//
// There is a limit on the number of data values that a ItemNode can contain,
// as specified in the SEMI Standard.
// The limit is expressed as following equation; n * b <= 16,777,215 (3 bytes),
// where n is the number of the data values in a ItemNode, and b is bytes to
// represent a data value which is different for each ItemNode type.
type ItemNode interface {
	// Size returns the array size of the data item.
	Size() int

	// Variables returns the variable names in the node, in the insertion order.
	Variables() []string

	// FillVariables returns a new ItemNode with the specified values filled into the variables.
	// The map input argument has variable name as its key, and fill-in value as its value.
	// Each fill-in value must be acceptable by the ItemNode's factory method.
	// If a variable in the ItemNode doesn't exist in the input map, the variable will remain unchanged.
	FillVariables(map[string]interface{}) ItemNode

	// ToBytes returns the byte representation of the data item.
	ToBytes() []byte
}

// EmptyItemNode is a immutable data type that represents a empty data item node.
// It will be used mostly on error cases.
type emptyItemNode struct{}

// NewEmptyItemNode creates a new empty data item node.
func NewEmptyItemNode() ItemNode {
	return emptyItemNode{}
}

// Size implements ItemNode.Size().
func (node emptyItemNode) Size() int {
	return 0
}

// Variables implements ItemNode.Variables().
func (node emptyItemNode) Variables() []string {
	return []string{}
}

// FillVariables implements ItemNode.FillVariables().
func (node emptyItemNode) FillVariables(values map[string]interface{}) ItemNode {
	return node
}

// ToBytes implements ItemNode.ToBytes()
func (node emptyItemNode) ToBytes() []byte {
	return []byte{}
}

// String returns the string representation of the node.
func (node emptyItemNode) String() string {
	return ""
}

// Helper functions

// isValidVarName checks that the variable name is valid as specified in the interface document.
func isValidVarName(name string) bool {
	re := regexp.MustCompile(`^[A-Za-z_]\w*(\[\d+\])*$`)
	return re.MatchString(name)
}

// isEllipsis checks whether a variable is ellipsis or not.
func isEllipsis(name string) bool {
	re := regexp.MustCompile(`^\.{3}(\[\d+\])?$`)
	return re.MatchString(name)
}

// getVariableNames returns variable names sorted by their positions.
// The input argument's key is a variable name and its value is the variable's position.
func getVariableNames(variablePosition map[string]int) []string {
	result := make([]string, 0, len(variablePosition))
	for name := range variablePosition {
		result = append(result, name)
	}
	sort.Slice(result, func(i, j int) bool {
		return variablePosition[result[i]] < variablePosition[result[j]]
	})
	return result
}

// getDataByteLength returns the number of bytes to represent a data with
// specified type and size.
//
// The input argument typ should be one of "list", "binary", "boolean", "ascii",
// "i8", "i1", "i2", "i4", "f8", "f4", "u8", "u1", "u2", or "u4".
// The input argument size means the number of values in a item node.
func getDataByteLength(typ string, size int) int {
	bytePerValue := map[string]int{
		"list":    1,
		"binary":  1,
		"boolean": 1,
		"ascii":   1,
		"i8":      8,
		"i1":      1,
		"i2":      2,
		"i4":      4,
		"f8":      8,
		"f4":      4,
		"u8":      8,
		"u1":      1,
		"u2":      2,
		"u4":      4,
	}
	return size * bytePerValue[typ]
}

// getHeaderBytes returns the header bytes, which consist of the format byte
// and the length bytes, of a SECS-II data item.
//
// The input argument typ should be one of "list", "binary", "boolean", "ascii",
// "i8", "i1", "i2", "i4", "f8", "f4", "u8", "u1", "u2", or "u4".
// The input argument size means the number of values in a item node.
// An error is returned when the header bytes cannot be created.
func getHeaderBytes(typ string, size int) ([]byte, error) {
	formatCode := map[string]int{
		"list":    0o00,
		"binary":  0o10,
		"boolean": 0o11,
		"ascii":   0o20,
		"i8":      0o30,
		"i1":      0o31,
		"i2":      0o32,
		"i4":      0o34,
		"f8":      0o40,
		"f4":      0o44,
		"u8":      0o50,
		"u1":      0o51,
		"u2":      0o52,
		"u4":      0o54,
	}

	dataByteLength := getDataByteLength(typ, size)
	if dataByteLength > MAX_BYTE_SIZE {
		return []byte{}, fmt.Errorf("size limit exceeded")
	}

	lengthBytes := []byte{
		byte(dataByteLength >> 16),
		byte(dataByteLength >> 8),
		byte(dataByteLength),
	}

	if lengthBytes[0] == 0 {
		if lengthBytes[1] == 0 {
			lengthBytes = lengthBytes[2:]
		} else {
			lengthBytes = lengthBytes[1:]
		}
	}

	result := []byte{}
	result = append(result, byte(formatCode[typ]<<2+len(lengthBytes)))
	result = append(result, lengthBytes...)
	return result, nil
}
