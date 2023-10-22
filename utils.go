package main

import (
	"errors"
	"os"
)

// returns true if track with ID trackID exists locally
func trackExists(trackID string) bool {
	if _, err := os.Stat(getTrackPath(trackID)); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
