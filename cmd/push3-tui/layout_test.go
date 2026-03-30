package main

import (
	"os"
	"strings"
	"testing"

	"github.com/loov/push/push3"
)

func TestLayoutMatchesReference(t *testing.T) {
	ref, err := os.ReadFile("ascii.txt")
	if err != nil {
		t.Fatalf("reading ascii.txt: %v", err)
	}

	m := model{
		buttons: make(map[push3.ButtonID]bool),
	}
	// The reference shows "( 1 )" etc for encoders — set encoder values to match.
	for i := range 8 {
		m.encoders[i+1] = i + 1
	}

	got := m.renderLayout()

	refLines := strings.Split(strings.TrimRight(string(ref), "\n"), "\n")
	gotLines := strings.Split(strings.TrimRight(got, "\n"), "\n")

	// Trim trailing whitespace from each line for comparison.
	for i := range refLines {
		refLines[i] = strings.TrimRight(refLines[i], " ")
	}
	for i := range gotLines {
		gotLines[i] = strings.TrimRight(gotLines[i], " ")
	}

	// Compare line counts.
	if len(gotLines) != len(refLines) {
		t.Errorf("line count: got %d, want %d", len(gotLines), len(refLines))
	}

	// Compare each line, showing first difference.
	maxLines := len(refLines)
	if len(gotLines) < maxLines {
		maxLines = len(gotLines)
	}

	diffs := 0
	for i := range maxLines {
		refLine := refLines[i]
		gotLine := gotLines[i]
		if gotLine != refLine {
			diffs++
			if diffs <= 10 {
				t.Errorf("line %d differs:\n  want: %q\n  got:  %q", i+1, refLine, gotLine)
			}
		}
	}
	if diffs > 10 {
		t.Errorf("... and %d more differing lines", diffs-10)
	}

	if diffs > 0 {
		// Write the actual output for debugging.
		os.WriteFile("ascii_got.txt", []byte(got), 0o644)
		t.Log("Wrote actual output to ascii_got.txt for comparison")
	}
}
