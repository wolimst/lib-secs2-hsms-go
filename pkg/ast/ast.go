// Package ast contains data types that represent abstract syntax tree of
// a SECS-II message, and data items in a SECS-II message.
package ast

import (
	"fmt"
	"unicode"
)

// DataMessage is a immutable data type that represents a SECS-II message.
// Implements HSMSMessage.
type DataMessage struct {
	name        string   // message name; should not contain whitespaces
	stream      int      // should be in range of [0, 128)
	function    int      // should be in range of [0, 256)
	waitBit     int      // 0 if wait bit is false, 1 if true, 2 if optional
	direction   string   // one of "H->E", "H<-E", "H<->E"
	dataItem    ItemNode // data item node that the message contains
	sessionID   int      // should be in range of [-1, 65536); -1 means not specified
	systemBytes []byte   // slice length should be 4

	// Rep invariants
	// - name should not contain whitespaces
	// - stream code should be in range of [0, 128)
	// - function code should be in range of [0, 256)
	// - waitBit should be either 0, 1, or 2
	// - waitBit should not be 1 (true) when function code is a even number
	// - direction should be either "H->E", "H<-E", or "H<->E"
	// - sessionID should be in range of [-1, 65536)
	// - systemBytes' length should be 4
}

// Factory methods

// NewDataMessage creates a new SECS-II message.
//
// The message can't be converted to HSMS format, before session id and system
// bytes are set, all variables in data item is set (if had any), and wait bit
// is set (if it had value of optional).
//
//
// Input argument specifications
//
// name is a identifier of this message node, that doesn't contain whitespaces.
//
// stream is a stream code of this message and should be in range of [0, 128).
//
// function is a function code of this message and should be in range of [0, 256).
//
// waitBit should be either 0 (false), 1 (true) or 2 (optional).
// waitBit cannot be 1 (true) when the function code is a even number.
//
// direction represents the direction of the message between the host and the equipment.
// direction should be either "H->E", "H<-E", or "H<->E".
//
// dataItem is the contents of this message.
func NewDataMessage(name string, stream int, function int, waitBit int, direction string, dataItem ItemNode) *DataMessage {
	message := &DataMessage{
		name:        name,
		stream:      stream,
		function:    function,
		waitBit:     waitBit,
		direction:   direction,
		dataItem:    dataItem,
		sessionID:   -1,
		systemBytes: []byte{0, 0, 0, 0},
	}
	message.checkRep()
	return message
}

// NewHSMSDataMessage creates a new SECS-II message, which can be converted to HSMS format.
//
// Input argument specifications
//
// name is a identifier of this message node, that doesn't contain whitespaces.
//
// stream is a stream code of this message and should be in range of [0, 128).
//
// function is a function code of this message and should be in range of [0, 256).
//
// waitBit should be either 0 (false) or 1 (true).
// waitBit cannot be 1 (true) when the function code is a even number.
//
// direction represents the direction of the message between the host and the equipment.
// direction should be either "H->E", "H<-E", or "H<->E".
//
// dataItem is the contents of this message, and it shouldn't contain any variable.
//
// sessionID should be in range of [0, 65535).
//
// systemBytes should have 4 bytes.
func NewHSMSDataMessage(name string, stream int, function int, waitBit int, direction string, dataItem ItemNode, sessionID int, systemBytes []byte) *DataMessage {
	if waitBit != 0 && waitBit != 1 {
		panic("wait bit should be 0 or 1 when creating HSMS convertible message")
	}

	if sessionID == -1 {
		panic("sessionID should be in range of [0, 65535) when creating HSMS convertible message")
	}

	if len(dataItem.Variables()) != 0 {
		panic("data item should not contain variables when creating HSMS convertible message")
	}

	systemBytesCopy := make([]byte, 4)
	for i, b := range systemBytes {
		if i >= 4 {
			break
		}
		systemBytesCopy[i] = b
	}

	message := &DataMessage{
		name:        name,
		stream:      stream,
		function:    function,
		waitBit:     waitBit,
		direction:   direction,
		dataItem:    dataItem,
		sessionID:   sessionID,
		systemBytes: systemBytesCopy,
	}
	message.checkRep()
	return message
}

// Public methods

// Name returns the name of the SECS-II message.
func (node *DataMessage) Name() string {
	return node.name
}

// StreamCode returns the stream code of the SECS-II message.
func (node *DataMessage) StreamCode() int {
	return node.stream
}

// FunctionCode returns the function code of the SECS-II message.
func (node *DataMessage) FunctionCode() int {
	return node.function
}

// WaitBit returns the wait bit status of the SECS-II message, which is one of "true", "false", "optional".
func (node *DataMessage) WaitBit() string {
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

// SetWaitBit sets the wait bit of the message, if it had optional wait bit.
// If the wait bit was not optional, return will be same as the original message.
//
// waitBit shouldn't be true if the message is a reply message, i.e. function code
// is a even number.
func (node *DataMessage) SetWaitBit(waitBit bool) *DataMessage {
	if node.waitBit != 2 {
		return node
	}

	waitBitAsNumber := 0
	if waitBit {
		waitBitAsNumber = 1
	}

	message := &DataMessage{
		name:        node.name,
		stream:      node.stream,
		function:    node.function,
		waitBit:     waitBitAsNumber,
		direction:   node.direction,
		dataItem:    node.dataItem,
		sessionID:   node.sessionID,
		systemBytes: node.systemBytes,
	}
	message.checkRep()
	return message
}

// Direction returns the direction of the SECS-II message.
func (node *DataMessage) Direction() string {
	return node.direction
}

// SessionID returns the session id of the SECS-II message.
// If the session id was not set, it will return -1.
func (node *DataMessage) SessionID() int {
	return node.sessionID
}

// SystemBytes returns the system bytes of the SECS-II message.
// If the system bytes was not set, it will return []byte{0, 0, 0, 0}.
func (node *DataMessage) SystemBytes() []byte {
	return node.systemBytes
}

// SetSessionIDAndSystemBytes sets session id and system bytes to the message.
//
// sessionID should be in range of [-1, 65536).
// systemBytes should have length of 4.
//
// Since DataMessage is a immutable type, this method will create and return a new message.
func (node *DataMessage) SetSessionIDAndSystemBytes(sessionID int, systemBytes []byte) *DataMessage {
	systemBytesCopy := make([]byte, 4)
	for i, b := range systemBytes {
		if i >= 4 {
			break
		}
		systemBytesCopy[i] = b
	}

	message := &DataMessage{
		name:        node.name,
		stream:      node.stream,
		function:    node.function,
		waitBit:     node.waitBit,
		direction:   node.direction,
		dataItem:    node.dataItem,
		sessionID:   sessionID,
		systemBytes: systemBytesCopy,
	}
	message.checkRep()
	return message
}

// Header returns the message header of the SECS-II message, e.g. "S6F11 W H<-E MessageName".
func (node *DataMessage) Header() string {
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
func (node *DataMessage) Variables() []string {
	return node.dataItem.Variables()
}

// FillVariables returns a new DataMessage with the specified values filled into the
// variables in this SECS-II message.
//
// The map input argument has variable name as its key, and fill-in value as its value.
// Each fill-in value must be acceptable by the ItemNode's factory method.
// If a variable in the ItemNode doesn't exist in the input map, the variable will remain unchanged.
func (node *DataMessage) FillVariables(values map[string]interface{}) *DataMessage {
	item := node.dataItem.FillVariables(values)

	message := &DataMessage{
		name:        node.name,
		stream:      node.stream,
		function:    node.function,
		waitBit:     node.waitBit,
		direction:   node.direction,
		dataItem:    item,
		sessionID:   node.sessionID,
		systemBytes: node.systemBytes,
	}
	message.checkRep()
	return message
}

// Type returns HSMS message type.
// Implements HSMSMessage.Type().
func (node *DataMessage) Type() string {
	return "data message"
}

// ToBytes returns the HSMS byte representation of the SECS-II message.
//
// It will return empty byte slice if the message can't be represented as HSMS format,
// i.e. wait bit is optional, the data item contains variable, or
// session id and message bytes are not set.
//
// Implements HSMSMessage.ToBytes().
func (node *DataMessage) ToBytes() []byte {
	if node.WaitBit() == "optional" || len(node.Variables()) != 0 || node.sessionID == -1 {
		return []byte{}
	}

	itemBytes := node.dataItem.ToBytes()
	result := make([]byte, 0, len(itemBytes)+14) // 4 length bytes, 10 header bytes

	// Message length bytes
	var msgLength uint32 = uint32(len(itemBytes) + 10) // 10 header bytes
	result = append(result, byte(msgLength>>24))
	result = append(result, byte(msgLength>>16))
	result = append(result, byte(msgLength>>8))
	result = append(result, byte(msgLength))
	// Header byte 0-1: device ID
	result = append(result, byte(node.sessionID>>8))
	result = append(result, byte(node.sessionID))
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
	result = append(result, node.systemBytes[:4]...)
	// Message text
	result = append(result, itemBytes...)

	return result
}

func (node *DataMessage) String() string {
	if _, ok := node.dataItem.(emptyItemNode); ok {
		return fmt.Sprintf("%s\n.", node.Header())
	}
	return fmt.Sprintf("%s\n%s\n.", node.Header(), node.dataItem)
}

// Private methods

func (node *DataMessage) checkRep() {
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

	if !(-1 <= node.sessionID && node.sessionID < 65536) {
		panic("session id out of range")
	}

	if len(node.systemBytes) != 4 {
		panic("system bytes length is not 4")
	}

	if node.direction != "H->E" && node.direction != "H<-E" && node.direction != "H<->E" {
		panic("invalid direction")
	}
}
