package scheduler

import (
	"encoding/json"
	"sort"

	"github.com/almerlucke/muse"
)

type Event struct {
	Messages []*muse.Message `json:"messages"`
	Time     float64         `json:"time"`
}

type Scheduler struct {
	*muse.BaseMessenger
	events []*Event
	index  int
}

func NewSchedulerWithJSONData(data []byte, identifier string) (*Scheduler, error) {
	var events []*Event

	err := json.Unmarshal(data, &events)
	if err != nil {
		return nil, err
	}

	return NewSchedulerWithEvents(events, identifier), nil
}

func NewSchedulerWithEvents(events []*Event, identifier string) *Scheduler {
	return &Scheduler{
		BaseMessenger: muse.NewBaseMessenger(identifier),
		events:        events,
	}
}

func (s *Scheduler) Merge(events []*Event) {
	s.events = append(s.events, events...)
	sort.Slice(s.events, func(i, j int) bool {
		return s.events[i].Time < s.events[j].Time
	})
}

func (s *Scheduler) Messages(timestamp int64, config *muse.Configuration) []*muse.Message {
	time := (float64(timestamp) / config.SampleRate) * 1000.0 // time in milliseconds
	numEvents := len(s.events)
	messages := []*muse.Message{}

	for s.index < numEvents {
		event := s.events[s.index]
		if event.Time > time {
			break
		}

		messages = append(messages, event.Messages...)
		s.index++
	}

	return messages
}
