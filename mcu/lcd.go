package mcu

import "github.com/loov/logic-push3/push3"

// ParseLCD extracts the position and text from an MCU LCD SysEx payload.
// The payload is the data between F0 and F7, validated as MCU SysEx.
// Returns the character position (0-111), the ASCII text, and whether parsing succeeded.
//
// LCD layout: 2 rows × 56 characters.
// Position 0-55 = top row, 56-111 = bottom row.
func ParseLCD(payload []byte) (position int, text string, ok bool) {
	if !IsMCUSysEx(payload) || SysExCommand(payload) != cmdLCD {
		return 0, "", false
	}
	if len(payload) < 7 {
		return 0, "", false
	}

	position = int(payload[5])
	// Text starts at byte 6
	textBytes := payload[6:]
	buf := make([]byte, len(textBytes))
	for i, b := range textBytes {
		if b >= 32 && b <= 126 {
			buf[i] = b
		} else {
			buf[i] = ' '
		}
	}
	return position, string(buf), true
}

// ApplyLCD writes LCD text into an LCDRow pair at the given position.
func ApplyLCD(lcd *[2]push3.LCDRow, position int, text string) {
	for i := 0; i < len(text); i++ {
		pos := position + i
		if pos < 0 {
			continue
		}
		row := 0
		col := pos
		if pos >= 56 {
			row = 1
			col = pos - 56
		}
		if col >= 56 {
			break
		}
		lcd[row][col] = text[i]
	}
}
