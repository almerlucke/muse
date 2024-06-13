package main

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/and"
	"github.com/almerlucke/genny/bucket"
	"github.com/almerlucke/genny/float"
	"github.com/almerlucke/genny/float/envelopes/adsr"
	"github.com/almerlucke/genny/float/interp"
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/float/shape/shapers/quantize"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/template"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/controls/gen"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/timer"
	"github.com/almerlucke/muse/modules"
	"github.com/almerlucke/muse/modules/effects/chorus"
	"github.com/almerlucke/muse/modules/effects/freeverb"
	"github.com/almerlucke/muse/modules/effects/pingpong"
	"github.com/almerlucke/muse/modules/filters/rbj"
	"github.com/almerlucke/muse/modules/wtscan"
	"github.com/almerlucke/muse/synths/basic"
	"github.com/almerlucke/muse/utils/notes"
	"log"
)

type Source struct {
	*muse.BasePatch
	osc1   *wtscan.Scanner
	osc2   *wtscan.Scanner
	ip     *interp.Interpolator
	gen    genny.Generator[float64]
	detune float64
}

func (s *Source) Activate(values map[string]any) {
	if freq, ok := values["frequency"].(float64); ok {
		s.osc1.SetFrequency(freq)
		s.osc2.SetFrequency(freq * s.detune)
	}

	if detune, ok := values["detune"].(float64); ok {
		s.detune = detune
	}

	if scanSpeed, ok := values["scanSpeed"].(float64); ok {
		s.ip.SetDelta(scanSpeed)
	}
}

func (s *Source) SetValues(values map[string]any) {

}

func (s *Source) SetValue(key string, value any) {

}

func (s *Source) New(initValues any) basic.Source {
	values := initValues.(map[string]any)
	ip := interp.New(float.ToFrame(function.NewRandom(0.0, 1.0)), interp.Linear, 0.0233)
	freq := values["frequency"].(float64)

	newSource := &Source{
		BasePatch: muse.NewPatch(0, 1),
		osc1:      wtscan.New(values["sf"].(*io.WaveTableSoundFile), freq, 0.0, 0.0, 1.0),
		osc2:      wtscan.New(values["sf"].(*io.WaveTableSoundFile), freq*1.01, 0.0, 0.0, 1.0),
		ip:        ip,
		gen:       float.FromFrame(ip, 0),
		detune:    1.075,
	}

	scanPos := gen.New[float64](newSource.gen, true).CtrlAddTo(newSource)

	newSource.SetSelf(newSource)

	newSource.AddModule(newSource.osc1)
	newSource.AddModule(newSource.osc2)

	newSource.In(modules.Amp(newSource.osc1, 0, 0.75).AddTo(newSource))
	newSource.In(modules.Amp(newSource.osc2, 0, 0.75).AddTo(newSource))

	newSource.osc1.CtrlIn(scanPos, 0, 2)
	newSource.osc2.CtrlIn(scanPos, 0, 2)

	return newSource
}

func main() {
	root := muse.New(2)

	sf, err := io.NewWaveTableSoundFile("resources/wavetables2048/FMAdditive1.wav", 2048)
	if err != nil {
		log.Fatalf("err loading sound file: %v", err)
	}

	cfg := map[string]any{
		"sf":        sf,
		"frequency": 100.0,
	}

	ampEnvSetting := adsr.NewSetting(1.0, 43.0, 0.2, 23.0, 40.0, 4000.0)
	filterEnvSetting := adsr.NewSetting(1.0, 46.0, 0.2, 26.0, 40.0, 4000.0)

	synth := basic.New(20, &Source{}, cfg, ampEnvSetting, filterEnvSetting, &rbj.Factory{}, rbj.DefaultConfig()).Named("synth").AddTo(root)

	root.AddMessenger(banger.NewTemplateBang([]string{"synth"}, template.Template{
		"command":   "trigger",
		"duration":  82.0,
		"amplitude": function.NewRandom(0.1, 0.8),
		"message": template.Template{
			"frequency": and.NewLoop[float64](
				bucket.New(bucket.Indexed, notes.PhrygianDominant.Freq(notes.C3)...),
				bucket.New(bucket.Indexed, notes.PhrygianDominant.Freq(notes.C5)...),
				bucket.New(bucket.Indexed, notes.PhrygianDominant.Freq(notes.C2)...),
				bucket.New(bucket.Indexed, notes.PhrygianDominant.Freq(notes.C4)...),
			),
			"attackDuration":  function.NewRandom(4.0, 50.0),
			"releaseDuration": function.NewRandom(3000.0, 7000.0),
			"pan":             function.NewRandom(0.0, 1.0),
			"filterFcMin":     function.NewRandom(30.0, 400.0),
			"filterFcMax":     function.NewRandom(2330.0, 12000.0),
			"filterResonance": function.NewRandom(0.1, 1.3),
			"detune":          function.NewRandom(1.001, 1.015),
			"scanSpeed":       function.NewRandom(0.0075, 0.075),
		},
	}).MsgrNamed("synthDriver"))

	timer.New(0.0, []string{"synthDriver"}, float.FromFrame(shape.New(float.ToFrame(function.NewRandom(250.0, 4000.0)), quantize.New(250.0)), 0)).MsgrAddTo(root)

	ch := chorus.New(0.3, 0.6, 0.7, 0.2, 1.0, 0.2, nil).AddTo(root).In(synth, synth, 1)

	pp := pingpong.New(2500.0, 1750.0, 0.4, 0.1).AddTo(root).In(ch, ch, 1)

	rvb := freeverb.New().AddTo(root).In(pp, pp, 1).Exec(func(obj any) {
		rvb := obj.(*freeverb.FreeVerb)
		rvb.SetWidth(1.0)
		rvb.SetDamp(0.4)
		rvb.SetRoomSize(0.9)
		rvb.SetDry(0.4)
		rvb.SetWet(0.1)
	})

	root.In(rvb, rvb, 1)

	_ = root.RenderAudio()
}
