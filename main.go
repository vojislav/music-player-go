package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/gdamore/tcell/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rivo/tview"
)

var sr = beep.SampleRate(44100)
var pages = tview.NewPages()

var artistList, albumList, trackList *tview.List

func listArtists() {
	rows := queryArtists()
	for rows.Next() {
		var artistID int
		var name string

		rows.Scan(&artistID, &name)
		artistList.AddItem(name, fmt.Sprint(artistID), 0, nil)
	}

	artistList.SetChangedFunc(listAlbums)
}

func listAlbums(_ int, artistName, artistIDString string, _ rune) {
	albumList.Clear()

	artistID, _ := strconv.ParseInt(artistIDString, 10, 32)
	rows := queryAlbums(int(artistID))
	for rows.Next() {
		var albumID, artistID, year int
		var name string
		rows.Scan(&albumID, &artistID, &name, &year)

		albumList.AddItem(fmt.Sprintf("(%d) %s", year, name), fmt.Sprint(albumID), 0, nil)
	}

	albumList.SetChangedFunc(listTracks)
}

func listTracks(_ int, albumName, albumIDString string, _ rune) {
	trackList.Clear()

	albumID, _ := strconv.ParseInt(albumIDString, 10, 32)

	rows := queryTracks(int(albumID))
	for rows.Next() {
		var trackID, albumID, artistID, track, duration int
		var title string
		rows.Scan(&trackID, &title, &albumID, &artistID, &track, &duration)

		trackList.AddItem(fmt.Sprintf("%d. %s", track, title), fmt.Sprint(trackID), 0, nil)
	}

	trackList.SetSelectedFunc(stream)
}

func updateCurrentTrack() {
	currentTrack.Clear()
	var status string
	if playerCtrl.Paused {
		status = "Paused"
	} else {
		status = "Playing"
	}

	fmt.Fprintf(currentTrack, "%s: %s - %s", status, queue.tracks[0].artist, queue.tracks[0].title)
}

func handleKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'p':
		speaker.Lock()
		playerCtrl.Paused = !playerCtrl.Paused
		updateCurrentTrack()
		speaker.Unlock()

	case 'q':
		app.Stop()

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
			currentTrackName, currentTrackID := trackList.GetItemText(currentTrackIndex)
			stream(currentTrackIndex, currentTrackName, currentTrackID, 0)
		}

		return nil

	}

	return event
}

var queue Queue
var app = tview.NewApplication()
var currentTrack *tview.TextView
var playerCtrl *beep.Ctrl

func main() {
	speaker.Init(sr, sr.N(time.Second/10))
	playerCtrl = &beep.Ctrl{Streamer: &queue, Paused: false}
	speaker.Play(playerCtrl)

	artistList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)
	artistList.SetBorder(true).SetTitle("Artist")

	albumList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)
	albumList.SetBorder(true).SetTitle("Albums")

	trackList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)
	trackList.SetBorder(true).SetTitle("Tracks")

	// artistList()

	listArtists()

	main, secondary := artistList.GetItemText(0)
	listAlbums(0, main, secondary, 0)

	main, secondary = albumList.GetItemText(0)
	listTracks(0, main, secondary, 0)

	currentTrack = tview.NewTextView()
	currentTrack.SetBorder(true)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(artistList, 0, 1, true).
			AddItem(albumList, 0, 1, false).
			AddItem(trackList, 0, 1, false), 0, 10, true).
		AddItem(currentTrack, 0, 1, false)

	pages.AddAndSwitchToPage("main", flex, true)

	app.SetInputCapture(handleKeys)

	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}
