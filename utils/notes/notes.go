package notes

import "math"

type Note int

type Chord []Note

type Scale []int

var Major = Scale{0, 2, 4, 5, 7, 9, 11}
var NaturalMinor = Scale{0, 2, 3, 5, 7, 8, 10}
var HarmonicMinor = Scale{0, 2, 3, 5, 7, 8, 11}
var MelodicMinor = Scale{0, 2, 3, 5, 7, 9, 11}
var HungarianMinor = Scale{0, 2, 3, 6, 7, 8, 11}
var PhrygianDominant = Scale{0, 1, 4, 5, 7, 8, 10}
var Persian = Scale{0, 1, 4, 5, 6, 8, 11}

const (
	C0  Note = 12
	Cs0 Note = 13
	Db0 Note = 13
	D0  Note = 14
	Ds0 Note = 15
	Eb0 Note = 15
	E0  Note = 16
	F0  Note = 17
	Fs0 Note = 18
	Gb0 Note = 18
	G0  Note = 19
	Gs0 Note = 20
	Ab0 Note = 20
	A0  Note = 21
	As0 Note = 22
	Bb0 Note = 22
	B0  Note = 23
	C1  Note = 24
	Cs1 Note = 25
	Db1 Note = 25
	D1  Note = 26
	Ds1 Note = 27
	Eb1 Note = 27
	E1  Note = 28
	F1  Note = 29
	Fs1 Note = 30
	Gb1 Note = 30
	G1  Note = 31
	Gs1 Note = 32
	Ab1 Note = 32
	A1  Note = 33
	As1 Note = 34
	Bb1 Note = 34
	B1  Note = 35
	C2  Note = 36
	Cs2 Note = 37
	Db2 Note = 37
	D2  Note = 38
	Ds2 Note = 39
	Eb2 Note = 39
	E2  Note = 40
	F2  Note = 41
	Fs2 Note = 42
	Gb2 Note = 42
	G2  Note = 43
	Gs2 Note = 44
	Ab2 Note = 44
	A2  Note = 45
	As2 Note = 46
	Bb2 Note = 46
	B2  Note = 47
	C3  Note = 48
	Cs3 Note = 49
	Db3 Note = 49
	D3  Note = 50
	Ds3 Note = 51
	Eb3 Note = 51
	E3  Note = 52
	F3  Note = 53
	Fs3 Note = 54
	Gb3 Note = 54
	G3  Note = 55
	Gs3 Note = 56
	Ab3 Note = 56
	A3  Note = 57
	As3 Note = 58
	Bb3 Note = 58
	B3  Note = 59
	C4  Note = 60
	Cs4 Note = 61
	Db4 Note = 61
	D4  Note = 62
	Ds4 Note = 63
	Eb4 Note = 63
	E4  Note = 64
	F4  Note = 65
	Fs4 Note = 66
	Gb4 Note = 66
	G4  Note = 67
	Gs4 Note = 68
	Ab4 Note = 68
	A4  Note = 69
	As4 Note = 70
	Bb4 Note = 70
	B4  Note = 71
	C5  Note = 72
	Cs5 Note = 73
	Db5 Note = 73
	D5  Note = 74
	Ds5 Note = 75
	Eb5 Note = 75
	E5  Note = 76
	F5  Note = 77
	Fs5 Note = 78
	Gb5 Note = 78
	G5  Note = 79
	Gs5 Note = 80
	Ab5 Note = 80
	A5  Note = 81
	As5 Note = 82
	Bb5 Note = 82
	B5  Note = 83
	C6  Note = 84
	Cs6 Note = 85
	Db6 Note = 85
	D6  Note = 86
	Ds6 Note = 87
	Eb6 Note = 87
	E6  Note = 88
	F6  Note = 89
	Fs6 Note = 90
	Gb6 Note = 90
	G6  Note = 91
	Gs6 Note = 92
	Ab6 Note = 92
	A6  Note = 93
	As6 Note = 94
	Bb6 Note = 94
	B6  Note = 95
	C7  Note = 96
	Cs7 Note = 97
	Db7 Note = 97
	D7  Note = 98
	Ds7 Note = 99
	Eb7 Note = 99
	E7  Note = 100
	F7  Note = 101
	Fs7 Note = 102
	Gb7 Note = 102
	G7  Note = 103
	Gs7 Note = 104
	Ab7 Note = 104
	A7  Note = 105
	As7 Note = 106
	Bb7 Note = 106
	B7  Note = 107
	C8  Note = 108
	Cs8 Note = 109
	Db8 Note = 109
	D8  Note = 110
	Ds8 Note = 111
	Eb8 Note = 111
	E8  Note = 112
	F8  Note = 113
	Fs8 Note = 114
	Gb8 Note = 114
	G8  Note = 115
	Gs8 Note = 116
	Ab8 Note = 116
	A8  Note = 117
	As8 Note = 118
	Bb8 Note = 118
	B8  Note = 119
	C9  Note = 120
	Cs9 Note = 121
	Db9 Note = 121
	D9  Note = 122
	Ds9 Note = 123
	Eb9 Note = 123
	E9  Note = 124
	F9  Note = 125
	Fs9 Note = 126
	Gb9 Note = 126
	G9  Note = 127
)

const (
	O0 Note = 0
	O1 Note = 12
	O2 Note = 24
	O3 Note = 36
	O4 Note = 48
	O5 Note = 60
	O6 Note = 72
	O7 Note = 84
	O8 Note = 96
	O9 Note = 108
)

var CMajor = Chord{C0, E0, G0}
var CMajorInv1 = Chord{E0, G0, C1}
var CMajor7 = Chord{C0, E0, G0, Bb0}
var CMajor7_3 = Chord{C0, E0, Bb0}
var CMinor = Chord{C0, Eb0, G0}
var DMajor = Chord{D0, Gb0, A0}
var EMinor = Chord{E0, G0, B0}
var EMajor = Chord{E0, Gs0, B0}
var EMajorInv1 = Chord{Gs0, B0, E1}
var EMajor7 = Chord{E0, Gs0, B0, D1}
var EMajor7_3 = Chord{E0, Gs0, D1}
var FMajor = Chord{F0, A0, C1}
var FMajorInv1 = Chord{A0, C1, F1}
var GMajor = Chord{G0, B0, D1}
var GMajorInv1 = Chord{B0, D1, G1}
var GMajor7 = Chord{G0, B0, D1, F1}
var GMajor7_3 = Chord{G0, B0, F1}
var AMinor = Chord{A0, C1, E1}
var AMinorInv1 = Chord{C0, E0, A0}
var AMajor7 = Chord{A0, Db1, E1, G1}
var AMajor7_3 = Chord{A0, Db1, G1}
var BMinor = Chord{B0, D1, Gb1}
var DMajor7 = Chord{D0, Gb0, A0, C1}
var DMajor7_3 = Chord{D0, Gb0, C1}

func (s Scale) Note(root Note, index int) Note {
	return root + Note(s[index]%len(s))
}

func (s Scale) Chord(root Note, index int) Chord {
	return Chord{root + Note(s[index]%len(s)), root + Note(s[(index+2)%len(s)]), root + Note(s[(index+4)%len(s)])}
}

func (s Scale) Freq(root Note) []float64 {
	f := make([]float64, len(s))

	for i, n := range s {
		f[i] = Mtof(n + int(root))
	}

	return f
}

func (s Scale) Root(root Note) []Note {
	t := make([]Note, len(s))
	for i, n := range s {
		t[i] = Note(n) + root
	}
	return t
}

func (c Chord) Freq(transpose Note) []float64 {
	f := make([]float64, len(c))
	for i, n := range c {
		f[i] = Mtof(int(n + transpose))
	}
	return f
}

func (c Chord) FreqAny(transpose Note) []any {
	f := make([]any, len(c))
	for i, n := range c {
		f[i] = Mtof(int(n + transpose))
	}
	return f
}

func (n Note) Freq() float64 {
	return Mtof(int(n))
}

func Ftom(freq float64) int {
	return int(12.0*math.Log2(freq/440.0)) + 69
}

func Mtof(midiNote int) float64 {
	return math.Pow(2, float64(midiNote-69)/12.0) * 440.0
}

func Mtofs(notes ...int) []float64 {
	f := make([]float64, len(notes))
	for i, note := range notes {
		f[i] = Mtof(note)
	}
	return f
}

func Mtofa(notes ...int) []any {
	f := make([]any, len(notes))
	for i, note := range notes {
		f[i] = Mtof(note)
	}
	return f
}
