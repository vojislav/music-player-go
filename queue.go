package main

import "fmt"

var queuePosition = -1

func addToQueue(currentTrackIndex int, currentTrackName, currentTrackID string, _ rune) {
	download(currentTrackID)
	tags := getTags(cacheDirectory + currentTrackID + ".mp3")
	itemText := fmt.Sprintf("%s - %s", tags.Artist(), tags.Title())
	queueList.AddItem(itemText, currentTrackID, 0, nil)
}

func addToQueueAndPlay(currentTrackIndex int, currentTrackName, currentTrackID string, _ rune) {
	download(currentTrackID)

	tags := getTags(cacheDirectory + currentTrackID + ".mp3")
	itemText := fmt.Sprintf("%s - %s", tags.Artist(), tags.Title())

	queueList.AddItem(itemText, currentTrackID, 0, nil)
	queuePosition = queueList.GetItemCount() - 1

	playTrack(queuePosition, currentTrackName, currentTrackID, 0)
}
