package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// position in queue of currently played track [0-indexed]. Should only be set setQueuePosition()
var queuePosition = -1

// the way in which the current playing track is marked.
var currentTrackMarker = "[::u]"

// the way in which the tracks that are in queue are marked
var trackInQueueMarker = "[::b]"

// the way in which the tracks that are not downloaded are marked
var trackNotDownloadedMarker = "[::d]"

func addToQueue(_ int, _, trackID string, _ rune) {
	tags := getTags(getTrackPath(trackID))
	itemText := fmt.Sprintf("%s - %s", tags.Artist(), tags.Title())
	queueList.AddItem(itemText, trackID, 0, nil)
}

func addToQueueAndPlay(_ int, _, trackID string, _ rune) {
	addToQueue(0, "", trackID, 0)

	setQueuePosition(queueList.GetItemCount() - 1)
	playTrack(queuePosition, "", trackID, 0)
}

// removes indicator that track is in queue. This fixes the situation where
// trackList/playlistTracks are not refreshed after removing track from queue
func removeInQueueMarks(list *tview.List, trackID string) {
	items := list.FindItems("", trackID, true, true)
	if len(items) == 0 { // sanity check, should always be false
		return
	}

	trackText, _ := list.GetItemText(items[0])
	if trackExists(trackID) { // if track didn't exist prior to adding to queue (and now exists), "not downloaded" mark should be removed
		trackText = strings.Replace(trackText, trackNotDownloadedMarker, "", 1)
	}
	list.SetItemText(items[0], strings.Replace(trackText, trackInQueueMarker, "", 1), trackID)
}

// if track is removed from queue list, search indexes are refreshed
func refreshSearchIndexes(trackIndex int) {
	var searchIndexesNew []int
	_, loc := binary_search(searchIndexes, trackIndex)

	searchIndexesNew = searchIndexes[:loc]
	for i := loc + 1; i < len(searchIndexes); i++ {
		searchIndexesNew = append(searchIndexesNew, searchIndexes[i]-1)
	}
	searchIndexes = searchIndexesNew
}

func removeFromQueue() {
	highlightedTrackIndex := queueList.GetCurrentItem()
	if highlightedTrackIndex < queuePosition {
		queuePosition -= 1
	} else if highlightedTrackIndex == queuePosition {
		stopTrack()
	}

	_, trackID := queueList.GetItemText(highlightedTrackIndex)
	removeInQueueMarks(trackList, trackID)
	removeInQueueMarks(playlistTracks, trackID)

	refreshSearchIndexes(highlightedTrackIndex)

	queueList.RemoveItem(highlightedTrackIndex)
}

func addAlbumToQueue(albumID string) {

}

// should only be called when current song is changed
func setQueuePosition(newQueuePosition int) {
	previousQueuePosition := queuePosition

	// if there was a previous song, remove the underline
	if previousQueuePosition != -1 {
		previousTrackText, trackID := queueList.GetItemText(previousQueuePosition)
		previousTrackText = strings.Replace(previousTrackText, currentTrackMarker, "", 1)
		queueList.SetItemText(previousQueuePosition, previousTrackText, trackID)
	}

	// if there is a next song, add underline
	if newQueuePosition != -1 {
		currentText, trackID := queueList.GetItemText(newQueuePosition)
		queueList.SetItemText(newQueuePosition,
			fmt.Sprintf("%s%s", currentTrackMarker, currentText),
			trackID)
	}

	// set new queuePosition
	queuePosition = newQueuePosition
}

// play the currently highlighted track in queue. No-op if queue is empty
func queuePlayHighlighted() {
	if queueList.GetItemCount() == 0 {
		return
	}
	currentTrackIndex := queueList.GetCurrentItem()
	currentTrackName, currentTrackID := queueList.GetItemText(currentTrackIndex)
	playTrack(currentTrackIndex, currentTrackName, currentTrackID, 0)
}

// returns string which is used to "mark" tracks which are either:
// in queue (bold);
// not downloaded (dim)
func markTrack(trackID string) string {
	if !trackExists(trackID) { // track doesn't exist locally in .cache
		return trackNotDownloadedMarker
	}

	indices := queueList.FindItems("", trackID, true, true)
	if len(indices) >= 1 { // track exists in queue
		return trackInQueueMarker
	}

	return ""
}

// marks tracks added to queue from track lists such as playlists or library
// TODO: this can be stacked infinitely. What are the implications?
func markList(list *tview.List, idx int) {
	prim, sec := list.GetItemText(idx)
	list.SetItemText(idx, trackInQueueMarker+prim, sec)
}

func queueInputHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRight, tcell.KeyEnter:
		queuePlayHighlighted()
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
		queuePlayHighlighted()
		return nil

	case 'x':
		removeFromQueue()
		return nil
	case 'o':
		findInLibrary(queueList)
		return nil
	}

	return event
}
