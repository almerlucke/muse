package main

import (
	"github.com/almerlucke/muse/utils/containers/list"
	"log"
)

func main() {
	l := list.New[int]()

	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)

	log.Printf("%d", l.Reduce(0, func(accum int, cur int) int {
		return accum + cur
	}))

	for it := l.Iterator(true); !it.Finished(); {
		v, _ := it.Next()
		log.Printf("elem: %d", v)
	}
}
