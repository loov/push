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
	pads     [8][8]uint8
	encoders [12]int
	touched  [12]bool
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
		if idx > 0 && idx < len(m.encoders) {
			m.encoders[idx] += msg.delta
		}
	case encoderTouchMsg:
		idx := int(msg.id)
		if idx > 0 && idx < len(m.touched) {
			m.touched[idx] = msg.touched
		}
	}
	return m, nil
}

// Grid dimensions.
const (
	colW = 8 // width of one grid column (encoder, pad, button)
)

func (m model) View() tea.View {
	var b strings.Builder

	// ── Row: Encoders ──
	b.WriteString(m.renderEncoderRow())
	b.WriteString("\n")

	// ── Row: Upper display buttons ──
	b.WriteString(m.renderDisplayBtnRow(
		push3.ButtonUpper1, push3.ButtonUpper2, push3.ButtonUpper3, push3.ButtonUpper4,
		push3.ButtonUpper5, push3.ButtonUpper6, push3.ButtonUpper7, push3.ButtonUpper8,
	))
	b.WriteString("\n")

	// ── Row: Display placeholder ──
	w := 8*colW + 7 // 8 columns + 7 gaps
	b.WriteString(displayStyle.Width(w).Render(""))
	b.WriteString("\n")

	// ── Row: Lower display buttons ──
	b.WriteString(m.renderDisplayBtnRow(
		push3.ButtonLower1, push3.ButtonLower2, push3.ButtonLower3, push3.ButtonLower4,
		push3.ButtonLower5, push3.ButtonLower6, push3.ButtonLower7, push3.ButtonLower8,
	))
	b.WriteString("\n\n")

	// ── Main section: left buttons | pad grid | right buttons ──
	left := m.renderLeftCol()
	pads := m.renderPadGrid()
	right := m.renderRightCol()
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", pads, "  ", right))

	b.WriteString("\n\n")

	// ── Bottom: nav + mode buttons ──
	b.WriteString(m.renderBottomRow())

	b.WriteString("\n")
	b.WriteString(dimStyle.Render(" q/ctrl+c to quit"))
	b.WriteString("\n")

	return tea.NewView(b.String())
}

// --- Encoder row ---

func (m model) renderEncoderRow() string {
	var cells []string
	for i := range 8 {
		enc := push3.EncoderID(i + 1)
		style := encStyle
		if m.touched[enc] {
			style = encTouchedStyle
		}
		val := m.encoders[enc]
		cells = append(cells, style.Render(fmt.Sprintf("%4d", val)))
	}
	return strings.Join(cells, " ")
}

// --- Display button row ---

func (m model) renderDisplayBtnRow(ids ...push3.ButtonID) string {
	var cells []string
	for i, id := range ids {
		label := fmt.Sprintf("%d", i+1)
		cells = append(cells, m.renderBtn(id, label))
	}
	return strings.Join(cells, " ")
}

// --- Left column: transport + mute/solo ---

func (m model) renderLeftCol() string {
	// 8 rows to match 8 pad rows. Each row is 1 line.
	leftW := 8
	rows := [8]string{
		m.renderBtn(push3.ButtonMute, "Mute"),
		m.renderBtn(push3.ButtonSolo, "Solo"),
		"",
		m.renderBtn(push3.ButtonStop, "Stop"),
		m.renderBtn(push3.ButtonPlay, " ▶  "),
		m.renderBtn(push3.ButtonRecord, " ●  "),
		"",
		m.renderBtn(push3.ButtonUndo, "Undo"),
	}
	var lines []string
	for _, r := range rows {
		if r == "" {
			lines = append(lines, strings.Repeat(" ", leftW))
		} else {
			lines = append(lines, padRight(r, leftW))
		}
	}
	return strings.Join(lines, "\n")
}

// --- Right column: scene/repeat intervals ---

func (m model) renderRightCol() string {
	rightW := 6
	labels := [8]struct {
		label string
		id    push3.ButtonID
	}{
		{"1/32t", push3.ButtonDiv1_32t},
		{"1/32", push3.ButtonDiv1_32},
		{"1/16t", push3.ButtonDiv1_16t},
		{"1/16", push3.ButtonDiv1_16},
		{"1/8t", push3.ButtonDiv1_8t},
		{"1/8", push3.ButtonDiv1_8},
		{"1/4t", push3.ButtonDiv1_4t},
		{"1/4", push3.ButtonDiv1_4},
	}
	var lines []string
	for _, entry := range labels {
		lines = append(lines, padRight(m.renderBtn(entry.id, entry.label), rightW))
	}
	return strings.Join(lines, "\n")
}

// --- 8x8 pad grid ---

func (m model) renderPadGrid() string {
	var rows []string
	for row := range 8 {
		var cells []string
		for col := range 8 {
			vel := m.pads[row][col]
			style := padOffStyle
			if vel > 0 {
				style = padVelStyle(vel)
			}
			cells = append(cells, style.Render("      "))
		}
		rows = append(rows, strings.Join(cells, " "))
	}
	return strings.Join(rows, "\n")
}

// --- Bottom row: nav + mode buttons ---

func (m model) renderBottomRow() string {
	nav := []struct {
		label string
		id    push3.ButtonID
	}{
		{"Shft", push3.ButtonShift},
		{"Sel", push3.ButtonSelect},
		{"  ◄ ", push3.ButtonLeft},
		{"  ▲ ", push3.ButtonUp},
		{"  ▼ ", push3.ButtonDown},
		{"  ► ", push3.ButtonRight},
		{"Note", push3.ButtonNote},
		{"Sess", push3.ButtonSession},
		{"Dev", push3.ButtonDevice},
		{"Mix", push3.ButtonMix},
		{"Brws", push3.ButtonBrowse},
		{"Clip", push3.ButtonClip},
		{"Del", push3.ButtonDelete},
		{"Qntz", push3.ButtonQuantize},
		{"Dup", push3.ButtonDuplicate},
	}
	var cells []string
	for _, n := range nav {
		cells = append(cells, m.renderBtn(n.id, n.label))
	}
	return " " + strings.Join(cells, " ")
}

// --- Button rendering ---

func (m model) renderBtn(id push3.ButtonID, label string) string {
	style := btnOffStyle
	if m.buttons[id] {
		switch id {
		case push3.ButtonPlay:
			style = btnGreenStyle
		case push3.ButtonRecord:
			style = btnRedStyle
		case push3.ButtonMute:
			style = btnCyanStyle
		case push3.ButtonSolo:
			style = btnYellowStyle
		default:
			style = btnOnStyle
		}
	}
	return style.Width(colW).Align(lipgloss.Center).Render(label)
}

// --- Styles ---

var (
	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	displayStyle = lipgloss.NewStyle().
			Height(3).
			Background(lipgloss.Color("#080808")).
			Foreground(lipgloss.Color("#222222")).
			Align(lipgloss.Center, lipgloss.Center)

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
			Width(colW).
			Foreground(lipgloss.Color("#555555")).
			Background(lipgloss.Color("#111111")).
			Align(lipgloss.Center)

	encTouchedStyle = lipgloss.NewStyle().
			Width(colW).
			Foreground(lipgloss.Color("#FF8800")).
			Background(lipgloss.Color("#222222")).
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
	if f < 0.5 {
		t := f * 2
		return 0, uint8(t * 255), uint8((1 - t) * 255)
	}
	t := (f - 0.5) * 2
	return uint8(t * 255), uint8((1 - t) * 255), 0
}

func padRight(s string, w int) string {
	n := lipgloss.Width(s)
	if n >= w {
		return s
	}
	return s + strings.Repeat(" ", w-n)
}
