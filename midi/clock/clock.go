package clock

import (
	"container/list"
	"context"
	"gitlab.com/gomidi/midi/v2"
	"time"
)

type Listener interface {
	ClockTick(tickCnt int64)
}

type Clock struct {
	startTime    time.Time
	nextTickTime float64
	timePerTick  float64
	tickCntr     int64
	ctx          context.Context
	stop         context.CancelFunc
	midiSendFunc func(msg midi.Message) error
	listeners    list.List
}

func New(bpm int, midiSendFunc func(msg midi.Message) error) *Clock {
	timePerTick := 60.0 / (float64(bpm) * 24.0)

	return &Clock{
		midiSendFunc: midiSendFunc,
		timePerTick:  timePerTick,
		nextTickTime: timePerTick,
	}
}

func (c *Clock) send() error {
	if c.midiSendFunc != nil {
		return c.midiSendFunc(midi.Start())
	}

	return nil
}

func (c *Clock) Start() error {
	if c.midiSendFunc != nil {
		err := c.midiSendFunc(midi.Start())
		if err != nil {
			return err
		}
	}

	c.startTime = time.Now()
	c.ctx, c.stop = context.WithCancel(context.Background())

	go c.run()

	return nil
}

func (c *Clock) Stop() error {
	c.stop()

	if c.midiSendFunc != nil {
		return c.midiSendFunc(midi.Stop())
	}

	return nil
}

func (c *Clock) run() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case t := <-time.After(time.Microsecond * 50):
			if t.Sub(c.startTime).Seconds() >= c.nextTickTime {
				err := c.send()
				if err != nil {
					c.stop()
					return
				}
				c.tickCntr++
				c.nextTickTime += c.timePerTick
			}
		}
	}
}
