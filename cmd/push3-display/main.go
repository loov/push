// Command push3-display sends test patterns to the Push 3 display over USB.
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

	"github.com/loov/push3/push"
)

func main() {
	pattern := flag.String("pattern", "bars", "test pattern: solid, bars, gradient, plasma")
	flag.Parse()

	if err := run(context.Background(), *pattern); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, pattern string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	d, err := push.OpenDisplay()
	if err != nil {
		return fmt.Errorf("opening display: %w", err)
	}
	defer d.Close()

	log.Printf("Display connected (%dx%d). Pattern: %s. Ctrl+C to exit.",
		push.DisplayWidth, push.DisplayHeight, pattern)

	switch pattern {
	case "solid":
		d.Fill(0, 100, 200)
		return sendStatic(ctx, d)
	case "bars":
		drawBars(d)
		return sendStatic(ctx, d)
	case "gradient":
		drawGradient(d)
		return sendStatic(ctx, d)
	case "plasma":
		return sendPlasma(ctx, d)
	default:
		return fmt.Errorf("unknown pattern %q", pattern)
	}
}

// sendStatic sends the same frame continuously at 60fps until cancelled.
func sendStatic(ctx context.Context, d *push.Display) error {
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.Clear()
			_ = d.SendFrame()
			return nil
		case <-ticker.C:
			if err := d.SendFrame(); err != nil {
				return err
			}
		}
	}
}

func drawBars(d *push.Display) {
	type bar struct{ r, g, b uint8 }
	colors := []bar{
		{255, 0, 0},     // red
		{0, 255, 0},     // green
		{0, 0, 255},     // blue
		{255, 255, 255}, // white
		{0, 0, 0},       // black
	}
	barWidth := push.DisplayWidth / len(colors)
	for y := range push.DisplayHeight {
		for x := range push.DisplayWidth {
			idx := x / barWidth
			if idx >= len(colors) {
				idx = len(colors) - 1
			}
			c := colors[idx]
			d.SetPixel(x, y, c.r, c.g, c.b)
		}
	}
}

func drawGradient(d *push.Display) {
	for y := range push.DisplayHeight {
		for x := range push.DisplayWidth {
			v := uint8(x * 255 / (push.DisplayWidth - 1))
			d.SetPixel(x, y, v, v, v)
		}
	}
}

func sendPlasma(ctx context.Context, d *push.Display) error {
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	frame := 0
	for {
		select {
		case <-ctx.Done():
			d.Clear()
			_ = d.SendFrame()
			return nil
		case <-ticker.C:
		}

		t := float64(frame) * 0.03
		for y := range push.DisplayHeight {
			fy := float64(y) / float64(push.DisplayHeight)
			for x := range push.DisplayWidth {
				fx := float64(x) / float64(push.DisplayWidth)
				v := math.Sin(fx*6 + t)
				v += math.Sin(fy*8 + t*1.3)
				v += math.Sin((fx+fy)*5 + t*0.7)
				v += math.Sin(math.Sqrt(fx*fx+fy*fy)*8 + t*1.1)
				// Map [-4, 4] → [0, 1]
				n := (v + 4) / 8
				r, g, b := hsvToRGB(n*360, 1, 1)
				d.SetPixel(x, y, r, g, b)
			}
		}
		if err := d.SendFrame(); err != nil {
			return err
		}
		frame++
	}
}

func hsvToRGB(h, s, v float64) (uint8, uint8, uint8) {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
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
