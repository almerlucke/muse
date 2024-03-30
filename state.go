package muse

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"
	"reflect"
)

// PersistToFile save object state to file
func PersistToFile(name string, obj any) error {
	var (
		b       bytes.Buffer
		zw      = gzip.NewWriter(&b)
		objType = reflect.TypeOf(obj).Elem()
		enc     = gob.NewEncoder(zw)
		err     error
	)

	err = enc.Encode(obj)
	if err != nil {
		_ = zw.Close()
		return fmt.Errorf("kratos failed to store state of %s to file %s: %w", objType, name, err)
	}

	err = zw.Close()
	if err != nil {
		return fmt.Errorf("kratos failed to store state of %s to file %s: %w", objType, name, err)
	}

	err = os.WriteFile(name, b.Bytes(), 0666)
	if err != nil {
		return fmt.Errorf("kratos failed to store state of %s to file %s: %w", objType, name, err)
	}

	return nil
}

// RestoreFromFile restore object state from file
func RestoreFromFile(name string, obj any) error {
	objType := reflect.TypeOf(obj).Elem()

	data, err := os.ReadFile(name)
	if err != nil {
		return fmt.Errorf("kratos failed to restore state of %s from file %s: %w", objType, name, err)
	}

	zr, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("kratos failed to restore state of %s from file %s: %w", objType, name, err)
	}

	defer func() {
		_ = zr.Close()
	}()

	err = gob.NewDecoder(zr).Decode(obj)
	if err != nil {
		return fmt.Errorf("kratos failed to restore state of %s from file %s: %w", objType.Name(), name, err)
	}

	return nil
}
