-- unPush 3 — Unofficial Push 3 MIDI Device Script for Logic Pro
-- Requires Lua 5.1 (Logic Pro's MDS runtime).

local PORT = 'Live Port'

--------------------------------------------------------------------------------
-- [1] Constants
--------------------------------------------------------------------------------

local BLACK     = 0
local GREEN     = 126

--------------------------------------------------------------------------------
-- [6] Items
--------------------------------------------------------------------------------

local nextID = 0
local function id()
	local v = nextID
	nextID = nextID + 1
	return v
end

local function button(name, cc)
	return {name=name, controlID=id(), midi={0xB0, cc, MIDI_LSB},
		inport=PORT, outport=PORT}
end

local function knob(name, cc)
	return {name=name, controlID=id(), midi={0xB0, cc, MIDI_LSB},
		fbType=FB_OFF, valueMode=kAssignScaled, selfFeedback=false,
		inport=PORT}
end

local function touch(name, note)
	return {name=name, controlID=id(), midi={MIDI_NoteOn, note, MIDI_LSB},
		inport=PORT, outport=PORT}
end

local function pad(name, note)
	return {name=name, controlID=id(), midi={MIDI_NoteOn, note, MIDI_LSB},
		inport=PORT, outport=PORT, valueMode=kAssignRotate}
end

local controls = {
	[0] = button('Play', 85),
}

-- Transport.
controls[#controls] = button('Record', 86)
controls[#controls] = button('Stop Clip', 29)
controls[#controls] = button('Metronome', 9)
controls[#controls] = button('Tap Tempo', 3)
controls[#controls] = button('Quantize', 116)
controls[#controls] = button('Automate', 89)
controls[#controls] = button('Capture', 65)
controls[#controls] = button('New', 92)
controls[#controls] = button('Save', 82)
controls[#controls] = button('Double Loop', 117)
controls[#controls] = button('Fixed Length', 90)

-- Action buttons.
controls[#controls] = button('Undo', 119)
controls[#controls] = button('Delete', 118)
controls[#controls] = button('Duplicate', 88)
controls[#controls] = button('Convert', 35)

-- Mode + modifier buttons.
controls[#controls] = button('Note', 50)
controls[#controls] = button('Mix', 112)
controls[#controls] = button('Device', 110)
controls[#controls] = button('Clip', 113)
controls[#controls] = button('Session Screen', 34)
controls[#controls] = button('Session Pad', 51)
controls[#controls] = button('Scale', 58)
controls[#controls] = button('Layout', 31)
controls[#controls] = button('Shift', 49)
controls[#controls] = button('Select', 48)
controls[#controls] = button('Accent', 57)
controls[#controls] = button('Repeat', 56)
controls[#controls] = button('Mute', 60)
controls[#controls] = button('Solo', 61)
controls[#controls] = button('Lock', 83)

-- Top-left row.
controls[#controls] = button('Sets', 80)
controls[#controls] = button('Setup', 30)
controls[#controls] = button('Learn', 81)
controls[#controls] = button('User', 59)

-- Display area.
controls[#controls] = button('Add', 32)
controls[#controls] = button('Swap', 33)

-- Navigation.
controls[#controls] = button('Up', 46)
controls[#controls] = button('Down', 47)
controls[#controls] = button('Left', 44)
controls[#controls] = button('Right', 45)
controls[#controls] = button('D-Pad Center', 91)
controls[#controls] = button('Octave Up', 55)
controls[#controls] = button('Octave Down', 54)
controls[#controls] = button('Page Left', 62)
controls[#controls] = button('Page Right', 63)

-- Encoder presses.
controls[#controls] = button('Volume Press', 111)
controls[#controls] = button('Swing/Tempo Press', 15)
controls[#controls] = button('Jog Click', 94)
controls[#controls] = button('Jog Push Left', 93)
controls[#controls] = button('Jog Push Right', 95)

-- Upper display buttons (CC 102-109).
for i = 1, 8 do
	controls[#controls] = button('Upper ' .. i, 101 + i)
end

-- Lower display buttons (CC 20-27).
for i = 1, 8 do
	controls[#controls] = button('Lower ' .. i, 19 + i)
end

-- Main Track button.
controls[#controls] = button('Main Track', 28)

-- Scene buttons (CC 43 down to 36).
for i = 1, 8 do
	controls[#controls] = button('Scene ' .. i, 44 - i)
end

-- Track encoders 1-8 (CC 71-78).
for i = 1, 8 do
	controls[#controls] = knob('Encoder ' .. i, 70 + i)
end

-- Volume encoder (CC 79).
controls[#controls] = knob('Volume', 79)

-- Swing/Tempo encoder (CC 14).
controls[#controls] = knob('Swing/Tempo', 14)

-- Jog wheel (CC 70).
controls[#controls] = knob('Jog Wheel', 70)

-- Encoder touch notes (0-7 = track, 8 = volume, 10 = swing, 11 = jog).
for i = 1, 8 do
	controls[#controls] = touch('Encoder Touch ' .. i, i - 1)
end
controls[#controls] = touch('Volume Touch', 8)
controls[#controls] = touch('Swing/Tempo Touch', 10)
controls[#controls] = touch('Jog Touch', 11)
controls[#controls] = touch('Touch Strip Touch', 12)
controls[#controls] = touch('D-Pad Center Touch', 13)

-- 64 pads (notes 36-99, 8x8 grid).
for row = 0, 7 do
	for col = 0, 7 do
		local note = 92 - row * 8 + col
		controls[#controls] = pad('Pad ' .. (row + 1) .. '/' .. (col + 1), note)
	end
end

--------------------------------------------------------------------------------
-- [8] MDS API
--------------------------------------------------------------------------------

function controller_info()
	return {
		manufacturer = 'Ableton',
		model = 'unPush 3',
		version = 1,

		items = controls,

		assignments = {
			{zone='Transport'},
			{control='Play', keyCmd=535},
			{control='Record', keyCmd=7},

			{zone='Global'},
			{mode='Global'},
			{control='Up', keyCmd=1272},
			{control='Down', keyCmd=1273},
			{control='Undo', keyCmd=14},
			{control='Save', keyCmd=9},
			{control='Metronome', keyCmd=56},
			{control='Quantize', keyCmd=68},
			{control='Capture', keyCmd=299},
			{control='Delete', keyCmd=31},
			{control='Duplicate', keyCmd=59},

			{zone='Rotaries'},
			{control='Note', setMode='Device'},
			{control='Mix', setMode='Mix'},

			{mode='Device'},
			{control='Encoder 1', CSTrack=true, trackParam=CS_SMARTCONTROL1, paramName='@tp'},
			{control='Encoder 2', CSTrack=true, trackParam=CS_SMARTCONTROL1, paramOffset=1, paramName='@tp'},
			{control='Encoder 3', CSTrack=true, trackParam=CS_SMARTCONTROL1, paramOffset=2, paramName='@tp'},
			{control='Encoder 4', CSTrack=true, trackParam=CS_SMARTCONTROL1, paramOffset=3, paramName='@tp'},
			{control='Encoder 5', CSTrack=true, trackParam=CS_SMARTCONTROL1, paramOffset=4, paramName='@tp'},
			{control='Encoder 6', CSTrack=true, trackParam=CS_SMARTCONTROL1, paramOffset=5, paramName='@tp'},
			{control='Encoder 7', CSTrack=true, trackParam=CS_SMARTCONTROL1, paramOffset=6, paramName='@tp'},
			{control='Encoder 8', CSTrack=true, trackParam=CS_SMARTCONTROL1, paramOffset=7, paramName='@tp'},

			{mode='Mix'},
			{control='Encoder 1', faderBankTrack=0, trackParam=AUVOLUME, paramName='@tp'},
			{control='Encoder 2', faderBankTrack=1, trackParam=AUVOLUME, paramName='@tp'},
			{control='Encoder 3', faderBankTrack=2, trackParam=AUVOLUME, paramName='@tp'},
			{control='Encoder 4', faderBankTrack=3, trackParam=AUVOLUME, paramName='@tp'},
			{control='Encoder 5', faderBankTrack=4, trackParam=AUVOLUME, paramName='@tp'},
			{control='Encoder 6', faderBankTrack=5, trackParam=AUVOLUME, paramName='@tp'},
			{control='Encoder 7', faderBankTrack=6, trackParam=AUVOLUME, paramName='@tp'},
			{control='Encoder 8', faderBankTrack=7, trackParam=AUVOLUME, paramName='@tp'},
		},

		alertAssignments = {},
	}
end

function controller_initialize(appName, newlyDetected)
	print("unPush3 INIT", appName, newlyDetected)
	return {
		midi={0xF0, 0x00, 0x21, 0x1D, 0x01, 0x01, 0x06, 0x7F, 0xF7},
		outport=PORT,
	}
end

function controller_finalize()
	return {}
end

function controller_midi_in(midi, port)
	if port ~= PORT then return nil end

	local ch = midi[0] % 16
	if ch > 0 then return {} end

	return nil
end
