package wtscan

import (
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/utils/mmath"
)

type Scanner struct {
	sf           *io.WaveTableSoundFile
	phase        float64
	inc          float64
	fc           float64
	sr           float64
	amp          float64
	newScanIndex float64
	scanIndex    float64
}

func New(sf *io.WaveTableSoundFile, fc, sr, phase, scanIndex, amp float64) *Scanner {
	return &Scanner{
		sf:           sf,
		phase:        phase,
		inc:          fc / sr,
		fc:           fc,
		sr:           sr,
		amp:          amp,
		scanIndex:    scanIndex,
		newScanIndex: scanIndex,
	}
}

func (sc *Scanner) wrap(v float64) float64 {
	for v >= 1.0 {
		v -= 1.0
	}

	for v < 0.0 {
		v += 1.0
	}

	return v
}

func (sc *Scanner) Frequency() float64 {
	return sc.fc
}

func (sc *Scanner) SetFrequency(fc float64) {
	sc.inc = fc / sc.sr
}

func (sc *Scanner) Phase() float64 {
	return sc.phase
}

func (sc *Scanner) SetPhase(phase float64) {
	sc.phase = sc.wrap(phase)
}

func (sc *Scanner) OffsetPhase(offset float64) {
	sc.phase += offset
	sc.phase = sc.wrap(sc.phase)
}

func (sc *Scanner) Amplitude() float64 {
	return sc.amp
}

func (sc *Scanner) SetAmplitude(amp float64) {
	sc.amp = amp
}

func (sc *Scanner) ScanIndex() float64 {
	return sc.scanIndex
}

func (sc *Scanner) SetScanIndex(index float64) {
	sc.newScanIndex = mmath.Limit(index, 0, 0.999)
}

func (sc *Scanner) Generate() float64 {
	tl := len(sc.sf.Tables)
	tf := sc.scanIndex * float64(tl)
	t1 := int(tf)
	t2 := t1 + 1
	if t2 >= tl {
		t2 = tl - 1
	}

	tfrct := tf - float64(t1)

	si := sc.phase * float64(sc.sf.TableSize)
	v1 := sc.sf.Tables[t1].Lookup(si, true)
	v2 := sc.sf.Tables[t2].Lookup(si, true)

	out := v1 + (v2-v1)*tfrct

	sc.phase += sc.inc
	if sc.phase >= 1.0 && sc.scanIndex != sc.newScanIndex {
		sc.scanIndex = sc.newScanIndex
	}

	sc.phase = sc.wrap(sc.phase)

	return out * sc.amp
}
