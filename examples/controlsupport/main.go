package main

import (
	"log"

	"github.com/almerlucke/muse"
)

type Logger struct {
	*muse.BaseControlSupport
}

func NewLogger() *Logger {
	return &Logger{BaseControlSupport: muse.NewBaseControlSupport("")}
}

func (l *Logger) ReceiveControlValue(value any, index int) {
	log.Printf("value %v received on index %d", value, index)
}

func main() {
	config := &muse.Configuration{SampleRate: 44100.0, BufferSize: 512}
	start := muse.NewControlThru()
	patch := muse.NewPatch(0, 0, config, "")
	end := NewLogger()
	start.ConnectControlOutput(0, patch, 0)
	patch.InternalControlInput().ConnectControlOutput(0, patch.InternalControlOutput(), 0)
	patch.ConnectControlOutput(0, end, 0)
	patch.ConnectControlOutput(0, end, 1)

	start.SendControlValue("test", 0)
}
