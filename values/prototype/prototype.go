package prototype

import (
	"reflect"

	"github.com/almerlucke/muse/values"
)

// Prototype is a prototype of a map. When Map() is called, a deep copy of the prototype is made with all Valuer values
// in the prototype replaced with the Value() from that Valuer. In the deep copy all placeholder values are replaced
// with the matching replacement values. If the value from a Valuer is a slice/array, the prototype is split into multiple
// maps based on the longest slice/array value found
type Prototype map[string]any

type Placeholder struct {
	Name  string
	Value any
}

type Replacement struct {
	Name  string
	Value any
}

func NewPlaceholder(name string) *Placeholder {
	return &Placeholder{Name: name, Value: nil}
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

func (p Prototype) intermediate() Prototype {
	pc := Prototype{}

	for k, v := range p {
		switch vt := v.(type) {
		case Prototype:
			pc[k] = vt.intermediate()
		case values.Valuer[any]:
			pc[k] = newIntermediateValue(vt.Value())
		case *Placeholder:
			pc[k] = newIntermediateValue(vt.Value)
		case Placeholder:
			pc[k] = newIntermediateValue(vt.Value)
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

func (p Prototype) Replace(replacements []*Replacement) {
	for _, v := range p {
		switch vt := v.(type) {
		case Prototype:
			vt.Replace(replacements)
		case *Placeholder:
			for _, replacement := range replacements {
				if vt.Name == replacement.Name {
					vt.Value = replacement.Value
					break
				}
			}
		case Placeholder:
			for _, replacement := range replacements {
				if vt.Name == replacement.Name {
					vt.Value = replacement.Value
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
func (p Prototype) SetState(m map[string]any) {
	for k, v := range m {
		switch vt := p[k].(type) {
		case Prototype:
			vt.SetState(v.(map[string]any))
		case values.Valuer[any]:
			vt.SetState(v.(map[string]any))
		case *Placeholder:
			vt.Value = v
		case Placeholder:
			vt.Value = v
		default:
			p[k] = v
		}
	}
}

func (p Prototype) GetState() map[string]any {
	m := map[string]any{}

	for k, v := range p {
		switch vt := v.(type) {
		case Prototype:
			m[k] = vt.GetState()
		case values.Valuer[any]:
			m[k] = vt.GetState()
		case *Placeholder:
			m[k] = vt.Value
		case Placeholder:
			m[k] = vt.Value
		default:
			m[k] = v
		}
	}

	return m
}

func (p Prototype) Value() []map[string]any {
	intermediate := p.intermediate()
	length := intermediate.length()
	maps := make([]map[string]any, length)

	for i := 0; i < length; i++ {
		maps[i] = intermediate.mapAtIndex(i)
	}

	return maps
}

func (p Prototype) Continuous() bool {
	return true
}

func (p Prototype) Reset() {
	for _, v := range p {
		switch vt := v.(type) {
		case Prototype:
			vt.Reset()
		case values.Valuer[any]:
			vt.Reset()
		default:
			break
		}
	}
}

func (p Prototype) Finished() bool {
	return false
}
