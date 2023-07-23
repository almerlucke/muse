package log

import (
	"log"

	"github.com/almerlucke/muse"
)

type Log struct {
	*muse.BaseControl
}

func NewLog() *Log {
	l := &Log{BaseControl: muse.NewBaseControl()}
	l.SetSelf(l)
	return l
}

func (l *Log) ReceiveControlValue(v any, i int) {
	log.Printf("Log %s: value %v at index %d", l.Identifier(), v, i)
}
