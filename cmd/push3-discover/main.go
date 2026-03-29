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

// Buttons to discover, in a logical order matching the physical layout.
var buttonsToDiscover = []string{
	// Top-left row
	"Sets", "Setup", "Learn", "User",
	// Top-right row
	"Device", "Mix", "Clip", "Session",
	// Display area
	"Undo", "Save", "Add", "Swap",
	// Bottom-left row
	"Lock", "Stop Clip", "Mute", "Solo",
	// Bottom-right
	"Master",
	// Left side (top to bottom)
	// Skipping encoder presses (Volume, Swing/Tempo) — handle separately.
	"Tap Tempo", "Metronome", "Quantize",
	"Fixed Length", "Automate",
	"New", "Capture", "Record", "Play",
	// Center
	// "Upper Display 1",
	"Upper Display 2", "Upper Display 3", "Upper Display 4",
	"Upper Display 5", "Upper Display 6", "Upper Display 7", "Upper Display 8",
	"Lower Display 1", "Lower Display 2", "Lower Display 3", "Lower Display 4",
	"Lower Display 5", "Lower Display 6", "Lower Display 7", "Lower Display 8",
	// Right side
	"1/32t", "1/32", "1/16t", "1/16", "1/8t", "1/8", "1/4t", "1/4",
	// D-pad
	"D-pad Up", "D-pad Down", "D-pad Left", "D-pad Right", "D-pad Center",
	// Right buttons
	"Note", "Session (right side)",
	"Scale", "Layout",
	"Repeat", "Accent",
	"Double Loop", "Duplicate",
	"Convert", "Delete",
	// Navigation
	"Octave Up", "Octave Down", "Page Left", "Page Right",
	// Bottom right
	"Shift", "Select",
	// Browse
	"Browse",
}

// Encoder rotation CCs to ignore (these fire when pressing encoder buttons).
var encoderCCs = map[uint8]bool{
	14: true, 15: true, // Tempo, Swing
	71: true, 72: true, 73: true, 74: true, // Track 1-4
	75: true, 76: true, 77: true, 78: true, // Track 5-8
	79: true, // Master
}

// Pad notes to ignore (36-99 are pads, not buttons).
func isPadNote(note uint8) bool {
	return note >= 36 && note <= 99
}

type result struct {
	name   string
	kind   string // "CC" or "Note"
	number int
	ch     int
}

// print writes a line with \r\n for raw terminal mode.
func print(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	fmt.Print(s + "\r\n")
}

func main() {
	source := flag.String("source", push.SourceName, "Push 3 MIDI source name")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Switch terminal to raw mode for single-keystroke backspace.
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

	print("Connected to %q", src.DisplayName())
	print("Press the prompted button on Push 3. Backspace to undo. Ctrl+C to exit.")
	print("")

	var mu sync.Mutex
	idx := 0
	waitingForRelease := false
	var results []result

	printPrompt := func() {
		if idx < len(buttonsToDiscover) {
			print("")
			print(">>> [%d/%d] Press: %s", idx+1, len(buttonsToDiscover), buttonsToDiscover[idx])
		} else {
			print("")
			print("=== All buttons discovered! ===")
			printResults(results)
		}
	}

	// Keyboard listener for backspace (undo).
	go func() {
		buf := make([]byte, 1)
		for {
			_, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			switch buf[0] {
			case 127, 8: // Backspace / Delete
				mu.Lock()
				if idx > 0 && !waitingForRelease {
					idx--
					if len(results) > 0 {
						removed := results[len(results)-1]
						results = results[:len(results)-1]
						print("    ↩ Undid: %s (%s %d)", removed.name, removed.kind, removed.number)
					}
					printPrompt()
				}
				mu.Unlock()
			case 3: // Ctrl+C
				stop()
				return
			}
		}
	}()

	printPrompt()

	_, err = client.OpenInput("discover", src, func(data []byte) {
		if len(data) == 0 || data[0] == 0xFE {
			return
		}

		mu.Lock()
		defer mu.Unlock()

		status := data[0]
		ch := int(status & 0x0F)
		kind := status & 0xF0

		switch kind {
		case 0xB0: // CC
			if len(data) < 3 {
				return
			}
			cc := data[1]
			val := data[2]

			// Ignore encoder rotation CCs.
			if encoderCCs[cc] {
				return
			}

			if val > 0 && !waitingForRelease {
				name := ""
				if idx < len(buttonsToDiscover) {
					name = buttonsToDiscover[idx]
				}
				print("    %-25s  CC  %3d  (0x%02X)  ch=%d", name, cc, cc, ch)
				results = append(results, result{name, "CC", int(cc), ch})
				waitingForRelease = true
			} else if val == 0 && waitingForRelease {
				waitingForRelease = false
				idx++
				printPrompt()
			}

		case 0x90: // Note On
			if len(data) < 3 {
				return
			}
			note := data[1]
			vel := data[2]

			// Ignore pad notes.
			if isPadNote(note) {
				return
			}
			if vel > 0 && !waitingForRelease {
				name := ""
				if idx < len(buttonsToDiscover) {
					name = buttonsToDiscover[idx]
				}
				print("    %-25s  Note %3d  (0x%02X)  vel=%d  ch=%d", name, note, note, vel, ch)
				results = append(results, result{name, "Note", int(note), ch})
				waitingForRelease = true
			} else if vel == 0 && waitingForRelease {
				waitingForRelease = false
				idx++
				printPrompt()
			}

		case 0x80: // Note Off
			if len(data) >= 2 && isPadNote(data[1]) {
				return
			}
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
	print("")
	printResults(results)
}

func printResults(results []result) {
	print("")
	print("── Results ──")
	for _, r := range results {
		print("  %-25s  %4s  %3d  (0x%02X)  ch=%d", r.name, r.kind, r.number, r.number, r.ch)
	}
}
