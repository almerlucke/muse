package delay

// Delay structure
type Delay struct {
	Buffer   []float64
	WriteLoc int
	Length   int
}

/*
#ifndef STMLIB_DSP_DELAY_LINE_H_
#define STMLIB_DSP_DELAY_LINE_H_

#include "stmlib/stmlib.h"
#include "stmlib/dsp/dsp.h"

#include <algorithm>

namespace stmlib {

template<typename T, size_t max_delay>
class DelayLine {
 public:
  DelayLine() { }
  ~DelayLine() { }

  void Init() {
    Reset();
  }

  void Reset() {
    std::fill(&line_[0], &line_[max_delay], T(0));
    delay_ = 1;
    write_ptr_ = 0;
  }

  inline void set_delay(size_t delay) {
    delay_ = delay;
  }

  inline void Write(const T sample) {
    line_[write_ptr_] = sample;
    write_ptr_ = (write_ptr_ - 1 + max_delay) % max_delay;
  }

  inline const T Allpass(const T sample, size_t delay, const T coefficient) {
    T read = line_[(write_ptr_ + delay) % max_delay];
    T write = sample + coefficient * read;
    Write(write);
    return -write * coefficient + read;
  }

  inline const T WriteRead(const T sample, float delay) {
    Write(sample);
    return Read(delay);
  }

  inline const T Read() const {
    return line_[(write_ptr_ + delay_) % max_delay];
  }

  inline const T Read(size_t delay) const {
    return line_[(write_ptr_ + delay) % max_delay];
  }

  inline const T Read(float delay) const {
    MAKE_INTEGRAL_FRACTIONAL(delay)
    const T a = line_[(write_ptr_ + delay_integral) % max_delay];
    const T b = line_[(write_ptr_ + delay_integral + 1) % max_delay];
    return a + (b - a) * delay_fractional;
  }

  inline const T ReadHermite(float delay) const {
    MAKE_INTEGRAL_FRACTIONAL(delay)
    int32_t t = (write_ptr_ + delay_integral + max_delay);
    const T xm1 = line_[(t - 1) % max_delay];
    const T x0 = line_[(t) % max_delay];
    const T x1 = line_[(t + 1) % max_delay];
    const T x2 = line_[(t + 2) % max_delay];
    const float c = (x1 - xm1) * 0.5f;
    const float v = x0 - x1;
    const float w = c + v;
    const float a = w + v + (x2 - x0) * 0.5f;
    const float b_neg = w + a;
    const float f = delay_fractional;
    return (((a * f) - b_neg) * f + c) * f + x0;
  }

 private:
  size_t write_ptr_;
  size_t delay_;
  T line_[max_delay];

  DISALLOW_COPY_AND_ASSIGN(DelayLine);
};

}  // namespace stmlib

#endif  // STM
*/

// New create a new delay
func New(length int) *Delay {
	return &Delay{
		Buffer: make([]float64, length),
		Length: length,
	}
}

// Write to delay
func (d *Delay) Write(in float64) {
	d.Buffer[d.WriteLoc] = in
	d.WriteLoc = (d.WriteLoc - 1 + d.Length) % d.Length
}

func (d *Delay) ReadLinear(loc float64) float64 {
	iloc := int(loc)
	v := d.Buffer[(d.WriteLoc+iloc)%d.Length]
	return v + (d.Buffer[(d.WriteLoc+iloc+1)%d.Length]-v)*(loc-float64(iloc))
}

func (d *Delay) ReadHermite(loc float64) float64 {
	iloc := int(loc)
	f := loc - float64(iloc)
	t := d.WriteLoc + iloc + d.Length
	xm1 := d.Buffer[(t-1)%d.Length]
	x0 := d.Buffer[t%d.Length]
	x1 := d.Buffer[(t+1)%d.Length]
	x2 := d.Buffer[(t+2)%d.Length]
	c := (x1 - xm1) * 0.5
	v := x0 - x1
	w := c + v
	a := w + v + (x2-x0)*0.5
	bNeg := w + a

	return (((a*f)-bNeg)*f+c)*f + x0
}
