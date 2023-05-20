package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var app = tview.NewApplication()
var pages = tview.NewPages()
var artistList, albumList, trackList, queueList *tview.List
var loadingPopup tview.Primitive
var currentTrackText, downloadProgressText, loadingTextBox, loginStatus *tview.TextView

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
	artistList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	artistList.SetBorder(true).SetTitle("Artist")
	artistList.SetChangedFunc(fillAlbumsList)

	albumList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	albumList.SetBorder(true).SetTitle("Albums")
	albumList.SetChangedFunc(fillTracksList)

	trackList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	trackList.SetBorder(true).SetTitle("Tracks")

	currentTrackText = tview.NewTextView()
	currentTrackText.SetBorder(true)

	downloadProgressText = tview.NewTextView()
	downloadProgressText.SetBorder(true)

	libraryFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(artistList, 0, 1, true).
			AddItem(albumList, 0, 1, false).
			AddItem(trackList, 0, 1, false), 0, 1, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(currentTrackText, 0, 1, false).
			AddItem(downloadProgressText, 10, 1, false), 3, 1, false)

	pages.AddPage("library", libraryFlex, true, true)

	// queue page
	queueList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true).SetWrapAround(false)
	queueList.SetBorder(true).SetTitle("Queue")
	queueList.SetSelectedFunc(playTrack)
	queueFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queueList, 0, 1, true).
		AddItem(currentTrackText, 3, 1, false)

	pages.AddPage("queue", queueFlex, true, false)

	// key handlers
	app.SetInputCapture(appInputHandler)
	libraryFlex.SetInputCapture(libraryKeyHandler)
	queueFlex.SetInputCapture(queueInputHandler)
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

	rows := queryTracks(int(albumID))
	for rows.Next() {
		var trackID, albumID, artistID, track, duration int
		var title string
		rows.Scan(&trackID, &title, &albumID, &artistID, &track, &duration)

		trackList.AddItem(fmt.Sprintf("%d. %s", track, title), fmt.Sprint(trackID), 0, nil)
	}

}

func appInputHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case '1':
		pages.SwitchToPage("queue")
		return nil
	case '2':
		pages.SwitchToPage("library")
		return nil
	}

	return event
}

func libraryKeyHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEnter:
		currentTrackIndex := trackList.GetCurrentItem()
		_, currentTrackID := trackList.GetItemText(currentTrackIndex)
		go downloadCallback(currentTrackID, addToQueueAndPlay)
		trackList.SetCurrentItem(currentTrackIndex + 1)
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
			addToQueueAndPlay(currentTrackIndex, currentTrackName, currentTrackID, 0)
		}
		return nil
	}

	switch event.Rune() {
	case 'q':
		app.Stop()
	case 'p':
		pauseTrack()

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

	case '>':
		nextTrack()
	case '<':
		previousTrack()
	case 's':
		stopTrack()

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
	}

	return event
}

func queueInputHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRight:
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
		return nil
	case 'p':
		pauseTrack()
		return nil
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case 'l':
		currentTrackIndex := queueList.GetCurrentItem()
		currentTrackName, currentTrackID := queueList.GetItemText(currentTrackIndex)
		playTrack(currentTrackIndex, currentTrackName, currentTrackID, 0)
		return nil
	case '>':
		nextTrack()
	case '<':
		previousTrack()
	case 's':
		stopTrack()
	}

	return event
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
		fmt.Fprintf(currentTrackText, "%s: %s - %s\t%s / %s\tQueue position: %d / %d", status, currentTrack.artist, currentTrack.title, currentTime, totalTime, queuePosition+1, queueList.GetItemCount())
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

		pages.SwitchToPage("library")
		initLibraryPage()
		app.Draw()
	}()
}

func gotoLibraryPage() {
	pages.SwitchToPage("library")
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
