package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/itchyny/gojq"
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
	case 'p':
		playPause()

	case '>':
		nextTrack()
	case '<':
		previousTrack()
	case 's':
		stopTrack()

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
	case 'N':
		previousSearchResult()

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
	}

	return event
}

func queueInputHandler(event *tcell.EventKey) *tcell.EventKey {
	focused := app.GetFocus()
	if focused == loginForm || focused == searchInput {
		return event
	}

	switch event.Key() {
	case tcell.KeyRight, tcell.KeyEnter:
		currentTrackIndex := queueList.GetCurrentItem()
		currentTrackName, currentTrackID := queueList.GetItemText(currentTrackIndex)
		playTrack(currentTrackIndex, currentTrackName, currentTrackID, 0)
		return nil
	case tcell.KeyLeft:
		return nil
	case tcell.KeyDelete:
		removeFromQueue()
	}

	switch event.Rune() {
	case 'q':
		app.Stop()
	case 'p':
		playPause()

	case '>':
		nextTrack()
	case '<':
		previousTrack()
	case 's':
		stopTrack()

	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case 'l':
		currentTrackIndex := queueList.GetCurrentItem()
		currentTrackName, currentTrackID := queueList.GetItemText(currentTrackIndex)
		playTrack(currentTrackIndex, currentTrackName, currentTrackID, 0)
		return nil
	case 'g':
		return tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone)

	case 'G':
		return tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone)

	case 'n':
		nextSearchResult()
	case 'N':
		previousSearchResult()

	case '/':
		searchIndexes = nil
		searchCurrentIndex = 0

		searchList = queueList
		app.SetFocus(bottomPanel)
		bottomPanel.SwitchToPage("search")
	}

	return event
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
		}
		return nil

	}

	switch event.Rune() {
	case 'q':
		app.Stop()
	case 'p':
		playPause()

	case '>':
		nextTrack()
	case '<':
		previousTrack()
	case 's':
		stopTrack()

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

	case 'N':
		previousSearchResult()

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

func nextSearchResult() {
	if len(searchIndexes) != 0 {
		searchCurrentIndex = (searchCurrentIndex + 1) % len(searchIndexes)
		searchList.SetCurrentItem(searchIndexes[searchCurrentIndex])
	}
}

func previousSearchResult() {
	if len(searchIndexes) != 0 {
		searchCurrentIndex = (searchCurrentIndex - 1) % len(searchIndexes)
		searchList.SetCurrentItem(searchIndexes[searchCurrentIndex])
	}
}

func searchInputHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		bottomPanel.SwitchToPage("current track info")
		app.SetFocus(searchList)
		return nil

	case tcell.KeyEnter:
		searchString := searchInput.GetText()
		if len(searchString) == 0 {
			searchIndexes = nil
			go searchStatus("Search cleared", "")
		} else {
			searchIndexes = searchList.FindItems(searchString, "-", false, true)
			if len(searchIndexes) == 0 {
				go searchStatus("No results found!", "")
			} else {
				searchList.SetCurrentItem(searchIndexes[0])
				go searchStatus("Searching: ", searchString)
			}
		}

		bottomPanel.SwitchToPage("current track info")
		app.SetFocus(searchList)
		searchInput.SetText("")

		return nil
	}

	return event
}

func searchStatus(message, searchString string) {
	previousText := currentTrackText.GetText(true)
	currentTrackText.Clear()
	fmt.Fprint(currentTrackText, message, searchString)
	time.Sleep(2 * time.Second)
	currentTrackText.Clear()
	fmt.Fprint(currentTrackText, previousText)
	app.Draw()
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

func getDownloadProgress(done chan bool, filePath string, fileSize int) {
	for {
		select {
		case <-done:
			downloadProgressText.Clear()
			return
		default:
			file, err := os.Open(filePath)
			if err != nil {
				continue
			}

			fi, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}

			size := fi.Size()

			if size == 0 {
				size = 1
			}

			downloadPercent = float64(size) / float64(fileSize) * 100
			downloadProgressText.Clear()
			fmt.Fprintf(downloadProgressText, "%.0f%%", downloadPercent)

			app.Draw()
		}
		time.Sleep(time.Second)
	}
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

func gotoLibraryPage() {
	pages.SwitchToPage("main")
	initLibraryPage()
}

func getTimeString(time int) string {
	minutes := fmt.Sprint(time / 60)
	seconds := fmt.Sprint(time % 60)

	if len(seconds) == 1 {
		seconds = "0" + seconds
	}

	return fmt.Sprint(minutes, ":", seconds)
}
