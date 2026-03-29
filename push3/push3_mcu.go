package push3

// MCU (Mackie Control Universal) protocol vocabulary types.

// MCUButton identifies an MCU protocol button by its MIDI note number.
type MCUButton uint8

// MCU transport buttons.
const (
	MCURewind  MCUButton = 91
	MCUFastFwd MCUButton = 92
	MCUStop    MCUButton = 93
	MCUPlay    MCUButton = 94
	MCURecord  MCUButton = 95
)

// MCU cursor buttons.
const (
	MCUCursorUp    MCUButton = 96
	MCUCursorDown  MCUButton = 97
	MCUCursorLeft  MCUButton = 98
	MCUCursorRight MCUButton = 99
)

// MCU channel strip buttons (base + channel 0-7).
const (
	MCURecArm0   MCUButton = 0  // 0-7
	MCUSolo0     MCUButton = 8  // 8-15
	MCUMute0     MCUButton = 16 // 16-23
	MCUSelect0   MCUButton = 24 // 24-31
	MCUVPotPush0 MCUButton = 32 // 32-39
)

// MCURecArm returns the REC ARM button for channel 0-7.
func MCURecArm(ch uint8) MCUButton { return MCURecArm0 + MCUButton(ch) }

// MCUSolo returns the SOLO button for channel 0-7.
func MCUSolo(ch uint8) MCUButton { return MCUSolo0 + MCUButton(ch) }

// MCUMute returns the MUTE button for channel 0-7.
func MCUMute(ch uint8) MCUButton { return MCUMute0 + MCUButton(ch) }

// MCUSelect returns the SELECT button for channel 0-7.
func MCUSelect(ch uint8) MCUButton { return MCUSelect0 + MCUButton(ch) }

// MCUVPotPush returns the V-Pot push button for channel 0-7.
func MCUVPotPush(ch uint8) MCUButton { return MCUVPotPush0 + MCUButton(ch) }

// MCU function keys.
const (
	MCUF1 MCUButton = 40
	MCUF2 MCUButton = 41
	MCUF3 MCUButton = 42
	MCUF4 MCUButton = 43
	MCUF5 MCUButton = 44
	MCUF6 MCUButton = 45
	MCUF7 MCUButton = 46
	MCUF8 MCUButton = 47
)

// MCU assign buttons.
const (
	MCUAssignTrack      MCUButton = 48
	MCUAssignSend       MCUButton = 49
	MCUAssignPan        MCUButton = 50
	MCUAssignPlugin     MCUButton = 51
	MCUAssignEQ         MCUButton = 52
	MCUAssignInstrument MCUButton = 53
)

// MCU modifier buttons.
const (
	MCUModShift  MCUButton = 54
	MCUModCtrl   MCUButton = 55
	MCUModOption MCUButton = 56
	MCUModAlt    MCUButton = 57
)

// MCU navigation.
const (
	MCUBankLeft     MCUButton = 68
	MCUBankRight    MCUButton = 69
	MCUChannelLeft  MCUButton = 70
	MCUChannelRight MCUButton = 71
)

// MCU miscellaneous.
const (
	MCUFlip  MCUButton = 50
	MCUZoom  MCUButton = 72
	MCUScrub MCUButton = 73
	MCUCycle MCUButton = 62
	MCUClick MCUButton = 65
)

// TransportState holds the current transport status from MCU.
type TransportState struct {
	Play   bool
	Stop   bool
	Record bool
	FFwd   bool
	Rew    bool
}

// TrackState holds the state of one MCU channel strip.
type TrackState struct {
	Name       string
	Mute       bool
	Solo       bool
	RecArm     bool
	Selected   bool
	FaderLevel uint16 // 14-bit, 0-16383
	Pan        int8   // -63 to +63
}

// LCDRow is one row of the MCU scribble strip (56 ASCII characters).
type LCDRow [56]byte

// String returns the LCD row as a trimmed string.
func (r LCDRow) String() string {
	return string(r[:])
}

// Cell returns the text for one of the 8 cells (7 chars each), trimmed.
func (r LCDRow) Cell(index int) string {
	if index < 0 || index > 7 {
		return ""
	}
	start := index * 7
	end := start + 7
	// Trim trailing spaces.
	for end > start && r[end-1] == ' ' {
		end--
	}
	return string(r[start:end])
}

// MCU SysEx protocol constants.
const (
	MCUModelIDLogic   byte = 0x14
	MCUModelIDGeneric byte = 0x12
)

// MCUSysExPrefix is the Mackie Control SysEx manufacturer prefix.
var MCUSysExPrefix = [3]byte{0x00, 0x00, 0x66}

// MCU assign mode, as reported by the host.
type MCUAssignMode uint8

const (
	MCUAssignModeUnknown MCUAssignMode = iota
	MCUAssignModeTrack
	MCUAssignModeSend
	MCUAssignModePan
	MCUAssignModePlugin
	MCUAssignModeEQ
	MCUAssignModeDynamics
)

func (m MCUAssignMode) String() string {
	switch m {
	case MCUAssignModeTrack:
		return "Track/Volume"
	case MCUAssignModeSend:
		return "Send"
	case MCUAssignModePan:
		return "Pan/Surround"
	case MCUAssignModePlugin:
		return "Plugin"
	case MCUAssignModeEQ:
		return "EQ"
	case MCUAssignModeDynamics:
		return "Dynamics"
	default:
		return "Unknown"
	}
}
