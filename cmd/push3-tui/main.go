// Command push3-tui shows a TUI mirroring the Push 3 layout.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/loov/push3/midi"
	"github.com/loov/push3/push"
	"github.com/loov/push3/push3"
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
	p.OnPadPressure = func(pos push3.PadPosition, pressure uint8) {
		prog.Send(padPressureMsg{pos: pos, pressure: pressure})
	}
	p.OnPadSlide = func(pos push3.PadPosition, value uint8) {
		prog.Send(padSlideMsg{pos: pos, value: value})
	}
	p.OnPadPitchBend = func(pos push3.PadPosition, value uint16) {
		prog.Send(padPitchBendMsg{pos: pos, value: value})
	}
	p.OnEncoder = func(id push3.EncoderID, delta int) {
		prog.Send(encoderMsg{id: id, delta: delta})
	}
	p.OnEncoderTouch = func(id push3.EncoderID, touched bool) {
		prog.Send(encoderTouchMsg{id: id, touched: touched})
	}
	p.OnTouchStrip = func(value uint16) {
		prog.Send(touchStripMsg{value: value})
	}
	p.OnTouchStripTouch = func(touched bool) {
		prog.Send(touchStripTouchMsg{touched: touched})
	}
	p.OnDPadCenterTouch = func(touched bool) {
		prog.Send(dpadCenterTouchMsg{touched: touched})
	}

	if _, err := prog.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Messages.
type (
	buttonMsg struct {
		id      push3.ButtonID
		pressed bool
	}
	padMsg struct {
		pos      push3.PadPosition
		velocity uint8
		pressed  bool
	}
	padPressureMsg struct {
		pos      push3.PadPosition
		pressure uint8
	}
	padSlideMsg struct {
		pos   push3.PadPosition
		value uint8
	}
	padPitchBendMsg struct {
		pos   push3.PadPosition
		value uint16
	}
	encoderMsg struct {
		id    push3.EncoderID
		delta int
	}
	encoderTouchMsg struct {
		id      push3.EncoderID
		touched bool
	}
	touchStripMsg      struct{ value uint16 }
	touchStripTouchMsg struct{ touched bool }
	dpadCenterTouchMsg struct{ touched bool }
)

// Model.
type model struct {
	push     *push.Push3
	buttons  map[push3.ButtonID]bool
	pads     [8][8]padState
	encoders [16]int
	touched  [16]bool

	// Touch strip
	touchStripValue   uint16
	touchStripTouched bool

	// D-pad center touch
	dpadCenterTouched bool

	// Last event info for status display
	lastEvent string
}

type padState struct {
	velocity  uint8
	pressure  uint8
	slide     uint8
	pitchBend uint16
}

func newModel(p *push.Push3) model {
	return model{push: p, buttons: make(map[push3.ButtonID]bool)}
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
		if msg.pressed {
			m.lastEvent = fmt.Sprintf("Button CC=%d pressed", msg.id)
		}
	case padMsg:
		if msg.pressed {
			m.pads[msg.pos.Row][msg.pos.Col].velocity = msg.velocity
			m.lastEvent = fmt.Sprintf("Pad (%d,%d) vel=%d", msg.pos.Row, msg.pos.Col, msg.velocity)
		} else {
			m.pads[msg.pos.Row][msg.pos.Col] = padState{}
		}
	case padPressureMsg:
		m.pads[msg.pos.Row][msg.pos.Col].pressure = msg.pressure
		m.lastEvent = fmt.Sprintf("Pad (%d,%d) pressure=%d", msg.pos.Row, msg.pos.Col, msg.pressure)
	case padSlideMsg:
		m.pads[msg.pos.Row][msg.pos.Col].slide = msg.value
		m.lastEvent = fmt.Sprintf("Pad (%d,%d) slide=%d", msg.pos.Row, msg.pos.Col, msg.value)
	case padPitchBendMsg:
		m.pads[msg.pos.Row][msg.pos.Col].pitchBend = msg.value
		m.lastEvent = fmt.Sprintf("Pad (%d,%d) bend=%d", msg.pos.Row, msg.pos.Col, msg.value)
	case encoderMsg:
		idx := int(msg.id)
		if idx > 0 && idx < len(m.encoders) {
			m.encoders[idx] += msg.delta
		}
		m.lastEvent = fmt.Sprintf("Encoder %d delta=%d val=%d", msg.id, msg.delta, m.encoders[idx])
	case encoderTouchMsg:
		idx := int(msg.id)
		if idx > 0 && idx < len(m.touched) {
			m.touched[idx] = msg.touched
		}
	case touchStripMsg:
		m.touchStripValue = msg.value
		m.lastEvent = fmt.Sprintf("Touch strip pos=%d", msg.value)
	case touchStripTouchMsg:
		m.touchStripTouched = msg.touched
		if msg.touched {
			m.lastEvent = "Touch strip touched"
		} else {
			m.lastEvent = "Touch strip released"
		}
	case dpadCenterTouchMsg:
		m.dpadCenterTouched = msg.touched
		if msg.touched {
			m.lastEvent = "D-pad center touched"
		} else {
			m.lastEvent = "D-pad center released"
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	var status string

	// Status line: last event
	status += fmt.Sprintf(" Last: %s\n", m.lastEvent)

	// Status line: active pad details
	for r := range 8 {
		for c := range 8 {
			ps := m.pads[r][c]
			if ps.velocity > 0 {
				status += fmt.Sprintf(" Pad(%d,%d): vel=%-3d prs=%-3d slide=%-3d bend=%-5d\n",
					r, c, ps.velocity, ps.pressure, ps.slide, ps.pitchBend)
			}
		}
	}

	return tea.NewView(m.renderLayout() + status + " q/ctrl+c to quit\n")
}
