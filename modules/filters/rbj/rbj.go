package rbj

import (
	"github.com/almerlucke/muse"
	rbjc "github.com/almerlucke/muse/components/filters/rbj"
)

type RBJFilter struct {
	*muse.BaseModule
	filter *rbjc.RBJFilter
}

func NewRBJFilter(filterType rbjc.RBJFilterType, freq float64, q float64, config *muse.Configuration, identifier string) *RBJFilter {
	return &RBJFilter{
		BaseModule: muse.NewBaseModule(3, 1, config, identifier),
		filter:     rbjc.NewRBJFilter(filterType, freq, q, 0, false, config.SampleRate),
	}
}

func (r *RBJFilter) Synthesize() bool {
	if !r.BaseModule.Synthesize() {
		return false
	}

	rawIn := r.Inputs[0].Buffer
	filterOut := r.Outputs[0].Buffer
	recalculate := false

	if r.Inputs[1].IsConnected() || r.Inputs[2].IsConnected() {
		recalculate = true
	}

	for i := 0; i < r.Config.BufferSize; i++ {
		if r.Inputs[1].IsConnected() {
			r.filter.Frequency = r.Inputs[1].Buffer[i]
		}

		if r.Inputs[2].IsConnected() {
			r.filter.Q = r.Inputs[2].Buffer[i]
		}

		if recalculate {
			r.filter.Update(r.Config.SampleRate)
		}

		filterOut[i] = r.filter.Filter(rawIn[i])
	}

	return true
}
