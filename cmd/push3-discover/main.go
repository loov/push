// Command push3-discover guides you through pressing each button on the
// Push 3 to discover its MIDI CC/note assignment.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"golang.org/x/term"

	"github.com/loov/logic-push3/midi"
	"github.com/loov/logic-push3/push"
)

// item describes something to discover.
type item struct {
	name string
	// If true, capture all events until Enter (for encoders with touch+press+rotate).
	// If false, auto-advance on press+release (for simple buttons).
	captureAll bool
}

var itemsToDiscover = []item{
	// Touch strip
	{"Touch strip: TOUCH (place finger, don't move)", true},
	{"Touch strip: SLIDE (drag finger up and down)", true},

	// Buttons (auto-advance)
	{"Sets", false}, {"Setup", false}, {"Learn", false}, {"User", false},
	{"Device", false}, {"Mix", false}, {"Clip", false}, {"Session", false},
	{"Undo", false}, {"Save", false}, {"Add", false}, {"Swap", false},
	{"Lock", false}, {"Stop Clip", false}, {"Mute", false}, {"Solo", false},
	{"Master", false},
	{"Tap Tempo", false}, {"Metronome", false}, {"Quantize", false},
	{"Fixed Length", false}, {"Automate", false},
	{"New", false}, {"Capture", false}, {"Record", false}, {"Play", false},
	{"Upper Display 1", false}, {"Upper Display 2", false},
	{"Upper Display 3", false}, {"Upper Display 4", false},
	{"Upper Display 5", false}, {"Upper Display 6", false},
	{"Upper Display 7", false}, {"Upper Display 8", false},
	{"Lower Display 1", false}, {"Lower Display 2", false},
	{"Lower Display 3", false}, {"Lower Display 4", false},
	{"Lower Display 5", false}, {"Lower Display 6", false},
	{"Lower Display 7", false}, {"Lower Display 8", false},
	{"1/32t", false}, {"1/32", false}, {"1/16t", false}, {"1/16", false},
	{"1/8t", false}, {"1/8", false}, {"1/4t", false}, {"1/4", false},
	{"D-pad Up", false}, {"D-pad Down", false},
	{"D-pad Left", false}, {"D-pad Right", false}, {"D-pad Center", false},
	{"Note", false}, {"Session (right side)", false},
	{"Scale", false}, {"Layout", false},
	{"Repeat", false}, {"Accent", false},
	{"Double Loop", false}, {"Duplicate", false},
	{"Convert", false}, {"Delete", false},
	{"Octave Up", false}, {"Octave Down", false},
	{"Page Left", false}, {"Page Right", false},
	{"Shift", false}, {"Select", false},
	{"Browse", false},
}

// Pad notes to ignore (36-99 are pads, not buttons).
func isPadNote(note uint8) bool {
	return note >= 36 && note <= 99
}

type event struct {
	kind   string // "CC", "Note", "NoteOff", "PitchBend", "ChanPressure", "PolyAT"
	ch     int
	number int
	value  int
}

func (e event) String() string {
	switch e.kind {
	case "CC":
		return fmt.Sprintf("CC  %3d (0x%02X) val=%3d ch=%d", e.number, e.number, e.value, e.ch)
	case "Note":
		return fmt.Sprintf("Note %3d (0x%02X) vel=%3d ch=%d", e.number, e.number, e.value, e.ch)
	case "NoteOff":
		return fmt.Sprintf("NoteOff %3d (0x%02X) ch=%d", e.number, e.number, e.ch)
	default:
		return fmt.Sprintf("%s num=%d val=%d ch=%d", e.kind, e.number, e.value, e.ch)
	}
}

type capturedItem struct {
	name   string
	events []event
}

// writeln writes a line with \r\n for raw terminal mode.
func writeln(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	fmt.Print(s + "\r\n")
}

func main() {
	source := flag.String("source", push.SourceName, "Push 3 MIDI source name")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	client, err := midi.NewClient("push3-discover")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	src, err := midi.FindSource(*source)
	if err != nil {
		log.Fatal(err)
	}

	writeln("Connected to %q", src.DisplayName())
	writeln("")
	writeln("For encoders: touch, press, and rotate — then press Enter to continue.")
	writeln("For buttons:  press and release — auto-advances.")
	writeln("Backspace to undo. Ctrl+C to exit.")
	writeln("")

	var mu sync.Mutex
	idx := 0
	waitingForRelease := false
	var captured []capturedItem     // completed items
	var currentEvents []event       // events for current captureAll item
	seen := map[string]bool{}       // dedup events for captureAll mode

	printPrompt := func() {
		if idx < len(itemsToDiscover) {
			it := itemsToDiscover[idx]
			mode := ""
			if it.captureAll {
				mode = " [Enter=next, show all events]"
			}
			writeln("")
			writeln(">>> [%d/%d] %s%s", idx+1, len(itemsToDiscover), it.name, mode)
			currentEvents = nil
			seen = map[string]bool{}
		} else {
			writeln("")
			writeln("=== All items discovered! ===")
			printAllResults(captured)
		}
	}

	// Keyboard listener.
	go func() {
		buf := make([]byte, 1)
		for {
			_, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			mu.Lock()
			switch buf[0] {
			case 13, 10: // Enter
				if idx < len(itemsToDiscover) && itemsToDiscover[idx].captureAll {
					// Save current events and advance.
					captured = append(captured, capturedItem{
						name:   itemsToDiscover[idx].name,
						events: append([]event{}, currentEvents...),
					})
					idx++
					printPrompt()
				}
			case 127, 8: // Backspace
				if idx > 0 && !waitingForRelease {
					idx--
					if len(captured) > 0 && captured[len(captured)-1].name == itemsToDiscover[idx].name {
						removed := captured[len(captured)-1]
						captured = captured[:len(captured)-1]
						writeln("    ↩ Undid: %s (%d events)", removed.name, len(removed.events))
					}
					printPrompt()
				}
			case 3: // Ctrl+C
				mu.Unlock()
				stop()
				return
			}
			mu.Unlock()
		}
	}()

	printPrompt()

	_, err = client.OpenInput("discover", src, func(data []byte) {
		if len(data) == 0 || data[0] == 0xFE {
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if idx >= len(itemsToDiscover) {
			return
		}

		ev := parseEvent(data)
		if ev == nil {
			return
		}

		it := itemsToDiscover[idx]

		if it.captureAll {
			// Show every unique event type.
			key := fmt.Sprintf("%s-%d", ev.kind, ev.number)
			if !seen[key] {
				seen[key] = true
				currentEvents = append(currentEvents, *ev)
				writeln("    %s", ev)
			}
			return
		}

		// Auto-advance mode for simple buttons.
		switch ev.kind {
		case "CC":
			if ev.value > 0 && !waitingForRelease {
				writeln("    %-25s  %s", it.name, ev)
				captured = append(captured, capturedItem{
					name:   it.name,
					events: []event{*ev},
				})
				waitingForRelease = true
			} else if ev.value == 0 && waitingForRelease {
				waitingForRelease = false
				idx++
				printPrompt()
			}
		case "Note":
			if ev.value > 0 && !waitingForRelease {
				writeln("    %-25s  %s", it.name, ev)
				captured = append(captured, capturedItem{
					name:   it.name,
					events: []event{*ev},
				})
				waitingForRelease = true
			} else if ev.value == 0 && waitingForRelease {
				waitingForRelease = false
				idx++
				printPrompt()
			}
		case "NoteOff":
			if waitingForRelease {
				waitingForRelease = false
				idx++
				printPrompt()
			}
		}
	})
	if err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
	writeln("")
	printAllResults(captured)
}

func parseEvent(data []byte) *event {
	if len(data) < 2 {
		return nil
	}
	status := data[0]
	ch := int(status & 0x0F)
	kind := status & 0xF0

	switch kind {
	case 0xB0: // CC
		if len(data) < 3 {
			return nil
		}
		return &event{"CC", ch, int(data[1]), int(data[2])}
	case 0x90: // Note On
		if len(data) < 3 {
			return nil
		}
		if isPadNote(data[1]) {
			return nil
		}
		vel := int(data[2])
		if vel == 0 {
			return &event{"NoteOff", ch, int(data[1]), 0}
		}
		return &event{"Note", ch, int(data[1]), vel}
	case 0x80: // Note Off
		if len(data) < 3 || isPadNote(data[1]) {
			return nil
		}
		return &event{"NoteOff", ch, int(data[1]), 0}
	case 0xE0: // Pitch Bend
		if len(data) < 3 {
			return nil
		}
		val := int(data[1]) | int(data[2])<<7
		return &event{"PitchBend", ch, 0, val}
	case 0xD0: // Channel Pressure
		return &event{"ChanPressure", ch, 0, int(data[1])}
	case 0xA0: // Poly Aftertouch
		if len(data) < 3 {
			return nil
		}
		return &event{"PolyAT", ch, int(data[1]), int(data[2])}
	}
	return nil
}

func printAllResults(items []capturedItem) {
	writeln("")
	writeln("── Results ──")
	for _, item := range items {
		if len(item.events) == 1 {
			writeln("  %-30s  %s", item.name, item.events[0])
		} else {
			writeln("  %s:", item.name)
			for _, ev := range item.events {
				writeln("    %s", ev)
			}
		}
	}
}
