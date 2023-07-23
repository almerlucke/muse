package modules

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/modules/functor"
)

func Mult(args ...any) muse.Module {
	iConns := muse.IConns(args...)
	return functor.NewMult(len(iConns)).IConns(iConns)
}

func Scale(mod muse.Module, outIndex int, scale float64, offset float64) muse.Module {
	return functor.NewScale(scale, offset).In(mod, outIndex)
}

func Amp(mod muse.Module, outIndex int, scale float64) muse.Module {
	return functor.NewScale(scale, 0).In(mod, outIndex)
}
