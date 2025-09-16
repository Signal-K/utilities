package main

import (
	"log"
	"time"
)

func main() {
	log.Println("Star Sailors background service starting...")

	// Test (getting started) - call a fake notification every minute
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ticker.C:
			log.Println("Checking for updates....")
		}
	}
}
