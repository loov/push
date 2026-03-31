package mcu

// MCU SysEx commands.
const (
	cmdKeepalive       byte = 0x00
	cmdTransportClick  byte = 0x0A
	cmdLCDBacklight    byte = 0x0B
	cmdTouchlessFaders byte = 0x0C
	cmdFaderSensitivity byte = 0x0E
	cmdMetersDump      byte = 0x10 // 0x10-0x17
	cmdLCD             byte = 0x12
	cmdKeepaliveAck    byte = 0x13 // also firmware version request
	cmdFirmwareReply   byte = 0x14
	cmdSerialReq       byte = 0x1A
	cmdSerialReply     byte = 0x1B
	cmdMeterMode       byte = 0x20
	cmdGlobalMeterMode byte = 0x21
	cmdFadersToMin     byte = 0x61
	cmdAllLEDsOff      byte = 0x62
	cmdReset           byte = 0x63
)

// DeviceInquiry is the Universal Device Inquiry SysEx request.
// Many DAWs respond to this with their vendor/model header.
var DeviceInquiry = []byte{0xF0, 0x7E, 0x7F, 0x06, 0x01, 0xF7}

// DefaultSerial is the serial number bytes sent in the handshake reply.
var DefaultSerial = [7]byte{0xDE, 0xAD, 0xBE, 0xEF, 0x00, 0x00, 0x01}

// IsSysEx checks whether a SysEx payload (without F0/F7) starts with
// the Mackie Control manufacturer prefix and an accepted model ID.
func IsSysEx(payload []byte) bool {
	if len(payload) < 4 {
		return false
	}
	if payload[0] != SysExPrefix[0] ||
		payload[1] != SysExPrefix[1] ||
		payload[2] != SysExPrefix[2] {
		return false
	}
	modelID := payload[3]
	return modelID == ModelIDMackieControl ||
		modelID == ModelIDMackieControlXT ||
		modelID == ModelIDLogicControl ||
		modelID == ModelIDLogicControlXT
}

// SysExCommand returns the command byte from an MCU SysEx payload.
// The payload must have been validated with IsSysEx first.
func SysExCommand(payload []byte) byte {
	if len(payload) < 5 {
		return 0
	}
	return payload[4]
}

// EncodeSerialReply creates the SysEx response to a serial number request (0x1A).
func EncodeSerialReply(modelID byte, serial [7]byte) []byte {
	msg := []byte{0xF0}
	msg = append(msg, SysExPrefix[:]...)
	msg = append(msg, modelID, cmdSerialReply)
	msg = append(msg, serial[:]...)
	msg = append(msg, 0xF7)
	return msg
}

// EncodeKeepaliveAck creates the SysEx keepalive acknowledgement (0x13 0x00).
func EncodeKeepaliveAck(modelID byte) []byte {
	return []byte{
		0xF0,
		SysExPrefix[0], SysExPrefix[1], SysExPrefix[2],
		modelID, cmdKeepaliveAck, 0x00,
		0xF7,
	}
}

// SysExKind classifies a validated MCU SysEx payload.
type SysExKind uint8

const (
	SysExUnknown         SysExKind = iota
	SysExKeepalive                 // Host ping (cmd 0x00)
	SysExTransportClick            // Transport button click on/off (cmd 0x0A)
	SysExLCDBacklight              // LCD backlight saver timeout (cmd 0x0B)
	SysExTouchlessFaders           // Touchless movable faders mode (cmd 0x0C)
	SysExFaderSensitivity          // Fader touch sensitivity (cmd 0x0E)
	SysExMetersDump                // Meter levels (cmd 0x10-0x17)
	SysExLCD                       // LCD text update (cmd 0x12)
	SysExSerialReq                 // Serial number request (cmd 0x1A)
	SysExMeterMode                 // Channel meter mode (cmd 0x20)
	SysExGlobalMeterMode           // Global LCD meter mode (cmd 0x21)
	SysExFadersToMin               // Faders to minimum (cmd 0x61)
	SysExAllLEDsOff                // All LEDs off (cmd 0x62)
	SysExReset                     // Reset (cmd 0x63)
)

// String returns a human-readable name for the SysEx kind.
func (k SysExKind) String() string {
	switch k {
	case SysExKeepalive:
		return "Keepalive"
	case SysExTransportClick:
		return "TransportClick"
	case SysExLCDBacklight:
		return "LCDBacklight"
	case SysExTouchlessFaders:
		return "TouchlessFaders"
	case SysExFaderSensitivity:
		return "FaderSensitivity"
	case SysExMetersDump:
		return "MetersDump"
	case SysExLCD:
		return "LCD"
	case SysExSerialReq:
		return "SerialReq"
	case SysExMeterMode:
		return "MeterMode"
	case SysExGlobalMeterMode:
		return "GlobalMeterMode"
	case SysExFadersToMin:
		return "FadersToMin"
	case SysExAllLEDsOff:
		return "AllLEDsOff"
	case SysExReset:
		return "Reset"
	default:
		return "Unknown"
	}
}

// ClassifySysEx determines the kind of an MCU SysEx payload.
func ClassifySysEx(payload []byte) SysExKind {
	if !IsSysEx(payload) {
		return SysExUnknown
	}
	cmd := SysExCommand(payload)
	switch {
	case cmd == cmdKeepalive:
		return SysExKeepalive
	case cmd == cmdTransportClick:
		return SysExTransportClick
	case cmd == cmdLCDBacklight:
		return SysExLCDBacklight
	case cmd == cmdTouchlessFaders:
		return SysExTouchlessFaders
	case cmd == cmdFaderSensitivity:
		return SysExFaderSensitivity
	case cmd == cmdLCD:
		return SysExLCD
	case cmd == cmdKeepaliveAck: // 0x13 — firmware version request / keepalive ACK
		return SysExUnknown
	case cmd == cmdFirmwareReply: // 0x14 — firmware version reply
		return SysExUnknown
	case cmd >= cmdMetersDump && cmd <= cmdMetersDump+7:
		return SysExMetersDump
	case cmd == cmdSerialReq:
		return SysExSerialReq
	case cmd == cmdMeterMode:
		return SysExMeterMode
	case cmd == cmdGlobalMeterMode:
		return SysExGlobalMeterMode
	case cmd == cmdFadersToMin:
		return SysExFadersToMin
	case cmd == cmdAllLEDsOff:
		return SysExAllLEDsOff
	case cmd == cmdReset:
		return SysExReset
	default:
		return SysExUnknown
	}
}

// HandleHandshake processes MCU SysEx handshake messages and returns
// any response that should be sent back. Returns nil if no response is needed.
func HandleHandshake(payload []byte) []byte {
	if !IsSysEx(payload) {
		return nil
	}
	modelID := payload[3]
	cmd := SysExCommand(payload)

	switch cmd {
	case cmdKeepalive:
		return EncodeKeepaliveAck(modelID)
	case cmdSerialReq:
		if len(payload) >= 6 && payload[5] == 0x00 {
			return EncodeSerialReply(modelID, DefaultSerial)
		}
	}
	return nil
}
