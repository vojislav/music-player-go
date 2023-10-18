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
		var title, album, artist, genre, suffix, albumID, artistID string
		var trackID, track, year, size, duration, bitrate int
		rows.Scan(&trackID, &title, &album, &artist, &track, &year, &genre, &size, &suffix, &duration, &bitrate, &albumID, &artistID)

		trackList.AddItem(fmt.Sprintf("%d. %s", track, title), fmt.Sprint(trackID), 0, nil)
	}
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
	app.SetFocus(trackList)
}

func libraryInputHandler(event *tcell.EventKey) *tcell.EventKey {
	focused := app.GetFocus()

	switch event.Key() {
	case tcell.KeyEnter:
		if focused == trackList {
			currentTrackIndex := trackList.GetCurrentItem()
			_, currentTrackID := trackList.GetItemText(currentTrackIndex)
			go downloadCallback(currentTrackID, addToQueueAndPlay)
			trackList.SetCurrentItem(currentTrackIndex + 1)
		}
		return nil

	case tcell.KeyLeft:
		if focused == albumList {
			app.SetFocus(artistList)
		} else if focused == trackList {
			app.SetFocus(albumList)
		}
		return nil

	case tcell.KeyRight:
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
	}

	switch event.Rune() {
	case 'h':
		if focused == albumList {
			app.SetFocus(artistList)
		} else if focused == trackList {
			app.SetFocus(albumList)
		}
		return nil

	case 'l':
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

	case ' ':
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
	}

	return event
}
