package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var app = tview.NewApplication()
var pages = tview.NewPages()
var mainPanel = tview.NewPages()
var bottomPanel = tview.NewPages()
var artistList, albumList, trackList, queueList *tview.List
var searchList *tview.List
var loadingPopup tview.Primitive
var currentTrackText, downloadProgressText, loadingTextBox, loginStatus *tview.TextView
var searchInput *tview.InputField
var playlistList, playlistTracks *tview.List
var loginGrid *tview.Grid

var searchIndexes []int
var searchCurrentIndex int

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

	loginGrid = tview.NewGrid().
		SetColumns(0, 40, 0).
		SetRows(0, 10, 3, 0).
		AddItem(loginForm, 1, 1, 1, 1, 0, 0, true).
		AddItem(loginStatus, 2, 1, 1, 1, 0, 0, false)

	loginGrid.SetBorderPadding(0, 0, 0, 0)

	pages.AddPage("login", loginGrid, true, false)

	// loading page
	loadingTextBox = tview.NewTextView()
	loadingTextBox.SetBorder(true)

	loadingPopup = popup(loadingTextBox, 40, 10)
	pages.AddPage("loading library", loadingPopup, true, false)

	// main panel
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainPanel, 0, 1, true).
		AddItem(bottomPanel, 3, 1, false)

	pages.AddPage("main", mainFlex, true, false)

	// bottom panel
	currentTrackText = tview.NewTextView()
	currentTrackText.SetBorder(true)

	searchInput = tview.NewInputField().
		SetLabel("Search: ")

	downloadProgressText = tview.NewTextView()
	downloadProgressText.SetBorder(true)

	bottomPanel.AddPage("current track info", tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(currentTrackText, 0, 1, false).
		AddItem(downloadProgressText, 10, 1, false), true, true)

	bottomPanel.AddPage("search", searchInput, true, false)

	// library page
	artistList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	artistList.SetBorder(true).SetTitle("Artist")
	artistList.SetChangedFunc(fillAlbumsList)

	albumList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	albumList.SetBorder(true).SetTitle("Albums")
	albumList.SetChangedFunc(fillTracksList)

	trackList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	trackList.SetBorder(true).SetTitle("Tracks")

	libraryFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(artistList, 0, 1, true).
			AddItem(albumList, 0, 1, false).
			AddItem(trackList, 0, 1, false), 0, 1, true)

	mainPanel.AddPage("library", libraryFlex, true, true)

	// queue page
	queueList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	queueList.SetBorder(true).SetTitle("Queue")
	queueList.SetSelectedFunc(playTrack)
	queueFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queueList, 0, 1, true)

	mainPanel.AddPage("queue", queueFlex, true, false)

	// playlist
	playlistList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	playlistList.SetBorder(true).SetTitle("Playlists")
	playlistList.SetChangedFunc(showPlaylist)

	playlistTracks = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	playlistTracks.SetBorder(true)
	playlistFlex := tview.NewFlex().
		AddItem(playlistList, 0, 1, true).
		AddItem(playlistTracks, 0, 3, false)

	mainPanel.AddPage("playlists", playlistFlex, true, false)

	// key handlers
	app.SetInputCapture(appInputHandler)
	libraryFlex.SetInputCapture(libraryInputHandler)
	queueFlex.SetInputCapture(queueInputHandler)
	playlistFlex.SetInputCapture(playlistInputHandler)
	searchInput.SetInputCapture(searchInputHandler)
}

func updateCurrentTrackText() {
	currentTrackText.Clear()
	if currentTrack.stream == nil {
		return
	}

	var status string
	if playerCtrl.Paused {
		status = "Paused"
	} else {
		status = "Playing"
	}

	currentTime := getTimeString(currentTrack.stream.Position() / sr.N(time.Second))
	totalTime := getTimeString(currentTrack.stream.Len() / sr.N(time.Second))

	// fmt.Fprintf(currentTrack, "%s: %s - %s", status, queue.tracks[0].artist, queue.tracks[0].title)
	if currentTrack.stream.Position() == currentTrack.stream.Len() {
		currentTrackText.Clear()
		nextTrack()
	} else {
		fmt.Fprintf(currentTrackText, "%s: %s - %s\t%s / %s\tQueue position: %d / %d", status, currentTrack.Artist, currentTrack.Title, currentTime, totalTime, queuePosition+1, queueList.GetItemCount())
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

func appInputHandler(event *tcell.EventKey) *tcell.EventKey {
	focused := app.GetFocus()
	if focused == loginGrid || focused == searchInput {
		return event
	}

	switch event.Rune() {
	case '1':
		mainPanel.SwitchToPage("queue")
		return nil
	case '2':
		mainPanel.SwitchToPage("library")
		return nil
	case '3':
		mainPanel.SwitchToPage("playlists")
		return nil

	}

	return event
}

func gotoLoadingPage() {
	pages.SwitchToPage("loading library")
	go func() {
		createDatabase()

		fmt.Fprint(loadingTextBox, "Loading...")
		app.Draw()

		loadDatabase()

		pages.SwitchToPage("main")
		initLibraryPage()
		app.Draw()
	}()
}
