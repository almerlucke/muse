package aifc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"reflect"
	"unsafe"

	"github.com/almerlucke/muse/io/sndfile/aifc/float80"
)

const (
	aifcVersion1        = uint32(0xA2805140)
	aifcCompressionName = "32-bit floating point"
	aifcCompressionType = "fl32"
)

var nativeEndian binary.ByteOrder

func init() {
	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xABCD)

	switch buf {
	case [2]byte{0xCD, 0xAB}:
		nativeEndian = binary.LittleEndian
	case [2]byte{0xAB, 0xCD}:
		nativeEndian = binary.BigEndian
	default:
		panic("Could not determine native endianness.")
	}
}

func swapBytes(b []byte) {
	if nativeEndian == binary.LittleEndian {
		n := len(b) / 2
		m := len(b) - 1

		for i := 0; i < n; i++ {
			tmp := b[i]
			b[i] = b[m-i]
			b[m-i] = tmp
		}
	}
}

func toPascalBytes(str string) []byte {
	strlen := len(str)
	size := strlen + 1
	ps := make([]byte, size+size%2) // pad 1 byte if uneven
	ps[0] = byte(strlen)
	for i, c := range str {
		ps[i+1] = byte(c)
	}
	return ps
}

type float interface {
	float32 | float64
}

func writeFloat[T float](aifc *AIFC, items []T) error {
	var buf bytes.Buffer

	numFrames := len(items) / int(aifc.numChannels)

	for _, item := range items {
		_ = binary.Write(&buf, binary.BigEndian, float32(item))
	}

	aifc.numSampleFrames += uint32(numFrames)

	_, err := aifc.file.Write(buf.Bytes())

	return err
}

type AIFC struct {
	numChannels     int16
	numSampleFrames uint32
	sampleRate      float64
	file            *os.File
}

func Open(filePath string, numChannels int16, sampleRate float64) (*AIFC, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	aifc := &AIFC{
		numChannels: numChannels,
		sampleRate:  sampleRate,
		file:        file,
	}

	err = aifc.writeHeader()
	if err != nil {
		file.Close()
		return nil, err
	}

	return aifc, nil
}

func (aifc *AIFC) Close() error {
	aifc.updateSizes()
	return aifc.file.Close()
}

func (aifc *AIFC) soundChunkSize() uint32 {
	return 8 + aifc.numSampleFrames*uint32(aifc.numChannels)*4
}

func (aifc *AIFC) commonChunkSize() uint32 {
	return 44
}

func (aifc *AIFC) versionChunkSize() uint32 {
	return 4
}

func (aifc *AIFC) formChunkSize() uint32 {
	return 4 + aifc.versionChunkSize() + aifc.commonChunkSize() + aifc.soundChunkSize() + 24
}

func (aifc *AIFC) updateSizes() error {
	// Seek form size
	_, err := aifc.file.Seek(4, os.SEEK_SET)
	if err != nil {
		return err
	}

	// Update form size
	size := aifc.formChunkSize()
	err = binary.Write(aifc.file, binary.BigEndian, uint32(size))
	if err != nil {
		return err
	}

	// Seek common number of frames
	_, err = aifc.file.Seek(34, os.SEEK_SET)
	if err != nil {
		return err
	}

	// Update number of frames
	err = binary.Write(aifc.file, binary.BigEndian, uint32(aifc.numSampleFrames))
	if err != nil {
		return err
	}

	// Seek ssnd size
	_, err = aifc.file.Seek(80, os.SEEK_SET)
	if err != nil {
		return err
	}

	// Update ssnd size
	err = binary.Write(aifc.file, binary.BigEndian, uint32(aifc.soundChunkSize()))
	if err != nil {
		return err
	}

	return nil
}

func (aifc *AIFC) Normalize(max float32) error {
	// Read and write buffers
	readerBuffer := make([]byte, 8192)
	writerBuffer := make([]byte, 0, 8192)

	// Seek to start of sound data
	_, err := aifc.file.Seek(92, os.SEEK_SET)
	if err != nil {
		return err
	}

	if max <= 0.0 {
		return nil
	}

	oom := 1.0 / max

	// Loop through all samples
	for {
		pos, err := aifc.file.Seek(0, os.SEEK_CUR)
		if err != nil {
			return err
		}

		// Read 8192 bytes if possible
		n, err := aifc.file.Read(readerBuffer)
		if err != nil {
			if err != io.EOF {
				return err
			}
			// Break if EOF
			break
		}

		// Create reader and writer objects for conversion
		byteReader := bytes.NewReader(readerBuffer[:n])
		writer := bytes.NewBuffer(writerBuffer[:0])

		// Normalize read bytes and convert back
		for {
			var f float32

			err := binary.Read(byteReader, binary.BigEndian, &f)
			if err != nil {
				if err != io.EOF {
					return err
				}
				break
			}

			binary.Write(writer, binary.BigEndian, f*oom)
		}

		// Seek last pos
		_, err = aifc.file.Seek(pos, os.SEEK_SET)
		if err != nil {
			return err
		}

		// Overwrite with normalized samples
		_, err = aifc.file.Write(writer.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func (aifc *AIFC) writeCommon() error {
	// COMM
	_, err := aifc.file.Write([]byte("COMM"))
	if err != nil {
		return err
	}

	compressionName := toPascalBytes(aifcCompressionName)

	// Size
	err = binary.Write(aifc.file, binary.BigEndian, uint32(len(compressionName)+22))
	if err != nil {
		return err
	}

	// Num Channels
	err = binary.Write(aifc.file, binary.BigEndian, aifc.numChannels)
	if err != nil {
		return err
	}

	// Write template number of frames, fill in later
	err = binary.Write(aifc.file, binary.BigEndian, uint32(0))
	if err != nil {
		return err
	}

	// Sample size
	err = binary.Write(aifc.file, binary.BigEndian, int16(32))
	if err != nil {
		return err
	}

	// Sample rate
	sampleRateBytes, _ := hex.DecodeString(float80.NewFromFloat64(aifc.sampleRate).String())
	_, err = aifc.file.Write(sampleRateBytes)
	if err != nil {
		return err
	}

	// Compression type
	_, err = aifc.file.Write([]byte(aifcCompressionType))
	if err != nil {
		return err
	}

	// Compression name
	_, err = aifc.file.Write(compressionName)
	if err != nil {
		return err
	}

	return nil
}

func (aifc *AIFC) writeVersion() error {
	// FVER
	_, err := aifc.file.Write([]byte("FVER"))
	if err != nil {
		return err
	}

	// Size
	err = binary.Write(aifc.file, binary.BigEndian, uint32(4))
	if err != nil {
		return err
	}

	// Version
	err = binary.Write(aifc.file, binary.BigEndian, aifcVersion1)
	if err != nil {
		return err
	}

	return nil
}

func (aifc *AIFC) writeDataHeader() error {
	// SSND
	_, err := aifc.file.Write([]byte("SSND"))
	if err != nil {
		return err
	}

	// Size template
	err = binary.Write(aifc.file, binary.BigEndian, uint32(0))
	if err != nil {
		return err
	}

	// Offset
	err = binary.Write(aifc.file, binary.BigEndian, uint32(0))
	if err != nil {
		return err
	}

	// Block allign
	err = binary.Write(aifc.file, binary.BigEndian, uint32(0))
	if err != nil {
		return err
	}

	return nil
}

func (aifc *AIFC) writeHeader() error {
	// FORM
	_, err := aifc.file.Write([]byte("FORM"))
	if err != nil {
		return err
	}

	// Write template size for now until we know full size
	err = binary.Write(aifc.file, binary.BigEndian, int32(0))
	if err != nil {
		return err
	}

	// AIFC
	_, err = aifc.file.Write([]byte("AIFC")) // 4 bytes
	if err != nil {
		return err
	}

	// Version chunk
	err = aifc.writeVersion() // 12 bytes
	if err != nil {
		return err
	}

	// Common chunk
	err = aifc.writeCommon() // 22 + 22 = 44
	if err != nil {
		return err
	}

	// Data header
	err = aifc.writeDataHeader()
	if err != nil {
		return err
	}

	return nil
}

func (aifc *AIFC) WriteItems(items any) error {
	t := reflect.TypeOf(items)
	if t.Kind() != reflect.Array && t.Kind() != reflect.Slice {
		return errors.New("expected a float array")
	}

	switch t.Elem().Kind() {
	case reflect.Float32:
		return writeFloat(aifc, items.([]float32))
	case reflect.Float64:
		return writeFloat(aifc, items.([]float64))
	}

	return nil
}
