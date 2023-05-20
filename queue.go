package main

import (
	"fmt"
)

var queuePosition = -1

func addToQueue(_ int, _, trackID string, _ rune) {
	tags := getTags(cacheDirectory + trackID + ".mp3")
	itemText := fmt.Sprintf("%s - %s", tags.Artist(), tags.Title())
	queueList.AddItem(itemText, trackID, 0, nil)
}

func addToQueueAndPlay(_ int, _, trackID string, _ rune) {
	tags := getTags(cacheDirectory + trackID + ".mp3")
	itemText := fmt.Sprintf("%s - %s", tags.Artist(), tags.Title())

	queueList.AddItem(itemText, trackID, 0, nil)
	queuePosition = queueList.GetItemCount() - 1

	playTrack(queuePosition, "", trackID, 0)
}

func removeFromQueue() {
	currentTrackIndex := queueList.GetCurrentItem()
	queueList.RemoveItem(currentTrackIndex)
	if currentTrackIndex < queuePosition {
		queuePosition--
	} else if currentTrackIndex == queuePosition {
		stopTrack()
		queuePosition = -1
	}
}

func addAlbumToQueue(albumID string) {

}
