package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/itchyny/gojq"
	"github.com/rivo/tview"
)

var playlistList, playlistTracks *tview.List

func initPlaylistPage() {
	fillPlaylists()
	if playlistList.GetItemCount() > 0 {
		main, secondary := playlistList.GetItemText(0)
		showPlaylist(0, main, secondary, 0)
	}
}

func fillPlaylists() {
	rows := queryPlaylists()
	for rows.Next() {
		var playlistID int
		var name string

		rows.Scan(&playlistID, &name)
		playlistList.AddItem(name, fmt.Sprint(playlistID), 0, nil)
	}
}

func showPlaylist(_ int, playlistName, playlistIDString string, _ rune) {
	playlistTracks.Clear()

	playlistTracksJSON, err := os.ReadFile(playlistDirectory + playlistName + ".json")
	if err != nil {
		printError(err)
	}

	query, err := gojq.Parse(`."subsonic-response".playlist.entry[]`)
	if err != nil {
		printError(err)
	}

	var resJSON map[string]interface{}
	json.Unmarshal(playlistTracksJSON, &resJSON)

	iter := query.Run(resJSON)

	for trackMap, ok := iter.Next(); ok; trackMap, ok = iter.Next() {
		track := Track{}
		trackJSON, _ := json.Marshal(trackMap)
		json.Unmarshal(trackJSON, &track)
		playlistTracks.AddItem(fmt.Sprintf("%s%s - %s", markTrack(track.ID), track.Artist, track.Title), fmt.Sprint(track.ID), 0, nil)
	}
}

func playlistInputHandler(event *tcell.EventKey) *tcell.EventKey {
	focused := app.GetFocus()

	switch event.Rune() {
	case 'h': // override 'h' to KeyLeft
		event = tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone)

	case 'l': // override 'l' to KeyRight
		event = tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone)

	case ' ':
		if focused == playlistList {
			listEnqueueSublist(playlistList, playlistTracks, false)
		} else if focused == playlistTracks {
			listEnqueueTrack(playlistTracks, false)
		}
		return nil

	case 'o':
		findInLibrary(playlistTracks)
		return nil
	}

	switch event.Key() {
	case tcell.KeyEnter:
		if focused == playlistList {
			listEnqueueSublist(playlistList, playlistTracks, true)
		} else if focused == playlistTracks {
			listEnqueueTrack(playlistTracks, true)
		}
		return nil

	case tcell.KeyLeft:
		if focused == playlistTracks {
			setAndSaveFocus(playlistList)
		}
		return nil

	case tcell.KeyRight:
		if focused == playlistList {
			setAndSaveFocus(playlistTracks)
		} else if focused == playlistTracks {
			listEnqueueTrack(playlistTracks, true)
		}
		return nil
	}

	return event
}
