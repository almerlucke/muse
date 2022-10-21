package main

import (
	"bufio"
	"log"
	"os"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/controls/clog"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/messengers/triggers/timer"
	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/template"
)

func main() {
	env := muse.NewEnvironment(0, 44100.0, 512)

	logger := clog.NewLog("test")
	lfo1 := env.AddControl(lfo.NewBasicControlLFO(2.0, 10.0, 30.0, env.Config, "lfo1"))
	lfo2 := env.AddControl(lfo.NewBasicControlLFO(1.0, 40.0, 60.0, env.Config, "lfo2"))
	timer := env.AddControl(timer.NewControlTimer(250.0, env.Config, "timer"))

	msgGen := env.AddControl(banger.NewControlTemplateGenerator(template.Template{
		"test": value.NewSequence([]any{1.0, 2.0, 3.0}),
		"lfo1": template.NewParameter("controlInput1", 0.0),
		"lfo2": template.NewParameter("controlInput2", 0.0),
	}, "template"))

	timer.ConnectToControl(0, msgGen, 0)
	lfo1.ConnectToControl(0, msgGen, 1)
	lfo2.ConnectToControl(0, msgGen, 2)
	msgGen.ConnectToControl(0, logger, 0)

	stream, err := env.InitializeAudio()
	if err != nil {
		log.Fatalf("error opening audio stream, %v", err)
	}

	defer env.TerminateAudio()

	stream.Start()

	reader := bufio.NewReader(os.Stdin)

	reader.ReadString('\n')
}
