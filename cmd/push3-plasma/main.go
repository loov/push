// Command push3-plasma runs an interactive plasma field demo on the Push 3 pad grid.
// Pad presses create ripple distortions that decay over time.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"time"

	"github.com/loov/push/midi"
	"github.com/loov/push/push3"
)

func main() {
	source := flag.String("source", push3.SourceName, "Push 3 MIDI source name")
	dest := flag.String("dest", push3.DestName, "Push 3 MIDI destination name")
	flag.Parse()

	if err := run(context.Background(), *source, *dest); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, sourceName, destName string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	client, err := midi.NewClient("push3-plasma")
	if err != nil {
		return fmt.Errorf("creating MIDI client: %w", err)
	}
	defer client.Close()

	p, err := push3.Connect(client, sourceName, destName)
	if err != nil {
		return fmt.Errorf("connecting to Push 3: %w", err)
	}

	log.Println("Uploading rainbow palette...")
	if err := uploadRainbowPalette(p); err != nil {
		return fmt.Errorf("uploading palette: %w", err)
	}

	log.Println("Running plasma demo... Press pads for ripples. Ctrl+C to exit.")

	state := &plasmaState{}

	p.OnPad = func(pos push3.PadPosition, velocity uint8, pressed bool) {
		if !pressed {
			return
		}
		t := float64(state.frame) * 0.08
		state.addRipple(pos.Row, pos.Col, velocity, t)
	}

	ticker := time.NewTicker(33 * time.Millisecond) // ~30fps
	defer ticker.Stop()

	var prev [8][8]uint8
	for {
		select {
		case <-ctx.Done():
			_ = p.ClearPads()
			log.Println("Shutting down")
			return nil
		case <-ticker.C:
		}

		t := float64(state.frame) * 0.08
		for row := range uint8(8) {
			for col := range uint8(8) {
				x := float64(col) / 3.5
				y := float64(row) / 3.5
				v := state.plasma(x, y, t)
				// Map [-1, 1] to palette index [1, 127].
				idx := uint8((v+1)/2*126) + 1
				if idx == prev[row][col] {
					continue
				}
				prev[row][col] = idx
				pos := push3.PadPosition{Row: row, Col: col}
				if err := p.SetPadColor(pos, idx); err != nil {
					return err
				}
			}
		}
		state.frame++
	}
}

// ripple tracks a pad press distortion.
type ripple struct {
	x, y  float64 // center position
	birth float64 // time of creation
	amp   float64 // initial amplitude (from velocity)
}

// plasmaState holds the mutable state for the plasma demo.
type plasmaState struct {
	frame   int
	ripples []ripple
}

// addRipple adds a ripple from a pad press.
func (s *plasmaState) addRipple(row, col uint8, velocity uint8, t float64) {
	s.ripples = append(s.ripples, ripple{
		x:     float64(col) / 3.5,
		y:     float64(row) / 3.5,
		birth: t,
		amp:   float64(velocity) / 127,
	})
}

// plasma computes a classic demoscene plasma value from overlapping sine waves,
// plus ripple distortions from pad presses.
func (s *plasmaState) plasma(x, y, t float64) float64 {
	v := math.Sin(x*1.3 + t*0.7)
	v += math.Sin(y*1.5 + t*0.9)
	v += math.Sin((x+y)*0.9 + t*0.5)
	v += math.Sin(math.Sqrt(x*x+y*y)*1.2 + t*1.1)
	v /= 4

	// Add ripple contributions.
	alive := 0
	for i := range s.ripples {
		r := &s.ripples[i]
		age := t - r.birth
		if age > 8 {
			continue
		}
		s.ripples[alive] = *r
		alive++

		decay := r.amp * math.Exp(-age*0.4)
		dx := x - r.x
		dy := y - r.y
		dist := math.Sqrt(dx*dx + dy*dy)
		v += decay * math.Sin(dist*6-age*4)
	}
	s.ripples = s.ripples[:alive]

	if v > 1 {
		v = 1
	} else if v < -1 {
		v = -1
	}
	return v
}

// hsvToRGB converts HSV (h in [0,360), s,v in [0,1]) to RGB bytes.
func hsvToRGB(h, s, v float64) (uint8, uint8, uint8) {
	c := v * s
	h2 := h / 60
	x := c * (1 - math.Abs(math.Mod(h2, 2)-1))
	m := v - c

	var r, g, b float64
	switch {
	case h2 < 1:
		r, g, b = c, x, 0
	case h2 < 2:
		r, g, b = x, c, 0
	case h2 < 3:
		r, g, b = 0, c, x
	case h2 < 4:
		r, g, b = 0, x, c
	case h2 < 5:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return uint8((r + m) * 255), uint8((g + m) * 255), uint8((b + m) * 255)
}

// uploadRainbowPalette writes a smooth HSV rainbow into palette indices 1-127.
func uploadRainbowPalette(p *push3.Device) error {
	for i := range 127 {
		h := float64(i) / 127 * 360
		r, g, b := hsvToRGB(h, 1, 1)
		if err := p.SetPaletteEntry(uint8(i+1), push3.Color{R: r, G: g, B: b}); err != nil {
			return fmt.Errorf("setting palette %d: %w", i+1, err)
		}
	}
	return p.ReapplyPalette()
}
