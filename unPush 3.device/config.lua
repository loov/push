-- unPush 3 — Unofficial Push 3 MIDI Device Script for Logic Pro
-- Requires Lua 5.1 (Logic Pro's MDS runtime).

local PORT = 'Ableton Push 3 Live Port'

function controller_info()
	local items = {}

	local function button(name, cc)
		items[#items+1] = {name=name, objectType='Button', midiType='Momentary',
			midi={0xB0, cc, MIDI_LSB}, inport=PORT, outport=PORT}
	end

	local function knob(name, cc)
		items[#items+1] = {name=name, objectType='Knob', midiType='RelativeSM',
			midi={0xB0, cc, MIDI_LSB}, inport=PORT, outport=PORT}
	end

	local function touch(name, note)
		items[#items+1] = {name=name, objectType='Button', midiType='Note',
			midi={0x90, note, MIDI_LSB}, inport=PORT, outport=PORT}
	end

	local function pad(name, note)
		items[#items+1] = {name=name, objectType='Drumpad', midiType='Note',
			midi={0x90, note, MIDI_LSB}, inport=PORT, outport=PORT}
	end

	-- Transport.
	button('Play or Stop', 85)
	button('Record', 86)
	button('Stop Clip', 29)
	button('Metronome', 9)
	button('Tap Tempo', 3)
	button('Quantize', 116)
	button('Automate', 89)
	button('Capture', 65)
	button('New', 92)
	button('Save', 82)
	button('Double Loop', 117)
	button('Fixed Length', 90)

	-- Action buttons.
	button('Undo', 119)
	button('Delete', 118)
	button('Duplicate', 88)
	button('Convert', 35)

	-- Mode + modifier buttons.
	button('Note', 50)
	button('Mix', 112)
	button('Device', 110)
	button('Clip', 113)
	button('Session Screen', 34)
	button('Session Pad', 51)
	button('Scale', 58)
	button('Layout', 31)
	button('Shift', 49)
	button('Select', 48)
	button('Accent', 57)
	button('Repeat', 56)
	button('Mute', 60)
	button('Solo', 61)
	button('Lock', 83)

	-- Top-left row.
	button('Sets', 80)
	button('Setup', 30)
	button('Learn', 81)
	button('User', 59)

	-- Display area.
	button('Add', 32)
	button('Swap', 33)

	-- Navigation.
	button('Up', 46)
	button('Down', 47)
	button('Left', 44)
	button('Right', 45)
	button('D-Pad Center', 91)
	button('Octave Up', 55)
	button('Octave Down', 54)
	button('Page Left', 62)
	button('Page Right', 63)

	-- Encoder presses.
	button('Volume Press', 111)
	button('Swing/Tempo Press', 15)
	button('Jog Click', 94)
	button('Jog Push Left', 93)
	button('Jog Push Right', 95)

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

	-- Scene buttons (CC 43 down to 36).
	for i = 1, 8 do
		button('Scene ' .. i, 44 - i)
	end

	-- Track encoders 1-8 (CC 71-78).
	for i = 1, 8 do
		knob('Encoder ' .. i, 70 + i)
	end

	-- Volume encoder (CC 79).
	knob('Volume', 79)

	-- Swing/Tempo encoder (CC 14).
	knob('Swing/Tempo', 14)

	-- Jog wheel (CC 70).
	knob('Jog Wheel', 70)

	-- Encoder touch notes.
	for i = 1, 8 do
		touch('Encoder Touch ' .. i, i - 1)
	end
	touch('Volume Touch', 8)
	touch('Swing/Tempo Touch', 10)
	touch('Jog Touch', 11)
	touch('Touch Strip Touch', 12)
	touch('D-Pad Center Touch', 13)

	-- 64 pads (notes 36-99, 8x8 grid).
	for row = 0, 7 do
		for col = 0, 7 do
			local note = 92 - row * 8 + col
			pad('Pad ' .. (row + 1) .. '/' .. (col + 1), note)
		end
	end

	-- Touch strip (pitch bend on channel 0).
	items[#items+1] = {name='Touch Strip', objectType='Wheel', midiType='PitchBend',
		midi={0xE0, MIDI_MSB, MIDI_LSB}, inport=PORT, outport=PORT}

	return {
		manufacturer = 'Ableton',
		model = 'Ableton Push 3',

		auto_passthrough = false,
		ignore_notes = true,

		-- Auto-detect Push 3 via Universal Device Inquiry.
		device_request = {0xF0, 0x7E, 0x7F, 0x06, 0x01, 0xF7},
		device_inquiry = {0xF0, 0x7E, 0x00, 0x06, 0x02,
			0x00, 0x21, 0x1D,  -- Ableton manufacturer ID
			MIDI_Wildcard, MIDI_Wildcard,
			MIDI_Wildcard, MIDI_Wildcard,
			MIDI_Wildcard, MIDI_Wildcard,
			MIDI_Wildcard, MIDI_Wildcard, 0xF7},

		items = items,
	}
end

function controller_initialize(appName, newlyDetected)
	print("unPush3 INIT", appName, newlyDetected)
	-- Set LED brightness to max.
	return {
		midi={0xF0, 0x00, 0x21, 0x1D, 0x01, 0x01, 0x06, 0x7F, 0xF7},
		outport=PORT,
	}
end

function controller_finalize()
	-- Reset LED brightness to default.
	return {
		midi={0xF0, 0x00, 0x21, 0x1D, 0x01, 0x01, 0x06, 0x40, 0xF7},
		outport=PORT,
	}
end

function controller_names(channel)
	if channel == 0 then
		return {
			[71] = "Encoder 1", [72] = "Encoder 2",
			[73] = "Encoder 3", [74] = "Encoder 4",
			[75] = "Encoder 5", [76] = "Encoder 6",
			[77] = "Encoder 7", [78] = "Encoder 8",
			[79] = "Volume",    [14] = "Swing/Tempo",
			[70] = "Jog Wheel",
		}
	end
	return {}
end

function controller_midi_in(midi, port)
	return nil
end
