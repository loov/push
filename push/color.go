package push

// SetPadColorAnimated sets a pad LED color with an animation mode.
// The animation type determines the MIDI channel used.
func (p *Push3) SetPadColorAnimated(pos PadPosition, paletteIndex uint8, anim Animation) error {
	status := 0x90 | (uint8(anim) & 0x0F)
	return p.output.Send([]byte{status, pos.PadNote(), paletteIndex})
}

// SetButtonColorAnimated sets a button LED color with an animation mode.
func (p *Push3) SetButtonColorAnimated(button ButtonID, paletteIndex uint8, anim Animation) error {
	status := 0xB0 | (uint8(anim) & 0x0F)
	return p.output.Send([]byte{status, byte(button), paletteIndex})
}

// SetPaletteEntry modifies one of the 128 color palette entries via SysEx.
// Call ReapplyPalette after modifying entries to apply changes.
func (p *Push3) SetPaletteEntry(index uint8, c Color) error {
	rL, rM := encodePaletteColor(c.R)
	gL, gM := encodePaletteColor(c.G)
	bL, bM := encodePaletteColor(c.B)
	// White channel set to 0 (RGB-only).
	return p.SendSysEx([]byte{0x03, index & 0x7F, rL, rM, gL, gM, bL, bM, 0, 0})
}

// ReapplyPalette tells the Push 3 to apply any modified palette entries.
func (p *Push3) ReapplyPalette() error {
	return p.SendSysEx([]byte{0x05})
}

// SetBrightness sets the global LED brightness (0-127).
func (p *Push3) SetBrightness(level uint8) error {
	if level > 127 {
		level = 127
	}
	return p.SendSysEx([]byte{0x06, level})
}

// SetTouchStripConfig configures the touch strip behavior via SysEx.
func (p *Push3) SetTouchStripConfig(cfg TouchStripConfig) error {
	return p.SendSysEx([]byte{0x17, encodeTouchStripConfig(cfg)})
}

// SetTouchStripLEDs sets the 31 touch strip LEDs via SysEx.
// Each LED uses a 3-bit color index (0-7). LEDs are ordered bottom to top.
func (p *Push3) SetTouchStripLEDs(leds [31]uint8) error {
	var buf [17]byte
	buf[0] = 0x19
	for i := range 16 {
		var lo, hi uint8
		idx := i * 2
		if idx < 31 {
			lo = leds[idx] & 0x07
		}
		if idx+1 < 31 {
			hi = leds[idx+1] & 0x07
		}
		buf[1+i] = (hi << 3) | lo
	}
	return p.SendSysEx(buf[:])
}

// SendMIDIClock sends a MIDI Clock tick (0xF8).
// LED animations synchronize to these ticks (24 per quarter note).
func (p *Push3) SendMIDIClock() error {
	return p.output.Send([]byte{0xF8})
}

// encodePaletteColor splits an 8-bit color value into two 7-bit MIDI bytes.
func encodePaletteColor(v uint8) (lsb, msb uint8) {
	return v & 0x7F, (v >> 7) & 0x01
}

// encodeTouchStripConfig converts a TouchStripConfig to a 7-bit flags byte.
func encodeTouchStripConfig(cfg TouchStripConfig) uint8 {
	var flags uint8
	if cfg.HostControl {
		flags |= 1 << 0
	}
	if cfg.HostSendsSysEx {
		flags |= 1 << 1
	}
	if cfg.ModWheel {
		flags |= 1 << 2
	}
	if cfg.PointMode {
		flags |= 1 << 3
	}
	if cfg.BarFromCenter {
		flags |= 1 << 4
	}
	if cfg.AutoReturn {
		flags |= 1 << 5
	}
	if cfg.ReturnToCenter {
		flags |= 1 << 6
	}
	return flags
}
