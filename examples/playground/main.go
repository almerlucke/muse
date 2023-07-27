package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/ops"
	"github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/timer"
	"github.com/almerlucke/muse/modules/fmsynth"
	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/template"
)

/*
 @param  {AudioBuffer} bufferNewSamples Microphone/MediaElement audio chunk
 * @return {Float32Array} 'audio/l16' chunk

 WebAudioL16Stream.prototype.downsample = function downsample(bufferNewSamples) {
  var buffer = null,
    newSamples = bufferNewSamples.length,
    unusedSamples = this.bufferUnusedSamples.length,
    i,
    offset;

  if (unusedSamples > 0) {
    buffer = new Float32Array(unusedSamples + newSamples);
    for (i = 0; i < unusedSamples; ++i) {
      buffer[i] = this.bufferUnusedSamples[i];
    }
    for (i = 0; i < newSamples; ++i) {
      buffer[unusedSamples + i] = bufferNewSamples[i];
    }
  } else {
    buffer = bufferNewSamples;
  }

  // downsampling variables
  var filter = [
      -0.037935, -0.00089024, 0.040173, 0.019989, 0.0047792, -0.058675, -0.056487,
      -0.0040653, 0.14527, 0.26927, 0.33913, 0.26927, 0.14527, -0.0040653, -0.056487,
      -0.058675, 0.0047792, 0.019989, 0.040173, -0.00089024, -0.037935
    ],
    samplingRateRatio = this.options.sourceSampleRate / TARGET_SAMPLE_RATE,
    nOutputSamples = Math.floor((buffer.length - filter.length) / (samplingRateRatio)) + 1,
    outputBuffer = new Float32Array(nOutputSamples);

  for (i = 0; i + filter.length - 1 < buffer.length; i++) {
    offset = Math.round(samplingRateRatio * i);
    var sample = 0;
    for (var j = 0; j < filter.length; ++j) {
      sample += buffer[offset + j] * filter[j];
    }
    outputBuffer[i] = sample;
  }

  var indexSampleAfterLastUsed = Math.round(samplingRateRatio * i);
  var remaining = buffer.length - indexSampleAfterLastUsed;
  if (remaining > 0) {
    this.bufferUnusedSamples = new Float32Array(remaining);
    for (i = 0; i < remaining; ++i) {
      this.bufferUnusedSamples[i] = buffer[indexSampleAfterLastUsed + i];
    }
  } else {
    this.bufferUnusedSamples = new Float32Array(0);
  }

  return outputBuffer;
};
*/

func main() {
	root := muse.New(1)

	fm := fmsynth.New(18, waveshaping.NewSineTable(2048)).Named("fm").(*fmsynth.FMSynth)
	fm.OperatorSettings[1].Level = 0.5
	fm.OperatorSettings[5].Level = 0.5
	fm.PitchEnvLevels = [4]float64{0.49, 0.51, 0.495, 0.5}
	fm.PitchEnvRates = [4]float64{0.95, 0.95, 0.95, 0.95}
	fm.ReleaseMode = ops.EnvelopeDurationRelease
	fm.ApplySettingsChange()
	fm.Add(root)

	banger.NewTemplateGenerator([]string{"fm"}, template.Template{
		"noteOn":   value.NewSequence([]any{36, 36, 48, 41, 51, 51, 49, 47, 32, 33}),
		"duration": value.NewSequence([]any{500.0, 300.0, 250.0, 150.0, 300.0, 125.0, 125.0, 500.0, 375.0}),
		"level":    1.0,
	}).MsgrNamed("notes").MsgrAdd(root)

	timer.NewTimer(250.0, []string{"notes"}).MsgrAdd(root)

	root.In(fm)

	root.RenderAudio()
}
