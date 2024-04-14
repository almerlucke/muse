package muse

import "container/list"

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
