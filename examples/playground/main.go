package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/controls/clog"
	"github.com/almerlucke/muse/controls/seq"
	"github.com/almerlucke/muse/messengers/triggers/timer"
	"github.com/almerlucke/muse/value"
)

func main() {
	env := muse.NewEnvironment(0, 44100.0, 512)

	logger := clog.NewLog("test")
	timer := env.AddControl(timer.NewControlTimer(250.0, env.Config, "timer"))
	s := env.AddControl(seq.NewSeq(value.NewSequenceNC([]float64{1.0, 2.0, 3.0, 4.0}), ""))

	timer.ConnectToControl(0, s, 0)
	s.ConnectToControl(1, s, 2)
	s.ConnectToControl(1, s, 1)
	s.ConnectToControl(0, logger, 0)

	env.QuickPlayAudio()
}
