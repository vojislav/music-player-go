package main

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var artistList, albumList, trackList, queueList *tview.List

func gotoLibraryPage() {
	pages.SwitchToPage("main")
	initLibraryPage()
}

func initLibraryPage() {
	fillArtistList()

	main, secondary := artistList.GetItemText(0)
	fillAlbumsList(0, main, secondary, 0)

	main, secondary = albumList.GetItemText(0)
	fillTracksList(0, main, secondary, 0)
}

func fillArtistList() {
	rows := queryArtists()
	for rows.Next() {
		var artistID int
		var name string

		rows.Scan(&artistID, &name)
		artistList.AddItem(name, fmt.Sprint(artistID), 0, nil)
	}
}

func fillAlbumsList(_ int, artistName, artistIDString string, _ rune) {
	albumList.Clear()

	artistID := toInt(artistIDString)
	rows := queryAlbums(int(artistID))
	for rows.Next() {
		var albumID, artistID, year int
		var name string
		rows.Scan(&albumID, &artistID, &name, &year)

		albumList.AddItem(fmt.Sprintf("(%d) %s", year, name), fmt.Sprint(albumID), 0, nil)
	}
}

func fillTracksList(_ int, albumName, albumIDString string, _ rune) {
	trackList.Clear()

	albumID := toInt(albumIDString)

	rows := queryAlbumTracks(albumID)
	for rows.Next() {
		var title, album, artist, genre, suffix, albumID, artistID string
		var trackID, track, disc, year, size, duration, bitrate int
		rows.Scan(&trackID, &title, &album, &artist, &track, &year, &genre, &size, &suffix, &duration, &bitrate, &disc, &albumID, &artistID)

		var trackText string

		alreadyInQueue := markTrack(strconv.FormatInt(int64(trackID), 10))

		trackText = alreadyInQueue

		if track != 0 {
			trackText += fmt.Sprintf("%d. ", track)
		}

		trackText += title

		trackList.AddItem(trackText, fmt.Sprint(trackID), 0, nil)
	}
}

// finds location of currently highlighted track in library.
// should only be used in queue or playlist.
// TODO: error handling;
func findInLibrary(list *tview.List) {
	focused := app.GetFocus()
	if focused != list || list.GetItemCount() == 0 {
		return
	}

	idx := list.GetCurrentItem()
	_, trackID := list.GetItemText(idx)
	var artist, album string
	queryArtistAndAlbum(toInt(trackID)).Scan(&artist, &album)

	artists := artistList.FindItems(artist, "", true, true)
	if len(artists) == 0 {
		return
	}
	artistList.SetCurrentItem(artists[0])

	albums := albumList.FindItems(album, "", true, true)
	if len(albums) == 0 {
		return
	}
	albumList.SetCurrentItem(albums[0])

	tracks := trackList.FindItems("", trackID, true, true)
	if len(tracks) == 0 {
		return
	}
	trackList.SetCurrentItem(tracks[0])

	mainPanel.SwitchToPage("library")
	setAndSaveFocus(trackList)
}

// enqueues (and downloads if necessary) a single track from current track list (either in "library" or in "playlists" views)
func listEnqueueTrack(list *tview.List, play bool) {
	currentTrackIndex := list.GetCurrentItem()
	currentTrackText, currentTrackID := list.GetItemText(currentTrackIndex)
	downloadAndEnqueueTrack(currentTrackText, currentTrackID, play)
	list.SetCurrentItem(currentTrackIndex + 1)
	markList(list, currentTrackIndex)
}

// enqueues (and downloads if necessary) all tracks from current album or playlist
func listEnqueueSublist(list *tview.List, sublist *tview.List, play bool) {
	currentListIndex := list.GetCurrentItem()

	for idx := 0; idx < sublist.GetItemCount(); idx++ {
		trackText, trackID := sublist.GetItemText(idx)
		downloadAndEnqueueTrack(trackText, trackID, play && idx == 0)
		markList(sublist, idx)
	}

	list.SetCurrentItem(currentListIndex + 1)
}

// adds dummy placeholder to queue, and downloads track and puts it on that place
func downloadAndEnqueueTrack(trackText string, trackID string, play bool) {
	queueList.AddItem(trackNotDownloadedMarker+trackText, trackID, 0, nil)
	idx := queueList.GetItemCount() - 1

	if play {
		go sendToPriorityDownloadQueue(trackID, idx)
	} else {
		go sendToDownloadQueue(trackID, idx)
	}
}

func libraryInputHandler(event *tcell.EventKey) *tcell.EventKey {
	focused := app.GetFocus()

	switch event.Rune() {
	case 'h': // override 'h' to KeyLeft
		event = tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone)

	case 'l': // override 'l' to KeyRight
		event = tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone)

	case ' ': // space key
		if focused == trackList {
			listEnqueueTrack(trackList, false)
		} else if focused == albumList {
			listEnqueueSublist(albumList, trackList, false)
		}
		return nil
	}

	switch event.Key() {
	case tcell.KeyEnter:
		if focused == trackList {
			listEnqueueTrack(trackList, true)
		} else if focused == albumList {
			listEnqueueSublist(albumList, trackList, true)
		}
		return nil

	case tcell.KeyLeft:
		if focused == albumList {
			setAndSaveFocus(artistList)
		} else if focused == trackList {
			setAndSaveFocus(albumList)
		}
		return nil

	case tcell.KeyRight:
		if focused == artistList {
			setAndSaveFocus(albumList)
		} else if focused == albumList {
			setAndSaveFocus(trackList)
		} else if focused == trackList {
			listEnqueueTrack(trackList, true)
		}
		return nil
	}

	return event
}
