package main

import "log"

type ICon struct {
	InIndex  int
	Module   Module
	OutIndex int
}

func ICons(rawIcons ...any) []*ICon {
	index := 0
	icons := make([]*ICon, len(rawIcons)/3)

	for index < len(rawIcons) {
		inIndex := rawIcons[index].(int)
		module := rawIcons[index+1].(Module)
		outIndex := rawIcons[index+2].(int)

		icons[index/3] = &ICon{
			InIndex:  inIndex,
			Module:   module,
			OutIndex: outIndex,
		}

		index += 3
	}

	return icons
}

type Module interface {
}

type TestModule struct {
}

func NewTestModule(inputs []*ICon) *TestModule {
	log.Printf("icons %v", inputs)
	return &TestModule{}
}

func main() {
	t1 := NewTestModule(nil)
	t2 := NewTestModule(ICons(0, t1, 1, 0, t1, 2))

	log.Printf("t %v", t2)
}
