package muse

type Identifiable interface {
	SetIdentifier(identifier string)
	Identifier() string
}
