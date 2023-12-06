package ast

import (
	"fmt"
	"strconv"
	"strings"
)

// UintNode is a immutable data type that represents an unsigned integer in a SECS-II message.
// Implements ItemNode.
type UintNode struct {
	byteSize  int            // Byte size of the unsigned integers; should be either 1, 2, 4, or 8
	values    []uint64       // Array of unsigned integers
	variables map[string]int // Variable name and its position in the data array

	// Rep invariants
	// - Each values[i] should be in range of [0, max], where max = 1<<(byteSize*8)-1
	// - If a variable exists in position i, values[i] will be zero-value (0) and should not be used.
	// - variable name should adhere to the variable naming rule; refer to interface.go
	// - variable positions should be unique, and be in range of [0, len(values))
}

// Factory methods

// NewUintNode creates a new UintNode that contains unsigned integer data.
//
// The byteSize should be either 1, 2, 4, or 8.
// Each input of the values should be an unsigned integer that could be represented within bytes of the byteSize,
// or it should be a string with a valid variable name as specified in the interface documentation.
func NewUintNode(byteSize int, values ...interface{}) ItemNode {
	if getDataByteLength(fmt.Sprintf("u%d", byteSize), len(values)) > MAX_BYTE_SIZE {
		panic("item node size limit exceeded")
	}

	var (
		nodeValues    []uint64       = make([]uint64, 0, len(values))
		nodeVariables map[string]int = make(map[string]int)
	)

	for i, value := range values {
		switch value := value.(type) {
		case int:
			nodeValues = append(nodeValues, uint64(value))
		case int8:
			nodeValues = append(nodeValues, uint64(value))
		case int16:
			nodeValues = append(nodeValues, uint64(value))
		case int32:
			nodeValues = append(nodeValues, uint64(value))
		case int64:
			nodeValues = append(nodeValues, uint64(value))
		case uint:
			nodeValues = append(nodeValues, uint64(value))
		case uint8:
			nodeValues = append(nodeValues, uint64(value))
		case uint16:
			nodeValues = append(nodeValues, uint64(value))
		case uint32:
			nodeValues = append(nodeValues, uint64(value))
		case uint64:
			nodeValues = append(nodeValues, value)
		case string:
			if _, ok := nodeVariables[value]; ok {
				panic("duplicated variable name found")
			}
			nodeVariables[value] = i
			nodeValues = append(nodeValues, 0)
		default:
			panic("input argument contains invalid type for UintNode")
		}
	}

	node := &UintNode{byteSize, nodeValues, nodeVariables}
	node.checkRep()
	return node
}

// Public methods

// Size implements ItemNode.Size().
func (node *UintNode) Size() int {
	return len(node.values)
}

func (node *UintNode) Type() string {
	return "uint"
}

func (node *UintNode) Value() []uint64 {
	return node.values
}

// Variables implements ItemNode.Variables().
func (node *UintNode) Variables() []string {
	return getVariableNames(node.variables)
}

// FillVariables implements ItemNode.FillVariables().
func (node *UintNode) FillVariables(values map[string]interface{}) ItemNode {
	if len(node.variables) == 0 {
		return node
	}

	nodeValues := make([]interface{}, 0, node.Size())
	for _, v := range node.values {
		nodeValues = append(nodeValues, v)
	}

	createNew := false
	for name, pos := range node.variables {
		if v, ok := values[name]; ok {
			nodeValues[pos] = v
			createNew = true
		} else {
			nodeValues[pos] = name
		}
	}

	if !createNew {
		return node
	}
	return NewUintNode(node.byteSize, nodeValues...)
}

// ToBytes implements ItemNode.ToBytes()
func (node *UintNode) ToBytes() []byte {
	if len(node.variables) != 0 {
		return []byte{}
	}

	result, err := getHeaderBytes(fmt.Sprintf("u%d", node.byteSize), node.Size())
	if err != nil {
		return []byte{}
	}

	for _, value := range node.values {
		for i := node.byteSize - 1; i >= 0; i-- {
			result = append(result, byte(value>>(i*8)))
		}
	}

	return result
}

// String returns the string representation of the node.
func (node *UintNode) String() string {
	if node.Size() == 0 {
		return fmt.Sprintf("<U%d[0]>", node.byteSize)
	}

	values := make([]string, 0, node.Size())
	for _, v := range node.values {
		values = append(values, strconv.FormatUint(v, 10))
	}

	for name, pos := range node.variables {
		values[pos] = name
	}

	return fmt.Sprintf("<U%d[%d] %v>", node.byteSize, node.Size(), strings.Join(values, " "))
}

// Private methods

func (node *UintNode) checkRep() {
	if node.byteSize != 1 && node.byteSize != 2 &&
		node.byteSize != 4 && node.byteSize != 8 {
		panic("invalid byte size")
	}

	for _, v := range node.values {
		if !(v <= uint64(1<<(node.byteSize*8)-1)) {
			panic("value overflow")
		}
	}

	visited := map[int]bool{}
	for name, pos := range node.variables {
		if node.values[pos] != 0 {
			panic("value in variable position isn't a zero-value")
		}

		if !isValidVarName(name) {
			panic("invalid variable name")
		}

		if _, ok := visited[pos]; ok {
			panic("variable position is not unique")
		}
		visited[pos] = true

		if !(0 <= pos && pos < node.Size()) {
			panic("variable position overflow")
		}
	}
}
