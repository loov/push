# Ableton Push 3 MIDI Reference

All data discovered empirically using the `cmd/push3-discover` tool
on the **Live Port** (`Ableton Push 3 Live Port`).

Push 3 exposes three MIDI ports:
- **Live Port** — control protocol (buttons, encoders, pads, LEDs)
- **User Port** — user-mappable MIDI
- **External Port** — pass-through for USB-connected MIDI devices

## Buttons

All buttons send **Control Change** on channel 0.
Press = val 127, release = val 0 (except Volume Press, which is inverted).

### Top-left row

| Button     | CC  | Hex  |
|------------|-----|------|
| Sets       |  80 | 0x50 |
| Setup      |  30 | 0x1E |
| Learn      |  81 | 0x51 |
| User       |  59 | 0x3B |

### Top-right row

| Button     | CC  | Hex  |
|------------|-----|------|
| Device     | 110 | 0x6E |
| Mix        | 112 | 0x70 |
| Clip       | 113 | 0x71 |
| Session    |  34 | 0x22 |

### Display area

| Button     | CC  | Hex  |
|------------|-----|------|
| Undo       | 119 | 0x77 |
| Save       |  82 | 0x52 |
| Add        |  32 | 0x20 |
| Swap       |  33 | 0x21 |

### Bottom-left row

| Button     | CC  | Hex  |
|------------|-----|------|
| Lock       |  83 | 0x53 |
| Stop Clip  |  29 | 0x1D |
| Mute       |  60 | 0x3C |
| Solo       |  61 | 0x3D |

### Left side (top to bottom)

| Button       | CC  | Hex  |
|--------------|-----|------|
| Tap Tempo    |   3 | 0x03 |
| Metronome    |   9 | 0x09 |
| Quantize     | 116 | 0x74 |
| Fixed Length |  90 | 0x5A |
| Automate     |  89 | 0x59 |
| New          |  92 | 0x5C |
| Capture      |  65 | 0x41 |
| Record       |  86 | 0x56 |
| Play         |  85 | 0x55 |

### Right side

| Button       | CC  | Hex  |
|--------------|-----|------|
| Note         |  50 | 0x32 |
| Session (R)  |  51 | 0x33 |
| Scale        |  58 | 0x3A |
| Layout       |  31 | 0x1F |
| Repeat       |  56 | 0x38 |
| Accent       |  57 | 0x39 |
| Double Loop  | 117 | 0x75 |
| Duplicate    |  88 | 0x58 |
| Convert      |  35 | 0x23 |
| Delete       | 118 | 0x76 |

### D-pad

| Button       | CC  | Hex  | Notes |
|--------------|-----|------|-------|
| Up           |  46 | 0x2E | |
| Down         |  47 | 0x2F | |
| Left         |  44 | 0x2C | |
| Right        |  45 | 0x2D | |
| Center click |  91 | 0x5B | CC press/release |
| Center touch |  —  |  —   | Note 13 on/off (touch only, not click) |

### Navigation (Octave/Page)

| Button       | CC  | Hex  |
|--------------|-----|------|
| Octave Up    |  55 | 0x37 |
| Octave Down  |  54 | 0x36 |
| Page Left    |  62 | 0x3E |
| Page Right   |  63 | 0x3F |

### Bottom-right

| Button     | CC  | Hex  |
|------------|-----|------|
| Shift      |  49 | 0x31 |
| Select     |  48 | 0x30 |

### Display buttons

| Button       | CC  | Hex  |
|--------------|-----|------|
| Upper 1      | 102 | 0x66 |
| Upper 2      | 103 | 0x67 |
| Upper 3      | 104 | 0x68 |
| Upper 4      | 105 | 0x69 |
| Upper 5      | 106 | 0x6A |
| Upper 6      | 107 | 0x6B |
| Upper 7      | 108 | 0x6C |
| Upper 8      | 109 | 0x6D |
| Lower 1      |  20 | 0x14 |
| Lower 2      |  21 | 0x15 |
| Lower 3      |  22 | 0x16 |
| Lower 4      |  23 | 0x17 |
| Lower 5      |  24 | 0x18 |
| Lower 6      |  25 | 0x19 |
| Lower 7      |  26 | 0x1A |
| Lower 8      |  27 | 0x1B |
| Master       |  28 | 0x1C |

### Time division / scene buttons

| Button     | CC  | Hex  |
|------------|-----|------|
| 1/32t      |  43 | 0x2B |
| 1/32       |  42 | 0x2A |
| 1/16t      |  41 | 0x29 |
| 1/16       |  40 | 0x28 |
| 1/8t       |  39 | 0x27 |
| 1/8        |  38 | 0x26 |
| 1/4t       |  37 | 0x25 |
| 1/4        |  36 | 0x24 |

## Encoders

All encoders send **Control Change** on channel 0 with relative two's
complement encoding: values 1-63 = clockwise, 65-127 = counter-clockwise.

| Encoder      | Rotation CC | Touch Note | Click       |
|--------------|-------------|------------|-------------|
| Track 1      |  71 (0x47)  | Note 0     | same as touch |
| Track 2      |  72 (0x48)  | Note 1     | same as touch |
| Track 3      |  73 (0x49)  | Note 2     | same as touch |
| Track 4      |  74 (0x4A)  | Note 3     | same as touch |
| Track 5      |  75 (0x4B)  | Note 4     | same as touch |
| Track 6      |  76 (0x4C)  | Note 5     | same as touch |
| Track 7      |  77 (0x4D)  | Note 6     | same as touch |
| Track 8      |  78 (0x4E)  | Note 7     | same as touch |
| Volume       |  79 (0x4F)  | Note 8     | CC 111 (inverted: val=0 press, val=127 release) |
| Swing/Tempo  |  14 (0x0E)  | Note 10    | CC 15 (val=127 press, val=0 release) |
| Jog Wheel    |  70 (0x46)  | Note 11    | CC 94 click, CC 93 push-left, CC 95 push-right |

### Volume encoder press (inverted polarity)

The Volume encoder click sends CC 111 with **inverted** polarity:
- Press down: `CC 111 val=0`
- Release: `CC 111 val=127`

Full event sequence for a Volume press:
1. `Note On 8 vel=127` (touch)
2. `CC 111 val=0` (press)
3. `CC 111 val=127` (release)
4. `Note Off 8` (untouch)

### Swing/Tempo encoder

Single encoder with one rotation CC (14). Click sends CC 15.
Note 9 is unused.

### Jog wheel actions

| Action      | Type | Number |
|-------------|------|--------|
| Touch       | Note | 11     |
| Rotate      | CC   | 70     |
| Click       | CC   | 94     |
| Push Left   | CC   | 93     |
| Push Right  | CC   | 95     |

## Pads

The 8×8 pad grid uses **MPE (MIDI Polyphonic Expression)**. Each pad
press is assigned a unique MIDI channel (1-15). Channel 0 is the MPE
manager channel and is not used for pad notes.

### Pad note mapping

Notes 36-99, mapped as: `note = 92 - row*8 + col`

| Position | Note |
|----------|------|
| (0,0) top-left | 92 |
| (0,7) top-right | 99 |
| (7,0) bottom-left | 36 |
| (7,7) bottom-right | 43 |

Reverse: `row = (99 - note) / 8`, `col = (note - 36) % 8`

### Events per pad press (on MPE channel N)

| Order | Message | Data |
|-------|---------|------|
| 1 | Channel Pressure (ch N) | Initial pressure (often 0) |
| 2 | CC 74 (ch N) | MPE Slide — vertical finger position (0-127) |
| 3 | Note On (ch N, note 36-99) | Velocity (1-127) |
| 4 | Pitch Bend (ch N) | Horizontal position (center = 8192) |
| 5 | Channel Pressure (ch N) | Ongoing pressure updates (0-127) |
| 6 | CC 74 (ch N) | Ongoing slide updates |
| 7 | Pitch Bend (ch N) | Ongoing horizontal updates |
| 8 | Note Off (ch N) | Release |

### Sliding across pad rows

When a finger slides from pad A to pad B (typically across rows), the
Push 3 reuses the same MPE channel and sends an unusual sequence —
**no Note On for the destination pad**:

```
Note On  A  ch=N  vel=V      ← finger touches pad A
CC 74       ch=N  val=...    ← slide value increases toward 127
ChanPressure ch=N val=...    ← pressure updates
...
CC 74       ch=N  val=127    ← finger reaches top edge of pad A
Note Off A  ch=N             ← pad A released (finger crossing boundary)
CC 74       ch=N  val=0      ← slide resets (entering pad B from bottom)
ChanPressure ch=N val=...    ← pressure continues on same channel
CC 74       ch=N  val=...    ← slide value increases (moving up pad B)
PitchBend    ch=N val=8192   ← horizontal position resets
...
Note Off B  ch=N             ← finger lifts off pad B
```

Key observations:
- The MPE channel (N) is **reused** for the entire slide gesture.
- **No Note On is sent for pad B** — the transition is implicit.
- CC 74 (Slide) going from 127 → 0 indicates crossing a pad boundary.
- The Note Off for pad B identifies the destination pad.
- Channel Pressure and CC 74 continue flowing between the two Note Offs
  (these are "orphan" events with no active Note On).

To detect slides programmatically:
1. Track which note is active per MPE channel.
2. On Note Off: if the note differs from the channel's active note,
   a slide occurred — the Note Off note is the destination pad.
3. Between the two Note Offs, CC 74 and pressure data on the channel
   relate to the finger position on the destination pad.

Use `cmd/push3-padlog` to visualize these sequences in real time.

## Touch Strip

| Event    | Message | Channel | Data |
|----------|---------|---------|------|
| Touch    | Note On 12 | 0 | vel=127 |
| Position | Pitch Bend | 0 | 0-16383 |
| Release  | Note Off 12 | 0 | |

The touch strip sends Pitch Bend on channel 0. MPE pad pitch bends
use channels 1-15, so there is no ambiguity.

## Summary of all Note numbers

| Note | Function |
|------|----------|
| 0-7  | Encoder touch (Track 1-8) |
| 8    | Encoder touch (Volume) |
| 9    | (unused) |
| 10   | Encoder touch (Swing/Tempo) |
| 11   | Encoder touch (Jog wheel) |
| 12   | Touch strip touch |
| 13   | D-pad center touch |
| 14-35 | (unused for notes) |
| 36-99 | Pad notes (8×8 grid, MPE channels) |

## Summary of all CC numbers

| CC  | Function |
|-----|----------|
| 3   | Tap Tempo |
| 9   | Metronome |
| 14  | Swing/Tempo rotation |
| 15  | Swing/Tempo click |
| 20-27 | Lower display buttons 1-8 |
| 28  | Master button |
| 29  | Stop Clip |
| 30  | Setup |
| 31  | Layout |
| 32  | Add |
| 33  | Swap |
| 34  | Session |
| 35  | Convert |
| 36-43 | Time division (1/4, 1/4t, 1/8, 1/8t, 1/16, 1/16t, 1/32, 1/32t) |
| 44  | Left |
| 45  | Right |
| 46  | Up |
| 47  | Down |
| 48  | Select |
| 49  | Shift |
| 50  | Note |
| 51  | Session (right) |
| 54  | Octave Down |
| 55  | Octave Up |
| 56  | Repeat |
| 57  | Accent |
| 58  | Scale |
| 59  | User |
| 60  | Mute |
| 61  | Solo |
| 62  | Page Left |
| 63  | Page Right |
| 65  | Capture |
| 70  | Jog wheel rotation |
| 71-78 | Track encoder 1-8 rotation |
| 74  | MPE Slide (on MPE channels, not ch 0) |
| 79  | Volume encoder rotation |
| 80  | Sets |
| 81  | Learn |
| 82  | Save |
| 83  | Lock |
| 85  | Play |
| 86  | Record |
| 88  | Duplicate |
| 89  | Automate |
| 90  | Fixed Length |
| 91  | D-pad center click |
| 92  | New |
| 93  | Jog push left |
| 94  | Jog click |
| 95  | Jog push right |
| 102-109 | Upper display buttons 1-8 |
| 110 | Device |
| 111 | Volume press (inverted polarity) |
| 112 | Mix |
| 113 | Clip |
| 116 | Quantize |
| 117 | Double Loop |
| 118 | Delete |
| 119 | Undo |
