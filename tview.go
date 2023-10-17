package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var app = tview.NewApplication()
var pages = tview.NewPages()
var mainPanel = tview.NewPages()
var bottomPanel = tview.NewPages()
var loadingPopup tview.Primitive
var currentTrackText, downloadProgressText, loadingTextBox, loginStatus, trackInfoTextBox,
	lyricsTextBox, nowPlayingTrackTextBox, nowPlayingTimeTextBox, progressBar *tview.TextView
var nowPlayingCover *tview.Image
var loginGrid *tview.Grid
var libraryFlex, queueFlex, playlistFlex, nowPlayingFlex *tview.Flex

var popup = func(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)
}

// generic handler used for track manipulation in [library, playlist, queue]
func trackInputHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case 'g':
		return tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone)
	case 'G':
		return tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone)
	case 'J':
		return tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone)
	case 'K':
		return tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone)

	case 'p':
		togglePlay()
		return nil
	case 's':
		stopTrack()
		return nil
	case '>':
		nextTrack()
		return nil
	case '<':
		previousTrack()
		return nil

	case '/':
		searchIndexes = nil
		searchCurrentIndex = 0

		searchList = app.GetFocus().(*tview.List)
		app.SetFocus(bottomPanel)
		bottomPanel.SwitchToPage("search")
		return nil
	case 'n':
		nextSearchResult()
		return nil
	case 'N':
		previousSearchResult()
		return nil

	case '=':
		changeVolume(volumeStep)
		return nil
	case '-':
		changeVolume(-volumeStep)
		return nil
	case 'm':
		toggleMute()
		return nil
	}
	return event
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

	// track info page
	trackInfoTextBox = tview.NewTextView()
	trackInfoTextBox.SetBorder(true).SetTitle("Track info")
	pages.AddPage("track info", trackInfoTextBox, true, false)

	// lyrics page
	lyricsTextBox = tview.NewTextView()
	lyricsTextBox.SetBorder(true)
	pages.AddPage("lyrics", lyricsTextBox, true, false)

	progressBar = tview.NewTextView().
		SetDynamicColors(true)

	nowPlayingCover = tview.NewImage()
	nowPlayingCover.SetSize(-90, 0)
	nowPlayingTrackTextBox = tview.NewTextView().
		SetTextAlign(tview.AlignCenter)
	nowPlayingTimeTextBox = tview.NewTextView().
		SetTextAlign(tview.AlignCenter)
	nowPlayingFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nowPlayingCover, 0, 1, false).
		AddItem(nowPlayingTrackTextBox, 2, 1, false).
		AddItem(nowPlayingTimeTextBox, 2, 1, false).
		AddItem(progressBar, 2, 1, false)
	pages.AddPage("nowplaying", nowPlayingFlex, true, false)

	// main panel
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainPanel, 0, 1, true).
		AddItem(progressBar, 1, 1, false).
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

	fmt.Fprintf(downloadProgressText, "%d%%", volumePercent)

	bottomPanel.AddPage("search", searchInput, true, false)

	selectedTextStyle := tcell.StyleDefault.Foreground(tview.Styles.PrimitiveBackgroundColor).
		Background(tview.Styles.PrimaryTextColor).Attributes(tcell.AttrBold)

	// library page
	artistList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	artistList.SetBorder(true).SetTitle("Artist")
	artistList.SetChangedFunc(fillAlbumsList)
	artistList.
		SetMainTextColor(tcell.ColorYellow).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedStyle(selectedTextStyle).
		SetSelectedBackgroundColor(tcell.ColorTeal).
		SetBorderColor(tcell.ColorYellow).
		SetTitleColor(tcell.ColorYellow)

	albumList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	albumList.SetBorder(true).SetTitle("Albums")
	albumList.SetChangedFunc(fillTracksList)
	albumList.
		SetMainTextColor(tcell.ColorYellow).
		SetSelectedStyle(selectedTextStyle).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorTeal).
		SetBorderColor(tcell.ColorYellow).
		SetTitleColor(tcell.ColorYellow)

	trackList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	trackList.SetBorder(true).SetTitle("Tracks")
	trackList.
		SetMainTextColor(tcell.ColorYellow).
		SetSelectedStyle(selectedTextStyle).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorTeal).
		SetBorderColor(tcell.ColorYellow).
		SetTitleColor(tcell.ColorYellow)

	libraryFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(artistList, 0, 1, true).
			AddItem(albumList, 0, 1, false).
			AddItem(trackList, 0, 1, false), 0, 1, true)

	mainPanel.AddPage("library", libraryFlex, true, true)

	// queue page
	queueList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	queueList.SetBorder(true).SetTitle("Queue")
	queueList.SetSelectedFunc(playTrack)
	queueFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queueList, 0, 1, true)

	mainPanel.AddPage("queue", queueFlex, true, false)

	// playlist
	playlistList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	playlistList.SetBorder(true).SetTitle("Playlists")
	playlistList.SetChangedFunc(showPlaylist)

	playlistTracks = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	playlistTracks.SetBorder(true)
	playlistFlex = tview.NewFlex().
		AddItem(playlistList, 0, 1, true).
		AddItem(playlistTracks, 0, 3, false)

	mainPanel.AddPage("playlists", playlistFlex, true, false)

	pages.SendToFront("track info")

	// key handlers
	app.SetInputCapture(appInputHandler)
	libraryFlex.SetInputCapture(libraryInputHandler)
	queueFlex.SetInputCapture(queueInputHandler)
	playlistFlex.SetInputCapture(playlistInputHandler)
	searchInput.SetInputCapture(searchInputHandler)
}

func toggleTrackInfo() {
	var list *tview.List
	focused := app.GetFocus()

	switch app.GetFocus() {
	case trackList:
		list = trackList
	case playlistTracks:
		list = playlistTracks
	case queueList:
		list = queueList
	case trackInfoTextBox:
		pages.HidePage("track info")
	}

	if focused != trackInfoTextBox {
		pages.ShowPage("track info")
		trackInfoTextBox.Clear()
		_, trackID := list.GetItemText(list.GetCurrentItem())
		var id, title, album, artist, genre, suffix, albumID, artistID string
		var track, year, size, duration, bitrate int
		queryTrackInfo(toInt(trackID)).Scan(&id, &title, &album, &artist, &track, &year, &genre, &size, &suffix, &duration, &bitrate, &albumID, &artistID)
		fmt.Fprintf(trackInfoTextBox, "Title: %s\nAlbum: %s\nArtist: %s\nYear: %d\nTrack: %d\nGenre: %s\nSize: %s\nDuration: %s\nSuffix: %s\nBit rate: %d kbps\n", title, album, artist, year, track, genre, getSizeString(size), getTimeString(duration), suffix, bitrate)
	}
}

func toggleLyrics() {
	var list *tview.List
	switch app.GetFocus() {
	case trackList:
		list = trackList
	case playlistTracks:
		list = playlistTracks
	case queueList:
		list = queueList
	case lyricsTextBox:
		pages.HidePage("lyrics")
		return
	}

	if app.GetFocus() != lyricsTextBox {
		go showLyrics(list)
	} else {
		pages.HidePage("lyrics")
	}
}

// clears and draw progress bar
func refreshProgressBar(currentTime int, totalTime int) {
	progressBar.Clear()

	_, _, width, _ := mainPanel.GetInnerRect()
	// width is reduced by 4 to account for " |" at the beginning and "| " at the end of progress bar
	width -= 4

	// amount of '=' characters in progress bar
	progressCount := int(float64(currentTime) / float64(totalTime) * float64(width))
	// amount of padding spaces in progress bar
	negativeProgressCount := width - progressCount

	progress := " |"
	negativeProgress := strings.Repeat(" ", negativeProgressCount) + "|"
	if progressCount != 0 {
		// leading '=' is replaced by '>'
		progress = " |" + strings.Repeat("=", progressCount-1) + ">"
	}

	fmt.Fprintf(progressBar, "%s%s", progress, negativeProgress)
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

	currentTimeInt := currentTrack.stream.Position() / sr.N(time.Second)
	totalTimeInt := currentTrack.stream.Len() / sr.N(time.Second)

	currentTime := getTimeString(currentTimeInt)
	totalTime := getTimeString(totalTimeInt)

	// TODO: lazy refresh (for progress bar aswell)
	nowPlayingTrackTextBox.Clear()
	fmt.Fprintf(nowPlayingTrackTextBox, "%s - %s", currentTrack.Artist, currentTrack.Title)

	nowPlayingTimeTextBox.Clear()
	fmt.Fprintf(nowPlayingTimeTextBox, "%s / %s", currentTime, totalTime)

	refreshProgressBar(currentTimeInt, totalTimeInt)

	fmt.Fprintf(currentTrackText, "%s: %s - %s\t%s / %s\tQueue position: %d / %d", status, currentTrack.Artist, currentTrack.Title, currentTime, totalTime, queuePosition+1, queueList.GetItemCount())
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

func getSizeString(size int) string {
	return fmt.Sprintf("%.1fM", float64(size)/(1024*1024))
}

func appInputHandler(event *tcell.EventKey) *tcell.EventKey {
	focused := app.GetFocus()
	frontPage, _ := pages.GetFrontPage()
	if frontPage == "login" || focused == searchInput {
		return event
	}

	mainFrontPage, _ := mainPanel.GetFrontPage()
	// if nil is returned from trackInputHandler, event has been handled. trackInputHandler() only applies to certain pages
	if mainFrontPage == "library" || mainFrontPage == "queue" || mainFrontPage == "playlists" || frontPage == "nowplaying" {
		if event = trackInputHandler(event); event == nil {
			return nil
		}
	}

	switch event.Rune() {
	case 'q':
		app.Stop()
		return nil
	case '1':
		pages.SwitchToPage("main")
		mainPanel.SwitchToPage("queue")
		return nil
	case '2':
		pages.SwitchToPage("main")
		mainPanel.SwitchToPage("library")
		return nil
	case '3':
		pages.SwitchToPage("main")
		mainPanel.SwitchToPage("playlists")
		return nil
	case '4':
		pages.SwitchToPage("nowplaying")
		displayNowPlaying()
		return nil

	case 'i':
		toggleTrackInfo()
		return nil
	case '.':
		toggleLyrics()
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
