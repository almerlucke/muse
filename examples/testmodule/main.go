package main

import (
	"log"

	"github.com/almerlucke/muse"
)

type TestModule struct {
	*muse.BaseModule
	Value float64
}

func NewTestModule(value float64, identifier string) *TestModule {
	return &TestModule{
		BaseModule: muse.NewBaseModule(1, 1, identifier),
		Value:      value,
	}
}

func (t *TestModule) Run(config *muse.Configuration) bool {
	if !t.BaseModule.Run(config) {
		return false
	}

	input := t.InputAtIndex(0)
	output := t.OutputAtIndex(0)

	if t.InputAtIndex(0).IsConnected() {
		for i := 0; i < config.BufferSize; i++ {
			output.Buffer[i] = input.Buffer[i] + t.Value
		}
	} else {
		for i := 0; i < config.BufferSize; i++ {
			output.Buffer[i] = t.Value
		}
	}

	return true
}

func main() {
	env := muse.NewEnvironment(1, 44100, 12)

	ip := muse.NewPatch(1, 1, "ip_patch")
	it1 := NewTestModule(0.25, "it1")

	ip.AddModule(it1)
	muse.Connect(ip, 0, it1, 0)
	muse.Connect(it1, 0, ip, 0)

	t1 := NewTestModule(1.25, "t1")
	t11 := NewTestModule(0.123, "t11")
	t2 := NewTestModule(3.4, "t2")

	env.AddModule(t1)
	env.AddModule(t11)
	env.AddModule(ip)
	env.AddModule(t2)

	log.Printf("lookup %v", env.Lookup("ip_patch.it1"))

	muse.Connect(t1, 0, ip, 0)
	muse.Connect(t11, 0, ip, 0)
	muse.Connect(ip, 0, t2, 0)
	muse.Connect(t2, 0, env, 0)

	env.PrepareBuffers()

	for i := 0; i < 12; i++ {
		env.RunOnce()
		for _, sample := range env.OutputAtIndex(0).Buffer {
			log.Printf("%v", sample)
		}
	}
}
