package ast

// HSMS data message is defined in ast.go

const (
	sTypeSelectReq   = 1
	sTypeSelectRsp   = 2
	sTypeDeselectReq = 3
	sTypeDeselectRsp = 4
	sTypeLinktestReq = 5
	sTypeLinktestRsp = 6
	sTypeRejectReq   = 7
	sTypeSeparateReq = 9
)

// HSMSMessage is a interface of immutable data types that represents a HSMS message.
//
// HSMSMessage contains two implementations, DataMessage and ControlMessage.
//
// DataMessage represents a SECS-II data message. Note that, some DataMessage
// might not be converted to HSMS format, i.e. its wait bit is in optional state,
// it's data item contain variables, or session id and system bytes are not set.
// Only a complete DataMessage, i.e. wait bit is true or false, no variables in data item,
// session id and system bytes are set, can be converted to HSMS format, and thus
// can be sent to the recipient.
//
// ControlMessage can represent one of select.req, select.rsp, deselect.req, deselect.rsp,
// linktest.req, linktest.rsp, reject.req, separate.req, and undefined control message.
type HSMSMessage interface {
	// Type returns HSMS message type.
	// Return will be one of "data message", "select.req", "select.rsp", "deselect.req", "deselect.rsp",
	// "linktest.req", "linktest.rsp", "reject.req", "separate.req", "undefined".
	Type() string

	// ToBytes returns byte representation of the HSMS message.
	ToBytes() []byte
}

// ControlMessage is a immutable data type that represents a HSMS control message.
// Implements HSMSMessage.
type ControlMessage struct {
	header []byte
	// Rep invariants
	// - header should have length of 10
	//
	// Safety from rep exposure
	// - header should not be exposed
}

// NewHSMSControlMessage creates HSMS control message from header bytes.
// header bytes should have appropriate values as specified in HSMS specification.
func NewHSMSControlMessage(header []byte) HSMSMessage {
	headerCopy := make([]byte, 10)
	for i, b := range header {
		if i > 10 {
			break
		}
		headerCopy[i] = b
	}
	return &ControlMessage{headerCopy}
}

// NewHSMSMessageSelectReq creates HSMS Select.req control message.
// systemBytes should have length of 4.
func NewHSMSMessageSelectReq(sessionID uint16, systemBytes []byte) HSMSMessage {
	header := make([]byte, 10)
	header[0] = byte(sessionID >> 8)
	header[1] = byte(sessionID)
	header[5] = sTypeSelectReq
	header[6] = systemBytes[0]
	header[7] = systemBytes[1]
	header[8] = systemBytes[2]
	header[9] = systemBytes[3]

	return &ControlMessage{header}
}

// NewHSMSMessageSelectRsp creates HSMS Select.rsp control message from Select.req message.
// selectStatus 0 means that communication is successfully established,
// 1 means that communication is already active,
// 2 means that communication is not ready,
// 3 means that connection that TCP/IP port is exhausted,
// 4-255 are reserved failure reason codes.
func NewHSMSMessageSelectRsp(selectReq HSMSMessage, selectStatus byte) HSMSMessage {
	if selectReq.Type() != "select.req" {
		panic("expected select.req message")
	}

	header := make([]byte, 10)
	msg, _ := selectReq.(*ControlMessage)
	header[0] = msg.header[0]
	header[1] = msg.header[1]
	header[3] = selectStatus
	header[5] = sTypeSelectRsp
	header[6] = msg.header[6]
	header[7] = msg.header[7]
	header[8] = msg.header[8]
	header[9] = msg.header[9]

	return &ControlMessage{header}
}

// NewHSMSMessageDeselectReq creates HSMS Deselect.req control message.
// systemBytes should have length of 4.
func NewHSMSMessageDeselectReq(sessionID uint16, systemBytes []byte) HSMSMessage {
	header := make([]byte, 10)
	header[0] = byte(sessionID >> 8)
	header[1] = byte(sessionID)
	header[5] = sTypeDeselectReq
	header[6] = systemBytes[0]
	header[7] = systemBytes[1]
	header[8] = systemBytes[2]
	header[9] = systemBytes[3]

	return &ControlMessage{header}
}

// NewHSMSMessageDeselectRsp creates HSMS Deselect.rsp control message from Deselect.req message.
// deselectStatus 0 means that the connection is successfully ended,
// 1 means that communication is not yet established,
// 2 means that communication is busy and cannot yet be relinquished,
// 3-255 are reserved failure reason codes.
func NewHSMSMessageDeselectRsp(deselectReq HSMSMessage, deselectStatus byte) HSMSMessage {
	if deselectReq.Type() != "deselect.req" {
		panic("expected deselect.req message")
	}

	header := make([]byte, 10)
	msg, _ := deselectReq.(*ControlMessage)
	header[0] = msg.header[0]
	header[1] = msg.header[1]
	header[3] = deselectStatus
	header[5] = sTypeDeselectRsp
	header[6] = msg.header[6]
	header[7] = msg.header[7]
	header[8] = msg.header[8]
	header[9] = msg.header[9]

	return &ControlMessage{header}
}

// NewHSMSMessageLinktestReq creates HSMS Linktest.req control message.
// systemBytes should have length of 4.
func NewHSMSMessageLinktestReq(systemBytes []byte) HSMSMessage {
	header := make([]byte, 10)
	header[0] = 0xFF
	header[1] = 0xFF
	header[5] = sTypeLinktestReq
	header[6] = systemBytes[0]
	header[7] = systemBytes[1]
	header[8] = systemBytes[2]
	header[9] = systemBytes[3]

	return &ControlMessage{header}
}

// NewHSMSMessageLinktestRsp creates HSMS Linktest.rsp control message from Linktest.req message.
func NewHSMSMessageLinktestRsp(linktestReq HSMSMessage) HSMSMessage {
	if linktestReq.Type() != "linktest.req" {
		panic("expected linktest.req message")
	}

	header := make([]byte, 10)
	msg, _ := linktestReq.(*ControlMessage)
	header[0] = 0xFF
	header[1] = 0xFF
	header[5] = sTypeLinktestRsp
	header[6] = msg.header[6]
	header[7] = msg.header[7]
	header[8] = msg.header[8]
	header[9] = msg.header[9]

	return &ControlMessage{header}
}

// NewHSMSMessageRejectReq creates HSMS Reject.req control message.
//
// sessionID, pType, sType, and systemBytes should be same as the HSMS message being rejected.
// systemBytes should have length of 4.
//
// reasonCode should be non-zero,
// 1 means that received message's sType is not supported,
// 2 means that received message's pType is not supported,
// 3 means that transaction is not open, i.e. response message was received without request,
// 4 means that data message is received in non-SELECTED state,
// 5-255 are reserved reason codes.
func NewHSMSMessageRejectReq(sessionID uint16, pType, sType byte, systemBytes []byte, reasonCode byte) HSMSMessage {
	header := make([]byte, 10)
	header[0] = byte(sessionID >> 8)
	header[1] = byte(sessionID)
	if reasonCode == 2 {
		header[2] = pType
	} else {
		header[2] = sType
	}
	header[3] = reasonCode
	header[5] = sTypeRejectReq
	header[6] = systemBytes[0]
	header[7] = systemBytes[1]
	header[8] = systemBytes[2]
	header[9] = systemBytes[3]

	return &ControlMessage{header}
}

// NewHSMSMessageSeparateReq creates HSMS Separate.req control message.
// systemBytes should have length of 4.
func NewHSMSMessageSeparateReq(sessionID uint16, systemBytes []byte) HSMSMessage {
	header := make([]byte, 10)
	header[0] = byte(sessionID >> 8)
	header[1] = byte(sessionID)
	header[5] = sTypeSeparateReq
	header[6] = systemBytes[0]
	header[7] = systemBytes[1]
	header[8] = systemBytes[2]
	header[9] = systemBytes[3]

	return &ControlMessage{header}
}

// Type returns the message type of the HSMS control message.
// Return will be one of "select.req", "select.rsp", "deselect.req", "deselect.rsp",
// "linktest.req", "linktest.rsp", "reject.req", "separate.req", "undefined".
func (msg *ControlMessage) Type() string {
	if msg.header[4] != 0 {
		return "undefined"
	}

	switch msg.header[5] {
	case 1:
		return "select.req"
	case 2:
		return "select.rsp"
	case 3:
		return "deselect.req"
	case 4:
		return "deselect.rsp"
	case 5:
		return "linktest.req"
	case 6:
		return "linktest.rsp"
	case 7:
		return "reject.req"
	case 9:
		return "separate.req"
	default:
		return "undefined"
	}
}

// ToBytes returns the HSMS byte representation of the control message.
func (msg *ControlMessage) ToBytes() []byte {
	result := make([]byte, 0, 14)
	result = append(result, 0, 0, 0, 10)
	result = append(result, msg.header...)
	return result
}
