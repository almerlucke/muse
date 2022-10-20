package main

import (
	"bufio"
	"log"
	"os"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/controls/clog"
	"github.com/almerlucke/muse/messengers/triggers/timer"
)

func main() {
	env := muse.NewEnvironment(0, 44100.0, 512)

	timer := env.AddControl(timer.NewControlTimer(250.0, env.Config, "timer"))
	logger := clog.NewLog("test")

	timer.ConnectToControl(0, logger, 0)

	stream, err := env.InitializeAudio()
	if err != nil {
		log.Fatalf("error opening audio stream, %v", err)
	}

	defer env.TerminateAudio()

	stream.Start()

	reader := bufio.NewReader(os.Stdin)

	reader.ReadString('\n')
}
