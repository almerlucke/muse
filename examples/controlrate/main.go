package main

import (
	"bufio"
	"log"
	"os"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/controls/clog"
	"github.com/almerlucke/muse/controls/msg"
	"github.com/almerlucke/muse/controls/rcv"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/timer"
	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/template"
)

func main() {
	env := muse.NewEnvironment(0, 44100.0, 512)

	logger := clog.NewLog("test")
	timer := env.AddControl(timer.NewControlTimer(250.0, env.Config, "timer"))

	msgGen := env.AddControl(banger.NewControlTemplateGenerator(template.Template{
		"test": value.NewSequence([]any{1.0, 2.0, 3.0}),
		"lfo1": 1.0,
		"lfo2": 2.0,
	}, ""))

	msgSend := msg.NewMsg(env, []string{"testRecv"}, "")
	msgRecv := env.AddControl(rcv.NewRcv("testRecv"))

	timer.ConnectToControl(0, msgGen, 0)
	msgGen.ConnectToControl(0, msgSend, 0)
	msgRecv.ConnectToControl(0, logger, 0)

	stream, err := env.InitializeAudio()
	if err != nil {
		log.Fatalf("error opening audio stream, %v", err)
	}

	defer env.TerminateAudio()

	stream.Start()

	reader := bufio.NewReader(os.Stdin)

	reader.ReadString('\n')
}
