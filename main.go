package main

import (
	"time"
)

func main() {
	go Server(25555)
	go Server(25554)
	time.Sleep(4 * time.Second)
	go Client(25555)
	go Client(25554)
	time.Sleep(50 * time.Second)
}
