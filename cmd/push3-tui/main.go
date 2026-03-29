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
	encoders [11]int     // accumulated value per encoder
	touched  [11]bool    // encoder touch state
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
		idx := int(msg.id) - 1
		if idx >= 0 && idx < len(m.encoders) {
			m.encoders[idx] += msg.delta
		}
	case encoderTouchMsg:
		idx := int(msg.id) - 1
		if idx >= 0 && idx < len(m.touched) {
			m.touched[idx] = msg.touched
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  ABLETON PUSH 3  "))
	b.WriteString("\n\n")

	// Upper display buttons
	b.WriteString(m.renderButtonRow(
		[]namedButton{
			{"Up1", push3.ButtonUpper1}, {"Up2", push3.ButtonUpper2},
			{"Up3", push3.ButtonUpper3}, {"Up4", push3.ButtonUpper4},
			{"Up5", push3.ButtonUpper5}, {"Up6", push3.ButtonUpper6},
			{"Up7", push3.ButtonUpper7}, {"Up8", push3.ButtonUpper8},
		},
	))

	// Encoders
	b.WriteString(m.renderEncoders())

	// Lower display buttons
	b.WriteString(m.renderButtonRow(
		[]namedButton{
			{"Lo1", push3.ButtonLower1}, {"Lo2", push3.ButtonLower2},
			{"Lo3", push3.ButtonLower3}, {"Lo4", push3.ButtonLower4},
			{"Lo5", push3.ButtonLower5}, {"Lo6", push3.ButtonLower6},
			{"Lo7", push3.ButtonLower7}, {"Lo8", push3.ButtonLower8},
		},
	))

	b.WriteString("\n")

	// Navigation + transport + mode buttons
	b.WriteString(m.renderControlButtons())
	b.WriteString("\n")

	// 8x8 pad grid
	b.WriteString(m.renderPadGrid())
	b.WriteString("\n")

	// Time division buttons
	b.WriteString(m.renderButtonRow(
		[]namedButton{
			{"1/4", push3.ButtonDiv1_4}, {"1/4t", push3.ButtonDiv1_4t},
			{"1/8", push3.ButtonDiv1_8}, {"1/8t", push3.ButtonDiv1_8t},
			{"1/16", push3.ButtonDiv1_16}, {"1/16t", push3.ButtonDiv1_16t},
			{"1/32", push3.ButtonDiv1_32}, {"1/32t", push3.ButtonDiv1_32t},
		},
	))

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  q/ctrl+c to quit"))
	b.WriteString("\n")

	return tea.NewView(b.String())
}

// --- Rendering helpers ---

type namedButton struct {
	label string
	id    push3.ButtonID
}

func (m model) renderButtonRow(buttons []namedButton) string {
	var parts []string
	for _, btn := range buttons {
		style := btnOffStyle
		if m.buttons[btn.id] {
			style = btnOnStyle
		}
		parts = append(parts, style.Render(centerPad(btn.label, 5)))
	}
	return " " + strings.Join(parts, " ") + "\n"
}

func (m model) renderEncoders() string {
	labels := []string{"Trk1", "Trk2", "Trk3", "Trk4", "Trk5", "Trk6", "Trk7", "Trk8"}
	var parts []string
	for i, label := range labels {
		val := m.encoders[i]
		style := encOffStyle
		if m.touched[i] {
			style = encOnStyle
		}
		text := fmt.Sprintf("%s\n%4d", label, val)
		parts = append(parts, style.Render(text))
	}
	return " " + strings.Join(parts, " ") + "\n"
}

func (m model) renderControlButtons() string {
	var b strings.Builder

	// Row 1: Navigation
	nav := []namedButton{
		{"  ▲ ", push3.ButtonUp},
	}
	b.WriteString(" ")
	b.WriteString(m.renderInlineButtons(nav))

	// Mode buttons on same line
	mode := []namedButton{
		{"Note", push3.ButtonNote}, {"Sess", push3.ButtonSession},
		{"Mix", push3.ButtonMix}, {"Brws", push3.ButtonBrowse},
		{"Dev", push3.ButtonDevice}, {"Clip", push3.ButtonClip},
	}
	b.WriteString("  ")
	b.WriteString(m.renderInlineButtons(mode))
	b.WriteString("\n")

	// Row 2: Left/Down/Right + Mute/Solo
	nav2 := []namedButton{
		{"  ◄ ", push3.ButtonLeft}, {"  ▼ ", push3.ButtonDown}, {"  ► ", push3.ButtonRight},
	}
	b.WriteString(" ")
	b.WriteString(m.renderInlineButtons(nav2))

	misc := []namedButton{
		{"Mute", push3.ButtonMute}, {"Solo", push3.ButtonSolo},
		{"Shft", push3.ButtonShift}, {"Sel", push3.ButtonSelect},
	}
	b.WriteString("  ")
	b.WriteString(m.renderInlineButtons(misc))
	b.WriteString("\n")

	// Row 3: Transport
	transport := []namedButton{
		{"Stop", push3.ButtonStop},
		{" ▶ ", push3.ButtonPlay},
		{" ● ", push3.ButtonRecord},
		{"Undo", push3.ButtonUndo},
		{"Del", push3.ButtonDelete},
		{"Qntz", push3.ButtonQuantize},
		{"Dup", push3.ButtonDuplicate},
	}
	b.WriteString(" ")
	b.WriteString(m.renderInlineButtons(transport))
	b.WriteString("\n")

	return b.String()
}

func (m model) renderInlineButtons(buttons []namedButton) string {
	var parts []string
	for _, btn := range buttons {
		style := btnOffStyle
		if m.buttons[btn.id] {
			switch btn.id {
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
		parts = append(parts, style.Render(centerPad(btn.label, 4)))
	}
	return strings.Join(parts, " ")
}

func (m model) renderPadGrid() string {
	var b strings.Builder
	for row := range 8 {
		b.WriteString(" ")
		for col := range 8 {
			vel := m.pads[row][col]
			style := padOffStyle
			if vel > 0 {
				style = padStyle(vel)
			}
			b.WriteString(style.Render("    "))
			if col < 7 {
				b.WriteString(" ")
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

// --- Styles ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF8800")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	btnOffStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 0)

	btnOnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 0)

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

	encOffStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555")).
			Background(lipgloss.Color("#111111")).
			Width(6).
			Align(lipgloss.Center)

	encOnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#333333")).
			Width(6).
			Align(lipgloss.Center)

	padOffStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#111111"))
)

// padStyle returns a colored style based on velocity.
func padStyle(velocity uint8) lipgloss.Style {
	// Map velocity to brightness/hue.
	r, g, b := velocityToRGB(velocity)
	bg := lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
	return lipgloss.NewStyle().Background(bg)
}

// velocityToRGB maps MIDI velocity to a color gradient.
func velocityToRGB(vel uint8) (r, g, b uint8) {
	if vel == 0 {
		return 0, 0, 0
	}
	// Gradient from blue (low velocity) through green to red (high velocity).
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

// centerPad pads a string to at least width, centering the content.
func centerPad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	total := width - len(s)
	left := total / 2
	right := total - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}
