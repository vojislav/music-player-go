package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
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

func queueInputHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRight, tcell.KeyEnter:
		currentTrackIndex := queueList.GetCurrentItem()
		currentTrackName, currentTrackID := queueList.GetItemText(currentTrackIndex)
		playTrack(currentTrackIndex, currentTrackName, currentTrackID, 0)
		return nil
	case tcell.KeyLeft:
		return nil
	case tcell.KeyDelete:
		removeFromQueue()
		return nil
	}

	switch event.Rune() {
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case 'l':
		currentTrackIndex := queueList.GetCurrentItem()
		currentTrackName, currentTrackID := queueList.GetItemText(currentTrackIndex)
		playTrack(currentTrackIndex, currentTrackName, currentTrackID, 0)
		return nil

	case 'x':
		removeFromQueue()
		return nil
	}

	return event
}
