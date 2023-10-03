package muse

type ThruModule struct {
	*BaseModule
}

func NewThruModule() *ThruModule {
	thru := &ThruModule{
		BaseModule: NewBaseModule(1, 1),
	}

	thru.SetSelf(thru)
	return thru
}

func (t *ThruModule) Synthesize() bool {
	if !t.BaseModule.Synthesize() {
		return false
	}

	for i := 0; i < t.Config.BufferSize; i++ {
		t.Outputs[0].Buffer[i] = t.Inputs[0].Buffer[i]
	}

	return true
}
