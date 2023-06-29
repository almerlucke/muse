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

func NewOsc(sf *io.WaveTableSoundFile, fc, sr, phase, tableIndex, amp float64) *Osc {
	tableSize := float64(sf.TableSize)

	return &Osc{
		sf:         sf,
		tableIndex: tableIndex * float64(len(sf.Tables)-1),
		phase:      phase * tableSize,
		inc:        fc / sr * tableSize,
		fc:         fc,
		sr:         sr,
		amp:        amp,
	}
}

func (osc *Osc) TableIndex() float64 {
	maxIndex := float64(len(osc.sf.Tables) - 1)

	return osc.tableIndex / maxIndex
}

func (osc *Osc) SetTableIndex(index float64) {
	maxIndex := float64(len(osc.sf.Tables) - 1)

	osc.tableIndex = index * maxIndex

	if osc.tableIndex > maxIndex {
		osc.tableIndex = maxIndex
	}

	if osc.tableIndex < 0 {
		osc.tableIndex = 0
	}
}

func (osc *Osc) Frequency() float64 {
	return osc.fc
}

func (osc *Osc) SetFrequency(fc float64) {
	osc.inc = fc / osc.sr * float64(osc.sf.TableSize)
}

func (osc *Osc) Phase() float64 {
	return osc.phase / float64(osc.sf.TableSize)
}

func (osc *Osc) SetPhase(phase float64) {
	tableSize := float64(osc.sf.TableSize)

	osc.phase = phase * tableSize

	for osc.phase >= tableSize {
		osc.phase -= tableSize
	}

	for osc.phase < 0.0 {
		osc.phase += tableSize
	}
}

func (osc *Osc) OffsetPhase(offset float64) {
	tableSize := float64(osc.sf.TableSize)

	osc.phase += offset * tableSize

	for osc.phase >= tableSize {
		osc.phase -= tableSize
	}

	for osc.phase < 0.0 {
		osc.phase += tableSize
	}
}

func (osc *Osc) Amplitude() float64 {
	return osc.amp
}

func (osc *Osc) SetAmplitude(amp float64) {
	osc.amp = amp
}

func (osc *Osc) Generate() float64 {
	t1 := int(osc.tableIndex)
	t2 := t1 + 1
	if t2 >= len(osc.sf.Tables) {
		t2 = len(osc.sf.Tables) - 1
	}

	tfract := osc.tableIndex - float64(t1)

	v1 := osc.sf.Tables[t1].Lookup(osc.phase, true)
	v2 := osc.sf.Tables[t2].Lookup(osc.phase, true)

	out := v1 + (v2-v1)*tfract

	tableSize := float64(osc.sf.TableSize)

	osc.phase += osc.inc

	for osc.phase >= tableSize {
		osc.phase -= tableSize
	}

	for osc.phase < 0.0 {
		osc.phase += tableSize
	}

	return out * osc.amp
}
