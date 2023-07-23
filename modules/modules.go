package modules

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/modules/functor"
)

func Mult(args ...any) muse.Module {
	iConns := muse.IConns(args...)
	// Use config from first input module
	config := iConns[0].Object.(muse.Module).Configuration()
	return functor.NewMult(len(iConns), config).IConns(iConns)
}

func Scale(mod muse.Module, outIndex int, scale float64, offset float64) muse.Module {
	return functor.NewScale(scale, offset, mod.Configuration()).In(mod, outIndex)
}

func Amp(mod muse.Module, outIndex int, scale float64) muse.Module {
	return functor.NewScale(scale, 0, mod.Configuration()).In(mod, outIndex)
}
