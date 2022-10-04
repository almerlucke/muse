package noise

import (
	"github.com/almerlucke/muse/utils/rand"

	"github.com/almerlucke/muse"
)

type Noise struct {
	*muse.BaseModule
	r *rand.Rand
}

func NewNoise(seed uint64, config *muse.Configuration) *Noise {
	return &Noise{
		BaseModule: muse.NewBaseModule(0, 1, config, ""),
		r:          rand.NewRandWithSeed(seed),
	}
}

func (n *Noise) Synthesize() bool {
	if !n.BaseModule.Synthesize() {
		return false
	}

	out := n.Outputs[0].Buffer

	for i := 0; i < n.Config.BufferSize; i++ {
		out[i] = n.r.RandFloat()*2.0 - 1.0
	}

	return true
}
