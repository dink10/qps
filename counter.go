package main

import (
	"github.com/paulbellamy/ratecounter"
	"time"
)

// NewCounter - Create counter
func NewCounter() (counter *ratecounter.RateCounter) {
	// We're recording marks-per-1second
	counter = ratecounter.NewRateCounter(1 * time.Second)

	return
}

// Increment - add +1 to value
func Increment(counter *ratecounter.RateCounter) {
	// Record an event happening
	counter.Incr(1)
}

// Count - get counter value
func Count(counter *ratecounter.RateCounter) int64 {
	return counter.Rate()
}
