package main

import (
	"errors"
	"os"
)

// returns true if file exists at location loc
func fileExists(loc string) bool {
	if _, err := os.Stat(loc); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

// returns true if track with ID trackID exists locally
func trackExists(trackID string) bool {
	return fileExists(getTrackPath(trackID))
}

// returns bool that indicates if exact target is found, and index of it
func binary_search(array []int, target int) (bool, int) {
	comp := func(v int) bool { return v >= target }
	idx := bisect(array, comp)
	return (idx < len(array) && array[idx] == target), idx
}

func bisect(array []int, comp func(v int) bool) int {
	lo, hi := 0, len(array)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if !comp(array[mid]) {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}
