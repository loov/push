package mcu

// MCU (Mackie Control Universal) protocol vocabulary types.

// Button identifies an MCU protocol button by its MIDI note number.
type Button uint8

// LEDState represents the three-state LED mode from MCU button velocity.
type LEDState uint8

const (
	LEDOff   LEDState = iota // velocity 0x00
	LEDBlink                 // velocity odd (not 0x7F)
	LEDOn                    // velocity 0x7F
)

func (s LEDState) String() string {
	switch s {
	case LEDOff:
		return "Off"
	case LEDBlink:
		return "Blink"
	case LEDOn:
		return "On"
	default:
		return "Unknown"
	}
}

// LEDStateFromVelocity converts a MIDI velocity to an LED state.
func LEDStateFromVelocity(vel uint8) LEDState {
	if vel == 0 {
		return LEDOff
	}
	if vel == 0x7F {
		return LEDOn
	}
	if vel%2 == 1 {
		return LEDBlink
	}
	return LEDOff
}

// Transport buttons.
const (
	Rewind  Button = 91
	FastFwd Button = 92
	Stop    Button = 93
	Play    Button = 94
	Record  Button = 95
)

// Cursor buttons.
const (
	CursorUp    Button = 96
	CursorDown  Button = 97
	CursorLeft  Button = 98
	CursorRight Button = 99
)

// Channel strip buttons (base + channel 0-7).
const (
	RecArm0   Button = 0  // 0-7
	Solo0     Button = 8  // 8-15
	Mute0     Button = 16 // 16-23
	Select0   Button = 24 // 24-31
	VPotPush0 Button = 32 // 32-39
)

// RecArm returns the REC ARM button for channel 0-7.
func RecArm(ch uint8) Button { return RecArm0 + Button(ch) }

// Solo returns the SOLO button for channel 0-7.
func Solo(ch uint8) Button { return Solo0 + Button(ch) }

// Mute returns the MUTE button for channel 0-7.
func Mute(ch uint8) Button { return Mute0 + Button(ch) }

// Select returns the SELECT button for channel 0-7.
func Select(ch uint8) Button { return Select0 + Button(ch) }

// VPotPush returns the V-Pot push button for channel 0-7.
func VPotPush(ch uint8) Button { return VPotPush0 + Button(ch) }

// Function keys.
const (
	F1 Button = 40
	F2 Button = 41
	F3 Button = 42
	F4 Button = 43
	F5 Button = 44
	F6 Button = 45
	F7 Button = 46
	F8 Button = 47
)

// Assign buttons.
const (
	AssignTrack      Button = 48
	AssignSend       Button = 49
	AssignPan        Button = 50
	AssignPlugin     Button = 51
	AssignEQ         Button = 52
	AssignInstrument Button = 53
)

// Modifier buttons.
const (
	ModShift  Button = 54
	ModCtrl   Button = 55
	ModOption Button = 56
	ModAlt    Button = 57
)

// Navigation.
const (
	BankLeft     Button = 68
	BankRight    Button = 69
	ChannelLeft  Button = 70
	ChannelRight Button = 71
)

// Miscellaneous.
const (
	Flip  Button = 50
	Zoom  Button = 72
	Scrub Button = 73
	Cycle Button = 62
	Click Button = 65
)

// Automation buttons.
const (
	AutoRead  Button = 74
	AutoWrite Button = 75
	AutoTrim  Button = 76
	AutoTouch Button = 77
	AutoLatch Button = 78
	AutoGroup Button = 79
)

// File/edit buttons.
const (
	Save   Button = 80
	Undo   Button = 81
	Cancel Button = 82
	Enter  Button = 83
)

// Editing buttons.
const (
	Markers    Button = 84
	Nudge      Button = 85
	Drop       Button = 87
	Replace    Button = 88
	GlobalSolo Button = 90
)

// User switches.
const (
	UserSwitch1 Button = 102
	UserSwitch2 Button = 103
)

// Fader touch buttons (base + channel 0-7, 8 = master).
const FaderTouch0 Button = 104 // 104-112

// FaderTouch returns the fader touch button for channel 0-7, or 8 for master.
func FaderTouch(ch uint8) Button { return FaderTouch0 + Button(ch) }

// Display LEDs.
const (
	SMPTELED    Button = 113
	BeatsLED    Button = 114
	RudeSoloLED Button = 115
)

// Hardware.
const RelayClick Button = 118

// Meter level special values.
const (
	MeterSetOverload   uint8 = 0x0E // set channel overload indicator
	MeterClearOverload uint8 = 0x0F // clear channel overload indicator
)

// TransportState holds the current transport status.
type TransportState struct {
	Play   bool
	Stop   bool
	Record bool
	FFwd   bool
	Rew    bool
}

// TrackState holds the state of one channel strip.
type TrackState struct {
	Name       string
	Mute       bool
	Solo       bool
	RecArm     bool
	Selected   bool
	FaderLevel uint16 // 14-bit, 0-16383
	Pan        int8   // -63 to +63
	MeterLevel uint8  // 0x0-0xD signal level
	Overload   bool   // channel overload indicator
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

// SysEx device model IDs.
const (
	ModelIDLogicControl    byte = 0x10
	ModelIDLogicControlXT  byte = 0x11
	ModelIDMackieControl   byte = 0x14
	ModelIDMackieControlXT byte = 0x15
)

// SysExPrefix is the Mackie Control SysEx manufacturer prefix.
var SysExPrefix = [3]byte{0x00, 0x00, 0x66}

// AssignMode is the current assign mode, as reported by the host.
type AssignMode uint8

const (
	AssignModeUnknown AssignMode = iota
	AssignModeTrack
	AssignModeSend
	AssignModePan
	AssignModePlugin
	AssignModeEQ
	AssignModeDynamics
)

func (m AssignMode) String() string {
	switch m {
	case AssignModeTrack:
		return "Track/Volume"
	case AssignModeSend:
		return "Send"
	case AssignModePan:
		return "Pan/Surround"
	case AssignModePlugin:
		return "Plugin"
	case AssignModeEQ:
		return "EQ"
	case AssignModeDynamics:
		return "Dynamics"
	default:
		return "Unknown"
	}
}
