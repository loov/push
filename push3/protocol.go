package push3

import "fmt"

// Color represents an RGB color for Push 3 LEDs and display.
type Color struct {
	R, G, B uint8
}

// PadPosition identifies a pad on the 8x8 grid.
type PadPosition struct {
	Row, Col uint8 // 0-7 each
}

// PadNote returns the MIDI note number for this pad position.
// Push 3 mapping: note = 92 - row*8 + col
func (p PadPosition) PadNote() uint8 {
	return 92 - p.Row*8 + p.Col
}

// PadPositionFromNote converts a MIDI note (36-99) to a pad position.
// Returns ok=false if the note is outside the pad range.
func PadPositionFromNote(note uint8) (PadPosition, bool) {
	if note < 36 || note > 99 {
		return PadPosition{}, false
	}
	row := (99 - note) / 8
	col := (note - 36) % 8
	return PadPosition{Row: row, Col: col}, true
}

// EncoderID identifies one of the rotary encoders.
type EncoderID uint8

// Encoder IDs (discovered via cmd/push3-discover).
const (
	EncoderTrack1 EncoderID = iota + 1
	EncoderTrack2
	EncoderTrack3
	EncoderTrack4
	EncoderTrack5
	EncoderTrack6
	EncoderTrack7
	EncoderTrack8
	EncoderVolume     // Large encoder on the left
	EncoderSwingTempo // Swing/Tempo encoder: rotation=CC 14, click=CC 15
	EncoderJog        // Jog wheel on the right
)

// String returns a human-readable name for the encoder.
func (e EncoderID) String() string {
	switch e {
	case EncoderTrack1:
		return "Track 1"
	case EncoderTrack2:
		return "Track 2"
	case EncoderTrack3:
		return "Track 3"
	case EncoderTrack4:
		return "Track 4"
	case EncoderTrack5:
		return "Track 5"
	case EncoderTrack6:
		return "Track 6"
	case EncoderTrack7:
		return "Track 7"
	case EncoderTrack8:
		return "Track 8"
	case EncoderVolume:
		return "Volume"
	case EncoderSwingTempo:
		return "Swing/Tempo"
	case EncoderJog:
		return "Jog Wheel"
	default:
		return fmt.Sprintf("Encoder(%d)", e)
	}
}

// EncoderCC returns the MIDI CC number for this encoder's rotation.
func (e EncoderID) EncoderCC() uint8 {
	switch {
	case e >= EncoderTrack1 && e <= EncoderTrack8:
		return 70 + uint8(e) // CC 71-78
	case e == EncoderVolume:
		return 79
	case e == EncoderSwingTempo:
		return 14 // Rotation is CC 14; click sends CC 15
	case e == EncoderJog:
		return 70
	default:
		return 0
	}
}

// EncoderFromCC returns the encoder for a given CC number.
// Returns ok=false if the CC doesn't map to an encoder.
func EncoderFromCC(cc uint8) (EncoderID, bool) {
	switch {
	case cc >= 71 && cc <= 78:
		return EncoderID(cc - 70), true
	case cc == 79:
		return EncoderVolume, true
	case cc == 14:
		return EncoderSwingTempo, true
	case cc == 70:
		return EncoderJog, true
	default:
		return 0, false
	}
}

// Encoder touch MIDI note numbers (discovered via cmd/push3-discover).
const (
	TouchTrack1 uint8 = 0
	TouchTrack2 uint8 = 1
	TouchTrack3 uint8 = 2
	TouchTrack4 uint8 = 3
	TouchTrack5 uint8 = 4
	TouchTrack6 uint8 = 5
	TouchTrack7 uint8 = 6
	TouchTrack8 uint8 = 7
	TouchVolume uint8 = 8
	// Note 9 is unused.
	TouchSwingTempo uint8 = 10 // Swing/Tempo encoder touch.
	TouchJog        uint8 = 11
	TouchTouchStrip uint8 = 12
	TouchDPadCenter uint8 = 13
)

// Touch strip sends Pitch Bend (channel 0) for position (0-16383)
// and Note 12 on/off for touch/release.

// EncoderTouchNote returns the MIDI note for this encoder's touch sensor.
func (e EncoderID) EncoderTouchNote() uint8 {
	switch {
	case e >= EncoderTrack1 && e <= EncoderTrack8:
		return TouchTrack1 + uint8(e) - 1
	case e == EncoderVolume:
		return TouchVolume
	case e == EncoderSwingTempo:
		return TouchSwingTempo
	case e == EncoderJog:
		return TouchJog
	default:
		return 0
	}
}

// DecodeRelative converts a sign-magnitude CC value to a signed delta.
// Bit 6 is the sign (0=positive, 1=negative), bits 0-5 are the magnitude.
// Values 1-63 = clockwise, 65-127 = counter-clockwise.
func DecodeRelative(value uint8) int {
	if value < 64 {
		return int(value)
	}
	return -int(value & 0x3F)
}

// ButtonID identifies a Push 3 button by its CC number.
type ButtonID uint8

// Push 3 button CC assignments (discovered via cmd/push3-discover).
const (
	// Top-left row
	ButtonSets  ButtonID = 80
	ButtonSetup ButtonID = 30
	ButtonLearn ButtonID = 81
	ButtonUser  ButtonID = 59

	// Top-right row
	ButtonDevice        ButtonID = 110
	ButtonMix           ButtonID = 112
	ButtonClip          ButtonID = 113
	ButtonSessionScreen ButtonID = 34

	// Display area
	ButtonUndo ButtonID = 119
	ButtonSave ButtonID = 82
	ButtonAdd  ButtonID = 32
	ButtonSwap ButtonID = 33

	// Bottom-left row
	ButtonLock     ButtonID = 83
	ButtonStopClip ButtonID = 29
	ButtonMute     ButtonID = 60
	ButtonSolo     ButtonID = 61

	// Transport / left side
	ButtonTapTempo    ButtonID = 3
	ButtonMetronome   ButtonID = 9
	ButtonQuantize    ButtonID = 116
	ButtonFixedLength ButtonID = 90
	ButtonAutomate    ButtonID = 89
	ButtonNew         ButtonID = 92
	ButtonCapture     ButtonID = 65
	ButtonRecord      ButtonID = 86
	ButtonPlay        ButtonID = 85

	// Right side
	ButtonNote       ButtonID = 50
	ButtonSessionPad ButtonID = 51 // Session button on the right side
	ButtonScale      ButtonID = 58
	ButtonLayout     ButtonID = 31
	ButtonRepeat     ButtonID = 56
	ButtonAccent     ButtonID = 57
	ButtonDoubleLoop ButtonID = 117
	ButtonDuplicate  ButtonID = 88
	ButtonConvert    ButtonID = 35
	ButtonDelete     ButtonID = 118

	// D-pad (directional CCs + center click CC 91, center touch Note 13)
	ButtonUp         ButtonID = 46
	ButtonDown       ButtonID = 47
	ButtonLeft       ButtonID = 44
	ButtonRight      ButtonID = 45
	ButtonDPadCenter ButtonID = 91

	// Navigation (Octave/Page)
	ButtonOctaveUp   ButtonID = 55
	ButtonOctaveDown ButtonID = 54
	ButtonPageLeft   ButtonID = 62
	ButtonPageRight  ButtonID = 63

	// Bottom-right
	ButtonShift  ButtonID = 49
	ButtonSelect ButtonID = 48

	// Encoder presses
	ButtonVolumePress     ButtonID = 111 // Volume encoder click
	ButtonSwingTempoPress ButtonID = 15  // Swing/Tempo encoder click (CC 15)

	// Jog wheel actions (discovered via push3-discover)
	ButtonJogClick     ButtonID = 94 // Press down
	ButtonJogPushLeft  ButtonID = 93 // Push sideways left
	ButtonJogPushRight ButtonID = 95 // Push sideways right
	// Jog rotation: CC 70 (EncoderJog), touch: Note 11

	// Upper display buttons
	ButtonUpper1 ButtonID = 102
	ButtonUpper2 ButtonID = 103
	ButtonUpper3 ButtonID = 104
	ButtonUpper4 ButtonID = 105
	ButtonUpper5 ButtonID = 106
	ButtonUpper6 ButtonID = 107
	ButtonUpper7 ButtonID = 108
	ButtonUpper8 ButtonID = 109

	// Lower display buttons
	ButtonLower1 ButtonID = 20
	ButtonLower2 ButtonID = 21
	ButtonLower3 ButtonID = 22
	ButtonLower4 ButtonID = 23
	ButtonLower5 ButtonID = 24
	ButtonLower6 ButtonID = 25
	ButtonLower7 ButtonID = 26
	ButtonLower8 ButtonID = 27

	ButtonMainTrack ButtonID = 28

	// Scene buttons (right column, top to bottom)
	ButtonScene1 ButtonID = 43 // labeled 1/32t
	ButtonScene2 ButtonID = 42 // labeled 1/32
	ButtonScene3 ButtonID = 41 // labeled 1/16t
	ButtonScene4 ButtonID = 40 // labeled 1/16
	ButtonScene5 ButtonID = 39 // labeled 1/8t
	ButtonScene6 ButtonID = 38 // labeled 1/8
	ButtonScene7 ButtonID = 37 // labeled 1/4t
	ButtonScene8 ButtonID = 36 // labeled 1/4
)

// String returns a human-readable name for the button.
func (b ButtonID) String() string {
	switch b {
	case ButtonSets:
		return "Sets"
	case ButtonSetup:
		return "Setup"
	case ButtonLearn:
		return "Learn"
	case ButtonUser:
		return "User"
	case ButtonDevice:
		return "Device"
	case ButtonMix:
		return "Mix"
	case ButtonClip:
		return "Clip"
	case ButtonSessionScreen:
		return "Session Screen"
	case ButtonUndo:
		return "Undo"
	case ButtonSave:
		return "Save"
	case ButtonAdd:
		return "Add"
	case ButtonSwap:
		return "Swap"
	case ButtonLock:
		return "Lock"
	case ButtonStopClip:
		return "Stop Clip"
	case ButtonMute:
		return "Mute"
	case ButtonSolo:
		return "Solo"
	case ButtonTapTempo:
		return "Tap Tempo"
	case ButtonMetronome:
		return "Metronome"
	case ButtonQuantize:
		return "Quantize"
	case ButtonFixedLength:
		return "Fixed Length"
	case ButtonAutomate:
		return "Automate"
	case ButtonNew:
		return "New"
	case ButtonCapture:
		return "Capture"
	case ButtonRecord:
		return "Record"
	case ButtonPlay:
		return "Play"
	case ButtonNote:
		return "Note"
	case ButtonSessionPad:
		return "Session Pad"
	case ButtonScale:
		return "Scale"
	case ButtonLayout:
		return "Layout"
	case ButtonRepeat:
		return "Repeat"
	case ButtonAccent:
		return "Accent"
	case ButtonDoubleLoop:
		return "Double Loop"
	case ButtonDuplicate:
		return "Duplicate"
	case ButtonConvert:
		return "Convert"
	case ButtonDelete:
		return "Delete"
	case ButtonUp:
		return "Up"
	case ButtonDown:
		return "Down"
	case ButtonLeft:
		return "Left"
	case ButtonRight:
		return "Right"
	case ButtonDPadCenter:
		return "D-Pad Center"
	case ButtonOctaveUp:
		return "Octave Up"
	case ButtonOctaveDown:
		return "Octave Down"
	case ButtonPageLeft:
		return "Page Left"
	case ButtonPageRight:
		return "Page Right"
	case ButtonShift:
		return "Shift"
	case ButtonSelect:
		return "Select"
	case ButtonVolumePress:
		return "Volume Press"
	case ButtonSwingTempoPress:
		return "Swing/Tempo Press"
	case ButtonJogClick:
		return "Jog Click"
	case ButtonJogPushLeft:
		return "Jog Push Left"
	case ButtonJogPushRight:
		return "Jog Push Right"
	case ButtonUpper1:
		return "Upper 1"
	case ButtonUpper2:
		return "Upper 2"
	case ButtonUpper3:
		return "Upper 3"
	case ButtonUpper4:
		return "Upper 4"
	case ButtonUpper5:
		return "Upper 5"
	case ButtonUpper6:
		return "Upper 6"
	case ButtonUpper7:
		return "Upper 7"
	case ButtonUpper8:
		return "Upper 8"
	case ButtonLower1:
		return "Lower 1"
	case ButtonLower2:
		return "Lower 2"
	case ButtonLower3:
		return "Lower 3"
	case ButtonLower4:
		return "Lower 4"
	case ButtonLower5:
		return "Lower 5"
	case ButtonLower6:
		return "Lower 6"
	case ButtonLower7:
		return "Lower 7"
	case ButtonLower8:
		return "Lower 8"
	case ButtonMainTrack:
		return "Main Track"
	case ButtonScene1:
		return "Scene 1"
	case ButtonScene2:
		return "Scene 2"
	case ButtonScene3:
		return "Scene 3"
	case ButtonScene4:
		return "Scene 4"
	case ButtonScene5:
		return "Scene 5"
	case ButtonScene6:
		return "Scene 6"
	case ButtonScene7:
		return "Scene 7"
	case ButtonScene8:
		return "Scene 8"
	default:
		return fmt.Sprintf("Button(CC %d)", b)
	}
}

// Animation determines how an LED transitions to a new color.
// The value maps directly to the MIDI channel used when setting the LED.
// Animations synchronize to MIDI clock messages (0xF8).
type Animation uint8

const (
	AnimStatic      Animation = 0  // Immediate color change, no animation
	AnimOneShot24   Animation = 1  // One-shot fade, 1/24 note
	AnimOneShot16   Animation = 2  // One-shot fade, 1/16 note
	AnimOneShot8    Animation = 3  // One-shot fade, 1/8 note
	AnimOneShot4    Animation = 4  // One-shot fade, 1/4 note
	AnimOneShotHalf Animation = 5  // One-shot fade, 1/2 note
	AnimPulse24     Animation = 6  // Continuous pulse, 1/24 note
	AnimPulse16     Animation = 7  // Continuous pulse, 1/16 note
	AnimPulse8      Animation = 8  // Continuous pulse, 1/8 note
	AnimPulse4      Animation = 9  // Continuous pulse, 1/4 note
	AnimPulseHalf   Animation = 10 // Continuous pulse, 1/2 note
	AnimBlink24     Animation = 11 // Continuous blink, 1/24 note
	AnimBlink16     Animation = 12 // Continuous blink, 1/16 note
	AnimBlink8      Animation = 13 // Continuous blink, 1/8 note
	AnimBlink4      Animation = 14 // Continuous blink, 1/4 note
	AnimBlinkHalf   Animation = 15 // Continuous blink, 1/2 note
)

// String returns a human-readable name for the animation.
func (a Animation) String() string {
	names := [16]string{
		"Static",
		"OneShot 1/24", "OneShot 1/16", "OneShot 1/8", "OneShot 1/4", "OneShot 1/2",
		"Pulse 1/24", "Pulse 1/16", "Pulse 1/8", "Pulse 1/4", "Pulse 1/2",
		"Blink 1/24", "Blink 1/16", "Blink 1/8", "Blink 1/4", "Blink 1/2",
	}
	if int(a) < len(names) {
		return names[a]
	}
	return fmt.Sprintf("Animation(%d)", a)
}

// TouchStripConfig holds touch strip configuration flags.
type TouchStripConfig struct {
	HostControl    bool // false=Push controls LEDs, true=host controls LEDs
	HostSendsSysEx bool // false=host sends values, true=host sends sysex
	ModWheel       bool // false=values as pitch bend, true=values as mod wheel
	PointMode      bool // false=LEDs show bar, true=LEDs show point
	BarFromCenter  bool // false=bar starts at bottom, true=bar starts at center
	AutoReturn     bool // false=no autoreturn, true=autoreturn enabled
	ReturnToCenter bool // false=return to bottom, true=return to center
}

// Well-known colors from the Push 3 default palette.
var (
	ColorBlack    = Color{0, 0, 0}
	ColorWhite    = Color{255, 255, 255}
	ColorRed      = Color{255, 0, 0}
	ColorGreen    = Color{0, 255, 0}
	ColorBlue     = Color{0, 0, 255}
	ColorYellow   = Color{255, 255, 0}
	ColorOrange   = Color{255, 153, 0}
	ColorPurple   = Color{153, 0, 255}
	ColorCyan     = Color{0, 255, 255}
	ColorPink     = Color{255, 0, 255}
	ColorLime     = Color{153, 255, 0}
	ColorGray     = Color{128, 128, 128}
	ColorDarkGray = Color{30, 30, 30}
)

// Default palette velocity indices.
const (
	PaletteBlack     uint8 = 0
	PaletteOrange    uint8 = 3
	PaletteYellow    uint8 = 8
	PaletteTurquoise uint8 = 15
	PalettePurple    uint8 = 22
	PalettePink      uint8 = 25
	PaletteWhite     uint8 = 122
	PaletteBlue      uint8 = 125
	PaletteGreen     uint8 = 126
	PaletteRed       uint8 = 127
)

// buttonCCs is the set of all Push 3 CC numbers that correspond to buttons.
var buttonCCs = map[uint8]bool{}

func init() {
	for _, id := range AllButtons {
		buttonCCs[byte(id)] = true
	}
}

// isButtonCC returns true if the CC number corresponds to a Push 3 button.
func isButtonCC(cc uint8) bool {
	return buttonCCs[cc]
}

// AllButtons lists all known Push 3 buttons.
var AllButtons = []ButtonID{
	// Top-left row
	ButtonSets,
	ButtonSetup,
	ButtonLearn,
	ButtonUser,

	// Top-right row
	ButtonDevice,
	ButtonMix,
	ButtonClip,
	ButtonSessionScreen,

	// Display area
	ButtonUndo,
	ButtonSave,
	ButtonAdd,
	ButtonSwap,

	// Bottom-left row
	ButtonLock,
	ButtonStopClip,
	ButtonMute,
	ButtonSolo,

	// Transport / left side
	ButtonTapTempo,
	ButtonMetronome,
	ButtonQuantize,
	ButtonFixedLength,
	ButtonAutomate,
	ButtonNew,
	ButtonCapture,
	ButtonRecord,
	ButtonPlay,

	// Right side
	ButtonNote,
	ButtonSessionPad,
	ButtonScale,
	ButtonLayout,
	ButtonRepeat,
	ButtonAccent,
	ButtonDoubleLoop,
	ButtonDuplicate,
	ButtonConvert,
	ButtonDelete,

	// D-pad
	ButtonUp,
	ButtonDown,
	ButtonLeft,
	ButtonRight,
	ButtonDPadCenter,

	// Navigation
	ButtonOctaveUp,
	ButtonOctaveDown,
	ButtonPageLeft,
	ButtonPageRight,

	// Bottom-right
	ButtonShift,
	ButtonSelect,

	// Encoder presses
	ButtonVolumePress,
	// Note: ButtonSwingTempoPress (CC 15) is NOT listed here because CC 15
	// is handled as EncoderSwingTempo rotation first. The click is detected
	// by value 127 on CC 15.

	// Jog wheel
	ButtonJogClick,
	ButtonJogPushLeft,
	ButtonJogPushRight,

	// Upper display buttons
	ButtonUpper1,
	ButtonUpper2,
	ButtonUpper3,
	ButtonUpper4,
	ButtonUpper5,
	ButtonUpper6,
	ButtonUpper7,
	ButtonUpper8,

	// Lower display buttons
	ButtonLower1,
	ButtonLower2,
	ButtonLower3,
	ButtonLower4,
	ButtonLower5,
	ButtonLower6,
	ButtonLower7,
	ButtonLower8,

	// Master
	ButtonMainTrack,

	// Time division
	ButtonScene1,
	ButtonScene2,
	ButtonScene3,
	ButtonScene4,
	ButtonScene5,
	ButtonScene6,
	ButtonScene7,
	ButtonScene8,
}

// encoderFromTouchNote maps a touch note to an EncoderID.
// Touch notes: 0-7 = Track 1-8, 8 = Volume, 10 = Tempo/Swing, 11 = Jog.
func encoderFromTouchNote(note uint8) (EncoderID, bool) {
	switch {
	case note <= 7:
		return EncoderID(note + 1), true // EncoderTrack1=1, note 0 → Track1
	case note == 8:
		return EncoderVolume, true
	case note == 10:
		return EncoderSwingTempo, true
	case note == 11:
		return EncoderJog, true
	default:
		return 0, false
	}
}
