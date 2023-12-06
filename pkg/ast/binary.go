package ast

import (
	"fmt"
	"strconv"
	"strings"
)

// BinaryNode is a immutable data type that represents a binary item in a SECS-II message.
// Implements ItemNode.
type BinaryNode struct {
	values    []int          // Array of binary values between [0, 255], represented as integers
	variables map[string]int // Variable name and its position in the data array

	// Rep invariants
	// - Each values[i] should be in range of [0, 255]
	// - If a variable exists in position i, values[i] will be zero-value (0) and should not be used.
	// - variable name should adhere to the variable naming rule; refer to interface.go
	// - variable positions should be unique, and be in range of [0, len(values))
}

// Factory methods

// NewBinaryNode creates a new BinaryNode.
//
// Each input argument should have one of following three forms.
// 1. An integer between [0, 255].
// 2. A string with binary format such as "0b1001" between [0, 255].
// 3. A string with a valid variable name as specified in the interface document.
func NewBinaryNode(values ...interface{}) ItemNode {
	if getDataByteLength("binary", len(values)) > MAX_BYTE_SIZE {
		panic("item node size limit exceeded")
	}

	var (
		nodeValues    []int          = make([]int, 0, len(values))
		nodeVariables map[string]int = make(map[string]int)
	)
	for i, value := range values {
		if v, ok := value.(int); ok {
			// value is a int
			nodeValues = append(nodeValues, v)
		} else if v, ok := value.(string); ok {
			if strings.HasPrefix(v, "0b") {
				// value is a binary string
				vAsInt64, _ := strconv.ParseInt(v, 0, 0)
				nodeValues = append(nodeValues, int(vAsInt64))
			} else {
				// value is a variable
				if _, ok := nodeVariables[v]; ok {
					panic("duplicated variable name found")
				}
				nodeVariables[v] = i
				nodeValues = append(nodeValues, 0)
			}
		} else {
			panic("input argument contains invalid type for BinaryNode")
		}
	}

	node := &BinaryNode{nodeValues, nodeVariables}
	node.checkRep()
	return node
}

// Public methods

// Size implements ItemNode.Size().
func (node *BinaryNode) Size() int {
	return len(node.values)
}

// Variables implements ItemNode.Variables().
func (node *BinaryNode) Variables() []string {
	return getVariableNames(node.variables)
}

func (node *BinaryNode) Type() string {
	return "binary"
}

func (node *BinaryNode) Value() []int {
	return node.values
}

// FillVariables implements ItemNode.FillVariables().
func (node *BinaryNode) FillVariables(values map[string]interface{}) ItemNode {
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
	return NewBinaryNode(nodeValues...)
}

// ToBytes implements ItemNode.ToBytes()
func (node *BinaryNode) ToBytes() []byte {
	if len(node.variables) != 0 {
		return []byte{}
	}

	result, err := getHeaderBytes("binary", node.Size())
	if err != nil {
		return []byte{}
	}

	for _, value := range node.values {
		result = append(result, byte(value))
	}

	return result
}

// String returns the string representation of the node.
func (node *BinaryNode) String() string {
	if node.Size() == 0 {
		return "<B[0]>"
	}

	values := make([]string, 0, node.Size())
	for _, value := range node.values {
		str := "0b" + strconv.FormatInt(int64(value), 2)
		values = append(values, str)
	}

	for name, pos := range node.variables {
		values[pos] = name
	}

	return fmt.Sprintf("<B[%d] %v>", node.Size(), strings.Join(values, " "))
}

// Private methods

func (node *BinaryNode) checkRep() {
	for _, v := range node.values {
		if !(0 <= v && v < 256) {
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
