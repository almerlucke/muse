package main

import (
	"bufio"
	"log"
	"os"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/controls/clog"
	"github.com/almerlucke/muse/messengers/lfo"
)

func main() {
	env := muse.NewEnvironment(0, 44100.0, 512)

	// timer := env.AddControl(timer.NewControlTimer(250.0, env.Config, "timer"))
	lfo := env.AddControl(lfo.NewBasicControlLFO(2.0, 10.0, 30.0, env.Config, "lfo"))
	logger := clog.NewLog("test")

	lfo.ConnectToControl(0, logger, 0)

	stream, err := env.InitializeAudio()
	if err != nil {
		log.Fatalf("error opening audio stream, %v", err)
	}

	defer env.TerminateAudio()

	stream.Start()

	reader := bufio.NewReader(os.Stdin)

	reader.ReadString('\n')
}
