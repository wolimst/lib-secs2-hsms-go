package ast

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// FloatNode is a immutable data type that represents a float in a SECS-II message.
// Implements ItemNode.
//
// Infinity and NaN are not supported.
//
// String representation of the float values will use the golang's %g formatting.
// Refer to the documentation of the fmt package (https://golang.org/pkg/fmt/).
type FloatNode struct {
	byteSize  int            // Byte size of the floats; should be either 4 or 8
	values    []float64      // Array of floats
	variables map[string]int // Variable name and its position in the data array

	// Rep invariants
	// - Each values[i] should be representable in bytes of byteSize
	// - math.IsInf(values[i], 0) == false && math.IsNaN(values[i]) == false
	// - If a variable exists in position i, values[i] will be zero-value (0) and should not be used
	// - variable name should adhere to the variable naming rule; refer to interface.go
	// - variable positions should be unique, and be in range of [0, len(values))
}

// Factory methods

// NewFloatNode creates a new FloatNode that contains float data.
//
// The byteSize should be either 4 or 8.
// Each input of the values should be a float that could be represented within bytes of the byteSize,
// or a string with a valid variable name as specified in the interface documentation.
func NewFloatNode(byteSize int, values ...interface{}) ItemNode {
	if getDataByteLength(fmt.Sprintf("f%d", byteSize), len(values)) > MAX_BYTE_SIZE {
		panic("item node size limit exceeded")
	}

	var (
		nodeValues    []float64      = make([]float64, 0, len(values))
		nodeVariables map[string]int = make(map[string]int)
	)

	for i, value := range values {
		switch value := value.(type) {
		case int:
			nodeValues = append(nodeValues, float64(value))
		case int8:
			nodeValues = append(nodeValues, float64(value))
		case int16:
			nodeValues = append(nodeValues, float64(value))
		case int32:
			nodeValues = append(nodeValues, float64(value))
		case int64:
			nodeValues = append(nodeValues, float64(value))
		case uint:
			nodeValues = append(nodeValues, float64(value))
		case uint8:
			nodeValues = append(nodeValues, float64(value))
		case uint16:
			nodeValues = append(nodeValues, float64(value))
		case uint32:
			nodeValues = append(nodeValues, float64(value))
		case uint64:
			nodeValues = append(nodeValues, float64(value))
		case float32:
			nodeValues = append(nodeValues, float64(value))
		case float64:
			nodeValues = append(nodeValues, value)
		case string:
			if _, ok := nodeVariables[value]; ok {
				panic("duplicated variable name found")
			}
			nodeVariables[value] = i
			nodeValues = append(nodeValues, 0)
		default:
			panic("input argument contains invalid type for FloatNode")
		}
	}

	node := &FloatNode{byteSize, nodeValues, nodeVariables}
	node.checkRep()
	return node
}

// Public methods

// Size implements ItemNode.Size().
func (node *FloatNode) Size() int {
	return len(node.values)
}

func (node *FloatNode) Type() string {
	return "float"
}

func (node *FloatNode) Value() []float64 {
	return node.values
}

// Variables implements ItemNode.Variables().
func (node *FloatNode) Variables() []string {
	return getVariableNames(node.variables)
}

// FillVariables implements ItemNode.FillVariables().
func (node *FloatNode) FillVariables(values map[string]interface{}) ItemNode {
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
	return NewFloatNode(node.byteSize, nodeValues...)
}

// ToBytes implements ItemNode.ToBytes()
func (node *FloatNode) ToBytes() []byte {
	if len(node.variables) != 0 {
		return []byte{}
	}

	result, err := getHeaderBytes(fmt.Sprintf("f%d", node.byteSize), node.Size())
	if err != nil {
		return []byte{}
	}

	if node.byteSize == 4 {
		for _, value := range node.values {
			bits := math.Float32bits(float32(value))
			result = append(result, byte(bits>>24))
			result = append(result, byte(bits>>16))
			result = append(result, byte(bits>>8))
			result = append(result, byte(bits))
		}
	} else {
		for _, value := range node.values {
			bits := math.Float64bits(value)
			result = append(result, byte(bits>>56))
			result = append(result, byte(bits>>48))
			result = append(result, byte(bits>>40))
			result = append(result, byte(bits>>32))
			result = append(result, byte(bits>>24))
			result = append(result, byte(bits>>16))
			result = append(result, byte(bits>>8))
			result = append(result, byte(bits))
		}
	}

	return result
}

// String returns the string representation of the node.
//
// The float values will be represented by the golang's %g formatting.
func (node *FloatNode) String() string {
	if node.Size() == 0 {
		return fmt.Sprintf("<F%d[0]>", node.byteSize)
	}

	values := make([]string, 0, node.Size())
	for _, v := range node.values {
		values = append(values, strconv.FormatFloat(v, 'g', -1, node.byteSize*8))
	}

	for name, pos := range node.variables {
		values[pos] = name
	}

	return fmt.Sprintf("<F%d[%d] %v>", node.byteSize, node.Size(), strings.Join(values, " "))
}

// Private methods

func (node *FloatNode) checkRep() {
	if node.byteSize != 4 && node.byteSize != 8 {
		panic("invalid byte size")
	}

	max := math.MaxFloat64
	if node.byteSize == 4 {
		max = math.MaxFloat32
	}
	for _, v := range node.values {
		if math.IsInf(v, 0) || math.IsNaN(v) {
			panic("invalid value")
		}

		if !(-max <= v && v <= max) {
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
