package utils

import (
	"encoding/json"
	"os"
)

func ToAnySlice[T any](s []T) []any {
	as := make([]any, len(s))
	for i, e := range s {
		as[i] = e
	}

	return as
}

type Factory[T any] interface {
	New(any) T
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
