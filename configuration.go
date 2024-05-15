package muse

import (
	"container/list"
	"github.com/almerlucke/muse/utils/duration"
)

var configList *list.List

func configurationInit() {
	configList = list.New()
	configList.PushFront(&Configuration{
		SampleRate: DefaultSamplerate,
		BufferSize: DefaultBufferSize,
	})
}

type Configuration struct {
	SampleRate float64
	BufferSize int
}

func (cfg *Configuration) MilliToSamps(milli float64) int64 {
	return duration.MilliToSamps(milli, cfg.SampleRate)
}

func (cfg *Configuration) MilliToSampsf(milli float64) float64 {
	return duration.MilliToSampsf(milli, cfg.SampleRate)
}

func (cfg *Configuration) SecToSamps(sec float64) int64 {
	return duration.SecToSamps(sec, cfg.SampleRate)
}

func (cfg *Configuration) SampsToSec(samps int64) float64 {
	return float64(samps) / cfg.SampleRate
}

func (cfg *Configuration) SampsToMilli(samps int64) float64 {
	return (float64(samps) / cfg.SampleRate) * 1000.0
}

func (cfg *Configuration) ControlRate() float64 {
	return cfg.SampleRate / float64(cfg.BufferSize)
}

func PushConfiguration(config *Configuration) {
	configList.PushFront(config)
}

func PopConfiguration() *Configuration {
	front := configList.Front()
	config := front.Value.(*Configuration)
	configList.Remove(front)
	return config
}

func CurrentConfiguration() *Configuration {
	return configList.Front().Value.(*Configuration)
}

func SampleRate() float64 {
	return CurrentConfiguration().SampleRate
}

func ControlRate() float64 {
	return CurrentConfiguration().ControlRate()
}

func BufferSize() int {
	return CurrentConfiguration().BufferSize
}
