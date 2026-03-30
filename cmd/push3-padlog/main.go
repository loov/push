// Command push3-padlog logs all pad-related MIDI messages with timestamps,
// channel tracking, and interpreted meanings. Useful for understanding
// how Push 3 handles pad slides across rows.
//
// It shows:
//   - Note On/Off with pad position
//   - Channel Pressure (aftertouch)
//   - CC 74 (MPE Slide / vertical finger position)
//   - Pitch Bend (horizontal position)
//   - Active channel state
//
// Usage:
//
//	go run ./cmd/push3-padlog
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/loov/push3/midi"
	"github.com/loov/push3/push"
)

type channelState struct {
	active  bool
	note    uint8
	pos     push.PadPosition
	started time.Time
}

func main() {
	source := flag.String("source", push.SourceName, "Push 3 MIDI source name")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	client, err := midi.NewClient("push3-padlog")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	src, err := midi.FindSource(*source)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Listening on %q — press/slide pads on Push 3\n", src.DisplayName())
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println()

	var channels [16]channelState
	start := time.Now()

	ts := func() string {
		return fmt.Sprintf("%8.3f", time.Since(start).Seconds())
	}

	_, err = client.OpenInput("padlog", src, func(data []byte) {
		if len(data) < 2 || data[0] == 0xFE {
			return
		}

		status := data[0] & 0xF0
		ch := data[0] & 0x0F

		switch status {
		case 0x90: // Note On
			if len(data) < 3 {
				return
			}
			note := data[1]
			vel := data[2]
			pos, isPad := push.PadPositionFromNote(note)
			if !isPad {
				return
			}

			if vel > 0 {
				channels[ch] = channelState{
					active:  true,
					note:    note,
					pos:     pos,
					started: time.Now(),
				}
				fmt.Printf("%s  ch=%2d  NOTE ON   note=%-3d  pad=(%d,%d)  vel=%-3d\n",
					ts(), ch, note, pos.Row, pos.Col, vel)
			} else {
				// Note On vel=0 is equivalent to Note Off.
				printNoteOff(ts(), ch, note, pos, &channels[ch])
			}

		case 0x80: // Note Off
			if len(data) < 3 {
				return
			}
			note := data[1]
			pos, isPad := push.PadPositionFromNote(note)
			if !isPad {
				return
			}
			printNoteOff(ts(), ch, note, pos, &channels[ch])

		case 0xD0: // Channel Pressure
			if len(data) < 2 {
				return
			}
			pressure := data[1]
			cs := &channels[ch]
			if cs.active {
				fmt.Printf("%s  ch=%2d  PRESSURE  val=%-3d  (pad=(%d,%d) note=%d)\n",
					ts(), ch, pressure, cs.pos.Row, cs.pos.Col, cs.note)
			} else {
				fmt.Printf("%s  ch=%2d  PRESSURE  val=%-3d  (no active pad — orphan)\n",
					ts(), ch, pressure)
			}

		case 0xB0: // CC
			if len(data) < 3 {
				return
			}
			cc := data[1]
			val := data[2]
			if cc != 74 { // Only show MPE Slide (CC 74)
				return
			}
			// Skip channel 0 (not MPE).
			if ch == 0 {
				return
			}
			cs := &channels[ch]
			if cs.active {
				fmt.Printf("%s  ch=%2d  SLIDE     val=%-3d  (pad=(%d,%d) note=%d)\n",
					ts(), ch, val, cs.pos.Row, cs.pos.Col, cs.note)
			} else {
				fmt.Printf("%s  ch=%2d  SLIDE     val=%-3d  (no active pad — orphan)\n",
					ts(), ch, val)
			}

		case 0xE0: // Pitch Bend
			if len(data) < 3 {
				return
			}
			if ch == 0 { // Channel 0 is touch strip, not pad.
				return
			}
			val := uint16(data[1]) | uint16(data[2])<<7
			cs := &channels[ch]
			if cs.active {
				fmt.Printf("%s  ch=%2d  BEND      val=%-5d  (pad=(%d,%d) note=%d)\n",
					ts(), ch, val, cs.pos.Row, cs.pos.Col, cs.note)
			} else {
				fmt.Printf("%s  ch=%2d  BEND      val=%-5d  (no active pad — orphan)\n",
					ts(), ch, val)
			}
		}
	})
	if err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
	fmt.Println("\nDone")
}

func printNoteOff(ts string, ch uint8, note uint8, pos push.PadPosition, cs *channelState) {
	if cs.active && cs.note == note {
		dur := time.Since(cs.started)
		fmt.Printf("%s  ch=%2d  NOTE OFF  note=%-3d  pad=(%d,%d)  (held %.3fs)\n",
			ts, ch, note, pos.Row, pos.Col, dur.Seconds())
		cs.active = false
	} else if cs.active && cs.note != note {
		fmt.Printf("%s  ch=%2d  NOTE OFF  note=%-3d  pad=(%d,%d)  *** SLIDE: was note=%d pad=(%d,%d) ***\n",
			ts, ch, note, pos.Row, pos.Col, cs.note, cs.pos.Row, cs.pos.Col)
		cs.active = false
	} else {
		fmt.Printf("%s  ch=%2d  NOTE OFF  note=%-3d  pad=(%d,%d)  (was not active)\n",
			ts, ch, note, pos.Row, pos.Col)
	}
}
