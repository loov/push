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
	// Transport
	byte(push3.ButtonPlay),
	byte(push3.ButtonRecord),
	byte(push3.ButtonStop),
	byte(push3.ButtonDuplicate),

	// Navigation
	byte(push3.ButtonUp),
	byte(push3.ButtonDown),
	byte(push3.ButtonLeft),
	byte(push3.ButtonRight),

	// Mode
	byte(push3.ButtonShift),
	byte(push3.ButtonSelect),
	byte(push3.ButtonNote),
	byte(push3.ButtonSession),

	// Function
	byte(push3.ButtonQuantize),
	byte(push3.ButtonDelete),
	byte(push3.ButtonUndo),

	// Device/Browse/Mix/Clip
	byte(push3.ButtonDevice),
	byte(push3.ButtonBrowse),
	byte(push3.ButtonMix),
	byte(push3.ButtonClip),

	// Mute/Solo
	byte(push3.ButtonMute),
	byte(push3.ButtonSolo),

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
