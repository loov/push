// Command push3-tui shows a TUI mirroring the Push 3 layout.
// Buttons, pads, and encoders light up in real-time as you interact
// with the physical device.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/loov/logic-push3/midi"
	"github.com/loov/logic-push3/push"
	"github.com/loov/logic-push3/push3"
)

func main() {
	source := flag.String("source", push.SourceName, "Push 3 MIDI source name")
	dest := flag.String("dest", push.DestName, "Push 3 MIDI destination name")
	flag.Parse()

	client, err := midi.NewClient("push3-tui")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	p, err := push.Connect(client, *source, *dest)
	if err != nil {
		log.Fatal(err)
	}

	m := newModel(p)
	prog := tea.NewProgram(m)

	// Wire Push 3 callbacks to send tea messages.
	p.OnButton = func(id push3.ButtonID, pressed bool) {
		prog.Send(buttonMsg{id: id, pressed: pressed})
	}
	p.OnPad = func(pos push3.PadPosition, velocity uint8, pressed bool) {
		prog.Send(padMsg{pos: pos, velocity: velocity, pressed: pressed})
	}
	p.OnEncoder = func(id push3.EncoderID, delta int) {
		prog.Send(encoderMsg{id: id, delta: delta})
	}
	p.OnEncoderTouch = func(id push3.EncoderID, touched bool) {
		prog.Send(encoderTouchMsg{id: id, touched: touched})
	}

	if _, err := prog.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// --- Messages ---

type buttonMsg struct {
	id      push3.ButtonID
	pressed bool
}

type padMsg struct {
	pos      push3.PadPosition
	velocity uint8
	pressed  bool
}

type encoderMsg struct {
	id    push3.EncoderID
	delta int
}

type encoderTouchMsg struct {
	id      push3.EncoderID
	touched bool
}

// --- Model ---

type model struct {
	push     *push.Push3
	buttons  map[push3.ButtonID]bool
	pads     [8][8]uint8 // velocity (0 = not pressed)
	encoders [12]int     // accumulated value per encoder (index 0 unused, 1-11)
	touched  [12]bool    // encoder touch state
}

func newModel(p *push.Push3) model {
	return model{
		push:    p,
		buttons: make(map[push3.ButtonID]bool),
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case buttonMsg:
		m.buttons[msg.id] = msg.pressed
	case padMsg:
		m.pads[msg.pos.Row][msg.pos.Col] = msg.velocity
	case encoderMsg:
		idx := int(msg.id)
		if idx >= 0 && idx < len(m.encoders) {
			m.encoders[idx] += msg.delta
		}
	case encoderTouchMsg:
		idx := int(msg.id)
		if idx >= 0 && idx < len(m.touched) {
			m.touched[idx] = msg.touched
		}
	}
	return m, nil
}

// Layout constants.
const (
	padW = 6 // characters wide per pad cell
	padH = 2 // lines tall per pad cell
	encW = 8 // characters wide per encoder column
)

func (m model) View() tea.View {
	var b strings.Builder

	// ── Encoders row ──
	b.WriteString("  ")
	for i := range 8 {
		enc := push3.EncoderID(i + 1) // EncoderTrack1..8
		style := encStyle
		if m.touched[enc] {
			style = encTouchedStyle
		}
		b.WriteString(style.Render(fmt.Sprintf("  (%d) ", m.encoders[enc])))
		b.WriteString(" ")
	}
	b.WriteString("\n")

	// ── Upper display buttons ──
	b.WriteString("  ")
	upBtns := [8]push3.ButtonID{
		push3.ButtonUpper1, push3.ButtonUpper2, push3.ButtonUpper3, push3.ButtonUpper4,
		push3.ButtonUpper5, push3.ButtonUpper6, push3.ButtonUpper7, push3.ButtonUpper8,
	}
	for i, id := range upBtns {
		label := fmt.Sprintf(" Up%d ", i+1)
		b.WriteString(m.btn(id, label))
		b.WriteString(" ")
	}
	b.WriteString("\n")

	// ── Display area (placeholder) ──
	dispStyle := lipgloss.NewStyle().
		Width(8*encW - 1).
		Height(3).
		Background(lipgloss.Color("#0a0a0a")).
		Foreground(lipgloss.Color("#333333")).
		Align(lipgloss.Center, lipgloss.Center)
	b.WriteString("  ")
	b.WriteString(dispStyle.Render("[ display ]"))
	b.WriteString("\n")

	// ── Lower display buttons ──
	b.WriteString("  ")
	loBtns := [8]push3.ButtonID{
		push3.ButtonLower1, push3.ButtonLower2, push3.ButtonLower3, push3.ButtonLower4,
		push3.ButtonLower5, push3.ButtonLower6, push3.ButtonLower7, push3.ButtonLower8,
	}
	for i, id := range loBtns {
		label := fmt.Sprintf(" Lo%d ", i+1)
		b.WriteString(m.btn(id, label))
		b.WriteString(" ")
	}
	b.WriteString("\n\n")

	// ── Left buttons │ Pad grid │ Scene/repeat buttons ──
	leftCol := m.renderLeftButtons()
	padGrid := m.renderPadGrid()
	rightCol := m.renderRightButtons()

	// Join left, pads, right side by side.
	leftLines := strings.Split(leftCol, "\n")
	padLines := strings.Split(padGrid, "\n")
	rightLines := strings.Split(rightCol, "\n")

	maxLines := max(len(leftLines), len(padLines), len(rightLines))
	for i := range maxLines {
		left := ""
		if i < len(leftLines) {
			left = leftLines[i]
		}
		pad := ""
		if i < len(padLines) {
			pad = padLines[i]
		}
		right := ""
		if i < len(rightLines) {
			right = rightLines[i]
		}
		// Pad left column to fixed width.
		left = fixWidth(left, 12)
		b.WriteString(left)
		b.WriteString(pad)
		b.WriteString("  ")
		b.WriteString(right)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  q/ctrl+c to quit"))
	b.WriteString("\n")

	return tea.NewView(b.String())
}

// --- Button rendering ---

func (m model) btn(id push3.ButtonID, label string) string {
	if m.buttons[id] {
		switch id {
		case push3.ButtonPlay:
			return btnGreenStyle.Render(label)
		case push3.ButtonRecord:
			return btnRedStyle.Render(label)
		case push3.ButtonMute:
			return btnCyanStyle.Render(label)
		case push3.ButtonSolo:
			return btnYellowStyle.Render(label)
		default:
			return btnOnStyle.Render(label)
		}
	}
	return btnOffStyle.Render(label)
}

// --- Left side buttons (transport, etc.) ---

func (m model) renderLeftButtons() string {
	var b strings.Builder
	// Vertical stack mirroring left side of Push 3.
	btns := []struct {
		label string
		id    push3.ButtonID
	}{
		{"Mute  ", push3.ButtonMute},
		{"Solo  ", push3.ButtonSolo},
		{"      ", 0}, // spacer
		{"Stop  ", push3.ButtonStop},
		{" ▶ Ply", push3.ButtonPlay},
		{" ● Rec", push3.ButtonRecord},
		{"      ", 0}, // spacer
		{"Undo  ", push3.ButtonUndo},
	}
	for _, entry := range btns {
		if entry.id == 0 {
			b.WriteString("          \n")
			continue
		}
		b.WriteString(m.btn(entry.id, entry.label))
		b.WriteString("\n")
	}
	return b.String()
}

// --- Right side buttons (scene/repeat, nav) ---

func (m model) renderRightButtons() string {
	var b strings.Builder
	// Time division / scene buttons stacked vertically on the right.
	sceneBtns := []struct {
		label string
		id    push3.ButtonID
	}{
		{" 1/32t", push3.ButtonDiv1_32t},
		{" 1/32 ", push3.ButtonDiv1_32},
		{" 1/16t", push3.ButtonDiv1_16t},
		{" 1/16 ", push3.ButtonDiv1_16},
		{" 1/8t ", push3.ButtonDiv1_8t},
		{" 1/8  ", push3.ButtonDiv1_8},
		{" 1/4t ", push3.ButtonDiv1_4t},
		{" 1/4  ", push3.ButtonDiv1_4},
	}
	for _, entry := range sceneBtns {
		b.WriteString(m.btn(entry.id, entry.label))
		b.WriteString("\n")
	}
	return b.String()
}

// --- Pad grid ---

func (m model) renderPadGrid() string {
	var b strings.Builder
	padCell := strings.Repeat(" ", padW)
	for row := range 8 {
		for col := range 8 {
			vel := m.pads[row][col]
			style := padOffStyle
			if vel > 0 {
				style = padVelStyle(vel)
			}
			b.WriteString(style.Render(padCell))
			if col < 7 {
				b.WriteString(" ")
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

// --- Navigation/control buttons row (below pads) ---

// --- Styles ---

var (
	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	btnOffStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555")).
			Background(lipgloss.Color("#1a1a1a"))

	btnOnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFFFFF")).
			Bold(true)

	btnGreenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#00FF00")).
			Bold(true)

	btnRedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#FF0000")).
			Bold(true)

	btnCyanStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#00FFFF")).
			Bold(true)

	btnYellowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFFF00")).
			Bold(true)

	encStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555")).
			Background(lipgloss.Color("#111111")).
			Width(encW - 1).
			Align(lipgloss.Center)

	encTouchedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF8800")).
			Background(lipgloss.Color("#222222")).
			Width(encW - 1).
			Bold(true).
			Align(lipgloss.Center)

	padOffStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#111111"))
)

func padVelStyle(velocity uint8) lipgloss.Style {
	r, g, b := velocityToRGB(velocity)
	bg := lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
	return lipgloss.NewStyle().Background(bg)
}

func velocityToRGB(vel uint8) (r, g, b uint8) {
	if vel == 0 {
		return 0, 0, 0
	}
	f := float64(vel) / 127.0
	switch {
	case f < 0.5:
		t := f * 2
		return 0, uint8(t * 255), uint8((1 - t) * 255)
	default:
		t := (f - 0.5) * 2
		return uint8(t * 255), uint8((1 - t) * 255), 0
	}
}

func fixWidth(s string, w int) string {
	// Crude: pad or truncate to w visible characters.
	n := lipgloss.Width(s)
	if n >= w {
		return s
	}
	return s + strings.Repeat(" ", w-n)
}
