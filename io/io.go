package io

import (
	"fmt"
	"github.com/almerlucke/muse/buffer"
	"github.com/almerlucke/sndfile"
	"os"
	"path/filepath"
)

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

func LoadSoundBankFromDirectory(root string, bank sndfile.SoundBank) error {
	var parentCnt = map[string]int{}

	err := filepath.Walk(root, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			dir := filepath.Dir(filePath)
			parentDir := filepath.Base(dir)
			cnt, ok := parentCnt[parentDir]
			if ok {
				cnt = cnt + 1
				parentCnt[parentDir] = cnt
			} else {
				cnt = 1
				parentCnt[parentDir] = 1
			}
			sf, err := sndfile.NewMipMapSoundFile(filePath, 4)
			if err != nil {
				return err
			}

			bank[fmt.Sprintf("%s%d", parentDir, cnt)] = sf
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
