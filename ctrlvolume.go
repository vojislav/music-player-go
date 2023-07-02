package main

import (
	"math"

	"github.com/faiface/beep"
)

type CtrlVolume struct {
	Streamer beep.Streamer
	Paused   bool
	Base     float64
	Volume   float64
	Silent   bool
}

var volumeStep = 0.2
var volumePercent = 100

func (cv *CtrlVolume) Stream(samples [][2]float64) (n int, ok bool) {
	if cv.Streamer == nil {
		return 0, false
	}

	if cv.Paused {
		for i := range samples {
			samples[i] = [2]float64{}
		}
		return len(samples), true
	}

	n, ok = cv.Streamer.Stream(samples)
	gain := 0.0

	if !cv.Silent {
		gain = math.Pow(cv.Base, cv.Volume)
	}

	for i := range samples[:n] {
		samples[i][0] *= gain
		samples[i][1] *= gain
	}
	return n, ok
}

func (cv *CtrlVolume) Err() error {
	if cv.Streamer == nil {
		return nil
	}
	return cv.Streamer.Err()
}
