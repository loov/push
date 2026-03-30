// Command push3-demo connects to an Ableton Push 3 and prints all incoming
// events. It also cycles pad colors through the palette on startup.
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
	"github.com/loov/push3/push3"
)

func main() {
	source := flag.String("source", push3.SourceName, "Push 3 MIDI source name")
	dest := flag.String("dest", push3.DestName, "Push 3 MIDI destination name")
	colorDemo := flag.Bool("colors", false, "cycle pad colors through palette on startup")
	animDemo := flag.Bool("anim", false, "show animated LED demo (pulse/blink)")
	brightness := flag.Int("brightness", -1, "set LED brightness (0-127)")
	raw := flag.Bool("raw", false, "print raw MIDI bytes for all messages")
	flag.Parse()

	if err := run(context.Background(), *source, *dest, *colorDemo, *animDemo, *brightness, *raw); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, sourceName, destName string, colorDemo, animDemo bool, brightness int, raw bool) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	client, err := midi.NewClient("push3-demo")
	if err != nil {
		return fmt.Errorf("creating MIDI client: %w", err)
	}
	defer client.Close()

	p, err := push3.Connect(client, sourceName, destName)
	if err != nil {
		return fmt.Errorf("connecting to Push 3: %w", err)
	}

	if raw {
		p.OnRawMIDI = func(data []byte) {
			// Filter out Active Sensing (0xFE) heartbeat.
			if len(data) == 1 && data[0] == 0xFE {
				return
			}
			log.Printf("[RAW] % X", data)
		}
	}

	p.OnButton = func(id push3.ButtonID, pressed bool) {
		action := "released"
		if pressed {
			action = "pressed"
		}
		log.Printf("[Button] CC=%d %s", id, action)
	}

	p.OnPad = func(pos push3.PadPosition, velocity uint8, pressed bool) {
		action := "released"
		if pressed {
			action = fmt.Sprintf("pressed vel=%d", velocity)
		}
		log.Printf("[Pad] (%d,%d) note=%d %s", pos.Row, pos.Col, pos.PadNote(), action)
	}

	p.OnEncoder = func(id push3.EncoderID, delta int) {
		log.Printf("[Encoder] %d delta=%d", id, delta)
	}

	p.OnEncoderTouch = func(id push3.EncoderID, touched bool) {
		action := "released"
		if touched {
			action = "touched"
		}
		log.Printf("[EncoderTouch] %d %s", id, action)
	}

	log.Printf("Connected to Push 3: source=%q dest=%q", sourceName, destName)
	log.Println("Press Ctrl+C to exit")

	if brightness >= 0 {
		level := uint8(min(brightness, 127))
		log.Printf("Setting LED brightness to %d", level)
		if err := p.SetBrightness(level); err != nil {
			log.Printf("Error setting brightness: %v", err)
		}
	}

	if colorDemo {
		log.Println("Running color demo...")
		palette := []uint8{
			push3.PaletteRed,
			push3.PaletteOrange,
			push3.PaletteYellow,
			push3.PaletteGreen,
			push3.PaletteTurquoise,
			push3.PaletteBlue,
			push3.PalettePurple,
			push3.PalettePink,
			push3.PaletteWhite,
		}
		for _, color := range palette {
			if err := p.SetAllPadsColor(color); err != nil {
				log.Printf("Error setting pad color: %v", err)
				break
			}
			if err := p.SetAllButtonsColor(color); err != nil {
				log.Printf("Error setting button color: %v", err)
				break
			}
			select {
			case <-time.After(500 * time.Millisecond):
			case <-ctx.Done():
				return nil
			}
		}
		if err := p.ClearPads(); err != nil {
			log.Printf("Error clearing pads: %v", err)
		}
		if err := p.ClearButtons(); err != nil {
			log.Printf("Error clearing buttons: %v", err)
		}
		log.Println("Color demo complete")
	}

	if animDemo {
		log.Println("Running animation demo...")
		// Row 0: pulsing colors
		for col := range uint8(8) {
			pos := push3.PadPosition{Row: 0, Col: col}
			if err := p.SetPadColorAnimated(pos, push3.PaletteRed+col, push3.AnimPulse4); err != nil {
				log.Printf("Error setting animated pad: %v", err)
				break
			}
		}
		// Row 1: blinking colors
		for col := range uint8(8) {
			pos := push3.PadPosition{Row: 1, Col: col}
			if err := p.SetPadColorAnimated(pos, push3.PaletteBlue+col, push3.AnimBlink4); err != nil {
				log.Printf("Error setting animated pad: %v", err)
				break
			}
		}
		// Row 2: one-shot fade
		for col := range uint8(8) {
			pos := push3.PadPosition{Row: 2, Col: col}
			if err := p.SetPadColorAnimated(pos, push3.PaletteGreen, push3.AnimOneShot4); err != nil {
				log.Printf("Error setting animated pad: %v", err)
				break
			}
		}
		log.Println("Animation demo active (top 3 rows). Press Ctrl+C to exit.")
	}

	<-ctx.Done()
	log.Println("Shutting down")
	return nil
}
