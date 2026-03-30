package mcu

import (
	"testing"

	"github.com/loov/push3/push3"
)

func TestParseLCD(t *testing.T) {
	tests := []struct {
		name     string
		payload  []byte
		wantPos  int
		wantText string
		wantOK   bool
	}{
		{
			name: "simple text at position 0",
			payload: []byte{
				0x00, 0x00, 0x66, 0x14, // prefix + model
				0x12,                    // LCD command
				0x00,                    // position 0
				'H', 'e', 'l', 'l', 'o', // text
			},
			wantPos:  0,
			wantText: "Hello",
			wantOK:   true,
		},
		{
			name: "text at position 7 (cell 1)",
			payload: []byte{
				0x00, 0x00, 0x66, 0x14,
				0x12,
				0x07,
				'T', 'r', 'a', 'c', 'k', '2', ' ',
			},
			wantPos:  7,
			wantText: "Track2 ",
			wantOK:   true,
		},
		{
			name: "non-printable chars replaced with spaces",
			payload: []byte{
				0x00, 0x00, 0x66, 0x14,
				0x12,
				0x00,
				'A', 0x01, 'B', 0x7F, 'C',
			},
			wantPos:  0,
			wantText: "A B C",
			wantOK:   true,
		},
		{
			name:    "wrong command",
			payload: []byte{0x00, 0x00, 0x66, 0x14, 0x13, 0x00},
			wantOK:  false,
		},
		{
			name:    "too short",
			payload: []byte{0x00, 0x00, 0x66, 0x14, 0x12},
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, text, ok := ParseLCD(tt.payload)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if pos != tt.wantPos {
				t.Errorf("pos = %d, want %d", pos, tt.wantPos)
			}
			if text != tt.wantText {
				t.Errorf("text = %q, want %q", text, tt.wantText)
			}
		})
	}
}

func TestApplyLCD(t *testing.T) {
	var lcd [2]push3.LCDRow
	for i := range lcd {
		for j := range lcd[i] {
			lcd[i][j] = ' '
		}
	}

	// Write "Track1 " to position 0 (top row, cell 0)
	ApplyLCD(&lcd, 0, "Track1 ")
	if got := lcd[0].Cell(0); got != "Track1" {
		t.Errorf("cell 0 = %q, want %q", got, "Track1")
	}

	// Write "Track2 " to position 7 (top row, cell 1)
	ApplyLCD(&lcd, 7, "Track2 ")
	if got := lcd[0].Cell(1); got != "Track2" {
		t.Errorf("cell 1 = %q, want %q", got, "Track2")
	}

	// Write to bottom row (position 56+)
	ApplyLCD(&lcd, 56, "Vol    ")
	if got := lcd[1].Cell(0); got != "Vol" {
		t.Errorf("bottom cell 0 = %q, want %q", got, "Vol")
	}
}

func TestLCDRowCell(t *testing.T) {
	var row push3.LCDRow
	copy(row[:], "Track1 Track2 Bass   Drums  Keys   Pad    Synth  Master ")

	tests := []struct {
		index int
		want  string
	}{
		{0, "Track1"},
		{1, "Track2"},
		{2, "Bass"},
		{3, "Drums"},
		{4, "Keys"},
		{5, "Pad"},
		{6, "Synth"},
		{7, "Master"},
		{-1, ""},
		{8, ""},
	}
	for _, tt := range tests {
		if got := row.Cell(tt.index); got != tt.want {
			t.Errorf("Cell(%d) = %q, want %q", tt.index, got, tt.want)
		}
	}
}
