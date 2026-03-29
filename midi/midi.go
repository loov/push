// Package midi provides a thin wrapper around CoreMIDI for sending and receiving MIDI messages.
package midi

import (
	"fmt"
	"strings"

	coremidi "github.com/youpy/go-coremidi"
)

// Client holds a CoreMIDI client reference.
type Client struct {
	client coremidi.Client
}

// NewClient creates a new CoreMIDI client with the given name.
func NewClient(name string) (*Client, error) {
	c, err := coremidi.NewClient(name)
	if err != nil {
		return nil, fmt.Errorf("midi: creating client %q: %w", name, err)
	}
	return &Client{client: c}, nil
}

// Close releases the CoreMIDI client.
func (c *Client) Close() error {
	return c.client.Close()
}

// Source represents a MIDI input source (something we read from).
type Source struct {
	source coremidi.Source
}

// Name returns the source's display name.
func (s Source) Name() string { return s.source.Name() }

// Destination represents a MIDI output destination (something we write to).
type Destination struct {
	dest coremidi.Destination
}

// Name returns the destination's display name.
func (d Destination) Name() string { return d.dest.Name() }

// Sources returns all available MIDI input sources.
func Sources() ([]Source, error) {
	all, err := coremidi.AllSources()
	if err != nil {
		return nil, fmt.Errorf("midi: listing sources: %w", err)
	}
	result := make([]Source, len(all))
	for i, s := range all {
		result[i] = Source{source: s}
	}
	return result, nil
}

// Destinations returns all available MIDI output destinations.
func Destinations() ([]Destination, error) {
	all, err := coremidi.AllDestinations()
	if err != nil {
		return nil, fmt.Errorf("midi: listing destinations: %w", err)
	}
	result := make([]Destination, len(all))
	for i, d := range all {
		result[i] = Destination{dest: d}
	}
	return result, nil
}

// FindSource finds a source whose name contains the given substring.
func FindSource(name string) (Source, error) {
	sources, err := Sources()
	if err != nil {
		return Source{}, err
	}
	for _, s := range sources {
		if strings.Contains(s.Name(), name) {
			return s, nil
		}
	}
	available := make([]string, len(sources))
	for i, s := range sources {
		available[i] = s.Name()
	}
	return Source{}, fmt.Errorf("midi: source %q not found, available: %v", name, available)
}

// FindDestination finds a destination whose name contains the given substring.
func FindDestination(name string) (Destination, error) {
	dests, err := Destinations()
	if err != nil {
		return Destination{}, err
	}
	for _, d := range dests {
		if strings.Contains(d.Name(), name) {
			return d, nil
		}
	}
	available := make([]string, len(dests))
	for i, d := range dests {
		available[i] = d.Name()
	}
	return Destination{}, fmt.Errorf("midi: destination %q not found, available: %v", name, available)
}

// InputPort receives MIDI messages from a connected source.
type InputPort struct {
	port coremidi.InputPort
}

// OpenInput creates an input port and connects it to the given source.
// The callback receives raw MIDI data for each incoming message.
func (c *Client) OpenInput(name string, source Source, callback func(data []byte)) (*InputPort, error) {
	port, err := coremidi.NewInputPort(c.client, name, func(_ coremidi.Source, packet coremidi.Packet) {
		callback(packet.Data)
	})
	if err != nil {
		return nil, fmt.Errorf("midi: opening input port %q: %w", name, err)
	}
	if _, err := port.Connect(source.source); err != nil {
		return nil, fmt.Errorf("midi: connecting input port %q to %q: %w", name, source.Name(), err)
	}
	return &InputPort{port: port}, nil
}

// OutputPort sends MIDI messages to a destination.
type OutputPort struct {
	port coremidi.OutputPort
	dest coremidi.Destination
}

// OpenOutput creates an output port connected to the given destination.
func (c *Client) OpenOutput(name string, dest Destination) (*OutputPort, error) {
	port, err := coremidi.NewOutputPort(c.client, name)
	if err != nil {
		return nil, fmt.Errorf("midi: opening output port %q: %w", name, err)
	}
	return &OutputPort{port: port, dest: dest.dest}, nil
}

// Send sends raw MIDI data to the destination.
func (o *OutputPort) Send(data []byte) error {
	packet := coremidi.NewPacket(data, 0)
	return packet.Send(&o.port, &o.dest)
}
