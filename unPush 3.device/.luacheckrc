std = "lua52"

-- Globals provided by Logic Pro's MDS runtime.
read_globals = {
	"MIDI_LSB",
	"MIDI_MSB",
	"MIDI_NoteOn",
	"MIDI_NoteOff",
	"MIDI_CtrChange",
	"MIDI_Wildcard",
	"FB_OFF",
	"FB_AUTO",
	"FB_HOR_SINGLE",
	"kAssignScaled",
	"kAssignRotate",
	"CS_SMARTCONTROL1",
	"AUVOLUME",
	"AUPAN",
	"AUSEND1",
	"AUMUTE",
	"AUSOLO",
	"CS_RECRDY",
	"settriggertimer",
}

-- MDS API functions that must be global.
globals = {
	"controller_info",
	"controller_initialize",
	"controller_finalize",
	"controller_midi_in",
	"controller_midi_out",
	"controller_timer_trigger",
	"controller_select_patch",
	"controller_select_patch_done",
	"controller_names",
	"CSFeedback",
	"CSFeedbackText",
}
