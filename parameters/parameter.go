package parameters

type Valuer interface {
	Value() any
	String() string
	Float() float64
	Int() int64
}

type Const struct {
	value any
}

func NewConst(c any) *Const {
	return &Const{value: c}
}

func (c *Const) Value() any {
	return c.value
}

func (c *Const) String() string {
	return c.value.(string)
}

func (c *Const) Float() float64 {
	return c.value.(float64)
}

func (c *Const) Int() int64 {
	return c.value.(int64)
}

type Sequence struct {
	values []any
	index  int
}

func NewSequence(values []any) *Sequence {
	return &Sequence{values: values}
}

func (s *Sequence) Value() any {
	v := s.values[s.index]
	s.index++
	if s.index >= len(s.values) {
		s.index = 0
	}
	return v
}

func (s *Sequence) String() string {
	return s.Value().(string)
}

func (s *Sequence) Float() float64 {
	return s.Value().(float64)
}

func (s *Sequence) Int() int64 {
	return s.Value().(int64)
}

type Function struct {
	f func() any
}

func NewFunction(f func() any) *Function {
	return &Function{f: f}
}

func (f *Function) Value() any {
	return f.f()
}

func (f *Function) String() string {
	return f.Value().(string)
}

func (f *Function) Float() float64 {
	return f.Value().(float64)
}

func (f *Function) Int() int64 {
	return f.Value().(int64)
}

type Map map[string]any
type Prototype map[string]any

func (p Prototype) Map() Map {
	m := Map{}

	for k, v := range p {
		if sub, ok := v.(Prototype); ok {
			m[k] = sub.Map()
		} else {
			m[k] = v.(Valuer).Value()
		}
	}

	return m
}

func (m Map) M(key string) Map {
	if sub, ok := m[key].(Map); ok {
		return sub
	}

	return nil
}

func (m Map) S(key string) string {
	if value, ok := m[key].(string); ok {
		return value
	}

	return ""
}

func (m Map) F(key string) float64 {
	if value, ok := m[key].(float64); ok {
		return value
	}

	return 0
}

func (m Map) I(key string) int64 {
	if value, ok := m[key].(int64); ok {
		return value
	}

	return 0
}
