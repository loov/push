package push3

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
	byte(ButtonSets),
	byte(ButtonSetup),
	byte(ButtonLearn),
	byte(ButtonUser),

	// Top-right row
	byte(ButtonDevice),
	byte(ButtonMix),
	byte(ButtonClip),
	byte(ButtonSession),

	// Display area
	byte(ButtonUndo),
	byte(ButtonSave),
	byte(ButtonAdd),
	byte(ButtonSwap),

	// Bottom-left row
	byte(ButtonLock),
	byte(ButtonStopClip),
	byte(ButtonMute),
	byte(ButtonSolo),

	// Transport / left side
	byte(ButtonTapTempo),
	byte(ButtonMetronome),
	byte(ButtonQuantize),
	byte(ButtonFixedLen),
	byte(ButtonAutomate),
	byte(ButtonNew),
	byte(ButtonCapture),
	byte(ButtonRecord),
	byte(ButtonPlay),

	// Right side
	byte(ButtonNote),
	byte(ButtonSessionR),
	byte(ButtonScale),
	byte(ButtonLayout),
	byte(ButtonRepeat),
	byte(ButtonAccent),
	byte(ButtonDoubleLoop),
	byte(ButtonDuplicate),
	byte(ButtonConvert),
	byte(ButtonDelete),

	// D-pad
	byte(ButtonUp),
	byte(ButtonDown),
	byte(ButtonLeft),
	byte(ButtonRight),
	byte(ButtonDPadCenter),

	// Navigation
	byte(ButtonOctaveUp),
	byte(ButtonOctaveDown),
	byte(ButtonPageLeft),
	byte(ButtonPageRight),

	// Bottom-right
	byte(ButtonShift),
	byte(ButtonSelect),

	// Encoder presses
	byte(ButtonVolumePress),
	// Note: ButtonSwingTempoPress (CC 15) is NOT listed here because CC 15
	// is handled as EncoderSwingTempo rotation first. The click is detected
	// by value 127 on CC 15.

	// Jog wheel
	byte(ButtonJogClick),
	byte(ButtonJogPushLeft),
	byte(ButtonJogPushRight),

	// Upper display buttons
	byte(ButtonUpper1),
	byte(ButtonUpper2),
	byte(ButtonUpper3),
	byte(ButtonUpper4),
	byte(ButtonUpper5),
	byte(ButtonUpper6),
	byte(ButtonUpper7),
	byte(ButtonUpper8),

	// Lower display buttons
	byte(ButtonLower1),
	byte(ButtonLower2),
	byte(ButtonLower3),
	byte(ButtonLower4),
	byte(ButtonLower5),
	byte(ButtonLower6),
	byte(ButtonLower7),
	byte(ButtonLower8),

	// Master
	byte(ButtonMaster),

	// Time division
	byte(ButtonDiv1_4),
	byte(ButtonDiv1_4t),
	byte(ButtonDiv1_8),
	byte(ButtonDiv1_8t),
	byte(ButtonDiv1_16),
	byte(ButtonDiv1_16t),
	byte(ButtonDiv1_32),
	byte(ButtonDiv1_32t),
}
