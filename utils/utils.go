package utils

import (
	"encoding/json"
	"os"
)

type Factory[T any] interface {
	New() T
}

func ReadJSON[T any](file string) (T, error) {
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

func ReadJSONNull[T any](file string) T {
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
