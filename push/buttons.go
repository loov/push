package push

import "github.com/loov/logic-push3/push3"

// buttonCCs is the set of all Push 3 CC numbers that correspond to buttons.
var buttonCCs = map[uint8]bool{}

func init() {
	for _, cc := range allButtonCCs {
		buttonCCs[cc] = true
	}
}

// isButtonCC returns true if the CC number corresponds to a Push 3 button.
func isButtonCC(cc uint8) bool {
	return buttonCCs[cc]
}

// All known Push 3 button CC numbers.
var allButtonCCs = []uint8{
	// Top-left row
	byte(push3.ButtonSets),
	byte(push3.ButtonSetup),
	byte(push3.ButtonLearn),
	byte(push3.ButtonUser),

	// Top-right row
	byte(push3.ButtonDevice),
	byte(push3.ButtonMix),
	byte(push3.ButtonClip),
	byte(push3.ButtonSession),

	// Display area
	byte(push3.ButtonUndo),
	byte(push3.ButtonSave),
	byte(push3.ButtonAdd),
	byte(push3.ButtonSwap),

	// Bottom-left row
	byte(push3.ButtonLock),
	byte(push3.ButtonStopClip),
	byte(push3.ButtonMute),
	byte(push3.ButtonSolo),

	// Transport / left side
	byte(push3.ButtonTapTempo),
	byte(push3.ButtonMetronome),
	byte(push3.ButtonQuantize),
	byte(push3.ButtonFixedLen),
	byte(push3.ButtonAutomate),
	byte(push3.ButtonNew),
	byte(push3.ButtonCapture),
	byte(push3.ButtonRecord),
	byte(push3.ButtonPlay),

	// Right side
	byte(push3.ButtonNote),
	byte(push3.ButtonSessionR),
	byte(push3.ButtonScale),
	byte(push3.ButtonLayout),
	byte(push3.ButtonRepeat),
	byte(push3.ButtonAccent),
	byte(push3.ButtonDoubleLoop),
	byte(push3.ButtonDuplicate),
	byte(push3.ButtonConvert),
	byte(push3.ButtonDelete),

	// D-pad
	byte(push3.ButtonUp),
	byte(push3.ButtonDown),
	byte(push3.ButtonLeft),
	byte(push3.ButtonRight),
	byte(push3.ButtonDPadCenter),

	// Navigation
	byte(push3.ButtonOctaveUp),
	byte(push3.ButtonOctaveDown),
	byte(push3.ButtonPageLeft),
	byte(push3.ButtonPageRight),

	// Bottom-right
	byte(push3.ButtonShift),
	byte(push3.ButtonSelect),

	// Encoder presses
	byte(push3.ButtonVolumePress),
	// Note: ButtonSwingTempoPress (CC 15) is NOT listed here because CC 15
	// is handled as EncoderSwingTempo rotation first. The click is detected
	// by value 127 on CC 15.

	// Jog wheel
	byte(push3.ButtonJogClick),
	byte(push3.ButtonJogPushLeft),
	byte(push3.ButtonJogPushRight),

	// Upper display buttons
	byte(push3.ButtonUpper1),
	byte(push3.ButtonUpper2),
	byte(push3.ButtonUpper3),
	byte(push3.ButtonUpper4),
	byte(push3.ButtonUpper5),
	byte(push3.ButtonUpper6),
	byte(push3.ButtonUpper7),
	byte(push3.ButtonUpper8),

	// Lower display buttons
	byte(push3.ButtonLower1),
	byte(push3.ButtonLower2),
	byte(push3.ButtonLower3),
	byte(push3.ButtonLower4),
	byte(push3.ButtonLower5),
	byte(push3.ButtonLower6),
	byte(push3.ButtonLower7),
	byte(push3.ButtonLower8),

	// Master
	byte(push3.ButtonMaster),

	// Time division
	byte(push3.ButtonDiv1_4),
	byte(push3.ButtonDiv1_4t),
	byte(push3.ButtonDiv1_8),
	byte(push3.ButtonDiv1_8t),
	byte(push3.ButtonDiv1_16),
	byte(push3.ButtonDiv1_16t),
	byte(push3.ButtonDiv1_32),
	byte(push3.ButtonDiv1_32t),
}
