package phaseshaper

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/phaseshaping"
)

type ParamMapFunction func(int, float64, *phaseshaping.PhaseDistortion)

type MessageMapFunction func(any, *phaseshaping.PhaseDistortion)

type PhaseShaper struct {
	*muse.BaseModule
	distortion  *phaseshaping.PhaseDistortion
	paramMapper ParamMapFunction
	msgMapper   MessageMapFunction
}

func NewPhaseShaper(freq float64, phase float64, distortion *phaseshaping.PhaseDistortion, numParams int, paramMapper ParamMapFunction, msgMapper MessageMapFunction, config *muse.Configuration, identifier string) *PhaseShaper {
	distortion.Generator.SetPhase(phase)
	distortion.Generator.SetFrequency(freq, config.SampleRate)

	return &PhaseShaper{
		BaseModule:  muse.NewBaseModule(1+numParams, 1, config, identifier),
		distortion:  distortion,
		paramMapper: paramMapper,
		msgMapper:   msgMapper,
	}
}

func (ps *PhaseShaper) ReceiveMessage(msg any) []*muse.Message {
	params, ok := msg.(map[string]any)
	if ok {
		f, ok := params["frequency"]
		if ok {
			ps.distortion.Generator.SetFrequency(f.(float64), ps.Config.SampleRate)
		}
	}

	if ps.msgMapper != nil {
		ps.msgMapper(msg, ps.distortion)
	}

	return nil
}

func (ps *PhaseShaper) Synthesize() bool {
	if !ps.BaseModule.Synthesize() {
		return false
	}

	output := ps.Outputs[0].Buffer

	for i := 0; i < ps.Config.BufferSize; i++ {
		if ps.Inputs[0].IsConnected() {
			ps.distortion.Generator.SetFrequency(ps.Inputs[0].Buffer[i], ps.Config.SampleRate)
		}

		if ps.paramMapper != nil {
			for j := 1; j < len(ps.Inputs); j++ {
				if ps.Inputs[j].IsConnected() {
					ps.paramMapper(j-1, ps.Inputs[j].Buffer[i], ps.distortion)
				}
			}
		}

		output[i] = ps.distortion.Tick()
	}

	return true
}
