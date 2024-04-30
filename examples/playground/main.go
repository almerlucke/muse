package main

import (
	"github.com/almerlucke/muse/components/scheduler"
	"log"
)

func main() {
	sched := scheduler.New(0, "1", 100, "2", 110, "3", 400, "4")
	timestamp := int64(0)

	for timestamp < 800 {
		items := sched.Schedule(timestamp)
		if len(items) > 0 {
			log.Printf("timestamp: %d -> items: %v", timestamp, items)
		}
		timestamp += 50
	}
}
