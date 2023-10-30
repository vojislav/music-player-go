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
	found, loc := binary_search(searchIndexes, trackIndex)
	if !found { // removed track is not necessarily in searchIndexes
		return
	}

	searchIndexesNew = searchIndexes[:loc]
	for i := loc + 1; i < len(searchIndexes); i++ {
		searchIndexesNew = append(searchIndexesNew, searchIndexes[i]-1)
	}
	searchIndexes = searchIndexesNew
}

// removes numbering, track text and duration from queue lists
func removeAllTrackInfoFromQueue(index int) {
	cnt := queueNumberList.GetItemCount()
	queueList.RemoveItem(index)
	queueNumberList.RemoveItem(cnt - 1)
	queueLengthList.RemoveItem(index)
}

func removeFromQueue() {
	if queueList.GetItemCount() == 0 {
		return
	}

	downloadMutex.RLock()
	if len(downloadMap) > 0 { // if anything is downloading; don't remove from queue
		// TODO: notif info
		downloadMutex.RUnlock()
		return
	}
	downloadMutex.RUnlock()

	highlightedTrackIndex := queueList.GetCurrentItem()
	if highlightedTrackIndex < queuePosition {
		queuePosition -= 1
	} else if highlightedTrackIndex == queuePosition {
		requestStopTrack()
	}

	_, trackID := queueList.GetItemText(highlightedTrackIndex)
	removeInQueueMarks(trackList, trackID)
	removeInQueueMarks(playlistTracks, trackID)

	refreshSearchIndexes(highlightedTrackIndex)

	removeAllTrackInfoFromQueue(highlightedTrackIndex)
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
	requestPlayTrack(currentTrackIndex, currentTrackName, currentTrackID, 0)
}

// returns string which is used to "mark" tracks which are either:
// in queue (bold);
// not downloaded (dim)
func markTrack(trackID string) string {
	indices := queueList.FindItems("", trackID, true, true)
	if len(indices) >= 1 { // track exists in queue
		return trackInQueueMarker
	}

	if !trackExists(trackID) { // track doesn't exist locally in .cache
		return trackNotDownloadedMarker
	}

	return ""
}

// marks tracks added to queue from track lists such as playlists or library
// TODO: this can be stacked infinitely. What are the implications?
func markList(list *tview.List, idx int) {
	prim, sec := list.GetItemText(idx)
	list.SetItemText(idx, trackInQueueMarker+prim, sec)
}

// adds song to queue (including numbering and track duration)
func addAllTrackInfoToQueue(trackText string, trackID string, position int, trackDurationString string) {
	queueList.AddItem(trackText, trackID, 0, nil)
	queueNumberList.AddItem(fmt.Sprintf("%d.", position+1), "", 0, nil)
	queueLengthList.AddItem(fmt.Sprintf("[%s[]", trackDurationString), "", 0, nil)
}

// adds dummy placeholder to queue, and downloads track and puts it on that place
// play argument indicates whether track should be played immediately upon download
func downloadAndEnqueueTrack(trackID string, play bool) {
	var artist, title string
	var duration int
	queryArtistAndTitleAndDuration(toInt(trackID)).Scan(&artist, &title, &duration)
	trackText := fmt.Sprintf("%s - %s", artist, title)

	// add placeholder and get its index
	idx := queueList.GetItemCount()
	addAllTrackInfoToQueue("_", trackID, idx, getTimeString(duration)) // this item must be added before playNext is set because of race condition

	if trackExists(trackID) { // no need to add it to download map if it exists
		queueList.SetItemText(idx, trackText, trackID)
		requestPlayIfNext(trackID, idx, play)
		return
	}

	if play {
		requestSetNext(idx)
	}

	// set placeholder text
	queueList.SetItemText(idx, trackNotDownloadedMarker+trackText, trackID)

	// request download
	downloadMutex.Lock()
	downloadMap[idx] = trackID
	downloadMutex.Unlock()

	downloadSemaphore <- true
}

// enqueues (and downloads if necessary) a single track from current track list (either in "library" or in "playlists" views)
func listEnqueueTrack(list *tview.List, play bool) {
	trackIndex := list.GetCurrentItem()
	_, trackID := list.GetItemText(trackIndex)
	downloadAndEnqueueTrack(trackID, play)
	list.SetCurrentItem(trackIndex + 1)
	markList(list, trackIndex)
}

// enqueues (and downloads if necessary) all tracks from current album or playlist
func listEnqueueSublist(list *tview.List, sublist *tview.List, play bool) {
	currentListIndex := list.GetCurrentItem()

	for idx := 0; idx < sublist.GetItemCount(); idx++ {
		_, trackID := sublist.GetItemText(idx)
		downloadAndEnqueueTrack(trackID, play && idx == 0)
		markList(sublist, idx)
	}

	list.SetCurrentItem(currentListIndex + 1)
}

// every time different track is highlighted also highlight its duration and number
func queueOnChange(index int, _ string, _ string, _ rune) {
	queueNumberList.SetCurrentItem(index)
	queueLengthList.SetCurrentItem(index)
}

func queueInputHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'l':
		event = tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)

	case 'x':
		event = tcell.NewEventKey(tcell.KeyDelete, 0, tcell.ModNone)
	case 'o':
		findInLibrary(queueList)
		return nil
	}

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

	return event
}
