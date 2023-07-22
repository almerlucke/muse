package main

import (
	"fmt"
	"log"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/osc"
)

func existingDependency(deps []muse.Module, m muse.Module) bool {
	for _, dep := range deps {
		if dep == m {
			return true
		}
	}

	return false
}

func moduleDependencies(m muse.Module) []muse.Module {
	deps := []muse.Module{}

	for i := 0; i < m.NumInputs(); i++ {
		input := m.InputAtIndex(i)

		for _, conn := range input.Connections {
			if !existingDependency(deps, conn.Module) {
				deps = append(deps, conn.Module)
			}
		}
	}

	return deps
}

func dependenciesString(deps []muse.Module) string {
	depStr := "["

	for _, dep := range deps {
		depStr += fmt.Sprintf("%v,", dep.Identifier())
	}

	depStr += "]"

	return depStr
}

func isInOutputChain(start muse.Module, search muse.Module, visitedMap modMap) bool {
	for i := 0; i < start.NumOutputs(); i++ {
		output := start.OutputAtIndex(i)
		for _, conn := range output.Connections {
			if conn.Module == search {
				return true
			}

			if _, visited := visitedMap[conn.Module]; !visited {
				visitedMap[conn.Module] = true
				if isInOutputChain(conn.Module, search, visitedMap) {
					return true
				}
			}
		}
	}

	return false
}

func getConnectionDepth(mod muse.Module, depthMap connectionDepthMap, colorMap NodeColorMap) int {
	if depth, ok := depthMap[mod]; ok {
		return depth
	}

	colorMap[mod] = Grey

	maxDepth := 0

	skipConnMap := map[*muse.Connection]bool{}

	for i := 0; i < mod.NumInputs(); i++ {
		input := mod.InputAtIndex(i)
		for _, conn := range input.Connections {
			existingColor, ok := colorMap[conn.Module]
			if ok {
				if existingColor == Grey {
					// Cycle detected, skip this connection
					skipConnMap[conn] = true
					continue
				}
			} else {
				colorMap[conn.Module] = Grey
			}
		}
	}

	for i := 0; i < mod.NumInputs(); i++ {
		input := mod.InputAtIndex(i)
		for _, conn := range input.Connections {
			if _, ok := skipConnMap[conn]; !ok {
				depth := getConnectionDepth(conn.Module, depthMap, colorMap)
				if depth > maxDepth {
					maxDepth = depth
				}
			}
		}
	}

	depthMap[mod] = maxDepth + 1
	colorMap[mod] = Black

	return depthMap[mod]
}

type NodeColor int

const (
	White NodeColor = iota
	Grey
	Black
)

type NodeColorMap map[muse.Module]NodeColor

type modMap map[muse.Module]bool

type connectionDepthMap map[muse.Module]int

type graphIndexMap map[muse.Module]int

type visitedNodes map[muse.Module]bool

func main() {
	sr := 44100.0
	bufferSize := 128
	config := &muse.Configuration{SampleRate: sr, BufferSize: bufferSize}

	o1 := osc.New(100.0, 0.0, config).Named("osc1")
	o2 := osc.New(100.0, 0.0, config).Named("osc2")
	o3 := osc.New(100.0, 0.0, config).Named("osc3")
	m1 := functor.NewMult(2, config)
	m1.SetIdentifier("m1")
	m2 := functor.NewMult(2, config)
	m2.SetIdentifier("m2")
	a3 := functor.NewAmp(1, config)
	a3.SetIdentifier("a3")
	off1 := functor.NewScale(1.0, 0.0, config)
	off1.SetIdentifier("off1")

	o1.Connect(0, m1, 0)
	o2.Connect(0, m1, 1)
	m1.Connect(0, m2, 1)
	o2.Connect(0, m2, 0)
	o3.Connect(0, m2, 1)
	o3.Connect(0, off1, 0)
	m2.Connect(0, a3, 0)
	off1.Connect(0, a3, 0)
	off1.Connect(0, o3, 0)

	depthMap := connectionDepthMap{}
	// indexMap := graphIndexMap{}

	getConnectionDepth(a3, depthMap, NodeColorMap{})
	// getGraphIndex(a3, 0, indexMap)

	for k, v := range depthMap {
		log.Printf("mod %v: %d", k.Identifier(), v)
	}

	// for k, v := range indexMap {
	// 	log.Printf("index %v: %d", k.Identifier(), v)
	// }
}
