package utils

import (
	"encoding/json"
	"math"
	"os"
)

type Factory[T any] interface {
	New() T
}

func Mtof(midiNote int) float64 {
	return math.Pow(2, float64(midiNote-69)/12.0) * 440.0
}

func ReadJSONObject[T any](file string) (T, error) {
	var empty T

	data, err := os.ReadFile(file)
	if err != nil {
		return empty, err
	}

	var object T

	err = json.Unmarshal(data, &object)
	if err != nil {
		return empty, err
	}

	return object, nil
}

func ReadJSONObjectNullable[T any](file string) T {
	var empty T

	data, err := os.ReadFile(file)
	if err != nil {
		return empty
	}

	var object T

	err = json.Unmarshal(data, &object)
	if err != nil {
		return empty
	}

	return object
}
