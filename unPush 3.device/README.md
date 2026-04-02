# unPush 3

Unofficial Logic Pro MIDI Device Script for Ableton Push 3. Turns Push 3
into a native Logic Pro control surface without requiring a bridge app.

## Installation

```
make install-midi-device-script
```

This copies the device bundle to `~/Music/Audio Music Apps/MIDI Device Scripts/`. Restart
Logic Pro to pick up changes.

## What works (Phase 1)

- **Device detection** — Logic Pro auto-detects Push 3 via USB (Live Port)
- **Transport** — Play, Record, Stop, Metronome, Tap Tempo, Quantize,
  Capture, Undo, Save, Delete, Duplicate, Double Loop, New
- **Encoders** — Track encoders 1-8, Volume, Swing/Tempo, Jog wheel
  (available in Controller Assignments via Cmd+K)
- **All buttons registered** — mode buttons, modifiers, navigation,
  d-pad, upper/lower display buttons, scene buttons
- **Touch sensors** — encoder touch, touch strip, d-pad center touch
- **Touch strip** — pitch bend on channel 0
- **64 pads** — registered as drumpads (pass-through only, no scale
  remapping yet)
- **LED feedback** — Note button lights green on connect

## Known limitations

- **No pad remapping** — pads pass raw MIDI notes, no scale-aware layouts yet
- **No display** — Push 3 display requires USB bulk transfer, not MIDI
- **No note repeat** — MDS cannot generate notes to Logic's instrument input
- **MPE blocked** — pad expression channels (1-15) are blocked until
  scale-aware remapping is implemented
