package main

import "log"

func main() {
	/*Read(location float64) float64 {
	buffer := delay.Buffer
	buflen := len(buffer)

	// if location >= float64(buflen) {
	// 	location = float64(buflen - 1)
	// }

	sampleLocation := float64(delay.WriteHead) - location

	for sampleLocation < 0.0 {
		sampleLocation += float64(buflen)
	}

	firstIndex := int(sampleLocation)
	fraction := sampleLocation - float64(firstIndex)
	secondIndex := firstIndex + 1

	if secondIndex >= buflen {
		secondIndex -= buflen
	}

	v1 := buffer[firstIndex]

	return v1 + (buffer[secondIndex]-v1)*fraction
	*/

	location := 44100.0 * 60000.0 / 105.0 * 1.25 * 0.001
	buflen := int(5000.0 * 44100.0 * 0.001)

	log.Printf("buflen %d", buflen)

	writeHead := 31500
	sampleLocation := float64(writeHead) - location

	log.Printf("location %f", location)

	log.Printf("sampleLocation1 %f", sampleLocation)

	for sampleLocation < 0.0 {
		sampleLocation += float64(buflen)
	}

	firstIndex := int(sampleLocation)
	fraction := sampleLocation - float64(firstIndex)
	secondIndex := firstIndex + 1

	if secondIndex >= buflen {
		secondIndex -= buflen
	}

	log.Printf("firstIndex %d", firstIndex)
	log.Printf("fraction %f", fraction)
	log.Printf("secondIndex %d", secondIndex)
}
