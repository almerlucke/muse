package filters

type ParamType int

const (
	FloatParam ParamType = iota
	IntParam
	BoolParam
	ArrayParam
)

type Param interface {
	Name() string
	Type() ParamType
}

/*
	MinValue float64
	MaxValue float64
*/

type Filter interface {
	SetCutoff(float64, float64)   // Set cutoff frequency and sample rate
	SetParam(string, float64)     // Set other parameter one by one
	SetParams(map[string]float64) // Set other parameters with a map
	Update()                      // Update internal coeffs
	Filter(float64) float64       // Filter input signal
}

/*

#include <memory.h>
#include <stdio.h>
#include <math.h>


#define polyin  float
#define polyout float

#define BUFSIZE 64

float delta_func [BUFSIZE];
float out_buffer [BUFSIZE];






void tick ( float in, float cf, float reso, float *out ) {


// start of sm code


// filter based on the text "Non linear digital implementation of the moog ladder filter" by Antti Houvilainen
// adopted from Csound code at http://www.kunstmusik.com/udo/cache/moogladder.udo
polyin input;
polyin cutoff;
polyin resonance;

polyout sigout;


// remove this line in sm
input = in; cutoff = cf; resonance = reso;


// resonance [0..1]
// cutoff from 0 (0Hz) to 1 (nyquist)

float pi; pi = 3.1415926535;
float v2; v2 = 40000;   // twice the 'thermal voltage of a transistor'
float sr; sr = 22100;

float  cutoff_hz;
cutoff_hz = cutoff * sr;

static float az1;
static float az2;
static float az3;
static float az4;
static float az5;
static float ay1;
static float ay2;
static float ay3;
static float ay4;
static float amf;



float x;         // temp var: input for taylor approximations
float xabs;
float exp_out;
float tanh1_out, tanh2_out;
float kfc;
float kf;
float kfcr;
float kacr;
float k2vg;

kfc  = cutoff_hz/sr; // sr is half the actual filter sampling rate
kf   = cutoff_hz/(sr*2);
// frequency & amplitude correction
kfcr = 1.8730*(kfc*kfc*kfc) + 0.4955*(kfc*kfc) - 0.6490*kfc + 0.9988;
kacr = -3.9364*(kfc*kfc)    + 1.8409*kfc       + 0.9968;

x  = -2.0 * pi * kfcr * kf;
exp_out  = expf(x);

k2vg = v2*(1-exp_out); // filter tuning


// cascade of 4 1st order sections
float x1 = (input - 4*resonance*amf*kacr) / v2;
float tanh1 = tanhf (x1);
float x2 = az1/v2;
float tanh2 = tanhf (x2);
ay1 = az1 + k2vg * ( tanh1 - tanh2);

// ay1  = az1 + k2vg * ( tanh( (input - 4*resonance*amf*kacr) / v2) - tanh(az1/v2) );
az1  = ay1;

ay2  = az2 + k2vg * ( tanh(ay1/v2) - tanh(az2/v2) );
az2  = ay2;

ay3  = az3 + k2vg * ( tanh(ay2/v2) - tanh(az3/v2) );
az3  = ay3;

ay4  = az4 + k2vg * ( tanh(ay3/v2) - tanh(az4/v2) );
az4  = ay4;

// 1/2-sample delay for phase compensation
amf  = (ay4+az5)*0.5;
az5  = ay4;



// oversampling (repeat same block)
ay1  = az1 + k2vg * ( tanh( (input - 4*resonance*amf*kacr) / v2) - tanh(az1/v2) );
az1  = ay1;

ay2  = az2 + k2vg * ( tanh(ay1/v2) - tanh(az2/v2) );
az2  = ay2;

ay3  = az3 + k2vg * ( tanh(ay2/v2) - tanh(az3/v2) );
az3  = ay3;

ay4  = az4 + k2vg * ( tanh(ay3/v2) - tanh(az4/v2) );
az4  = ay4;

// 1/2-sample delay for phase compensation
amf  = (ay4+az5)*0.5;
az5  = ay4;


sigout = amf;



// end of sm code


*out   = sigout;

} // tick


int main ( int argc, char *argv[] )  {

    // set delta function
    memset ( delta_func, 0, sizeof(delta_func));
    delta_func[0] = 1.0;

    int i = 0;
    for ( i = 0; i < BUFSIZE; i++ ) {
            tick ( delta_func[i], 0.6, 0.7, out_buffer+i );
    }
    for ( i = 0; i < BUFSIZE; i++ ) {
            printf ("%f;", out_buffer[i] );
    }
    printf ( "\n" );


} // main
*/

/*
//Init
cutoff = cutoff freq in Hz
fs = sampling frequency //(e.g. 44100Hz)
res = resonance [0 - 1] //(minimum - maximum)

f = 2 * cutoff / fs; //[0 - 1]
k = 3.6*f - 1.6*f*f -1; //(Empirical tunning)
p = (k+1)*0.5;
scale = e^((1-p)*1.386249;
r = res*scale;
y4 = output;

y1=y2=y3=y4=oldx=oldy1=oldy2=oldy3=0;

//Loop
//--Inverted feed back for corner peaking
x = input - r*y4;

//Four cascaded onepole filters (bilinear transform)
y1=x*p + oldx*p - k*y1;
y2=y1*p+oldy1*p - k*y2;
y3=y2*p+oldy2*p - k*y3;
y4=y3*p+oldy3*p - k*y4;

//Clipper band limited sigmoid
y4 = y4 - (y4^3)/6;

oldx = x;
oldy1 = y1;
oldy2 = y2;
oldy3 = y3;
*/

/*
This filter works and sounds fine in my VST.
I've re-written the code using templates, which makes life easier when switching between <float> and <double> implementation.



#pragma once

namespace DistoCore
{
  template<class T>
  class MoogFilter
  {
  public:
    MoogFilter();
    ~MoogFilter() {};

    T getSampleRate() const { return sampleRate; }
    void setSampleRate(T fs) { sampleRate = fs; calc(); }
    T getResonance() const { return resonance; }
    void setResonance(T filterRezo) { resonance = filterRezo; calc(); }
    T getCutoff() const { return cutoff; }
    T getCutoffHz() const { return cutoff * sampleRate * 0.5; }
    void setCutoff(T filterCutoff) { cutoff = filterCutoff; calc(); }

    void init();
    void calc();
    T process(T input);
    // filter an input sample using normalized params
    T filter(T input, T cutoff, T resonance);

  protected:
    // cutoff and resonance [0 - 1]
    T cutoff;
    T resonance;
    T sampleRate;
    T fs;
    T y1,y2,y3,y4;
    T oldx;
    T oldy1,oldy2,oldy3;
    T x;
    T r;
    T p;
    T k;
  };

    /*
		 template<class T>
		 MoogFilter<T>::MoogFilter()
		 : sampleRate(T(44100.0))
		 , cutoff(T(1.0))
		 , resonance(T(0.0))
		 {
			 init();
		 }

		 template<class T>
		 void MoogFilter<T>::init()
		 {
			 // initialize values
			 y1=y2=y3=y4=oldx=oldy1=oldy2=oldy3=T(0.0);
			 calc();
		 }

		 template<class T>
		 void MoogFilter<T>::calc()
		 {
			 // TODO: replace with your constant
			 const double kPi = 3.1415926535897931;

			 // empirical tuning
			 p = cutoff * (T(1.8) - T(0.8) * cutoff);
			 // k = p + p - T(1.0);
			 // A much better tuning seems to be:
			 k = T(2.0) * sin(cutoff * kPi * T(0.5)) - T(1.0);

			 T t1 = (T(1.0) - p) * T(1.386249);
			 T t2 = T(12.0) + t1 * t1;
			 r = resonance * (t2 + T(6.0) * t1) / (t2 - T(6.0) * t1);
		 };

		 template<class T>
		 T MoogFilter<T>::process(T input)
		 {
			 // process input
			 x = input - r * y4;

			 // four cascaded one-pole filters (bilinear transform)
			 y1 =  x * p + oldx  * p - k * y1;
			 y2 = y1 * p + oldy1 * p - k * y2;
			 y3 = y2 * p + oldy2 * p - k * y3;
			 y4 = y3 * p + oldy3 * p - k * y4;

			 // clipper band limited sigmoid
			 y4 -= (y4 * y4 * y4) / T(6.0);

			 oldx = x; oldy1 = y1; oldy2 = y2; oldy3 = y3;

			 return y4;
		 }

		 template<class T>
		 T MoogFilter<T>::filter(T input, T filterCutoff, T filterRezo)
		 {
			 // set params first
			 cutoff = filterCutoff;
			 resonance = filterRezo;
			 calc();

			 return process(input);
		 }
	 }*/

/*
	 TLP24DB = class
constructor create;
procedure process(inp,Frq,Res:single;SR:integer);
private
t, t2, x, f, k, p, r, y1, y2, y3, y4, oldx, oldy1, oldy2, oldy3: Single;
public outlp:single;
end;
----------------------------------------
implementation

constructor TLP24DB.create;
begin
  y1:=0;
  y2:=0;
  y3:=0;
  y4:=0;
  oldx:=0;
  oldy1:=0;
  oldy2:=0;
  oldy3:=0;
end;
procedure TLP24DB.process(inp: Single; Frq: Single; Res: Single; SR: Integer);
begin
  f := (Frq+Frq) / SR;
  p:=f*(1.8-0.8*f);
  k:=p+p-1.0;
  t:=(1.0-p)*1.386249;
  t2:=12.0+t*t;
  r := res*(t2+6.0*t)/(t2-6.0*t);
  x := inp - r*y4;
  y1:=x*p + oldx*p - k*y1;
  y2:=y1*p+oldy1*p - k*y2;
  y3:=y2*p+oldy2*p - k*y3;
  y4:=y3*p+oldy3*p - k*y4;
  y4 := y4 - ((y4*y4*y4)/6.0);
  oldx := x;
  oldy1 := y1+_kd;
  oldy2 := y2+_kd;;
  oldy3 := y3+_kd;;
  outlp := y4;
end;

// the result is in outlp
// 1/ call MyTLP24DB.Process
// 2/then get the result from outlp.
// this filter have a fantastic sound w/a very special res
// _kd is the denormal killer value.
*/
