package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var app = tview.NewApplication()
var pages = tview.NewPages()
var mainPanel = tview.NewPages()
var bottomPage = tview.NewPages()
var loadingPopup tview.Primitive
var currentTrackText, currentTrackTime, downloadProgressText, loadingTextBox, loginStatus, trackInfoTextBox,
	lyricsTextBox, helpWindowTextBox, nowPlayingTrackTextBox, nowPlayingTimeTextBox, progressBar *tview.TextView
var nowPlayingCover *tview.Image
var loginGrid *tview.Grid
var libraryFlex, queueFlex, playlistFlex, nowPlayingFlex, bottomPanel *tview.Flex

// remembers which page was last before going to help page
var lastPage string

// for each page remember which *tview.List was focused last so context can be restored
// IMPORTANT: this and setAndSaveFocus() and restoreFocus() work only on mainFrontPage.
// if focus is on bottomPage (i.e. during search) using these functions would lead to undefined behaviour
var pageFocus map[string]*tview.List = make(map[string]*tview.List)

// every time focus is set save that information
func setAndSaveFocus(list *tview.List) {
	mainFrontPage, _ := mainPanel.GetFrontPage()
	pageFocus[mainFrontPage] = list
	app.SetFocus(list)
}

// on certain multi-listed pages [library, playlists] context should be restored
func restoreFocus() {
	mainFrontPage, _ := mainPanel.GetFrontPage()
	list, ok := pageFocus[mainFrontPage]
	if ok && list != nil {
		app.SetFocus(list)
	}
}

// saves context of caller of track info or lyrics
var focusedList *tview.List

var popup = func(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)
}

func Center(width, height int, p tview.Primitive) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}

var selectedTextStyle = tcell.StyleDefault.Foreground(tview.Styles.PrimitiveBackgroundColor).
	Background(tview.Styles.PrimaryTextColor).Attributes(tcell.AttrBold)

func setColor(list *tview.List) {
	list.SetMainTextColor(tcell.ColorYellow).
		SetSelectedStyle(selectedTextStyle).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorTeal).
		SetBorderColor(tcell.ColorYellow).
		SetTitleColor(tcell.ColorYellow)
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
		requestTogglePlay()
		return nil
	case 's':
		requestStopTrack()
		return nil
	case '>':
		requestNextTrack()
		return nil
	case '<':
		requestPreviousTrack()
		return nil

	case '/':
		searchIndexes = nil

		searchList = app.GetFocus().(*tview.List)

		// this prevents bug which happens when search is called while pageFocus map is empty
		setAndSaveFocus(searchList)

		if searchStartContext == -1 {
			searchStartContext = searchList.GetCurrentItem()
		}

		// freak situation where we don't use setAndSaveFocus() because we don't want to return to
		// bottomPage, but to context before bottomPage is focused.
		app.SetFocus(bottomPage)
		bottomPage.SwitchToPage("search")
		return nil
	case 'n':
		nextSearchResult()
		return nil
	case 'N':
		previousSearchResult()
		return nil

	case '=':
		requestChangeVolume(volumeStep)
		return nil
	case '-':
		requestChangeVolume(-volumeStep)
		return nil
	case 'm':
		requestMute()
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
	trackInfoTextBox = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(false)
	trackInfoTextBox.
		SetBorder(true).
		SetTitle(" Track info ")
	trackInfoTextBox.
		SetBorderColor(tcell.ColorYellow).
		SetTitleColor(tcell.ColorYellow)
	pages.AddPage("track info", Center(50, 20, trackInfoTextBox), true, false)

	// lyrics page
	lyricsTextBox = tview.NewTextView()
	lyricsTextBox.
		SetDynamicColors(true).
		SetBorder(true)
	lyricsTextBox.
		SetBorderColor(tcell.ColorYellow).
		SetTitleColor(tcell.ColorYellow)
	pages.AddPage("lyrics", Center(75, 30, lyricsTextBox), true, false)

	// help window
	helpWindowTextBox = tview.NewTextView()
	helpWindowTextBox.
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Help ")
	helpWindowTextBox.
		SetBorderColor(tcell.ColorYellow).
		SetTitleColor(tcell.ColorYellow)
	initHelpWindow()
	pages.AddPage("help", helpWindowTextBox, true, false)

	progressBar = tview.NewTextView().
		SetDynamicColors(true)
	progressBar.SetTextColor(tcell.ColorPeachPuff)

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
		AddItem(bottomPage, 3, 1, false)

	pages.AddPage("main", mainFlex, true, false)

	// bottom panel that contains info about current track and download progress
	bottomPanel = tview.NewFlex().
		SetDirection(tview.FlexColumn)
	bottomPanel.SetBorder(false)

	// panel that contains track info
	currentTrackPanel := tview.NewFlex().
		SetDirection(tview.FlexColumn)
	currentTrackPanel.
		SetBorderColor(tcell.ColorDarkGray).
		SetBorder(true)

	// TODO: add "..." if text cannot fit
	currentTrackText = tview.NewTextView().
		SetDynamicColors(true)
	currentTrackTime = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	currentTrackPanel.
		AddItem(currentTrackText, 0, 1, false).
		AddItem(currentTrackTime, 13, 1, false) // maximum string length for time format len([xx:xx/xx:xx]) = 13

	downloadProgressText = tview.NewTextView()
	downloadProgressText.
		SetBorder(true).
		SetBorderColor(tcell.ColorDarkGrey)
	fmt.Fprintf(downloadProgressText, "%d%%", volumePercent)

	bottomPanel.
		AddItem(currentTrackPanel, 0, 1, false).
		AddItem(downloadProgressText, 10, 1, false)

	bottomPage.AddPage("current track info", bottomPanel, true, true)

	searchInput = tview.NewInputField().
		SetLabel("Search: ")

	bottomPage.AddPage("search", searchInput, true, false)

	// library page
	artistList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	artistList.SetBorder(true).SetTitle(" Artist ")
	artistList.SetChangedFunc(fillAlbumsList)
	setColor(artistList)

	albumList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	albumList.SetBorder(true).SetTitle(" Albums ")
	albumList.SetChangedFunc(fillTracksList)
	setColor(albumList)

	trackList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	trackList.SetBorder(true).SetTitle(" Tracks ")
	setColor(trackList)

	libraryFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(artistList, 0, 1, true).
			AddItem(albumList, 0, 1, false).
			AddItem(trackList, 0, 1, false), 0, 1, true)

	mainPanel.AddPage("library", libraryFlex, true, true)

	// queue page
	queueList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	queueList.SetBorder(true).SetTitle(" Queue ")
	queueList.SetSelectedFunc(requestPlayTrack)
	setColor(queueList)
	queueFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queueList, 0, 1, true)

	mainPanel.AddPage("queue", queueFlex, true, false)

	// playlist
	playlistList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	playlistList.SetBorder(true).SetTitle(" Playlists ")
	playlistList.SetChangedFunc(showPlaylist)
	setColor(playlistList)

	playlistTracks = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	playlistTracks.SetBorder(true)
	setColor(playlistTracks)
	playlistFlex = tview.NewFlex().
		AddItem(playlistList, 0, 1, true).
		AddItem(playlistTracks, 0, 3, false)

	mainPanel.AddPage("playlists", playlistFlex, true, false)

	// key handlers
	app.SetInputCapture(appInputHandler)
	libraryFlex.SetInputCapture(libraryInputHandler)
	queueFlex.SetInputCapture(queueInputHandler)
	playlistFlex.SetInputCapture(playlistInputHandler)
	searchInput.SetInputCapture(searchInputHandler)
	searchInput.SetChangedFunc(searchIncremental)
}

func toggleTrackInfo() {
	var list *tview.List

	switch app.GetFocus() {
	case trackList:
		list = trackList
	case playlistTracks:
		list = playlistTracks
	case queueList:
		list = queueList
	case trackInfoTextBox:
		pages.HidePage("track info")
		setAndSaveFocus(focusedList)
		focusedList = nil
		return
	default:
		return
	}
	focusedList = list

	pages.ShowPage("track info")
	pages.SendToFront("track info")
	trackInfoTextBox.Clear()
	_, trackID := list.GetItemText(list.GetCurrentItem())
	var id, title, album, artist, genre, suffix, albumID, artistID string
	var track, disc, year, size, duration, bitrate int
	queryTrackInfo(toInt(trackID)).Scan(&id, &title, &album, &artist, &track, &year, &genre, &size, &suffix, &duration, &bitrate, &disc, &albumID, &artistID)
	if genre == "" {
		genre = "-"
	}
	fmt.Fprintf(trackInfoTextBox, "[yellow::b]Title[-::B]: %s\n[yellow::b]Album[-::B]: %s\n[yellow::b]Artist[-::B]: %s\n[yellow::b]Year[-::B]: %d\n[yellow::b]Track[-::B]: %d\n[yellow::b]Disc[-::B]: %d\n[yellow::b]Genre[-::B]: %s\n[yellow::b]Size[-::B]: %s\n[yellow::b]Duration[-::B]: %s\n[yellow::b]Suffix[-::B]: %s\n[yellow::b]Bit rate[-::B]: %d kbps\n", title, album, artist, year, track, disc, genre, getSizeString(size), getTimeString(duration), suffix, bitrate)
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
		setAndSaveFocus(focusedList)
		focusedList = nil
		return
	default:
		return
	}
	focusedList = list

	go showLyrics(list)
}

// toggles between whichever page is current and help page
func toggleHelpPage() {
	frontPage, _ := pages.GetFrontPage()
	if frontPage == "help" {
		pages.SwitchToPage(lastPage)
		restoreFocus()
	} else {
		lastPage = frontPage
		pages.SwitchToPage("help")
	}
}

func initHelpWindow() {
	readme, err := os.ReadFile(readmeFile)
	if err != nil {
		log.Fatal(err)
	}

	r, _ := regexp.Compile("(?s)keyboard shortcuts.*")
	shortcuts := r.Find(readme)

	shortcutsString := string(shortcuts)
	shortcutsString = strings.Replace(shortcutsString, "keyboard shortcuts", "[yellow::b]Keyboard shortcuts:[-::-]", 1)

	itemBegin, _ := regexp.Compile(`\*\s*` + "`")
	itemEnd, _ := regexp.Compile("`" + `\s`)
	shortcutsString = string(itemBegin.ReplaceAll([]byte(shortcutsString), []byte("[yellow::b]")))
	shortcutsString = string(itemEnd.ReplaceAll([]byte(shortcutsString), []byte("[-::-] ")))

	fmt.Fprint(helpWindowTextBox, shortcutsString)

	fmt.Fprintf(helpWindowTextBox, "\n\n---\n\n[yellow::b]Memory usage:[-::-]\n\n")
	fmt.Fprintf(helpWindowTextBox, "Tracks cache: %s\n", getSizeString(getDirSize(cacheDirectory)))
	fmt.Fprintf(helpWindowTextBox, "Covers cache: %s\n", getSizeString(getDirSize(coversDirectory)))
	fmt.Fprintf(helpWindowTextBox, "Lyrics cache: %s\n", getSizeString(getDirSize(lyricsDirectory)))
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

	progress := " [::d]|[::D]"
	negativeProgress := strings.Repeat(" ", negativeProgressCount) + "[::d]|[::D]"
	if progressCount != 0 {
		// leading '=' is replaced by '>'
		progress += strings.Repeat("=", progressCount-1) + ">"
	}

	fmt.Fprintf(progressBar, "%s%s", progress, negativeProgress)
}

func updateCurrentTrackText() { // TODO: better name as this function is getting bloated and affects other stuff
	currentTrackText.Clear()
	currentTrackTime.Clear()

	nowPlayingTrackTextBox.Clear()
	nowPlayingTimeTextBox.Clear()

	if currentTrack.stream == nil {
		progressBar.Clear()
		nowPlayingTrackTextBox.SetText("No currently playing track.")
		removeCoverArt()
		app.Draw()
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

	// update bottomPanel info
	fmt.Fprintf(currentTrackText, `[red::b]%s:[-::-] "%s" [::d]by[::D] [yellow]%s[-] [::d]in[::D] [teal]%s (%d)[-]`, status, currentTrack.Title, currentTrack.Artist, currentTrack.Album, currentTrack.Year)
	fmt.Fprintf(currentTrackTime, "[blue::b][%s/%s]", currentTime, totalTime)

	// update progress bar
	refreshProgressBar(currentTimeInt, totalTimeInt)

	// update nowplaying
	fmt.Fprintf(nowPlayingTrackTextBox, "%s - %s", currentTrack.Artist, currentTrack.Title)
	fmt.Fprintf(nowPlayingTimeTextBox, "%s / %s", currentTime, totalTime)
	displayCoverArt()

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
		stopApp()
		return nil
	case '1':
		pages.SwitchToPage("main")
		mainPanel.SwitchToPage("queue")
		return nil
	case '2':
		pages.SwitchToPage("main")
		mainPanel.SwitchToPage("library")
		restoreFocus()
		return nil
	case '3':
		pages.SwitchToPage("main")
		mainPanel.SwitchToPage("playlists")
		restoreFocus()
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

	switch event.Key() {
	case tcell.KeyF1:
		toggleHelpPage()
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

		gotoLibraryPage()
		initPlaylistPage()
		app.Draw()
	}()
}
