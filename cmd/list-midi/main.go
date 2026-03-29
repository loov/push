// Command list-midi prints all available MIDI sources and destinations.
package main

import (
	"fmt"
	"log"

	"github.com/loov/logic-push3/midi"
)

func main() {
	sources, err := midi.Sources()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("=== MIDI Sources (inputs we can read from) ===")
	for i, s := range sources {
		fmt.Printf("  [%d] name=%q display=%q manufacturer=%q\n", i, s.Name(), s.DisplayName(), s.Manufacturer())
	}

	dests, err := midi.Destinations()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\n=== MIDI Destinations (outputs we can write to) ===")
	for i, d := range dests {
		fmt.Printf("  [%d] name=%q display=%q manufacturer=%q\n", i, d.Name(), d.DisplayName(), d.Manufacturer())
	}
}
