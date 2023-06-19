package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/itchyny/gojq"
)

func initPlaylistPage() {
	fillPlaylists()
	main, secondary := playlistList.GetItemText(0)
	showPlaylist(0, main, secondary, 0)
}

func fillPlaylists() {
	for _, playlist := range getPlaylists() {
		playlistList.AddItem(playlist.Name, fmt.Sprint(playlist.ID), 0, nil)
	}
}

func showPlaylist(_ int, playlistName, playlistIDString string, _ rune) {
	playlistTracks.Clear()

	playlistTracksJSON, err := os.ReadFile(playlistDirectory + playlistName + ".json")
	if err != nil {
		log.Fatal(err)
	}

	query, err := gojq.Parse(`."subsonic-response".playlist.entry[]`)
	if err != nil {
		log.Fatal(err)
	}

	var resJSON map[string]interface{}
	json.Unmarshal(playlistTracksJSON, &resJSON)

	iter := query.Run(resJSON)

	for trackMap, ok := iter.Next(); ok; trackMap, ok = iter.Next() {
		track := Track{}
		trackJSON, _ := json.Marshal(trackMap)
		json.Unmarshal(trackJSON, &track)
		playlistTracks.AddItem(fmt.Sprintf("%s - %s", track.Artist, track.Title), fmt.Sprint(track.ID), 0, nil)
	}
}

func playlistInputHandler(event *tcell.EventKey) *tcell.EventKey {
	focused := app.GetFocus()
	if focused == loginForm || focused == searchInput {
		return event
	}

	switch event.Key() {
	case tcell.KeyEnter:
		focused := app.GetFocus()
		if focused == playlistTracks {
			currentTrackIndex := playlistTracks.GetCurrentItem()
			_, currentTrackID := playlistTracks.GetItemText(currentTrackIndex)
			go downloadCallback(currentTrackID, addToQueueAndPlay)
			playlistTracks.SetCurrentItem(currentTrackIndex + 1)
		}
		return nil

	case tcell.KeyLeft:
		focused = app.GetFocus()
		if focused == playlistTracks {
			app.SetFocus(playlistList)
		}
		return nil

	case tcell.KeyRight:
		focused = app.GetFocus()
		if focused == playlistList {
			app.SetFocus(playlistTracks)
		} else if focused == playlistTracks {
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
		focused = app.GetFocus()
		if focused == playlistTracks {
			app.SetFocus(playlistList)
		}
		return nil

	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)

	case 'l':
		focused = app.GetFocus()
		if focused == playlistList {
			app.SetFocus(playlistTracks)
		} else if focused == playlistTracks {
			currentTrackIndex := playlistTracks.GetCurrentItem()
			_, currentTrackID := playlistTracks.GetItemText(currentTrackIndex)
			go downloadCallback(currentTrackID, addToQueueAndPlay)
			playlistTracks.SetCurrentItem(currentTrackIndex + 1)
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

	case '/':
		searchIndexes = nil
		searchCurrentIndex = 0

		switch app.GetFocus() {
		case playlistList:
			searchList = playlistList
		case playlistTracks:
			searchList = playlistTracks
		}

		app.SetFocus(bottomPanel)
		bottomPanel.SwitchToPage("search")
		return nil

	case ' ':
		focused := app.GetFocus()
		if focused == playlistTracks {
			currentTrackIndex := playlistTracks.GetCurrentItem()
			_, currentTrackID := playlistTracks.GetItemText(currentTrackIndex)
			go downloadCallback(currentTrackID, addToQueue)
			playlistTracks.SetCurrentItem(currentTrackIndex + 1)
		}
		return nil
	}

	return event
}
