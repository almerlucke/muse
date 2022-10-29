package clog

import (
	"log"

	"github.com/almerlucke/muse"
)

type Log struct {
	*muse.BaseControl
}

func NewLog(id string) *Log {
	l := &Log{BaseControl: muse.NewBaseControl(id)}
	// Always set self so we can reference embedding struct from base control
	l.SetSelf(l)
	return l
}

func (l *Log) ReceiveControlValue(v any, i int) {
	log.Printf("Log %s: value %v at index %d", l.Identifier(), v, i)
}
