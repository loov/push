package mcu

import (
	"strings"
)

// DetectMode infers the current assign mode from LCD content.
// It checks the first cell of the top LCD row for known mode keywords.
func DetectMode(lcd [2]LCDRow) AssignMode {
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

	return AssignModeUnknown
}

var modeKeywords = map[string]AssignMode{
	"volume": AssignModeTrack,
	"trkfmt": AssignModeTrack,
	"pan":    AssignModePan,
	"surr":   AssignModePan,
	"send":   AssignModeSend,
	"plug":   AssignModePlugin,
	"eq":     AssignModeEQ,
	"dyn":    AssignModeDynamics,
}

// DetectModeFromAssign maps an assign button note to a mode.
func DetectModeFromAssign(note byte) AssignMode {
	switch Button(note) {
	case AssignTrack:
		return AssignModeTrack
	case AssignSend:
		return AssignModeSend
	case AssignPan:
		return AssignModePan
	case AssignPlugin:
		return AssignModePlugin
	case AssignEQ:
		return AssignModeEQ
	case AssignInstrument:
		return AssignModeDynamics
	default:
		return AssignModeUnknown
	}
}
