package main

import (
	"fmt"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/loov/logic-push3/push3"
)

// A region that should be highlighted when a button is pressed.
// Covers all rows of the button box (top border through bottom border).
type region struct {
	rowStart, rowEnd int // inclusive range of rows
	col, width       int
	id               push3.ButtonID
}

// All interactive button regions, positions from ascii.txt (0-indexed).
// Each region covers the full inner width between │ borders, across all rows.
var buttonRegions = []region{
	// Top row (rows 3-5): left buttons — inner cols 2..6 (width 5) per cell
	// Top row — center (display buttons 1-8) — inner cols 31..33 (width 3)
	{3, 5, 2, 5, push3.ButtonUpper1}, {3, 5, 8, 5, push3.ButtonUpper2},
	{3, 5, 14, 5, push3.ButtonUpper3}, {3, 5, 20, 5, push3.ButtonUpper4},

	{3, 5, 31, 5, push3.ButtonUpper5}, {3, 5, 37, 5, push3.ButtonUpper6},
	{3, 5, 43, 5, push3.ButtonUpper7}, {3, 5, 49, 5, push3.ButtonUpper8},

	// Hmm wait, let me re-derive from the actual template positions.
}

func init() {
	// Clear and rebuild from scratch using exact positions from ascii.txt.
	buttonRegions = nil

	// Helper: button spanning rows [r0, r1] inclusive, inner content cols [c, c+w).
	add := func(r0, r1, c, w int, id push3.ButtonID) {
		buttonRegions = append(buttonRegions, region{r0, r1, c, w, id})
	}

	// ── Top button row (rows 3-5) ──
	// Left: Sets, Setup, Learn, User
	add(3, 5, 2, 5, push3.ButtonSets)
	add(3, 5, 8, 5, push3.ButtonSetup)
	add(3, 5, 14, 5, push3.ButtonLearn)
	add(3, 5, 20, 5, push3.ButtonUser)

	// Center: display buttons 1-8
	topIDs := [8]push3.ButtonID{
		push3.ButtonUpper1, push3.ButtonUpper2, push3.ButtonUpper3, push3.ButtonUpper4,
		push3.ButtonUpper5, push3.ButtonUpper6, push3.ButtonUpper7, push3.ButtonUpper8,
	}
	for i, id := range topIDs {
		add(3, 5, 30+i*6, 5, id)
	}

	// Right: Device, Mix, Clip, Session
	add(3, 5, 81, 5, push3.ButtonDevice)
	add(3, 5, 87, 5, push3.ButtonMix)
	add(3, 5, 93, 4, push3.ButtonClip)
	add(3, 5, 99, 5, push3.ButtonSession)

	// ── Display area buttons (rows 6-12) ──
	add(6, 8, 20, 5, push3.ButtonUndo)
	add(10, 12, 20, 5, push3.ButtonSave)
	add(6, 8, 82, 5, push3.ButtonAdd)
	add(10, 12, 81, 5, push3.ButtonSwap)

	// ── Bottom button row (rows 13-15) ──
	add(13, 15, 2, 5, push3.ButtonLock)
	add(13, 15, 8, 5, push3.ButtonStopClip)
	add(13, 15, 14, 5, push3.ButtonMute)
	add(13, 15, 20, 5, push3.ButtonSolo)

	// Center: display buttons 1-8
	botIDs := [8]push3.ButtonID{
		push3.ButtonLower1, push3.ButtonLower2, push3.ButtonLower3, push3.ButtonLower4,
		push3.ButtonLower5, push3.ButtonLower6, push3.ButtonLower7, push3.ButtonLower8,
	}
	for i, id := range botIDs {
		add(13, 15, 30+i*6, 5, id)
	}

	// Master
	add(13, 15, 81, 5, push3.ButtonMaster)

	// ── Left pad-area buttons ──
	// Swing/Tempo box (rows 16-18)
	// Tap/Tempo box (rows 19-22)
	add(16, 18, 3, 12, push3.ButtonTapTempo) // Swing/Tempo uses same button?
	add(19, 22, 3, 12, push3.ButtonTapTempo)
	// Metronome (row 23, standalone label)
	add(23, 23, 3, 13, push3.ButtonMetronome)
	// Quantize (row 24, standalone label)
	add(24, 24, 3, 12, push3.ButtonQuantize)
	// Fixed Length (row 26)
	add(26, 26, 3, 14, push3.ButtonFixedLen)
	// Automate (row 27)
	add(27, 27, 3, 12, push3.ButtonAutomate)
	// New (rows 29-31)
	add(29, 31, 3, 14, push3.ButtonNew)
	// Capture (rows 31-33)
	add(31, 33, 3, 14, push3.ButtonCapture)
	// Record (rows 34-37)
	add(34, 37, 3, 14, push3.ButtonRecord)
	// Play (rows 38-40)
	add(38, 40, 3, 14, push3.ButtonPlay)

	// ── Scene/repeat buttons (right col, rows 16-40) ──
	sceneIDs := [8]push3.ButtonID{
		push3.ButtonDiv1_32t, push3.ButtonDiv1_32,
		push3.ButtonDiv1_16t, push3.ButtonDiv1_16,
		push3.ButtonDiv1_8t, push3.ButtonDiv1_8,
		push3.ButtonDiv1_4t, push3.ButtonDiv1_4,
	}
	sceneRows := [8]int{16, 19, 22, 25, 28, 31, 34, 37}
	for i, id := range sceneIDs {
		add(sceneRows[i], sceneRows[i]+2, 81, 5, id)
	}

	// ── D-pad (rows 13-17, cols 92-99) ──
	add(14, 14, 93, 5, push3.ButtonUp)
	add(15, 15, 91, 1, push3.ButtonLeft)
	add(15, 15, 97, 1, push3.ButtonRight)
	add(16, 16, 93, 5, push3.ButtonDown)

	// ── Right side button pairs ──
	// Note/Session (rows 19-21)
	add(19, 21, 90, 6, push3.ButtonNote)
	add(19, 21, 97, 6, push3.ButtonSessionR)
	// Scale/Layout (rows 21-23)
	add(21, 23, 90, 6, push3.ButtonScale)
	add(21, 23, 97, 6, push3.ButtonLayout)
	// Repeat/Accent (rows 24-26)
	add(24, 26, 90, 6, push3.ButtonRepeat)
	add(24, 26, 97, 6, push3.ButtonAccent)
	// 2Loop/Dup (rows 26-28)
	add(26, 28, 90, 6, push3.ButtonDoubleLoop)
	add(26, 28, 97, 6, push3.ButtonDuplicate)
	// Conv/Del (rows 28-30)
	add(28, 30, 90, 6, push3.ButtonConvert)
	add(28, 30, 97, 6, push3.ButtonDelete)

	// ── Nav pad (rows 32-36, cols 92-99) ──
	add(33, 33, 93, 5, push3.ButtonOctaveUp)
	add(34, 34, 91, 1, push3.ButtonPageLeft)
	add(34, 34, 97, 1, push3.ButtonPageRight)
	add(35, 35, 93, 5, push3.ButtonOctaveDown)

	// Shift/Select (rows 38-40)
	add(38, 40, 90, 6, push3.ButtonShift)
	add(38, 40, 97, 6, push3.ButtonSelect)
}

// renderLayout renders the Push 3 TUI with highlighting for pressed controls.
func (m model) renderLayout() string {
	lines := loadTemplate()

	// Stamp encoder values.
	for i := range 8 {
		enc := push3.EncoderID(i + 1)
		val := fmt.Sprintf("%d", m.encoders[enc])
		col := 31 + i*6
		putStr(lines, 1, col, fmt.Sprintf(" %-2s", val))
	}

	// Stamp pad fills.
	for pr := range 8 {
		for pc := range 8 {
			vel := m.pads[pr][pc]
			if vel > 0 {
				col := 30 + pc*6
				// Each pad has 2 content rows.
				putStr(lines, 17+pr*3, col, "     ")
				putStr(lines, 18+pr*3, col, "     ")
			}
		}
	}

	// Build highlight map: for each (row, col) store a style index.
	highlights := map[[2]int]int{} // key: [row, col], value: style index
	var styles []lipgloss.Style

	addHighlight := func(row, col, width int, style lipgloss.Style) {
		idx := len(styles)
		styles = append(styles, style)
		for c := col; c < col+width; c++ {
			highlights[[2]int{row, c}] = idx
		}
	}

	// Encoder highlights.
	for i := range 8 {
		enc := push3.EncoderID(i + 1)
		if m.touched[enc] {
			addHighlight(1, 31+i*6, 3, encTouchedStyle)
		}
	}

	// Button highlights — cover all rows of the button box.
	for _, r := range buttonRegions {
		if r.id != 0 && m.buttons[r.id] {
			style := activeStyle(r.id)
			for row := r.rowStart; row <= r.rowEnd; row++ {
				addHighlight(row, r.col, r.width, style)
			}
		}
	}

	// Pad highlights.
	for pr := range 8 {
		for pc := range 8 {
			vel := m.pads[pr][pc]
			if vel > 0 {
				addHighlight(17+pr*3, 30+pc*6, 5, padVelStyle(vel))
				addHighlight(18+pr*3, 30+pc*6, 5, padVelStyle(vel))
			}
		}
	}

	// Render with highlighting.
	var out strings.Builder
	for row, line := range lines {
		runes := []rune(line)
		col := 0
		for col < len(runes) {
			if idx, ok := highlights[[2]int{row, col}]; ok {
				// Find run of consecutive chars with the same style index.
				end := col + 1
				for end < len(runes) {
					if idx2, ok2 := highlights[[2]int{row, end}]; !ok2 || idx2 != idx {
						break
					}
					end++
				}
				out.WriteString(styles[idx].Render(string(runes[col:end])))
				col = end
			} else {
				out.WriteRune(runes[col])
				col++
			}
		}
		out.WriteRune('\n')
	}
	return out.String()
}

func activeStyle(id push3.ButtonID) lipgloss.Style {
	switch id {
	case push3.ButtonPlay:
		return btnGreen
	case push3.ButtonRecord:
		return btnRed
	case push3.ButtonMute:
		return btnCyan
	case push3.ButtonSolo:
		return btnYellow
	default:
		return btnOn
	}
}

func padVelStyle(velocity uint8) lipgloss.Style {
	r, g, b := velocityToRGB(velocity)
	return lipgloss.NewStyle().
		Background(lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))).
		Foreground(lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b)))
}

func velocityToRGB(vel uint8) (r, g, b uint8) {
	if vel == 0 {
		return 0, 0, 0
	}
	f := float64(vel) / 127.0
	if f < 0.5 {
		t := f * 2
		return 0, uint8(t * 255), uint8((1 - t) * 255)
	}
	t := (f - 0.5) * 2
	return uint8(t * 255), uint8((1 - t) * 255), 0
}

// ─── Styles ───

var (
	encTouchedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF8800")).
			Bold(true)

	btnOn = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#FFFFFF")).
		Bold(true)

	btnGreen = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#00FF00")).
			Bold(true)

	btnRed = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#FF0000")).
		Bold(true)

	btnCyan = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#00FFFF")).
		Bold(true)

	btnYellow = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFFF00")).
			Bold(true)
)

// ─── Template loading ───

func loadTemplate() []string {
	data, err := os.ReadFile("ascii.txt")
	if err != nil {
		data, err = os.ReadFile("cmd/push3-tui/ascii.txt")
		if err != nil {
			return []string{"[ascii.txt not found]"}
		}
	}
	raw := strings.TrimRight(string(data), "\n")
	lines := strings.Split(raw, "\n")

	maxW := 0
	for _, l := range lines {
		if n := len([]rune(l)); n > maxW {
			maxW = n
		}
	}
	for i := range lines {
		runes := []rune(lines[i])
		if len(runes) < maxW {
			lines[i] = string(runes) + strings.Repeat(" ", maxW-len(runes))
		}
	}
	return lines
}

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
