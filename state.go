package muse

type Stater interface {
	GetState() map[string]any
	SetState(map[string]any)
}
