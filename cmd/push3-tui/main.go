// Command push3-tui shows a TUI mirroring the Push 3 layout.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"

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

// Messages.
type (
	buttonMsg       struct{ id push3.ButtonID; pressed bool }
	padMsg          struct{ pos push3.PadPosition; velocity uint8; pressed bool }
	encoderMsg      struct{ id push3.EncoderID; delta int }
	encoderTouchMsg struct{ id push3.EncoderID; touched bool }
)

// Model.
type model struct {
	push     *push.Push3
	buttons  map[push3.ButtonID]bool
	pads     [8][8]uint8
	encoders [12]int
	touched  [12]bool
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

func (m model) View() tea.View {
	return tea.NewView(m.renderLayout() + " q/ctrl+c to quit\n")
}
