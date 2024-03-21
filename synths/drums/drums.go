package drums

import (
	"github.com/almerlucke/muse/modules/player"
	"github.com/almerlucke/muse/modules/polyphony"
	"github.com/almerlucke/sndfile"
)

func NewDrums(soundBank sndfile.SoundBank, numVoices int) *polyphony.Polyphony {
	var initSound sndfile.SoundFiler

	for _, v := range soundBank {
		if initSound == nil {
			initSound = v
			break
		}
	}

	voices := make([]polyphony.Voice, numVoices)
	for i := 0; i < numVoices; i++ {
		p := player.New(initSound, 1.0, 1.0, true)
		p.SetSoundBank(soundBank)
		voices[i] = p
	}

	return polyphony.New(initSound.NumChannels(), voices)
}
