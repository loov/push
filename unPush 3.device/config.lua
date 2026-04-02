-- unPush 3 — Unofficial Push 3 MIDI Device Script for Logic Pro
--
-- Phase 1: Device Registration + Transport + Encoders

--------------------------------------------------------------------------------
-- [1] Constants
--------------------------------------------------------------------------------

local PORT = 'Ableton Push 3 Live Port'

-- SysEx prefix for Push 3 commands.
local SYSEX_PREFIX = {0xF0, 0x00, 0x21, 0x1D, 0x01, 0x01}

-- Palette indices (default Push 3 palette).
local BLACK     = 0
local ORANGE    = 3
local YELLOW    = 8
local TURQUOISE = 15
local PURPLE    = 22
local PINK      = 25
local WHITE     = 122
local BLUE      = 125
local GREEN     = 126
local RED       = 127

-- Button CCs.
local CC_SETS           = 80
local CC_SETUP          = 30
local CC_LEARN          = 81
local CC_USER           = 59
local CC_DEVICE         = 110
local CC_MIX            = 112
local CC_CLIP           = 113
local CC_SESSION_SCREEN = 34
local CC_UNDO           = 119
local CC_SAVE           = 82
local CC_ADD            = 32
local CC_SWAP           = 33
local CC_LOCK           = 83
local CC_STOP_CLIP      = 29
local CC_MUTE           = 60
local CC_SOLO           = 61
local CC_TAP_TEMPO      = 3
local CC_METRONOME      = 9
local CC_QUANTIZE       = 116
local CC_FIXED_LENGTH   = 90
local CC_AUTOMATE       = 89
local CC_NEW            = 92
local CC_CAPTURE        = 65
local CC_RECORD         = 86
local CC_PLAY           = 85
local CC_NOTE           = 50
local CC_SESSION_PAD    = 51
local CC_SCALE          = 58
local CC_LAYOUT         = 31
local CC_REPEAT         = 56
local CC_ACCENT         = 57
local CC_DOUBLE_LOOP    = 117
local CC_DUPLICATE      = 88
local CC_CONVERT        = 35
local CC_DELETE          = 118
local CC_UP             = 46
local CC_DOWN           = 47
local CC_LEFT           = 44
local CC_RIGHT          = 45
local CC_DPAD_CENTER    = 91
local CC_OCTAVE_UP      = 55
local CC_OCTAVE_DOWN    = 54
local CC_PAGE_LEFT      = 62
local CC_PAGE_RIGHT     = 63
local CC_SHIFT          = 49
local CC_SELECT         = 48
local CC_VOLUME_PRESS   = 111
local CC_SWING_PRESS    = 15
local CC_JOG_CLICK      = 94
local CC_JOG_PUSH_LEFT  = 93
local CC_JOG_PUSH_RIGHT = 95

-- Upper display buttons: CC 102-109.
-- Lower display buttons: CC 20-27.
-- Main Track: CC 28.
-- Scene buttons: CC 36-43 (scene 8=36, scene 1=43).

-- Encoder rotation CCs.
-- Track 1-8: CC 71-78, Volume: CC 79, Swing/Tempo: CC 14, Jog: CC 70.

-- Encoder touch notes.
-- Track 1-8: Notes 0-7, Volume: Note 8, Swing/Tempo: Note 10, Jog: Note 11.

-- Touch strip touch: Note 12.
-- D-pad center touch: Note 13.

-- Pad notes: 36-99 (note = 92 - row*8 + col).

--------------------------------------------------------------------------------
-- [6] Items — build all controller items
--------------------------------------------------------------------------------

local function build_items()
    local items = {}

    local function button(name, cc)
        items[#items+1] = {
            name = name,
            objectType = 'Button',
            midiType = 'Momentary',
            midi = {0xB0, cc, MIDI_LSB},
            inport = PORT,
            outport = PORT,
        }
    end

    local function knob(name, cc)
        items[#items+1] = {
            name = name,
            objectType = 'Knob',
            midiType = 'RelativeSigned',
            midi = {0xB0, cc, MIDI_LSB},
            inport = PORT,
            outport = PORT,
        }
    end

    local function touch(name, note)
        items[#items+1] = {
            name = name,
            objectType = 'Button',
            midiType = 'Note',
            midi = {0x90, note, MIDI_LSB},
            inport = PORT,
            outport = PORT,
        }
    end

    -- Transport buttons.
    button('Play', CC_PLAY)
    button('Record', CC_RECORD)
    button('Stop Clip', CC_STOP_CLIP)
    button('Metronome', CC_METRONOME)
    button('Tap Tempo', CC_TAP_TEMPO)
    button('Quantize', CC_QUANTIZE)
    button('Automate', CC_AUTOMATE)
    button('Capture', CC_CAPTURE)
    button('New', CC_NEW)
    button('Save', CC_SAVE)
    button('Double Loop', CC_DOUBLE_LOOP)
    button('Fixed Length', CC_FIXED_LENGTH)

    -- Action buttons.
    button('Undo', CC_UNDO)
    button('Delete', CC_DELETE)
    button('Duplicate', CC_DUPLICATE)
    button('Convert', CC_CONVERT)

    -- Mode + modifier buttons.
    button('Note', CC_NOTE)
    button('Mix', CC_MIX)
    button('Device', CC_DEVICE)
    button('Clip', CC_CLIP)
    button('Session Screen', CC_SESSION_SCREEN)
    button('Session Pad', CC_SESSION_PAD)
    button('Scale', CC_SCALE)
    button('Layout', CC_LAYOUT)
    button('Shift', CC_SHIFT)
    button('Select', CC_SELECT)
    button('Accent', CC_ACCENT)
    button('Repeat', CC_REPEAT)
    button('Mute', CC_MUTE)
    button('Solo', CC_SOLO)
    button('Lock', CC_LOCK)

    -- Top-left row.
    button('Sets', CC_SETS)
    button('Setup', CC_SETUP)
    button('Learn', CC_LEARN)
    button('User', CC_USER)

    -- Display area.
    button('Add', CC_ADD)
    button('Swap', CC_SWAP)

    -- Navigation.
    button('Up', CC_UP)
    button('Down', CC_DOWN)
    button('Left', CC_LEFT)
    button('Right', CC_RIGHT)
    button('D-Pad Center', CC_DPAD_CENTER)
    button('Octave Up', CC_OCTAVE_UP)
    button('Octave Down', CC_OCTAVE_DOWN)
    button('Page Left', CC_PAGE_LEFT)
    button('Page Right', CC_PAGE_RIGHT)

    -- Encoder presses.
    button('Volume Press', CC_VOLUME_PRESS)
    button('Swing/Tempo Press', CC_SWING_PRESS)
    button('Jog Click', CC_JOG_CLICK)
    button('Jog Push Left', CC_JOG_PUSH_LEFT)
    button('Jog Push Right', CC_JOG_PUSH_RIGHT)

    -- Upper display buttons (CC 102-109).
    for i = 1, 8 do
        button('Upper ' .. i, 101 + i)
    end

    -- Lower display buttons (CC 20-27).
    for i = 1, 8 do
        button('Lower ' .. i, 19 + i)
    end

    -- Main Track button.
    button('Main Track', 28)

    -- Scene buttons (CC 43 down to 36, scene 1=43, scene 8=36).
    for i = 1, 8 do
        button('Scene ' .. i, 44 - i)
    end

    -- Track encoders 1-8 (CC 71-78, sign-magnitude relative).
    for i = 1, 8 do
        knob('Encoder ' .. i, 70 + i)
    end

    -- Volume encoder (CC 79).
    knob('Volume', 79)

    -- Swing/Tempo encoder (CC 14).
    knob('Swing/Tempo', 14)

    -- Jog wheel (CC 70).
    knob('Jog Wheel', 70)

    -- Encoder touch notes (0-7 = track, 8 = volume, 10 = swing, 11 = jog).
    for i = 1, 8 do
        touch('Encoder Touch ' .. i, i - 1)
    end
    touch('Volume Touch', 8)
    touch('Swing/Tempo Touch', 10)
    touch('Jog Touch', 11)

    -- Touch strip touch (Note 12).
    touch('Touch Strip Touch', 12)

    -- D-Pad center touch (Note 13).
    touch('D-Pad Center Touch', 13)

    -- Touch strip (pitch bend on channel 0).
    items[#items+1] = {
        name = 'Touch Strip',
        objectType = 'Wheel',
        midiType = 'PitchBend',
        midi = {0xE0, MIDI_MSB, MIDI_LSB},
        inport = PORT,
        outport = PORT,
    }

    -- 64 pads (notes 36-99, 8x8 grid).
    for row = 0, 7 do
        for col = 0, 7 do
            local note = 92 - row * 8 + col
            items[#items+1] = {
                name = 'Pad ' .. (row + 1) .. '/' .. (col + 1),
                objectType = 'Drumpad',
                midiType = 'Note',
                midi = {0x90, note, MIDI_LSB},
                inport = PORT,
                outport = PORT,
            }
        end
    end

    return items
end

--------------------------------------------------------------------------------
-- [8] MDS API
--------------------------------------------------------------------------------

function controller_info()
    return {
        manufacturer = 'Ableton',
        model = 'unPush 3',
        auto_passthrough = false,
        supports_feedback = true,
        preset_name = 'unPush 3 \xe2\x80\x94 Connect via USB, use Live Port',
        inport = PORT,
        outport = PORT,

        -- Auto-detect Push 3 via Universal Device Inquiry.
        device_request = {0xF0, 0x7E, 0x7F, 0x06, 0x01, 0xF7},
        device_inquiry = {0xF0, 0x7E, 0x00, 0x06, 0x02,
                          0x00, 0x21, 0x1D,
                          MIDI_Wildcard, MIDI_Wildcard,
                          MIDI_Wildcard, MIDI_Wildcard,
                          MIDI_Wildcard, MIDI_Wildcard,
                          MIDI_Wildcard, MIDI_Wildcard, 0xF7},

        items = build_items(),
    }
end

function controller_initialize(appName, newlyDetected)
    -- Set LED brightness to max.
    local msg = {0xF0, 0x00, 0x21, 0x1D, 0x01, 0x01, 0x06, 0x7F, 0xF7}

    -- Light the Note button to indicate active mode.
    local note_led = {0xB0, CC_NOTE, GREEN}

    return {midi = msg}, {midi = note_led}
end

function controller_finalize()
    -- Turn off Note button LED.
    return {midi = {0xB0, CC_NOTE, BLACK}}
end

function controller_midi_in(midiEvent, portName)
    -- Only process Live Port events.
    if portName and not portName:match('Live') then
        return nil
    end

    local status = midiEvent[1] & 0xF0
    local ch = midiEvent[1] & 0x0F

    -- Block MPE channels (1-15) for now — Phase 2 handles pad remapping.
    if ch > 0 then
        return {}  -- block
    end

    -- Pass through channel 0 events (buttons, encoders, touch strip).
    return midiEvent
end

function controller_midi_out(midiEvent, name, valueString, color)
    return midiEvent
end

function controller_timer_trigger()
    return nil
end
