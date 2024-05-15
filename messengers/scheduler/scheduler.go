package scheduler

import (
	"sort"

	"github.com/almerlucke/muse"
)

type ControlMessage struct {
	OutIndex int `json:"outIndex"`
	Content  any `json:"content"`
}

type Event struct {
	Messages        []*muse.Message   `json:"messages"`
	ControlMessages []*ControlMessage `json:"controlMessages"`
	When            float64           `json:"when"`
}

type Scheduler struct {
	*muse.BaseMessenger
	events     []*Event
	eventIndex int
	eventMap   map[float64]*Event
}

func New() *Scheduler {
	s := &Scheduler{
		BaseMessenger: muse.NewBaseMessenger(),
		eventMap:      map[float64]*Event{},
	}

	s.SetSelf(s)

	return s
}

func NewWithEvents(events []*Event) *Scheduler {
	s := &Scheduler{
		BaseMessenger: muse.NewBaseMessenger(),
		events:        events,
		eventMap:      map[float64]*Event{},
	}

	for _, event := range events {
		s.eventMap[event.When] = event
	}

	s.SetSelf(s)

	return s
}

func (s *Scheduler) reorderEvents() {
	sort.Slice(s.events, func(i, j int) bool {
		return s.events[i].When < s.events[j].When
	})
}

func (s *Scheduler) ScheduleControlMessage(when float64, content any, outIndex int) {
	msg := &ControlMessage{
		OutIndex: outIndex,
		Content:  content,
	}

	if event, ok := s.eventMap[when]; ok {
		event.ControlMessages = append(event.ControlMessages, msg)
	} else {
		newEvent := &Event{
			ControlMessages: []*ControlMessage{msg},
			When:            when,
		}
		s.events = append(s.events, newEvent)
		s.reorderEvents()
		s.eventMap[when] = newEvent
	}
}

func (s *Scheduler) ScheduleMessage(when float64, msg *muse.Message) {
	if event, ok := s.eventMap[when]; ok {
		event.Messages = append(event.Messages, msg)
	} else {
		newEvent := &Event{
			Messages: []*muse.Message{msg},
			When:     when,
		}
		s.events = append(s.events, newEvent)
		s.reorderEvents()
		s.eventMap[when] = newEvent
	}
}

func (s *Scheduler) ScheduleEvents(events []*Event) {
	for _, event := range events {
		if existingEvent, ok := s.eventMap[event.When]; ok {
			existingEvent.ControlMessages = append(existingEvent.ControlMessages, event.ControlMessages...)
			existingEvent.Messages = append(existingEvent.Messages, event.Messages...)
		} else {
			s.events = append(s.events, event)
			s.eventMap[event.When] = event
		}
	}

	s.reorderEvents()
}

func (s *Scheduler) Tick(timestamp int64, config *muse.Configuration) {
	_ = s.Messages(timestamp, config)
}

func (s *Scheduler) Messages(timestamp int64, config *muse.Configuration) []*muse.Message {
	var (
		when      = config.SampsToMilli(timestamp)
		numEvents = len(s.events)
		messages  []*muse.Message
	)

	for s.eventIndex < numEvents {
		event := s.events[s.eventIndex]
		if event.When > when {
			break
		}

		for _, controlMessage := range event.ControlMessages {
			s.SendControlValue(controlMessage.Content, controlMessage.OutIndex)
		}

		messages = append(messages, event.Messages...)
		s.eventIndex++
	}

	return messages
}
