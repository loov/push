package mcu

import (
	"strings"

	"github.com/loov/logic-push3/push3"
)

// DetectMode infers the current MCU assign mode from LCD content.
// It checks the first cell of the top LCD row for known mode keywords.
func DetectMode(lcd [2]push3.LCDRow) push3.MCUAssignMode {
	// Check the first cell (7 chars) of the top row for mode hints.
	cell0 := strings.ToLower(strings.TrimSpace(lcd[0].Cell(0)))

	for keyword, mode := range modeKeywords {
		if strings.Contains(cell0, keyword) {
			return mode
		}
	}

	// Also check the full top row for mode keywords if cell 0 didn't match.
	topRow := strings.ToLower(strings.TrimSpace(lcd[0].String()))
	for keyword, mode := range modeKeywords {
		if strings.Contains(topRow, keyword) {
			return mode
		}
	}

	return push3.MCUAssignModeUnknown
}

var modeKeywords = map[string]push3.MCUAssignMode{
	"volume": push3.MCUAssignModeTrack,
	"trkfmt": push3.MCUAssignModeTrack,
	"pan":    push3.MCUAssignModePan,
	"surr":   push3.MCUAssignModePan,
	"send":   push3.MCUAssignModeSend,
	"plug":   push3.MCUAssignModePlugin,
	"eq":     push3.MCUAssignModeEQ,
	"dyn":    push3.MCUAssignModeDynamics,
}

// DetectModeFromAssign maps an MCU assign button note to a mode.
func DetectModeFromAssign(note byte) push3.MCUAssignMode {
	switch push3.MCUButton(note) {
	case push3.MCUAssignTrack:
		return push3.MCUAssignModeTrack
	case push3.MCUAssignSend:
		return push3.MCUAssignModeSend
	case push3.MCUAssignPan:
		return push3.MCUAssignModePan
	case push3.MCUAssignPlugin:
		return push3.MCUAssignModePlugin
	case push3.MCUAssignEQ:
		return push3.MCUAssignModeEQ
	case push3.MCUAssignInstrument:
		return push3.MCUAssignModeDynamics
	default:
		return push3.MCUAssignModeUnknown
	}
}
