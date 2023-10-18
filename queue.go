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

func removeFromQueue() {
	highlightedTrackIndex := queueList.GetCurrentItem()
	if highlightedTrackIndex < queuePosition {
		queuePosition -= 1
	} else if highlightedTrackIndex == queuePosition {
		stopTrack()
	}
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

// finds location of currently highlighted track in library.
// should only be used in queue or playlist.
// TODO: error handling; TODO: what if files aren't downloaded? getTags breaks
func findInLibrary(list *tview.List) {
	focused := app.GetFocus()
	if focused != list || list.GetItemCount() == 0 {
		return
	}

	idx := list.GetCurrentItem()
	_, secondary := list.GetItemText(idx)
	tags := getTags(getTrackPath(secondary))

	artists := artistList.FindItems(tags.Artist(), "", true, true)
	if len(artists) == 0 {
		return
	}
	artistList.SetCurrentItem(artists[0])

	albums := albumList.FindItems(tags.Album(), "", true, true)
	if len(albums) == 0 {
		return
	}
	albumList.SetCurrentItem(albums[0])

	tracks := trackList.FindItems("", secondary, true, true)
	if len(tracks) == 0 {
		return
	}
	trackList.SetCurrentItem(tracks[0])

	mainPanel.SwitchToPage("library")
	app.SetFocus(trackList)
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
