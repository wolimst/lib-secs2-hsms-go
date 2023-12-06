package ast

import (
	"fmt"
	"strings"
)

// ListNode is a immutable data type that represents a list data in a SECS-II message.
// Implements ItemNode.
//
// It contains other item nodes, and the size of ListNode is equal to the number
// of items it contains, counted *non-recursively*.
//
// A ListNode can contain a special variable, ellipsis, represented as three dots "...".
// An ellipsis means that the item nodes before it can be repeated arbitrary times.
// Each ListNode can contain one ellipsis at most, and the ellipsis should not be the first item
// of the ListNode.
//
// When filling in values into variables, the ellipsis variables will be filled in at first,
// over non-ellipsis variables.
// For nested ListNodes containing multiple ellipsis, they will be filled in appearing order
// on the top ListNode's string representation.
//
// When a ellipsis is filled in with a value, and a item node that contains variables is repeated,
// the variable names will become array-like notation.
// For example, <L[4] <U1 var> varNode ... <A "text">> will be
// <L[5] <U1 var[0]> varNode[0] <U1 var[1]> varNode[1] <A "text">>,
// when 1 is filled into the ellipsis (1 repeat).
//
// The multi-dimensional array-like notation is also possible, when there was nested ListNodes with ellipsis,
// therefore, repeating nested ListNode multiple times.
// Nested ellipsis can be named and identified also with the array-like notation, e.g. ...[0], ...[1].
//
// The size of the ListNode in it's string representation, will be only specified when the size is deterministic,
// which means there is no ellipsis and ItemNode variable.
type ListNode struct {
	values    []ItemNode     // Array of ItemNodes that this ListNode contains
	variables map[string]int // Variable name and its position in the data array

	// Rep invariants
	// - If a variable exists in position i, values[i] will be zero-value (emptyItemNode) and should not be used
	// - The first item of the list node should not be an ellipsis
	// - Variable names should adhere to the variable naming rule; refer to interface.go
	// - All variable names in a ListNode, including its child item nodes' variables, should be unique
	// - Each ListNode can contain at most one ellipsis variable, counted *non-recursively*
	// - Variable positions should be unique, and be in range of [0, len(values))
}

// Factory methods

// NewListNode creates a new ListNode that contains multiple data item nodes.
//
// Each input of the values should be a ItemNode,
// or a string with valid variable name as specified in the interface documentation.
func NewListNode(values ...interface{}) ItemNode {
	if getDataByteLength("list", len(values)) > MAX_BYTE_SIZE {
		panic("item node size limit exceeded")
	}

	var (
		nodeValues    []ItemNode     = make([]ItemNode, 0, len(values))
		nodeVariables map[string]int = make(map[string]int)
		emptyNode     ItemNode       = NewEmptyItemNode()
	)

	for i, value := range values {
		if v, ok := value.(ItemNode); ok {
			nodeValues = append(nodeValues, v)
		} else if v, ok := value.(string); ok {
			nodeValues = append(nodeValues, emptyNode)
			if _, ok := nodeVariables[v]; ok {
				panic("duplicated variable name found")
			}
			nodeVariables[v] = i
		} else {
			panic("input argument contains invalid type for ListNode")
		}
	}

	node := &ListNode{nodeValues, nodeVariables}
	node.checkRep()
	return node
}

// Public methods

// Size implements ItemNode.Size().
func (node *ListNode) Size() int {
	return len(node.values)
}

func (node *ListNode) Type() string {
	return "list"
}

func (node *ListNode) Value() []ItemNode {
	return node.values
}

// Variables implements ItemNode.Variables().
func (node *ListNode) Variables() []string {
	result := []string{}

	var posVar map[int]string = node.variablesSwapKeyValue()
	for i, item := range node.values {
		if _, ok := item.(emptyItemNode); ok {
			// Contains item node variable
			result = append(result, posVar[i])
		} else {
			// Call Variables() of child node recursively
			result = append(result, item.Variables()...)
		}
	}

	return result
}

// FillVariables implements ItemNode.FillVariables().
func (node *ListNode) FillVariables(values map[string]interface{}) ItemNode {
	ellipsisValues, otherValues := node.splitValues(values)

	// Fill in ellipsis
	ellipsisToFill, ellipsisRemaining := node.ellipsisAnalysis(ellipsisValues)
	nodeEllipsisFilled := node
	if ellipsisToFill > 0 {
		nodeEllipsisFilled = node.fillEllipsis(ellipsisValues, newFillState(ellipsisRemaining)).(*ListNode)
	}

	// Fill in non-ellipsis variables with specified values
	nodeValues := make([]interface{}, 0, nodeEllipsisFilled.Size())
	for _, item := range nodeEllipsisFilled.values {
		nodeValues = append(nodeValues, item.FillVariables(otherValues))
	}
	for name, pos := range nodeEllipsisFilled.variables {
		if v, ok := otherValues[name]; ok {
			nodeValues[pos] = v
		} else {
			nodeValues[pos] = name
		}
	}

	return NewListNode(nodeValues...)
}

// ToBytes implements ItemNode.ToBytes()
func (node *ListNode) ToBytes() []byte {
	if len(node.variables) != 0 {
		return []byte{}
	}

	result, err := getHeaderBytes("list", node.Size())
	if err != nil {
		return []byte{}
	}

	for _, item := range node.values {
		// Call ToBytes() of child node recursively
		childResult := item.ToBytes()
		if len(childResult) == 0 {
			return []byte{}
		}
		result = append(result, childResult...)
	}

	return result
}

// String returns the string representation of the node.
func (node *ListNode) String() string {
	return node.stringIndented(0)
}

// Private methods

func (node *ListNode) checkRep() {
	ellipsisExist := false
	visitedIndex := map[int]bool{}
	for name, pos := range node.variables {
		if _, ok := node.values[pos].(emptyItemNode); !ok {
			panic("value in variable position isn't a zero-value")
		}

		if !isValidVarName(name) {
			if isEllipsis(name) {
				if pos == 0 {
					panic("ellipsis shouldn't be the first item in ListNode")
				}

				if ellipsisExist {
					panic("multiple ellipsis is not supported")
				} else {
					ellipsisExist = true
				}
			} else {
				panic("invalid variable name")
			}
		}

		if _, ok := visitedIndex[pos]; ok {
			panic("variable position is not unique")
		}
		visitedIndex[pos] = true

		if !(0 <= pos && pos < node.Size()) {
			panic("variable position overflow")
		}
	}

	// Check duplicated variables including child item nodes
	variables := node.Variables()
	foundVarName := map[string]bool{}
	for _, v := range variables {
		if _, ok := foundVarName[v]; ok {
			panic("duplicated variable name found in child item node")
		}
		foundVarName[v] = true
	}
}

// stringIndented returns the indented string representation of this list node.
// Each indent level adds 2 spaces as prefix to each line.
// The indent level should be non-negative.
func (node *ListNode) stringIndented(level int) string {
	indentStr := strings.Repeat("  ", level)
	if node.Size() == 0 {
		return fmt.Sprintf("%v<L[0]>", indentStr)
	}

	var (
		posVar         map[int]string = node.variablesSwapKeyValue()
		sizeDetermined bool           = true
		sb             strings.Builder
	)
	for i, val := range node.values {
		if v, ok := val.(*ListNode); ok {
			// Nested ListNode
			fmt.Fprintln(&sb, v.stringIndented(level+1))
		} else if varName, ok := posVar[i]; ok {
			// Variable in ListNode
			if isEllipsis(varName) {
				varName = "..."
			}
			fmt.Fprintf(&sb, "%v  %v\n", indentStr, varName)
			sizeDetermined = false
		} else {
			// Child ItemNode
			fmt.Fprintf(&sb, "%v  %v\n", indentStr, val)
		}
	}

	sizeStr := ""
	if sizeDetermined {
		sizeStr = fmt.Sprintf("[%d]", node.Size())
	}
	return fmt.Sprintf("%v<L%v\n%v%v>", indentStr, sizeStr, sb.String(), indentStr)
}

// variablesSwapKeyValue returns a new map with the keys and the values of node.variables swapped.
// The key and the value of the node.variables are guaranteed to be unique, by the rep invariant.
func (node *ListNode) variablesSwapKeyValue() map[int]string {
	result := map[int]string{}
	for k, v := range node.variables {
		result[v] = k
	}
	return result
}

// splitValues splits input map into two independent map, one with ellipsis key and one without.
func (node *ListNode) splitValues(values map[string]interface{}) (ellipsisValues, otherValues map[string]interface{}) {
	ellipsisValues = map[string]interface{}{}
	otherValues = map[string]interface{}{}
	for k, v := range values {
		if isEllipsis(k) {
			ellipsisValues[k] = v
		} else {
			otherValues[k] = v
		}
	}
	return ellipsisValues, otherValues
}

// ellipsisAnalysis returns the number of ellipsis to be filled in, and the number
// of remaining ellipsis after filling in target ellipsis.
func (node *ListNode) ellipsisAnalysis(values map[string]interface{}) (int, int) {
	var (
		ellipsisToFill    int
		ellipsisRemaining int
		ellipsisValue     int
	)
	for name := range node.variables {
		if isEllipsis(name) {
			if v, ok := values[name]; ok {
				ellipsisToFill = 1
				ellipsisValue = v.(int)
			} else {
				ellipsisRemaining = 1
			}
		}
	}
	for _, item := range node.values {
		if listNode, ok := item.(*ListNode); ok {
			ef, er := listNode.ellipsisAnalysis(values)
			ellipsisToFill += (ellipsisValue + 1) * ef
			ellipsisRemaining += (ellipsisValue + 1) * er
		}
	}
	return ellipsisToFill, ellipsisRemaining
}

// fillEllipsis fills in ellipsis variables with specified number of repeated
// item nodes in the ListNode. Ellipsis will be filled in appearing order on
// the top ListNode's string representation.
func (node *ListNode) fillEllipsis(values map[string]interface{}, state *fillState) ItemNode {

	// Check whether this ListNode have a ellipsis to fill
	var (
		ellipsisPosition int = -1
		ellipsisValue    int = 0
	)
	for name, pos := range node.variables {
		if _, ok := values[name]; ok && isEllipsis(name) {
			ellipsisPosition = pos
			ellipsisValue = values[name].(int)
			if ellipsisValue > 0 {
				state.growDimension()
			}
			break
		}
	}

	nodeValues := []interface{}{}
	posVar := node.variablesSwapKeyValue()
	for i := 0; i < node.Size(); i++ {
		// Repeat handling
		if i == ellipsisPosition {
			if ellipsisValue == 0 {
				// No state change; use as is and leave as is
				continue
			}

			if state.getCurrentDimensionIndex() < ellipsisValue {
				// Repeat items before ellipsis
				state.growIndex()
				i = 0
			} else {
				// Repeat finished
				state.exitDimension()
				continue
			}
		}

		// Handle each item in the list node
		item := node.values[i]
		switch itemTyped := item.(type) {
		case *ListNode:
			nodeValues = append(nodeValues, itemTyped.fillEllipsis(values, state))
		case *ASCIINode:
			if len(item.Variables()) == 0 {
				nodeValues = append(nodeValues, item)
			} else {
				varName := state.getNewVariableName(item.Variables()[0])
				minLength := itemTyped.variable.minLength
				maxLength := itemTyped.variable.maxLength
				nodeValues = append(nodeValues, NewASCIINodeVariable(varName, minLength, maxLength))
			}
		case emptyItemNode:
			varName := state.getNewVariableName(posVar[i])
			nodeValues = append(nodeValues, varName)
		default:
			variables := item.Variables()
			if len(variables) == 0 {
				nodeValues = append(nodeValues, item)
			} else {
				fill := map[string]interface{}{}
				for _, v := range variables {
					fill[v] = state.getNewVariableName(v)
				}
				nodeValues = append(nodeValues, item.FillVariables(fill))
			}
		}
	}
	return NewListNode(nodeValues...)
}

// fillState is a mutable data type that contains state information for ListNode.fillEllipsis().
// It should be created on ListNode.FillVariables(), should be used on fillEllipsis() method call
// inside FillVariables(), and should not be exposed to elsewhere.
type fillState struct {
	currentDimension int   // current dimension of indices, currentIndices[:currentDimension]
	currentIndices   []int // current indices which describe suffix of next variable name
	ellipsisCount    int   // Number of encountered ellipsis that is not filled in
	multipleEllipsis bool  // true if there will be multiple remaining ellipsis after fillEllipsis() call
}

func newFillState(remainingEllipsisCount int) *fillState {
	multipleEllipsis := false
	if remainingEllipsisCount > 1 {
		multipleEllipsis = true
	}
	return &fillState{0, []int{}, 0, multipleEllipsis}
}

func (state *fillState) growDimension() {
	if state.currentDimension == len(state.currentIndices) {
		state.currentIndices = append(state.currentIndices, 0)
	} else {
		state.currentIndices[state.currentDimension] = 0
	}
	state.currentDimension += 1
}

func (state *fillState) exitDimension() {
	state.currentDimension -= 1
}

func (state *fillState) getCurrentDimensionIndex() int {
	return state.currentIndices[state.currentDimension-1]
}

func (state *fillState) growIndex() {
	state.currentIndices[state.currentDimension-1] += 1
}

func (state *fillState) getNewVariableName(name string) string {
	if isEllipsis(name) {
		if state.multipleEllipsis {
			name = fmt.Sprintf("...[%d]", state.ellipsisCount)
			state.ellipsisCount += 1
			return name
		} else {
			return "..."
		}
	}

	for i := 0; i < state.currentDimension; i++ {
		name += fmt.Sprintf("[%d]", state.currentIndices[i])
	}
	return name
}
