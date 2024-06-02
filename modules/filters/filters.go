package filters

import "github.com/almerlucke/muse"

type Filter interface {
	muse.Module
	SetResonance(float64)
	Resonance() float64
	SetDrive(float64)
	Drive() float64
	SetFrequency(float64)
	Frequency() float64
	SetType(int)
	Type() int
}

type FilterConfig struct {
	Frequency float64
	Resonance float64
	Drive     float64
	Type      int
}

func NewFilterConfig(frequency float64, resonance float64, drive float64, t int) *FilterConfig {
	return &FilterConfig{
		Frequency: frequency,
		Resonance: resonance,
		Drive:     drive,
		Type:      t,
	}
}
