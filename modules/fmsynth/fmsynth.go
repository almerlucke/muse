package fmsynth

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/ops"
	"github.com/almerlucke/muse/utils/notes"
)

type voice struct {
	ops        *ops.Ops
	identifier int
}

type OperatorSetting struct {
	LevelEnvLevels [4]float64
	LevelEnvRates  [4]float64
	FrequencyMode  ops.FrequencyMode
	Frequency      float64
	FrequencyRatio float64
	Level          float64
}

type FMSynth struct {
	*muse.BaseModule
	voices           []*voice
	OperatorSettings [6]OperatorSetting
	PitchEnvLevels   [4]float64
	PitchEnvRates    [4]float64
	ReleaseMode      ops.EnvelopeReleaseMode
}

func NewFMSynth(numVoices int, table []float64, config *muse.Configuration) *FMSynth {
	voices := make([]*voice, numVoices)

	for i := 0; i < numVoices; i++ {
		voices[i] = &voice{
			identifier: 0,
			ops:        ops.NewOps(table, 400.0, config.SampleRate),
		}
	}

	fmSynth := &FMSynth{
		BaseModule: muse.NewBaseModule(0, 1, config, ""),
		voices:     voices,
	}

	fmSynth.SetSelf(fmSynth)

	fmSynth.OperatorSettings[0].FrequencyRatio = 2.01
	fmSynth.OperatorSettings[0].Level = 0.2
	fmSynth.OperatorSettings[1].FrequencyRatio = 1.02
	fmSynth.OperatorSettings[1].Level = 0.5
	fmSynth.OperatorSettings[2].FrequencyRatio = 4.01
	fmSynth.OperatorSettings[2].Level = 0.2
	fmSynth.OperatorSettings[3].FrequencyRatio = 3.02
	fmSynth.OperatorSettings[3].Level = 0.1
	fmSynth.OperatorSettings[4].FrequencyRatio = 1.53
	fmSynth.OperatorSettings[4].Level = 0.2
	fmSynth.OperatorSettings[5].FrequencyRatio = 3.01
	fmSynth.OperatorSettings[5].Level = 0.5

	for i := 0; i < 6; i++ {
		fmSynth.OperatorSettings[i].FrequencyMode = ops.TrackFrequency
		fmSynth.OperatorSettings[i].LevelEnvLevels = ops.DefaultEnvLevels
		fmSynth.OperatorSettings[i].LevelEnvRates = ops.DefaultEnvRates
	}

	fmSynth.PitchEnvLevels = ops.DefaultPitchEnvLevels
	fmSynth.PitchEnvRates = ops.DefaultPitchEnvRates
	fmSynth.ReleaseMode = ops.EnvelopeNoteOffRelease

	return fmSynth
}

func (fm *FMSynth) ApplySettingsChange() {
	for _, voice := range fm.voices {
		voice.ops.PitchEnvelope().Levels = fm.PitchEnvLevels
		voice.ops.PitchEnvelope().Rates = fm.PitchEnvRates
		voice.ops.SetReleaseMode(fm.ReleaseMode)
		for i := 0; i < 6; i++ {
			setting := fm.OperatorSettings[i]
			op := voice.ops.Operator(i)
			op.SetFrequencyMode(setting.FrequencyMode, fm.Config.SampleRate)
			op.SetFrequency(setting.Frequency, fm.Config.SampleRate)
			op.SetFrequencyRatio(setting.FrequencyRatio, fm.Config.SampleRate)
			op.SetLevel(setting.Level)
			op.LevelEnvelope().Levels = setting.LevelEnvLevels
			op.LevelEnvelope().Rates = setting.LevelEnvRates
		}
	}
}

func (fm *FMSynth) ApplyAlgo(algo *ops.Algo) {
	for _, voice := range fm.voices {
		voice.ops.Apply(algo)
	}
}

func (fm *FMSynth) getVoice() *voice {
	for _, v := range fm.voices {
		if v.ops.Idle() {
			return v
		}
	}

	return nil
}

func (fm *FMSynth) noteOff(identifier int) {
	for _, v := range fm.voices {
		if v.identifier == identifier {
			v.ops.NoteOff()
		}
	}
}

func (fm *FMSynth) ReceiveControlValue(value any, index int) {
	// switch index {
	// case 0: // NoteOn
	// 	o.SetFrequency(value.(float64))
	// }
}

func (fm *FMSynth) ReceiveMessage(msg any) []*muse.Message {
	if params, ok := msg.(map[string]any); ok {
		if noteOnIdentifier, ok := params["noteOn"]; ok {
			duration := 0.0
			if durationRaw, ok := params["duration"]; ok {
				duration = durationRaw.(float64) / 1000.0
			}
			level := params["level"].(float64)
			pitch := noteOnIdentifier.(int)
			voice := fm.getVoice()
			if voice != nil {
				voice.identifier = pitch
				voice.ops.NoteOn(notes.Mtof(pitch), level, duration)
			}
		} else if noteOffIdentifier, ok := params["noteOff"]; ok {
			pitch := noteOffIdentifier.(int)
			fm.noteOff(pitch)
		}
	}

	return nil
}

func (fm *FMSynth) Synthesize() bool {
	if !fm.BaseModule.Synthesize() {
		return false
	}

	out := fm.OutputAtIndex(0).Buffer

	for i := 0; i < fm.Config.BufferSize; i++ {
		accum := 0.0
		for _, voice := range fm.voices {
			if !voice.ops.Idle() {
				voice.ops.PrepareRun()
				accum += voice.ops.Run()
			}
		}
		out[i] = accum
	}

	return true
}
