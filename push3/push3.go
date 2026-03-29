// Package push3 defines shared vocabulary types for the logic-push3 application.
package push3

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
	EncoderVolume // Large encoder on the left
	EncoderTempo  // Left half of Swing/Tempo encoder
	EncoderSwing  // Right half of Swing/Tempo encoder
	EncoderJog    // Jog wheel on the right
)

// EncoderCC returns the MIDI CC number for this encoder's rotation.
func (e EncoderID) EncoderCC() uint8 {
	switch {
	case e >= EncoderTrack1 && e <= EncoderTrack8:
		return 70 + uint8(e) // CC 71-78
	case e == EncoderVolume:
		return 79
	case e == EncoderTempo:
		return 14
	case e == EncoderSwing:
		return 15
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
		return EncoderTempo, true
	case cc == 15:
		return EncoderSwing, true
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
	TouchTempo uint8 = 10 // Tempo and Swing share the same physical knob.
	TouchJog   uint8 = 11
)

// EncoderTouchNote returns the MIDI note for this encoder's touch sensor.
func (e EncoderID) EncoderTouchNote() uint8 {
	switch {
	case e >= EncoderTrack1 && e <= EncoderTrack8:
		return TouchTrack1 + uint8(e) - 1
	case e == EncoderVolume:
		return TouchVolume
	case e == EncoderTempo, e == EncoderSwing:
		return TouchTempo
	case e == EncoderJog:
		return TouchJog
	default:
		return 0
	}
}

// DecodeRelative converts a two's complement CC value to a signed delta.
// Values 1-63 = clockwise, 65-127 = counter-clockwise.
func DecodeRelative(value uint8) int {
	if value < 64 {
		return int(value)
	}
	return int(value) - 128
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
	ButtonDevice  ButtonID = 110
	ButtonMix     ButtonID = 112
	ButtonClip    ButtonID = 113
	ButtonSession ButtonID = 34

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
	ButtonTapTempo   ButtonID = 3
	ButtonMetronome  ButtonID = 9
	ButtonQuantize   ButtonID = 116
	ButtonFixedLen   ButtonID = 90
	ButtonAutomate   ButtonID = 89
	ButtonNew        ButtonID = 92
	ButtonCapture    ButtonID = 65
	ButtonRecord     ButtonID = 86
	ButtonPlay       ButtonID = 85

	// Right side
	ButtonNote       ButtonID = 50
	ButtonSessionR   ButtonID = 51 // Session button on the right side
	ButtonScale      ButtonID = 58
	ButtonLayout     ButtonID = 31
	ButtonRepeat     ButtonID = 56
	ButtonAccent     ButtonID = 57
	ButtonDoubleLoop ButtonID = 117
	ButtonDuplicate  ButtonID = 88
	ButtonConvert    ButtonID = 35
	ButtonDelete     ButtonID = 118

	// D-pad
	ButtonUp    ButtonID = 46
	ButtonDown  ButtonID = 47
	ButtonLeft  ButtonID = 44
	ButtonRight ButtonID = 45

	// Navigation (Octave/Page)
	ButtonOctaveUp   ButtonID = 55
	ButtonOctaveDown ButtonID = 54
	ButtonPageLeft   ButtonID = 62
	ButtonPageRight  ButtonID = 63

	// Bottom-right
	ButtonShift  ButtonID = 49
	ButtonSelect ButtonID = 48

	// Encoder presses
	ButtonVolumePress ButtonID = 111 // Volume encoder click (was incorrectly ButtonBrowse)

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

	ButtonMaster ButtonID = 28

	// Time division / scene buttons
	ButtonDiv1_4   ButtonID = 36
	ButtonDiv1_4t  ButtonID = 37
	ButtonDiv1_8   ButtonID = 38
	ButtonDiv1_8t  ButtonID = 39
	ButtonDiv1_16  ButtonID = 40
	ButtonDiv1_16t ButtonID = 41
	ButtonDiv1_32  ButtonID = 42
	ButtonDiv1_32t ButtonID = 43
)

// Well-known colors from the Push 3 default palette.
var (
	ColorBlack     = Color{0, 0, 0}
	ColorWhite     = Color{255, 255, 255}
	ColorRed       = Color{255, 0, 0}
	ColorGreen     = Color{0, 255, 0}
	ColorBlue      = Color{0, 0, 255}
	ColorYellow    = Color{255, 255, 0}
	ColorOrange    = Color{255, 153, 0}
	ColorPurple    = Color{153, 0, 255}
	ColorCyan      = Color{0, 255, 255}
	ColorPink      = Color{255, 0, 255}
	ColorLime      = Color{153, 255, 0}
	ColorGray      = Color{128, 128, 128}
	ColorDarkGray  = Color{30, 30, 30}
)

// Default palette velocity indices.
const (
	PaletteBlack   uint8 = 0
	PaletteOrange  uint8 = 3
	PaletteYellow  uint8 = 8
	PaletteTurquoise uint8 = 15
	PalettePurple  uint8 = 22
	PalettePink    uint8 = 25
	PaletteWhite   uint8 = 122
	PaletteBlue    uint8 = 125
	PaletteGreen   uint8 = 126
	PaletteRed     uint8 = 127
)
