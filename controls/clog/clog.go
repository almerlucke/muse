package clog

import (
	"log"

	"github.com/almerlucke/muse"
)

type Log struct {
	*muse.BaseControl
}

func NewLog(id string) *Log {
	return &Log{BaseControl: muse.NewBaseControl(id)}
}

func (l *Log) ReceiveControlValue(v any, i int) {
	log.Printf("Log %s: value %v at index %d", l.Identifier(), v, i)
}
