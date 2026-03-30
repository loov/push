package push

import "github.com/loov/push3/push3"

// Encoder touch note to EncoderID mapping.
// Touch notes: 0-7 = Track 1-8, 8 = Volume, 10 = Tempo/Swing, 11 = Jog.
func encoderFromTouchNote(note uint8) (push3.EncoderID, bool) {
	switch {
	case note <= 7:
		return push3.EncoderID(note + 1), true // EncoderTrack1=1, note 0 → Track1
	case note == 8:
		return push3.EncoderVolume, true
	case note == 10:
		return push3.EncoderSwingTempo, true
	case note == 11:
		return push3.EncoderJog, true
	default:
		return 0, false
	}
}
