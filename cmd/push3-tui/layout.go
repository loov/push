package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/loov/logic-push3/push3"
)

// The layout is based on the ascii.txt reference file.
// We load the template, then stamp dynamic values into known positions.

// renderLayout renders the Push 3 TUI by starting from the ASCII template
// and filling in dynamic values (encoder values, pad states, etc).
func (m model) renderLayout() string {
	lines := loadTemplate()

	// Encoders: values on line 1 (0-indexed), positions from template "( 1 )" etc.
	// The template has "( 1 )" at specific columns. We overwrite the number.
	for i := range 8 {
		enc := push3.EncoderID(i + 1)
		val := fmt.Sprintf("%d", m.encoders[enc])
		// The template has "( 1 )" — value sits at col 31+i*6, 3 chars wide.
		col := 31 + i*6
		padded := fmt.Sprintf(" %-2s", val) // " 1 " style
		putStr(lines, 1, col, padded)
	}

	// Pad grid: each pad occupies a 5-char inner area.
	// Pad row 0 starts at template line 17, each pad row is 3 template lines.
	// Pad col 0 starts at template col 31, each pad col is 6 chars.
	// Inner content is on the middle line, cols +1 to +4.
	for pr := range 8 {
		for pc := range 8 {
			vel := m.pads[pr][pc]
			row := 17 + pr*3 // middle line of pad cell
			col := 31 + pc*6 // first inner col
			if vel > 0 {
				// Show velocity as a bar/block.
				putStr(lines, row, col, "█████")
			}
		}
	}

	return strings.Join(lines, "\n") + "\n"
}

// loadTemplate reads the ascii.txt template file.
// Falls back to a minimal template if the file is not found.
func loadTemplate() []string {
	data, err := os.ReadFile("ascii.txt")
	if err != nil {
		// Try relative to binary location.
		data, err = os.ReadFile("cmd/push3-tui/ascii.txt")
		if err != nil {
			return []string{"[ascii.txt not found]"}
		}
	}
	raw := strings.TrimRight(string(data), "\n")
	lines := strings.Split(raw, "\n")

	// Ensure all lines are padded to the same width for safe indexing.
	maxW := 0
	for _, l := range lines {
		if len(l) > maxW {
			maxW = len(l)
		}
	}
	for i := range lines {
		if len(lines[i]) < maxW {
			lines[i] += strings.Repeat(" ", maxW-len(lines[i]))
		}
	}
	return lines
}

// putStr overwrites characters in lines[row] starting at col.
func putStr(lines []string, row, col int, s string) {
	if row < 0 || row >= len(lines) {
		return
	}
	line := []rune(lines[row])
	for i, ch := range s {
		c := col + i
		if c >= 0 && c < len(line) {
			line[c] = ch
		}
	}
	lines[row] = string(line)
}
