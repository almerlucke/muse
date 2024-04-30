package scheduler

import (
	"cmp"
	"log"
	"reflect"
	"slices"
)

type Slot struct {
	Timestamp int64
	Value     any
}

type Scheduler struct {
	slots     []Slot
	lastIndex int
}

func New(args ...any) *Scheduler {
	var (
		sched = &Scheduler{}
		slt   Slot
	)

	if (len(args) % 2) != 0 {
		log.Panicf("expected even number of arguments, got %d", len(args))
	}

	sched.slots = make([]Slot, len(args)>>1)

	for index, arg := range args {
		if index%2 == 0 {
			switch t := arg.(type) {
			case int:
				slt.Timestamp = int64(t)
			case int64:
				slt.Timestamp = t
			default:
				log.Panicf("expected int or int64 timestamp, got %d", reflect.TypeOf(arg))
			}
		} else {
			slt.Value = arg
			sched.slots[index>>1] = slt
		}
	}

	slices.SortFunc(sched.slots, func(i, j Slot) int {
		return cmp.Compare(i.Timestamp, j.Timestamp)
	})

	return sched
}

func (sched *Scheduler) Add(timestamp int64, value any) {
	sched.slots = append(sched.slots, Slot{Timestamp: timestamp, Value: value})
}

func (sched *Scheduler) Sort() {
	slices.SortFunc(sched.slots, func(i, j Slot) int {
		return cmp.Compare(i.Timestamp, j.Timestamp)
	})
}

func (sched *Scheduler) Schedule(timestamp int64) []any {
	var (
		index = sched.lastIndex
		items []any
	)

	for index < len(sched.slots) && sched.slots[index].Timestamp < timestamp {
		items = append(items, sched.slots[index].Value)
		index++
	}

	sched.lastIndex = index

	return items
}

func (sched *Scheduler) Reset() {
	sched.lastIndex = 0
}
