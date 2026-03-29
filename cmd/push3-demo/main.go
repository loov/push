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

	"github.com/loov/logic-push3/midi"
	"github.com/loov/logic-push3/push"
	"github.com/loov/logic-push3/push3"
)

func main() {
	source := flag.String("source", push.SourceName, "Push 3 MIDI source name")
	dest := flag.String("dest", push.DestName, "Push 3 MIDI destination name")
	colorDemo := flag.Bool("colors", false, "cycle pad colors through palette on startup")
	flag.Parse()

	if err := run(context.Background(), *source, *dest, *colorDemo); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, sourceName, destName string, colorDemo bool) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	client, err := midi.NewClient("push3-demo")
	if err != nil {
		return fmt.Errorf("creating MIDI client: %w", err)
	}
	defer client.Close()

	p, err := push.Connect(client, sourceName, destName)
	if err != nil {
		return fmt.Errorf("connecting to Push 3: %w", err)
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
			select {
			case <-time.After(500 * time.Millisecond):
			case <-ctx.Done():
				return nil
			}
		}
		if err := p.ClearPads(); err != nil {
			log.Printf("Error clearing pads: %v", err)
		}
		log.Println("Color demo complete")
	}

	<-ctx.Done()
	log.Println("Shutting down")
	return nil
}
