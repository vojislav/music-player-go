package main

import (
	"fmt"
	"time"

	"github.com/faiface/beep/speaker"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var app = tview.NewApplication()
var pages = tview.NewPages()
var artistList, albumList, trackList *tview.List
var loadingPopup tview.Primitive
var currentTrackText, loadingTextBox, loginStatus *tview.TextView

var popup = func(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)
}

func initView() {
	// login page
	loginForm = tview.NewForm().
		AddInputField("Username", "", 20, nil, func(username string) { config.Username = username }).
		AddPasswordField("Password", "", 20, '*', setToken).
		AddButton("Login", loginUser)
	loginForm.SetBorder(true)

	loginStatus = tview.NewTextView()
	loginStatus.SetBorder(true)

	loginGrid := tview.NewGrid().
		SetColumns(0, 40, 0).
		SetRows(0, 10, 3, 0).
		AddItem(loginForm, 1, 1, 1, 1, 0, 0, true).
		AddItem(loginStatus, 2, 1, 1, 1, 0, 0, false)

	loginGrid.SetBorderPadding(0, 0, 0, 0)

	pages.AddPage("login", loginGrid, true, false)

	// load library page
	loadingTextBox = tview.NewTextView()
	loadingTextBox.SetBorder(true)

	loadingPopup = popup(loadingTextBox, 40, 10)
	pages.AddPage("loading library", loadingPopup, true, false)

	// library page
	artistList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)
	artistList.SetBorder(true).SetTitle("Artist")

	albumList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)
	albumList.SetBorder(true).SetTitle("Albums")

	trackList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)
	trackList.SetBorder(true).SetTitle("Tracks")

	currentTrackText = tview.NewTextView()

	libraryFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(artistList, 0, 1, true).
			AddItem(albumList, 0, 1, false).
			AddItem(trackList, 0, 1, false), 0, 1, true).
		AddItem(currentTrackText, 3, 1, false)

	pages.AddPage("library", libraryFlex, true, true)

	// key handlers
	libraryFlex.SetInputCapture(libraryKeyHandler)
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

	artistList.SetChangedFunc(fillAlbumsList)
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

	albumList.SetChangedFunc(fillTracksList)
}

func fillTracksList(_ int, albumName, albumIDString string, _ rune) {
	trackList.Clear()

	albumID := toInt(albumIDString)

	rows := queryTracks(int(albumID))
	for rows.Next() {
		var trackID, albumID, artistID, track, duration int
		var title string
		rows.Scan(&trackID, &title, &albumID, &artistID, &track, &duration)

		trackList.AddItem(fmt.Sprintf("%d. %s", track, title), fmt.Sprint(trackID), 0, nil)
	}

	trackList.SetSelectedFunc(playTrack)
}

func libraryKeyHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
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
			currentTrackIndex := trackList.GetCurrentItem()
			currentTrackName, currentTrackID := trackList.GetItemText(currentTrackIndex)
			playTrack(currentTrackIndex, currentTrackName, currentTrackID, 0)
		}
		return nil
	}

	switch event.Rune() {
	case 'q':
		app.Stop()
	case 'p':
		speaker.Lock()
		playerCtrl.Paused = !playerCtrl.Paused
		if playerCtrl.Paused {
			killTicker <- true
		} else {
			go trackTime()
		}
		speaker.Unlock()

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
			playTrack(currentTrackIndex, currentTrackName, currentTrackID, 0)
		}

		return nil

	}

	return event
}

func updateCurrentTrackText() {
	currentTrackText.Clear()
	var status string
	if playerCtrl.Paused {
		status = "Paused"
	} else {
		status = "Playing"
	}

	currentTime := getTimeString(currentTrack.stream.Position() / sr.N(time.Second))
	totalTime := getTimeString(currentTrack.stream.Len() / sr.N(time.Second))

	// fmt.Fprintf(currentTrack, "%s: %s - %s", status, queue.tracks[0].artist, queue.tracks[0].title)
	if currentTime == totalTime {
		currentTrackText.Clear()
	} else {
		fmt.Fprintf(currentTrackText, "%s: %s - %s\t%s / %s\n", status, currentTrack.artist, currentTrack.title, currentTime, totalTime)
	}
	app.Draw()
}

func getTimeString(time int) string {
	minutes := fmt.Sprint(time / 60)
	seconds := fmt.Sprint(time % 60)

	if len(seconds) == 1 {
		seconds = "0" + seconds
	}

	return fmt.Sprint(minutes, ":", seconds)
}
