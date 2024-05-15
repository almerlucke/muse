package duration

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
