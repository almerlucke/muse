package prototype

import (
	"reflect"

	"github.com/almerlucke/muse/values"
)

// Prototype is a prototype of a map. When Map() is called, a deep copy of the prototype is made with all Valuer values
// in the prototype replaced with the Value() from that Valuer. In the deep copy all placeholder values are replaced
// with the matching replacement values
type Prototype map[string]any

type Placeholder struct {
	Name string
}

type Replacement struct {
	Name  string
	Value any
}

func NewPlaceholder(name string) *Placeholder {
	return &Placeholder{Name: name}
}

func NewReplacement(name string, value any) *Replacement {
	return &Replacement{Name: name, Value: value}
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

func (p Prototype) intermediate(replacements []*Replacement) Prototype {
	pc := Prototype{}

	for k, v := range p {
		switch vt := v.(type) {
		case Prototype:
			pc[k] = vt.intermediate(replacements)
		case values.Valuer[any]:
			vtv := vt.Value()
			pc[k] = newIntermediateValue(vtv)
		case *Placeholder:
			for _, replacement := range replacements {
				if vt.Name == replacement.Name {
					pc[k] = newIntermediateValue(replacement.Value)
					break
				}
			}
		case Placeholder:
			for _, replacement := range replacements {
				if vt.Name == replacement.Name {
					pc[k] = newIntermediateValue(replacement.Value)
					break
				}
			}
		default:
			pc[k] = v
		}
	}

	return pc
}

func (p Prototype) length() int {
	max := 0

	for _, v := range p {
		switch vt := v.(type) {
		case Prototype:
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

func (p Prototype) mapAtIndex(i int) map[string]any {
	m := map[string]any{}

	for k, v := range p {
		switch vt := v.(type) {
		case Prototype:
			m[k] = vt.mapAtIndex(i)
		case *intermediateValue:
			m[k] = vt.index(i)
		default:
			m[k] = v
		}
	}

	return m
}

func (p Prototype) Map(replacements []*Replacement) []map[string]any {
	intermediate := p.intermediate(replacements)
	length := intermediate.length()
	maps := make([]map[string]any, length)

	for i := 0; i < length; i++ {
		maps[i] = intermediate.mapAtIndex(i)
	}

	return maps
}
