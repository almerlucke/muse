package timing

const (
	Second = 1000.0
	Minute = Second * 60
	Hour   = Minute * 60
)

func MilliToSamps(milli float64, sr float64) int64 {
	return int64(milli * 0.001 * sr)
}

func MilliToSampsf(milli float64, sr float64) float64 {
	return milli * 0.001 * sr
}

func SecToSamps(sec float64, sr float64) int64 {
	return int64(sec * sr)
}

func SecToSampsf(sec float64, sr float64) float64 {
	return sec * sr
}

func BPMToMilli(bpm int) float64 {
	return 60000.0 / float64(bpm)
}
