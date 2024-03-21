package wtosc

import "github.com/almerlucke/muse/io"

type Osc struct {
	sf         *io.WaveTableSoundFile
	tableIndex float64
	phase      float64
	inc        float64
	fc         float64
	sr         float64
	amp        float64
}

func New(sf *io.WaveTableSoundFile, fc, sr, phase, tableIndex, amp float64) *Osc {
	return &Osc{
		sf:         sf,
		tableIndex: tableIndex,
		phase:      phase,
		inc:        fc / sr,
		fc:         fc,
		sr:         sr,
		amp:        amp,
	}
}

func (osc *Osc) TableIndex() float64 {
	return osc.tableIndex
}

func (osc *Osc) wrap(v float64) float64 {
	for v >= 1.0 {
		v -= 1.0
	}

	for v < 0.0 {
		v += 1.0
	}

	return v
}

func (osc *Osc) SetTableIndex(index float64) {
	osc.tableIndex = osc.wrap(index)
}

func (osc *Osc) Frequency() float64 {
	return osc.fc
}

func (osc *Osc) SetFrequency(fc float64) {
	osc.inc = fc / osc.sr
}

func (osc *Osc) Phase() float64 {
	return osc.phase
}

func (osc *Osc) SetPhase(phase float64) {
	osc.phase = osc.wrap(phase)
}

func (osc *Osc) OffsetPhase(offset float64) {
	osc.phase += offset
	osc.phase = osc.wrap(osc.phase)
}

func (osc *Osc) Amplitude() float64 {
	return osc.amp
}

func (osc *Osc) SetAmplitude(amp float64) {
	osc.amp = amp
}

func (osc *Osc) Generate() float64 {
	tl := len(osc.sf.Tables)
	tf := osc.tableIndex * float64(tl)
	t1 := int(tf)
	t2 := t1 + 1
	if t2 >= tl {
		t2 = tl - 1
	}

	tfrct := tf - float64(t1)

	si := osc.phase * float64(osc.sf.TableSize)
	v1 := osc.sf.Tables[t1].Lookup(si, true)
	v2 := osc.sf.Tables[t2].Lookup(si, true)

	out := v1 + (v2-v1)*tfrct

	osc.phase += osc.inc
	osc.phase = osc.wrap(osc.phase)

	return out * osc.amp
}
