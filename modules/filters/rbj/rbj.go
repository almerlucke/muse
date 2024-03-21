package rbj

import (
	"github.com/almerlucke/muse"
	rbjc "github.com/almerlucke/muse/components/filters/rbj"
)

type Filter struct {
	*muse.BaseModule
	filter *rbjc.Filter
	fc     float64
	q      float64
}

func New(filterType rbjc.FilterType, fc float64, q float64) *Filter {
	rbj := &Filter{
		BaseModule: muse.NewBaseModule(3, 1),
		filter:     rbjc.NewFilter(filterType, fc, q, 0, false, muse.SampleRate()),
		fc:         fc,
		q:          q,
	}

	rbj.SetSelf(rbj)

	return rbj
}

func (r *Filter) Resonance() float64 {
	return r.q
}

func (r *Filter) SetResonance(q float64) {
	r.filter.Q = q
	r.filter.Update(r.Config.SampleRate)
}

func (r *Filter) Frequency() float64 {
	return r.fc
}

func (r *Filter) SetFrequency(fc float64) {
	r.filter.Frequency = fc
	r.filter.Update(r.Config.SampleRate)
}

func (r *Filter) FilterType() rbjc.FilterType {
	return r.filter.FilterType
}

func (r *Filter) SetFilterType(t rbjc.FilterType) {
	r.filter.FilterType = t
	r.filter.Update(r.Config.SampleRate)
}

func (r *Filter) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Cutoff Frequency
		r.SetFrequency(value.(float64))
	case 1: // Resonance
		r.SetResonance(value.(float64))
	case 2: // Filter Mode
		if fltVal, ok := value.(float64); ok {
			r.SetFilterType(rbjc.FilterType(fltVal))
		} else if intVal, ok := value.(int); ok {
			r.SetFilterType(rbjc.FilterType(intVal))
		}
	}
}

func (r *Filter) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	if fc, ok := content["frequency"]; ok {
		r.SetFrequency(fc.(float64))
	}

	if res, ok := content["resonance"]; ok {
		r.SetResonance(res.(float64))
	}

	if t, ok := content["filterType"]; ok {
		if fltVal, ok := t.(float64); ok {
			r.SetFilterType(rbjc.FilterType(fltVal))
		} else if intVal, ok := t.(int); ok {
			r.SetFilterType(rbjc.FilterType(intVal))
		}
	}

	return nil
}

func (r *Filter) Synthesize() bool {
	if !r.BaseModule.Synthesize() {
		return false
	}

	rawIn := r.Inputs[0].Buffer
	filterOut := r.Outputs[0].Buffer
	recalculate := false

	for i := 0; i < r.Config.BufferSize; i++ {
		if r.Inputs[1].IsConnected() {
			recalculate = true
			r.filter.Frequency = r.Inputs[1].Buffer[i]
		}

		if r.Inputs[2].IsConnected() {
			recalculate = true
			r.filter.Q = r.Inputs[2].Buffer[i]
		}

		if recalculate {
			r.filter.Update(r.Config.SampleRate)
		}

		filterOut[i] = r.filter.Process(rawIn[i])
	}

	return true
}
