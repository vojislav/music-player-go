package main

import (
	"fmt"
	"time"
)

type StatusCustomMessage struct {
	message string
	sleepMs int
}

var statusChannel = make(chan bool, 1024)

var statusCustomMessageChannel = make(chan StatusCustomMessage)

// update status and do it every 1 second again
func asyncRequestStatusUpdate() {
	statusChannel <- true
}

// update status and stop periodic updating
func asyncRequestStatusPause() {
	statusChannel <- false
}

// send custom status message
func syncRequestCustomStatus(message string, sleepMs int) {
	statusCustomMessageChannel <- StatusCustomMessage{message, sleepMs}
}

// Activates every second while track is playing.
// Checks if track should be scrobbled and if next track should be played.
// Returns true if there is current track and false if there is not
func tickTrack() bool {
	if currentTrack.stream == nil {
		return false
	}

	if currentTrack.stream.Position() >= currentTrack.stream.Len()/2 {
		scrobble(toInt(currentTrack.ID), "true")
	}
	if currentTrack.stream.Position() == currentTrack.stream.Len() {
		requestNextTrack()
	}
	return true
}

// thread that updates status text
func trackTime() {
	updatePeriodically := false

	for {
		select {
		case <-time.After(1 * time.Second): // timeout after 1 second
			if !updatePeriodically || !tickTrack() {
				continue
			}
			updateCurrentTrackText()

		case request := <-statusChannel: // we have update pending
			if request {
				updatePeriodically = true
				tickTrack()
			} else {
				updatePeriodically = false
			}
			updateCurrentTrackText()

		case request := <-statusCustomMessageChannel:
			currentTrackText.Clear()
			fmt.Fprint(currentTrackText, request.message)
			time.Sleep(time.Millisecond * time.Duration(request.sleepMs))

			updateCurrentTrackText()
		}
	}
}
