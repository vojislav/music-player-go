package main

import (
	"fmt"

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
		var trackID, albumID, artistID, track, duration int
		var title string
		rows.Scan(&trackID, &title, &albumID, &artistID, &track, &duration)

		trackList.AddItem(fmt.Sprintf("%d. %s", track, title), fmt.Sprint(trackID), 0, nil)
	}
}

func libraryInputHandler(event *tcell.EventKey) *tcell.EventKey {
	focused := app.GetFocus()
	if focused == loginForm || focused == searchInput {
		return event
	}

	switch event.Key() {
	case tcell.KeyEnter:
		focused := app.GetFocus()
		if focused == trackList {
			currentTrackIndex := trackList.GetCurrentItem()
			_, currentTrackID := trackList.GetItemText(currentTrackIndex)
			go downloadCallback(currentTrackID, addToQueueAndPlay)
			trackList.SetCurrentItem(currentTrackIndex + 1)
		}
		return nil

	case tcell.KeyLeft:
		focused := app.GetFocus()
		if focused == albumList {
			app.SetFocus(artistList)
		} else if focused == trackList {
			app.SetFocus(albumList)
		}
		return nil

	case tcell.KeyRight:
		focused := app.GetFocus()
		if focused == artistList {
			app.SetFocus(albumList)
		} else if focused == albumList {
			app.SetFocus(trackList)
		} else if focused == trackList {
			currentTrackIndex := playlistTracks.GetCurrentItem()
			_, currentTrackID := playlistTracks.GetItemText(currentTrackIndex)
			go downloadCallback(currentTrackID, addToQueueAndPlay)
			playlistTracks.SetCurrentItem(currentTrackIndex + 1)
		}
		return nil
	}

	switch event.Rune() {
	case 'q':
		app.Stop()
		return nil
	case 'p':
		playPause()
		return nil

	case '>':
		nextTrack()
		return nil
	case '<':
		previousTrack()
		return nil
	case 's':
		stopTrack()
		return nil

	case 'h':
		focused := app.GetFocus()
		if focused == albumList {
			app.SetFocus(artistList)
		} else if focused == trackList {
			app.SetFocus(albumList)
		}
		return nil

	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case 'l':
		focused := app.GetFocus()
		if focused == artistList {
			app.SetFocus(albumList)
		} else if focused == albumList {
			app.SetFocus(trackList)
		} else if focused == trackList {
			currentTrackIndex := trackList.GetCurrentItem()
			_, currentTrackID := trackList.GetItemText(currentTrackIndex)
			go downloadCallback(currentTrackID, addToQueueAndPlay)
			trackList.SetCurrentItem(currentTrackIndex + 1)
		}
		return nil

	case 'g':
		return tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone)

	case 'G':
		return tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone)

	case 'n':
		nextSearchResult()
		return nil
	case 'N':
		previousSearchResult()
		return nil

	case ' ':
		focused := app.GetFocus()
		if focused == trackList {
			currentTrackIndex := trackList.GetCurrentItem()
			_, currentTrackID := trackList.GetItemText(currentTrackIndex)
			go downloadCallback(currentTrackID, addToQueue)
			trackList.SetCurrentItem(currentTrackIndex + 1)
		} else if focused == albumList {
			currentAlbumIndex := albumList.GetCurrentItem()
			_, currentAlbumID := trackList.GetItemText(currentAlbumIndex)
			addAlbumToQueue(currentAlbumID)
		}
		return nil

	case '/':
		searchIndexes = nil
		searchCurrentIndex = 0

		switch app.GetFocus() {
		case artistList:
			searchList = artistList
		case albumList:
			searchList = albumList
		case trackList:
			searchList = trackList
		}
		app.SetFocus(bottomPanel)
		bottomPanel.SwitchToPage("search")
		return nil
	}

	return event
}
