package messengers

import (
	"encoding/json"
	"sort"

	"github.com/almerlucke/muse"
)

type Event struct {
	Message *muse.Message `json:"message"`
	Time    float64       `json:"time"`
}

type Scheduler struct {
	events []*Event
	index  int
}

func NewSchedulerWithJSONData(data []byte) (*Scheduler, error) {
	var events []*Event

	err := json.Unmarshal(data, &events)
	if err != nil {
		return nil, err
	}

	return NewSchedulerWithEvents(events), nil
}

func NewSchedulerWithEvents(events []*Event) *Scheduler {
	return &Scheduler{
		events: events,
	}
}

func (s *Scheduler) Merge(events []*Event) {
	s.events = append(s.events, events...)
	sort.Slice(s.events, func(i, j int) bool {
		return s.events[i].Time < s.events[j].Time
	})
}

func (s *Scheduler) Messages(timestamp int64, config *muse.Configuration) []*muse.Message {
	time := float64(timestamp) / config.SampleRate
	numEvents := len(s.events)
	messages := []*muse.Message{}

	for s.index < numEvents {
		event := s.events[s.index]
		if event.Time > time {
			break
		}

		messages = append(messages, event.Message)
		s.index++
	}

	return messages
}
