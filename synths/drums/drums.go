package drums

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/modules/player"
	"github.com/almerlucke/muse/modules/polyphony"
)

func NewDrums(soundBank io.SoundBank, numVoices int, config *muse.Configuration) *polyphony.Polyphony {
	var initSound io.SoundFiler

	for _, v := range soundBank {
		if initSound == nil {
			initSound = v
			break
		}
	}

	voices := make([]polyphony.Voice, numVoices)
	for i := 0; i < numVoices; i++ {
		player := player.New(initSound, 1.0, 1.0, true, config)
		player.SetSoundBank(soundBank)
		voices[i] = player
	}

	return polyphony.New(initSound.NumChannels(), voices, config)
}
