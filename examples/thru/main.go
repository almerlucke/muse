package main

import "github.com/almerlucke/muse"

func main() {
	m := muse.NewWithInputs(1, 1)
	m.Connect(0, m, 0)
	_ = m.RenderAudio()
}
