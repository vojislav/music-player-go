package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/itchyny/gojq"
	"github.com/rivo/tview"
)

var playlistList, playlistTracks *tview.List

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
		playlistTracks.AddItem(fmt.Sprintf("%s%s - %s", markTrack(track.ID), track.Artist, track.Title), fmt.Sprint(track.ID), 0, nil)
	}
}

func playlistInputHandler(event *tcell.EventKey) *tcell.EventKey {
	focused := app.GetFocus()

	switch event.Key() {
	case tcell.KeyEnter:
		if focused == playlistTracks {
			currentTrackIndex := playlistTracks.GetCurrentItem()
			_, currentTrackID := playlistTracks.GetItemText(currentTrackIndex)
			go downloadCallback(currentTrackID, addToQueueAndPlay)
			playlistTracks.SetCurrentItem(currentTrackIndex + 1)
			markList(playlistTracks, currentTrackIndex)
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
			currentTrackIndex := playlistTracks.GetCurrentItem()
			_, currentTrackID := playlistTracks.GetItemText(currentTrackIndex)
			go downloadCallback(currentTrackID, addToQueueAndPlay)
			playlistTracks.SetCurrentItem(currentTrackIndex + 1)
			markList(playlistTracks, currentTrackIndex)
		}
		return nil
	}

	switch event.Rune() {
	case 'h':
		if focused == playlistTracks {
			setAndSaveFocus(playlistList)
		}
		return nil
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case 'l':
		if focused == playlistList {
			setAndSaveFocus(playlistTracks)
		} else if focused == playlistTracks {
			currentTrackIndex := playlistTracks.GetCurrentItem()
			_, currentTrackID := playlistTracks.GetItemText(currentTrackIndex)
			go downloadCallback(currentTrackID, addToQueueAndPlay)
			playlistTracks.SetCurrentItem(currentTrackIndex + 1)
			markList(playlistTracks, currentTrackIndex)
		}
		return nil

	case ' ':
		if focused == playlistTracks {
			currentTrackIndex := playlistTracks.GetCurrentItem()
			_, currentTrackID := playlistTracks.GetItemText(currentTrackIndex)
			go downloadCallback(currentTrackID, addToQueue)
			playlistTracks.SetCurrentItem(currentTrackIndex + 1)
			markList(playlistTracks, currentTrackIndex)
		}
		return nil

	case 'o':
		findInLibrary(playlistTracks)
		return nil
	}

	return event
}
