package push

import "github.com/loov/logic-push3/push3"

// Encoder touch note to EncoderID mapping.
// Touch notes: 0-7 = Track 1-8, 8 = Master, 9 = Swing, 10 = Tempo.
func encoderFromTouchNote(note uint8) (push3.EncoderID, bool) {
	switch {
	case note <= 7:
		return push3.EncoderID(note + 1), true // EncoderTrack1=1, note 0 → Track1
	case note == 8:
		return push3.EncoderMaster, true
	case note == 9:
		return push3.EncoderSwing, true
	case note == 10:
		return push3.EncoderTempo, true
	default:
		return 0, false
	}
}
