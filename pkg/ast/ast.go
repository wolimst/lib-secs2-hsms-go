// Package ast contains data types that represent abstract syntax tree of
// a SECS-II message, and data items in a SECS-II message.
package ast

import (
	"fmt"
	"unicode"
)

// MessageNode is a immutable data type that represents a SECS-II message.
type MessageNode struct {
	name      string   // message name
	stream    int      // stream code of the message
	function  int      // function code of the message
	waitBit   int      // wait bit of the message; 0 if wait bit is false, 1 if true, 2 if optional
	direction string   // message direction; one of "H->E", "H<-E", "H<->E"
	dataItem  ItemNode // data item node that the message contains

	// Rep invariants
	// - name should not contain whitespaces
	// - stream code should be in range of [0, 128)
	// - function code should be in range of [0, 256)
	// - waitBit should be either 0, 1, or 2
	// - waitBit should not be 1 (true) when function code is a even number
	// - direction should be either "H->E", "H<-E", or "H<->E"
}

// Factory methods

// NewMessageNode creates a new MessageNode that represents a SECS-II message.
//
// The input argument name is a identifier of this message node.
// name should not contain whitespaces.
// stream is a stream code of this message and should be in range of [0, 128).
// function is a function code of this message and should be in range of [0, 256).
// waitBit should be either 0 (false), 1 (true), 2 (optional).
// waitBit cannot be 1 (true) when the function code is a even number.
// direction represents the direction of the message between the host and the equipment.
// direction should be either "H->E", "H<-E", or "H<->E".
// dataItem is a ItemNode that the message node will contain.
func NewMessageNode(name string, stream int, function int, waitBit int, direction string, dataItem ItemNode) *MessageNode {
	messageNode := &MessageNode{name, stream, function, waitBit, direction, dataItem}
	messageNode.checkRep()
	return messageNode
}

// NewEmptyMessageNode creates a new SECS-II message node that contains no data item.
//
// The input argument name is a identifier of this message node.
// stream is a stream code of this message and should be in range of [0, 128).
// function is a function code of this message and should be in range of [0, 256).
// waitBit should be either 0 (false), 1 (true), 2 (optional).
// waitBit cannot be 1 (true) when the function code is a even number.
// direction represents the direction of the message between the host and the equipment.
// direction should be either "H->E", "H<-E", or "H<->E".
func NewEmptyMessageNode(name string, stream int, function int, waitBit int, direction string) *MessageNode {
	return NewMessageNode(name, stream, function, waitBit, direction, emptyItemNode{})
}

// Public methods

// Name returns the name of the SECS-II message.
func (node *MessageNode) Name() string {
	return node.name
}

// StreamCode returns the stream code of the SECS-II message.
func (node *MessageNode) StreamCode() int {
	return node.stream
}

// FunctionCode returns the function code of the SECS-II message.
func (node *MessageNode) FunctionCode() int {
	return node.function
}

// WaitBit returns the wait bit status of the SECS-II message, which is one of "true", "false", "optional".
func (node *MessageNode) WaitBit() string {
	switch node.waitBit {
	case 0:
		return "false"
	case 1:
		return "true"
	case 2:
		return "optional"
	}
	panic("rep invariant broken")
}

// Direction returns the direction of the SECS-II message.
func (node *MessageNode) Direction() string {
	return node.direction
}

// Header returns the message header of the SECS-II message, e.g. "S6F11 W H<-E MessageName".
func (node *MessageNode) Header() string {
	header := fmt.Sprintf("S%dF%d", node.stream, node.function)

	switch node.waitBit {
	case 1:
		header += " W"
	case 2:
		header += " [W]"
	}

	header += " " + node.direction

	if len(node.name) > 0 {
		header += " " + node.name
	}

	return header
}

// Variables returns the variable names in this SECS-II message.
func (node *MessageNode) Variables() []string {
	return node.dataItem.Variables()
}

// FillVariables returns a new MessageNode with the specified values filled into the
// variables in this SECS-II message.
//
// The map input argument has variable name as its key, and fill-in value as its value.
// Each fill-in value must be acceptable by the ItemNode's factory method.
// If a variable in the ItemNode doesn't exist in the input map, the variable will remain unchanged.
func (node *MessageNode) FillVariables(values map[string]interface{}) *MessageNode {
	item := node.dataItem.FillVariables(values)
	return NewMessageNode(node.name, node.stream, node.function, node.waitBit, node.direction, item)
}

// ToHSMS returns the HSMS representation of the SECS-II message.
//
// It will return empty byte slice if the wait bit is optional or the data item
// in the message has variables.
//
// systemBytes is a transaction identifier, and it should have exactly 4 elements.
// If the message is a reply message, the system bytes should be equal to the primary
// message's system bytes, otherwise the recipient of this HSMS might not recognize it.
func (node *MessageNode) ToHSMS(deviceID uint16, systemBytes []byte) []byte {
	if node.WaitBit() == "optional" || len(node.Variables()) != 0 {
		return []byte{}
	}

	itemBytes := node.dataItem.ToBytes()
	result := make([]byte, 0, len(itemBytes)+14)

	// Message length bytes
	var msgLength uint32 = uint32(len(itemBytes) + 10)
	result = append(result, byte(msgLength>>24))
	result = append(result, byte(msgLength>>16))
	result = append(result, byte(msgLength>>8))
	result = append(result, byte(msgLength))
	// Header byte 0-1: device ID
	result = append(result, byte(deviceID>>8))
	result = append(result, byte(deviceID))
	// Header byte 2-3: wait bit + stream code, function code
	headerByte2 := node.StreamCode()
	if node.WaitBit() == "true" {
		headerByte2 += 0b10000000
	}
	result = append(result, byte(headerByte2))
	result = append(result, byte(node.FunctionCode()))
	// Header byte 4-5: PType, SType
	result = append(result, 0, 0)
	// Header byte 6-9: system bytes
	result = append(result, systemBytes[:4]...)
	// Message text
	result = append(result, itemBytes...)

	return result
}

func (node *MessageNode) String() string {
	if _, ok := node.dataItem.(emptyItemNode); ok {
		return fmt.Sprintf("%s\n.", node.Header())
	}
	return fmt.Sprintf("%s\n%s\n.", node.Header(), node.dataItem)
}

// Private methods

func (node *MessageNode) checkRep() {
	for _, ch := range node.name {
		if unicode.IsSpace(ch) {
			panic("message name shouldn't contain whitespaces")
		}
	}

	if !(0 <= node.stream && node.stream < 128) {
		panic("stream code out of range")
	}

	if !(0 <= node.function && node.function < 256) {
		panic("function code out of range")
	}

	if node.waitBit == 1 && node.function%2 == 0 {
		panic("wait bit = true is not valid for reply message")
	}

	if !(0 <= node.waitBit && node.waitBit <= 2) {
		panic("invalid wait bit")
	}

	if node.direction != "H->E" && node.direction != "H<-E" && node.direction != "H<->E" {
		panic("invalid direction")
	}
}
