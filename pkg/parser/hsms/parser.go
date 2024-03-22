package hsms

import (
	"encoding/binary"
	"math"

	"github.com/GunsonJack/lib-secs2-hsms-go/pkg/ast"
)

const (
	sTypeDataMessage = 0
	sTypeSelectReq   = 1
	sTypeSelectRsp   = 2
	sTypeDeselectReq = 3
	sTypeDeselectRsp = 4
	sTypeLinktestReq = 5
	sTypeLinktestRsp = 6
	sTypeRejectReq   = 7
	sTypeSeparateReq = 9

	formatCodeList    = 0o00
	formatCodeBinary  = 0o10
	formatCodeBoolean = 0o11
	formatCodeASCII   = 0o20
	formatCodeI8      = 0o30
	formatCodeI1      = 0o31
	formatCodeI2      = 0o32
	formatCodeI4      = 0o34
	formatCodeF8      = 0o40
	formatCodeF4      = 0o44
	formatCodeU8      = 0o50
	formatCodeU1      = 0o51
	formatCodeU2      = 0o52
	formatCodeU4      = 0o54
)

// Parse parses the input bytes that represent a HSMS message.
//
// input should contain only one HSMS message.
//
// If parsing fails, ok == false will be returned.
func Parse(input []byte) (msg ast.HSMSMessage, ok bool) {
	// Handle panics on abstract syntax tree creation
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()

	p := &parser{input: input}
	if ok := p.parseMessageLength(); !ok {
		return p.msg, false
	}
	if ok := p.parseMessage(); !ok {
		return p.msg, false
	}
	return p.msg, true
}

type parser struct {
	input     []byte          // a HSMS input message in bytes
	pos       int             // current position in input
	msgLength int             // message length (excluding length bytes)
	msg       ast.HSMSMessage // parsed HSMS message
}

// parseMessageLength parses the message length which is the first 4 bytes of
// HSMS byte input, and store the result in the parser struct.
func (p *parser) parseMessageLength() (ok bool) {
	if len(p.input) < 14 { // length bytes + header bytes
		return false
	}

	lengthBytes := p.input[0:4]
	p.pos += 4

	p.msgLength = int(binary.BigEndian.Uint32(lengthBytes))
	return len(p.input[p.pos:]) == p.msgLength
}

// parseMessage parses the message header and the message text, and store it
// in the parser struct.
func (p *parser) parseMessage() (ok bool) {
	headerBytes := p.input[p.pos : p.pos+10]
	p.pos += 10

	if headerBytes[4] != 0 { // PType
		// Not a SECS-II message
		return false
	}

	switch headerBytes[5] { // SType
	case sTypeDataMessage:
		stream := int(headerBytes[2] & 0b01111111)
		function := int(headerBytes[3])
		waitBit := int(headerBytes[2] >> 7)
		sessionID := int(binary.BigEndian.Uint16(headerBytes[:2]))
		systemBytes := headerBytes[6:10]
		dataItem, ok := p.parseMessageText()
		if !ok {
			return false
		}
		p.msg = ast.NewHSMSDataMessage("", stream, function, waitBit, "H<->E", dataItem, sessionID, systemBytes)
		return true

	case sTypeSelectReq, sTypeSelectRsp, sTypeDeselectReq, sTypeDeselectRsp,
		sTypeLinktestReq, sTypeLinktestRsp, sTypeRejectReq, sTypeSeparateReq:
		p.msg = ast.NewHSMSControlMessage(headerBytes)
		return true

	default:
		// Undefined SType
		return false
	}
	// should not reach here
}

// parseMessageText creates ast.ItemNode from binary HSMS message text.
// Return empty item node and ok == false if the message cannot be parsed.
func (p *parser) parseMessageText() (dataItem ast.ItemNode, ok bool) {
	if p.msgLength == 10 {
		return ast.NewEmptyItemNode(), true
	}

	formatCode := p.input[p.pos] >> 2
	lengthBytesCount := int(p.input[p.pos] & 0b00000011)
	if lengthBytesCount == 0 {
		return ast.NewEmptyItemNode(), false
	}
	p.pos += 1

	lengthBytes := p.input[p.pos : p.pos+lengthBytesCount]
	var length int
	for i, b := range lengthBytes {
		shift := (lengthBytesCount - i - 1) * 8
		length += int(b << shift)
	}
	p.pos += lengthBytesCount

	switch formatCode {
	case formatCodeList:
		values := make([]interface{}, length)
		for i := 0; i < length; i++ {
			values[i], ok = p.parseMessageText()
			if !ok {
				return ast.NewEmptyItemNode(), false
			}
		}
		return ast.NewListNode(values...), true

	case formatCodeASCII:
		var str string
		for _, v := range p.input[p.pos : p.pos+length] {
			str += string(v)
		}
		p.pos += length
		return ast.NewASCIINode(str), true

	case formatCodeBinary:
		values := make([]interface{}, length)
		for i, v := range p.input[p.pos : p.pos+length] {
			values[i] = v
		}
		p.pos += length
		return ast.NewBinaryNode(values...), true

	case formatCodeBoolean:
		values := make([]interface{}, length)
		for i, v := range p.input[p.pos : p.pos+length] {
			if v == 0 {
				values[i] = false
			} else {
				values[i] = true
			}
		}
		p.pos += length
		return ast.NewBooleanNode(values...), true

	case formatCodeF4:
		return p.parseFloat(4, length)
	case formatCodeF8:
		return p.parseFloat(8, length)

	case formatCodeI1:
		return p.parseInt(1, length)
	case formatCodeI2:
		return p.parseInt(2, length)
	case formatCodeI4:
		return p.parseInt(4, length)
	case formatCodeI8:
		return p.parseInt(8, length)

	case formatCodeU1:
		return p.parseUint(1, length)
	case formatCodeU2:
		return p.parseUint(2, length)
	case formatCodeU4:
		return p.parseUint(4, length)
	case formatCodeU8:
		return p.parseUint(8, length)

	default:
		return ast.NewEmptyItemNode(), false
	}
	// should not reach here
}

func (p *parser) parseFloat(byteSize int, length int) (dataItem ast.ItemNode, ok bool) {
	if length%byteSize != 0 {
		return ast.NewEmptyItemNode(), false
	}

	valueCounts := length / byteSize
	values := make([]interface{}, valueCounts)
	for i := 0; i < valueCounts; i++ {
		start, end := p.pos+byteSize*i, p.pos+byteSize*(i+1)
		if byteSize == 4 {
			bits := binary.BigEndian.Uint32(p.input[start:end])
			values[i] = math.Float32frombits(bits)
		} else if byteSize == 8 {
			bits := binary.BigEndian.Uint64(p.input[start:end])
			values[i] = math.Float64frombits(bits)
		}
	}
	p.pos += length
	return ast.NewFloatNode(byteSize, values...), true
}

func (p *parser) parseInt(byteSize int, length int) (dataItem ast.ItemNode, ok bool) {
	if length%byteSize != 0 {
		return ast.NewEmptyItemNode(), false
	}

	valueCounts := length / byteSize
	values := make([]interface{}, valueCounts)
	for i := 0; i < valueCounts; i++ {
		start, end := p.pos+byteSize*i, p.pos+byteSize*(i+1)
		if byteSize == 1 {
			values[i] = int8(p.input[p.pos+i])
		} else if byteSize == 2 {
			bits := binary.BigEndian.Uint16(p.input[start:end])
			values[i] = int16(bits)
		} else if byteSize == 4 {
			bits := binary.BigEndian.Uint32(p.input[start:end])
			values[i] = int32(bits)
		} else if byteSize == 8 {
			bits := binary.BigEndian.Uint64(p.input[start:end])
			values[i] = int64(bits)
		}
	}
	p.pos += length
	return ast.NewIntNode(byteSize, values...), true
}

func (p *parser) parseUint(byteSize int, length int) (dataItem ast.ItemNode, ok bool) {
	if length%byteSize != 0 {
		return ast.NewEmptyItemNode(), false
	}

	valueCounts := length / byteSize
	values := make([]interface{}, valueCounts)
	for i := 0; i < valueCounts; i++ {
		start, end := p.pos+byteSize*i, p.pos+byteSize*(i+1)
		if byteSize == 1 {
			values[i] = uint8(p.input[p.pos+i])
		} else if byteSize == 2 {
			values[i] = binary.BigEndian.Uint16(p.input[start:end])
		} else if byteSize == 4 {
			values[i] = binary.BigEndian.Uint32(p.input[start:end])
		} else if byteSize == 8 {
			values[i] = binary.BigEndian.Uint64(p.input[start:end])
		}
	}
	p.pos += length
	return ast.NewUintNode(byteSize, values...), true
}
