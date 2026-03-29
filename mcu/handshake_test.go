package mcu

import (
	"testing"

	"github.com/loov/logic-push3"
)

func TestIsMCUSysEx(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		want    bool
	}{
		{"logic model", []byte{0x00, 0x00, 0x66, 0x14, 0x00}, true},
		{"generic model", []byte{0x00, 0x00, 0x66, 0x12, 0x00}, true},
		{"wrong prefix", []byte{0x00, 0x00, 0x67, 0x14, 0x00}, false},
		{"wrong model", []byte{0x00, 0x00, 0x66, 0x15, 0x00}, false},
		{"too short", []byte{0x00, 0x00, 0x66}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMCUSysEx(tt.payload); got != tt.want {
				t.Errorf("IsMCUSysEx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifySysEx(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		want    SysExKind
	}{
		{"keepalive", []byte{0x00, 0x00, 0x66, 0x14, 0x00}, SysExKeepalive},
		{"lcd", []byte{0x00, 0x00, 0x66, 0x14, 0x12, 0x00, 'A'}, SysExLCD},
		{"serial req", []byte{0x00, 0x00, 0x66, 0x14, 0x1A, 0x00}, SysExSerialReq},
		{"vpot ring", []byte{0x00, 0x00, 0x66, 0x14, 0x20, 0x03, 0x05}, SysExVPotRing},
		{"meters", []byte{0x00, 0x00, 0x66, 0x14, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, SysExMetersDump},
		{"unknown cmd", []byte{0x00, 0x00, 0x66, 0x14, 0x99}, SysExUnknown},
		{"not mcu", []byte{0x00, 0x00, 0x67, 0x14, 0x12}, SysExUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClassifySysEx(tt.payload); got != tt.want {
				t.Errorf("ClassifySysEx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleHandshake(t *testing.T) {
	t.Run("keepalive", func(t *testing.T) {
		payload := []byte{0x00, 0x00, 0x66, 0x14, 0x00}
		resp := HandleHandshake(payload)
		if resp == nil {
			t.Fatal("expected response for keepalive")
		}
		// Should be a keepalive ACK: F0 00 00 66 14 13 00 F7
		if resp[0] != 0xF0 || resp[len(resp)-1] != 0xF7 {
			t.Error("response should be wrapped in SysEx")
		}
		if resp[5] != 0x13 {
			t.Errorf("expected ACK command 0x13, got 0x%02X", resp[5])
		}
	})

	t.Run("serial request", func(t *testing.T) {
		payload := []byte{0x00, 0x00, 0x66, 0x14, 0x1A, 0x00}
		resp := HandleHandshake(payload)
		if resp == nil {
			t.Fatal("expected response for serial request")
		}
		if resp[5] != 0x1B {
			t.Errorf("expected serial reply command 0x1B, got 0x%02X", resp[5])
		}
	})

	t.Run("lcd is not handshake", func(t *testing.T) {
		payload := []byte{0x00, 0x00, 0x66, 0x14, 0x12, 0x00, 'A'}
		resp := HandleHandshake(payload)
		if resp != nil {
			t.Error("LCD should not trigger a handshake response")
		}
	})
}

func TestEncodeSerialReply(t *testing.T) {
	resp := EncodeSerialReply(push3.MCUModelIDLogic, DefaultSerial)
	if resp[0] != 0xF0 {
		t.Error("must start with F0")
	}
	if resp[len(resp)-1] != 0xF7 {
		t.Error("must end with F7")
	}
	if resp[4] != push3.MCUModelIDLogic {
		t.Errorf("model ID = 0x%02X, want 0x%02X", resp[4], push3.MCUModelIDLogic)
	}
	if resp[5] != 0x1B {
		t.Errorf("command = 0x%02X, want 0x1B", resp[5])
	}
	// Serial bytes should follow
	for i, b := range DefaultSerial {
		if resp[6+i] != b {
			t.Errorf("serial[%d] = 0x%02X, want 0x%02X", i, resp[6+i], b)
		}
	}
}
