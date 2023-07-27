package oversampler

import "github.com/almerlucke/muse"

/*
Run inner module at higher samplerate, downsample output
*/

type Oversampler struct {
	*muse.BaseModule
	module muse.Module
}
