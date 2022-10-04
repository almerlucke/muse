package rand

const (
	defaultSeed uint64 = 1
)

type Rand struct {
	u, v, w uint64
}

func NewRand() *Rand {
	return NewRandWithSeed(defaultSeed)
}

func NewRandWithSeed(seed uint64) *Rand {
	r := &Rand{}
	r.seed(seed)
	return r
}

func (r *Rand) seed(s uint64) {
	r.v = 4101842887655102017
	r.w = 1
	r.u = s ^ r.v
	r.RandInt()
	r.v = r.u
	r.RandInt()
	r.w = r.v
	r.RandInt()
}

// RandInt Random number between [0, 2^64 - 1] (Numerical Recipes in C" Third Edition)
func (r *Rand) RandInt() uint64 {
	r.u = r.u*2862933555777941757 + 7046029254386353087
	r.v ^= r.v >> 17
	r.v ^= r.v << 31
	r.v ^= r.v >> 8
	r.w = 4294957665*(r.w&0xffffffff) + (r.w >> 32)
	x := r.u ^ (r.u << 21)
	x ^= x >> 35
	x ^= x << 4

	return (x + r.v) ^ r.w
}

// RandFloat Random number between [0, 1)
func (r *Rand) RandFloat() float64 {
	return 5.42101086242752217e-20 * float64(r.RandInt())
}
