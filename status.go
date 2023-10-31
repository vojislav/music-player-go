package main

import (
	"fmt"
	"time"
)

type StatusOperation int64

const (
	StartUpdate StatusOperation = iota
	PauseUpdate
	CustomStatus
)

type StatusChangeRequest struct {
	opcode  StatusOperation
	message string
	sleepMs int
}

var statusChannel = make(chan StatusChangeRequest, 1024)

func requestStatusChange(op StatusOperation, message string, sleepMs int) {
	statusChannel <- StatusChangeRequest{op, message, sleepMs}
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

		case request := <-statusChannel: // we have request pending
			switch request.opcode {
			case StartUpdate: // continue updating
				updatePeriodically = true
				tickTrack()

			case PauseUpdate: // pause updating
				updatePeriodically = false

			case CustomStatus:
				currentTrackText.Clear()
				fmt.Fprint(currentTrackText, request.message)
				time.Sleep(time.Millisecond * time.Duration(request.sleepMs))
			}
			updateCurrentTrackText()
		}
	}
}
