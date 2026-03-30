// Command logic-push3 bridges Ableton Push 3 to Apple Logic Pro via the
// Mackie Control Universal (MCU) protocol over IAC Driver virtual MIDI ports.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/loov/push/mcu"
	"github.com/loov/push/midi"
)

func main() {
	mcuIn := flag.String("mcu-in", "Logic Pro Virtual In", "MIDI destination name (app→Logic)")
	mcuOut := flag.String("mcu-out", "Logic Pro Virtual Out", "MIDI source name (Logic→app)")
	debug := flag.Bool("debug", false, "print all incoming MIDI messages")
	flag.Parse()

	if err := run(context.Background(), *mcuIn, *mcuOut, *debug); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, mcuInName, mcuOutName string, debug bool) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	client, err := midi.NewClient("logic-push3")
	if err != nil {
		return fmt.Errorf("creating MIDI client: %w", err)
	}
	defer client.Close()

	// Find IAC Driver ports.
	// MCU "In" port is where Logic *reads from* (we write to it).
	// MCU "Out" port is where Logic *writes to* (we read from it).
	outDest, err := midi.FindDestination(mcuInName)
	if err != nil {
		return fmt.Errorf("finding MCU output destination: %w", err)
	}
	inSource, err := midi.FindSource(mcuOutName)
	if err != nil {
		return fmt.Errorf("finding MCU input source: %w", err)
	}

	output, err := client.OpenOutput("mcu-out", outDest)
	if err != nil {
		return fmt.Errorf("opening MCU output: %w", err)
	}

	state := mcu.NewState()

	// Open input and process incoming MCU messages.
	_, err = client.OpenInput("mcu-in", inSource, func(data []byte) {
		if debug {
			log.Printf("[MIDI] % X", data)
		}

		msg := mcu.Parse(data)

		// Handle handshake messages (keepalive, serial request).
		if msg.Kind == mcu.MsgSysEx {
			if resp := mcu.HandleHandshake(msg.SysExData); resp != nil {
				if err := output.Send(resp); err != nil {
					log.Printf("[MCU] handshake send error: %v", err)
				}
				if debug {
					log.Printf("[MCU] handshake response sent: % X", resp)
				}
				return
			}
		}

		// Update state and log changes.
		if desc := state.Handle(msg); desc != "" {
			log.Printf("[MCU] %s", desc)
		}
	})
	if err != nil {
		return fmt.Errorf("opening MCU input: %w", err)
	}

	// Send Device Inquiry to initiate handshake.
	if err := output.Send(mcu.DeviceInquiry); err != nil {
		return fmt.Errorf("sending device inquiry: %w", err)
	}
	log.Println("[MCU] Device Inquiry sent, waiting for Logic Pro...")

	log.Printf("[MCU] Connected: reading from %q, writing to %q", inSource.Name(), outDest.Name())
	log.Println("[MCU] Press Ctrl+C to exit")

	// Wait for interrupt.
	<-ctx.Done()
	log.Println("[MCU] Shutting down")
	return nil
}
