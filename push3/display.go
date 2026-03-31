package push3

import (
	"errors"
	"fmt"

	"github.com/loov/push/usb"
)

// Display constants.
const (
	DisplayWidth  = 960
	DisplayHeight = 160
)

const (
	pixelBytes    = 2                                         // RGB565LE
	pixelsPerLine = DisplayWidth * pixelBytes                 // 1920
	fillerPerLine = 128                                       // padding per line
	bytesPerLine  = pixelsPerLine + fillerPerLine             // 2048
	headerSize    = 16                                        // frame header
	frameSize     = headerSize + DisplayHeight*bytesPerLine   // 327,696
	frameBufSize  = DisplayWidth * DisplayHeight * pixelBytes // 307,200
)

// Push 3 USB identifiers.
const (
	displayVendorID  = 0x2982
	displayProductID = 0x1969
	displayEndpoint  = 0x01
)

// frameHeader is the 16-byte frame header.
var frameHeader = [headerSize]byte{0xFF, 0xCC, 0xAA, 0x88}

// xorMask is the 4-byte repeating XOR pattern applied to pixel data per line.
var xorMask = [4]byte{0xE7, 0xF3, 0xE7, 0xFF}

// Display drives the Push 3's 960x160 BGR565 display over USB.
type Display struct {
	dev     *usb.Device
	frame   []byte // 307,200 bytes: pixel framebuffer (BGR565LE, row-major)
	sendbuf []byte // 16,384 bytes: one chunk of XOR-encoded lines with filler
}

// OpenDisplay opens the Push 3 USB display.
func OpenDisplay() (*Display, error) {
	dev, err := usb.Open(displayVendorID, displayProductID, displayEndpoint)
	if err != nil {
		return nil, fmt.Errorf("push: opening display: %w", err)
	}
	d := &Display{
		dev:     dev,
		frame:   make([]byte, frameBufSize),
		sendbuf: make([]byte, linesPerChunk*bytesPerLine),
	}
	return d, nil
}

// Close releases the USB device.
func (d *Display) Close() error {
	return d.dev.Close()
}

// SetPixel sets a pixel in the framebuffer.
// No bounds checking — caller must ensure 0 <= x < 960 and 0 <= y < 160.
func (d *Display) SetPixel(x, y int, r, g, b uint8) {
	off := (y*DisplayWidth + x) * pixelBytes
	lo, hi := rgb565(r, g, b)
	d.frame[off] = lo
	d.frame[off+1] = hi
}

// Clear fills the framebuffer with black.
func (d *Display) Clear() {
	clear(d.frame)
}

// Fill fills the entire framebuffer with one color.
func (d *Display) Fill(r, g, b uint8) {
	lo, hi := rgb565(r, g, b)
	for i := 0; i < len(d.frame); i += 2 {
		d.frame[i] = lo
		d.frame[i+1] = hi
	}
}

// Pix returns the raw framebuffer for direct manipulation.
// Layout is RGB565LE, row-major, stride = 1920 bytes per row.
func (d *Display) Pix() []byte { return d.frame }

// Stride returns the byte stride per row (1920).
func (d *Display) Stride() int { return pixelsPerLine }

// linesPerChunk controls how many lines are batched per USB transfer.
// 8 lines × 2048 bytes = 16,384 bytes per chunk.
const linesPerChunk = 8

// SendFrame assembles the frame (header + XOR-encoded lines + filler)
// and writes it to the USB device.
//
// If the USB pipe stalls, the stall is cleared and the frame is dropped.
// This is normal under load and the next frame will succeed.
func (d *Display) SendFrame() error {
	// Send header.
	if _, err := d.dev.Write(frameHeader[:]); err != nil {
		if errors.Is(err, usb.ErrPipeStalled) {
			return nil // drop frame
		}
		return fmt.Errorf("push: sending frame header: %w", err)
	}

	// Encode lines into sendbuf and send in chunks.
	for y := range DisplayHeight {
		src := y * pixelsPerLine
		dst := (y % linesPerChunk) * bytesPerLine
		for i := range pixelsPerLine {
			d.sendbuf[dst+i] = d.frame[src+i] ^ xorMask[i&3]
		}

		if (y+1)%linesPerChunk == 0 {
			chunk := d.sendbuf[:linesPerChunk*bytesPerLine]
			if _, err := d.dev.Write(chunk); err != nil {
				if errors.Is(err, usb.ErrPipeStalled) {
					return nil // drop frame
				}
				return fmt.Errorf("push: sending lines %d-%d: %w", y-linesPerChunk+1, y, err)
			}
		}
	}
	return nil
}

// rgb565 converts 8-bit RGB to the Push 3's 16-bit BGR565 pixel format
// in little-endian byte order.
// Layout: LSB = GGGRRRRR, MSB = BBBBBGGG.
func rgb565(r, g, b uint8) (lo, hi byte) {
	v := uint16(b>>3)<<11 | uint16(g>>2)<<5 | uint16(r>>3)
	return byte(v), byte(v >> 8)
}
