package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests the HSMS control messages
//
// Testing Strategy:
//
// Create each control message and test the result of public observer methods.
//
// Partitions:
//
// - sessionID: 0, ..., 65535
// - systemBytes: 0x00000000, ..., 0xFFFFFFFF
// - selectStatus: 0, 1, 2, 3
// - deselectStatus: 0, 1, 2
// - reject reasonCode: 1, 2, 3, 4
// - reject pType, sType: no partition

func TestHSMSControlMessage(t *testing.T) {
	msg := NewHSMSControlMessage([]byte{1, 2, 0, 0, 0, 1, 0, 1, 2, 3})
	assert.Equal(t, "select.req", msg.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 1, 2, 0, 0, 0, 1, 0, 1, 2, 3}, msg.ToBytes())
}

func TestHSMSControlMessage_SelectReqRsp(t *testing.T) {
	req1 := NewHSMSMessageSelectReq(0, []byte{0, 0, 0, 0})
	assert.Equal(t, "select.req", req1.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0}, req1.ToBytes())

	req2 := NewHSMSMessageSelectReq(1, []byte{0, 0, 0, 1})
	assert.Equal(t, "select.req", req2.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1}, req2.ToBytes())

	req3 := NewHSMSMessageSelectReq(0x0100, []byte{0xFC, 0xFD, 0xFE, 0xFF})
	assert.Equal(t, "select.req", req3.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 1, 0, 0, 0, 0, 1, 0xFC, 0xFD, 0xFE, 0xFF}, req3.ToBytes())

	req4 := NewHSMSMessageSelectReq(0xFFFF, []byte{0xFF, 0xFF, 0xFF, 0xFF})
	assert.Equal(t, "select.req", req4.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0xFF, 0xFF, 0, 0, 0, 1, 0xFF, 0xFF, 0xFF, 0xFF}, req4.ToBytes())

	// select status 0
	rsp1 := NewHSMSMessageSelectRsp(req1, 0)
	assert.Equal(t, "select.rsp", rsp1.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0}, rsp1.ToBytes())

	// select status 1
	rsp2 := NewHSMSMessageSelectRsp(req2, 1)
	assert.Equal(t, "select.rsp", rsp2.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0, 1, 0, 1, 0, 2, 0, 0, 0, 1}, rsp2.ToBytes())

	// select status 2
	rsp3 := NewHSMSMessageSelectRsp(req3, 2)
	assert.Equal(t, "select.rsp", rsp3.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 1, 0, 0, 2, 0, 2, 0xFC, 0xFD, 0xFE, 0xFF}, rsp3.ToBytes())

	// select status 3
	rsp4 := NewHSMSMessageSelectRsp(req4, 3)
	assert.Equal(t, "select.rsp", rsp4.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0xFF, 0xFF, 0, 3, 0, 2, 0xFF, 0xFF, 0xFF, 0xFF}, rsp4.ToBytes())
}

func TestHSMSControlMessage_DeselectReqRsp(t *testing.T) {
	req1 := NewHSMSMessageDeselectReq(0, []byte{0, 0, 0, 0})
	assert.Equal(t, "deselect.req", req1.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0}, req1.ToBytes())

	req2 := NewHSMSMessageDeselectReq(1, []byte{0, 0, 0, 1})
	assert.Equal(t, "deselect.req", req2.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0, 1, 0, 0, 0, 3, 0, 0, 0, 1}, req2.ToBytes())

	req3 := NewHSMSMessageDeselectReq(0xAABB, []byte{0xFC, 0xFD, 0xFE, 0xFF})
	assert.Equal(t, "deselect.req", req3.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0xAA, 0xBB, 0, 0, 0, 3, 0xFC, 0xFD, 0xFE, 0xFF}, req3.ToBytes())

	// deselect status 0
	rsp1 := NewHSMSMessageDeselectRsp(req1, 0)
	assert.Equal(t, "deselect.rsp", rsp1.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0}, rsp1.ToBytes())

	// deselect status 1
	rsp2 := NewHSMSMessageDeselectRsp(req2, 1)
	assert.Equal(t, "deselect.rsp", rsp2.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0, 1, 0, 1, 0, 4, 0, 0, 0, 1}, rsp2.ToBytes())

	// deselect status 2
	rsp3 := NewHSMSMessageDeselectRsp(req3, 2)
	assert.Equal(t, "deselect.rsp", rsp3.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0xAA, 0xBB, 0, 2, 0, 4, 0xFC, 0xFD, 0xFE, 0xFF}, rsp3.ToBytes())
}

func TestHSMSControlMessage_LinktestReqRsp(t *testing.T) {
	req1 := NewHSMSMessageLinktestReq([]byte{0, 0, 0, 0})
	assert.Equal(t, "linktest.req", req1.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0xFF, 0xFF, 0, 0, 0, 5, 0, 0, 0, 0}, req1.ToBytes())

	req2 := NewHSMSMessageLinktestReq([]byte{0xFC, 0xFD, 0xFE, 0xFF})
	assert.Equal(t, "linktest.req", req2.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0xFF, 0xFF, 0, 0, 0, 5, 0xFC, 0xFD, 0xFE, 0xFF}, req2.ToBytes())

	req3 := NewHSMSMessageLinktestReq([]byte{0xFF, 0xFF, 0xFF, 0xFF})
	assert.Equal(t, "linktest.req", req3.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0xFF, 0xFF, 0, 0, 0, 5, 0xFF, 0xFF, 0xFF, 0xFF}, req3.ToBytes())

	rsp1 := NewHSMSMessageLinktestRsp(req1)
	assert.Equal(t, "linktest.rsp", rsp1.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0xFF, 0xFF, 0, 0, 0, 6, 0, 0, 0, 0}, rsp1.ToBytes())

	rsp2 := NewHSMSMessageLinktestRsp(req2)
	assert.Equal(t, "linktest.rsp", rsp2.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0xFF, 0xFF, 0, 0, 0, 6, 0xFC, 0xFD, 0xFE, 0xFF}, rsp2.ToBytes())

	rsp3 := NewHSMSMessageLinktestRsp(req3)
	assert.Equal(t, "linktest.rsp", rsp3.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0xFF, 0xFF, 0, 0, 0, 6, 0xFF, 0xFF, 0xFF, 0xFF}, rsp3.ToBytes())
}

func TestHSMSControlMessage_RejectReq(t *testing.T) {
	// reason code 1, sType 8
	req1 := NewHSMSMessageRejectReq(0, 0, 8, []byte{0, 0, 0, 0}, 1)
	assert.Equal(t, "reject.req", req1.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0, 0, 8, 1, 0, 7, 0, 0, 0, 0}, req1.ToBytes())

	// reason code 2, pType 1
	req2 := NewHSMSMessageRejectReq(1, 1, 0, []byte{0, 0, 0, 1}, 2)
	assert.Equal(t, "reject.req", req2.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0, 1, 1, 2, 0, 7, 0, 0, 0, 1}, req2.ToBytes())

	// reason code 3, sType 9
	req3 := NewHSMSMessageRejectReq(0x1234, 0, 9, []byte{0xFC, 0xFD, 0xFE, 0xFF}, 3)
	assert.Equal(t, "reject.req", req3.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0x12, 0x34, 9, 3, 0, 7, 0xFC, 0xFD, 0xFE, 0xFF}, req3.ToBytes())

	// reason code 4
	req4 := NewHSMSMessageRejectReq(0xFFFF, 0, 0, []byte{0xFF, 0xFF, 0xFF, 0xFF}, 4)
	assert.Equal(t, "reject.req", req4.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0xFF, 0xFF, 0, 4, 0, 7, 0xFF, 0xFF, 0xFF, 0xFF}, req4.ToBytes())
}

func TestHSMSControlMessage_SeparateReq(t *testing.T) {
	req1 := NewHSMSMessageSeparateReq(0, []byte{0, 0, 0, 0})
	assert.Equal(t, "separate.req", req1.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0, 0, 0, 0, 0, 9, 0, 0, 0, 0}, req1.ToBytes())

	req2 := NewHSMSMessageSeparateReq(1, []byte{0, 0, 0, 1})
	assert.Equal(t, "separate.req", req2.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0, 1, 0, 0, 0, 9, 0, 0, 0, 1}, req2.ToBytes())

	req3 := NewHSMSMessageSeparateReq(0xFFFE, []byte{0x12, 0x34, 0x56, 0x78})
	assert.Equal(t, "separate.req", req3.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0xFF, 0xFE, 0, 0, 0, 9, 0x12, 0x34, 0x56, 0x78}, req3.ToBytes())

	req4 := NewHSMSMessageSeparateReq(0xFFFF, []byte{0xFF, 0xFF, 0xFF, 0xFF})
	assert.Equal(t, "separate.req", req4.Type())
	assert.Equal(t, []byte{0, 0, 0, 10, 0xFF, 0xFF, 0, 0, 0, 9, 0xFF, 0xFF, 0xFF, 0xFF}, req4.ToBytes())
}
