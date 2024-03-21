package io

import (
	"fmt"
	"github.com/almerlucke/muse/buffer"
	"github.com/almerlucke/sndfile"
	"github.com/almerlucke/sndfile/writer"
)

type InputBuffer struct {
	conv    *writer.ChannelConverter[float64]
	buffers [][]float64
}

func NewInputBuffer(frameSize int, numChannels int) *InputBuffer {
	return &InputBuffer{
		conv:    writer.NewChannelConverter[float64](frameSize, numChannels),
		buffers: make([][]float64, numChannels),
	}
}

func (ib *InputBuffer) Convert(input any) []float32 {
	for i, b := range input.([]buffer.Buffer) {
		ib.buffers[i] = b
	}

	return ib.conv.Convert(ib.buffers)
}

func (ib *InputBuffer) FrameSize() int {
	return ib.conv.FrameSize()
}

type WaveTableSoundFile struct {
	Tables    []buffer.Buffer
	TableSize int
}

func NewWaveTableSoundFile(filePath string, tableSize int) (*WaveTableSoundFile, error) {
	sndFile, err := sndfile.NewSoundFile(filePath)
	if err != nil {
		return nil, err
	}

	numTables := int(sndFile.NumFrames()) / tableSize
	remaining := int(sndFile.NumFrames()) % tableSize

	if remaining != 0 {
		return nil, fmt.Errorf(
			"wavetable file did not contain exact multiple of table size %d: numTables = %d,  remaining = %d", tableSize, numTables, remaining,
		)
	}

	wsf := &WaveTableSoundFile{
		Tables:    make([]buffer.Buffer, numTables),
		TableSize: tableSize,
	}

	buf := sndFile.Buffer(0, 0)
	offset := 0

	for i := 0; i < numTables; i++ {
		wsf.Tables[i] = buf[offset : offset+tableSize]
		offset += tableSize
	}

	return wsf, nil
}
