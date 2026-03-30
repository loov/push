package mcu

import (
	"strings"
)

// DetectMode infers the current MCU assign mode from LCD content.
// It checks the first cell of the top LCD row for known mode keywords.
func DetectMode(lcd [2]LCDRow) MCUAssignMode {
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

	return MCUAssignModeUnknown
}

var modeKeywords = map[string]MCUAssignMode{
	"volume": MCUAssignModeTrack,
	"trkfmt": MCUAssignModeTrack,
	"pan":    MCUAssignModePan,
	"surr":   MCUAssignModePan,
	"send":   MCUAssignModeSend,
	"plug":   MCUAssignModePlugin,
	"eq":     MCUAssignModeEQ,
	"dyn":    MCUAssignModeDynamics,
}

// DetectModeFromAssign maps an MCU assign button note to a mode.
func DetectModeFromAssign(note byte) MCUAssignMode {
	switch MCUButton(note) {
	case MCUAssignTrack:
		return MCUAssignModeTrack
	case MCUAssignSend:
		return MCUAssignModeSend
	case MCUAssignPan:
		return MCUAssignModePan
	case MCUAssignPlugin:
		return MCUAssignModePlugin
	case MCUAssignEQ:
		return MCUAssignModeEQ
	case MCUAssignInstrument:
		return MCUAssignModeDynamics
	default:
		return MCUAssignModeUnknown
	}
}
