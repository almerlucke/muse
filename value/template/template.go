package template

import (
	"reflect"

	"github.com/almerlucke/muse/value"
)

// Template is a template of a map. When Value() is called, a deep copy of the template is made with all Valuer values
// in the template replaced with the Value() from that Valuer. In the deep copy all parameter entries are replaced
// with the matching replacement values. If the value from a Valuer is a slice/array, the output of the template is split into multiple
// maps based on the longest slice/array value found
type Template map[string]any

type Parameter struct {
	Name  string
	Value any
}

func NewParameter(name string, val any) *Parameter {
	return &Parameter{Name: name, Value: val}
}

// intermediateValue is created in prototype to check if prototype needs to split if value is an array or slice
type intermediateValue struct {
	value        any
	reflectValue reflect.Value
	length       int
	indexable    bool
}

func newIntermediateValue(v any) *intermediateValue {
	reflectValue := reflect.ValueOf(v)
	kind := reflectValue.Kind()
	length := 1
	indexable := false

	if kind == reflect.Slice || kind == reflect.Array {
		length = reflectValue.Len()
		indexable = true
	}

	return &intermediateValue{value: v, reflectValue: reflectValue, length: length, indexable: indexable}
}

func (iv *intermediateValue) index(i int) any {
	// return normal value if not indexable
	if !iv.indexable {
		return iv.value
	}

	// clip index to length
	if i >= iv.length {
		i = iv.length - 1
	}

	// return element at index
	return iv.reflectValue.Index(i).Interface()
}

func (t Template) intermediate() Template {
	tc := Template{}

	for k, v := range t {
		switch vt := v.(type) {
		case Template:
			tc[k] = vt.intermediate()
		case value.Valuer[any]:
			tc[k] = newIntermediateValue(vt.Value())
		case *Parameter:
			tc[k] = newIntermediateValue(vt.Value)
		case Parameter:
			tc[k] = newIntermediateValue(vt.Value)
		default:
			tc[k] = v
		}
	}

	return tc
}

func (t Template) length() int {
	max := 1

	for _, v := range t {
		switch vt := v.(type) {
		case Template:
			pmax := vt.length()
			if pmax > max {
				max = pmax
			}
		case *intermediateValue:
			if vt.length > max {
				max = vt.length
			}
		default:
			break
		}
	}

	return max
}

func (t Template) mapAtIndex(i int) map[string]any {
	m := map[string]any{}

	for k, v := range t {
		switch vt := v.(type) {
		case Template:
			m[k] = vt.mapAtIndex(i)
		case *intermediateValue:
			m[k] = vt.index(i)
		default:
			m[k] = v
		}
	}

	return m
}

func (t Template) SetParameter(name string, val any) {
	t.SetParameters([]*Parameter{NewParameter(name, val)})
}

func (t Template) SetParameters(parameters []*Parameter) {
	for _, v := range t {
		switch vt := v.(type) {
		case Template:
			vt.SetParameters(parameters)
		case *Parameter:
			for _, parameter := range parameters {
				if vt.Name == parameter.Name {
					vt.Value = parameter.Value
					break
				}
			}
		case Parameter:
			for _, parameter := range parameters {
				if vt.Name == parameter.Name {
					vt.Value = parameter.Value
					break
				}
			}
		default:
			break
		}
	}
}

/*
	Valuer interface methods
*/

func (t Template) Value() []map[string]any {
	intermediate := t.intermediate()
	length := intermediate.length()
	maps := make([]map[string]any, length)

	for i := 0; i < length; i++ {
		maps[i] = intermediate.mapAtIndex(i)
	}

	return maps
}

func (t Template) Continuous() bool {
	return true
}

func (t Template) Reset() {
	for _, v := range t {
		switch vt := v.(type) {
		case Template:
			vt.Reset()
		case value.Valuer[any]:
			vt.Reset()
		default:
			break
		}
	}
}

func (t Template) Finished() bool {
	return false
}
