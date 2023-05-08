package main

import (
	"fmt"
	"os"
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
var loadingPopup tview.Primitive

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

	artistID := toInt(artistIDString)
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

	albumID := toInt(albumIDString)

	rows := queryTracks(int(albumID))
	for rows.Next() {
		var trackID, albumID, artistID, track, duration int
		var title string
		rows.Scan(&trackID, &title, &albumID, &artistID, &track, &duration)

		trackList.AddItem(fmt.Sprintf("%d. %s", track, title), fmt.Sprint(trackID), 0, nil)
	}

	trackList.SetSelectedFunc(stream)
}

func getTimeString(time int) string {
	minutes := fmt.Sprint(time / 60)
	seconds := fmt.Sprint(time % 60)

	if len(seconds) == 1 {
		seconds = "0" + seconds
	}

	return fmt.Sprint(minutes, ":", seconds)
}

func updateCurrentTrack() {
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
		fmt.Fprintf(currentTrackText, "%s: %s - %s\t%s / %s", status, currentTrack.artist, currentTrack.title, currentTime, totalTime)
	}
	app.Draw()
}

func handleKeys(event *tcell.EventKey) *tcell.EventKey {
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
			stream(currentTrackIndex, currentTrackName, currentTrackID, 0)
		}
		return nil
	}

	switch event.Rune() {
	case 'p':
		speaker.Lock()
		playerCtrl.Paused = !playerCtrl.Paused
		if playerCtrl.Paused {
			killTicker <- true
		} else {
			go makeTicker()
		}
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

// var queue Queue
var app = tview.NewApplication()
var currentTrackText *tview.TextView
var loadingTextBox *tview.TextView
var playerCtrl *beep.Ctrl
var currentTrack Track
var ticker *time.Ticker
var killTicker = make(chan bool)

var cacheDirectory, configDirectory, databaseFile string

func makeTicker() {
	updateCurrentTrack()
	scrobble(currentTrack.id, "false")
	ticker = time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			if currentTrack.stream.Position() >= currentTrack.stream.Len()/2 {
				scrobble(currentTrack.id, "true")
			}
			updateCurrentTrack()
		case <-killTicker:
			ticker.Stop()
			updateCurrentTrack()
			return
		}
	}
}

func play() {
	speaker.Clear()
	playerCtrl = &beep.Ctrl{Streamer: currentTrack.stream, Paused: false}
	speaker.Play(playerCtrl)
	go makeTicker()
}

func init() {
	homeDirectory, _ := os.UserHomeDir()

	cacheDirectory = homeDirectory + "/.cache/music-player-go/"
	if _, err := os.Stat(cacheDirectory); err != nil {
		os.Mkdir(cacheDirectory, 0755)
	}

	configDirectory = homeDirectory + "/.config/music-player-go/"
	if _, err := os.Stat(configDirectory); err != nil {
		os.Mkdir(configDirectory, 0755)
	}

	databaseFile = configDirectory + "database.db"
}

func initView() {
	artistList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)
	artistList.SetBorder(true).SetTitle("Artist")

	albumList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)
	albumList.SetBorder(true).SetTitle("Albums")

	trackList = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)
	trackList.SetBorder(true).SetTitle("Tracks")

	currentTrackText = tview.NewTextView()
	currentTrackText.SetBorder(true)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(artistList, 0, 1, true).
			AddItem(albumList, 0, 1, false).
			AddItem(trackList, 0, 1, false), 0, 10, true).
		AddItem(currentTrackText, 0, 1, false)

	pages.AddPage("main", flex, true, true)

	loadingTextBox = tview.NewTextView()
	loadingTextBox.SetBorder(true)

	popup := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewGrid().
			SetColumns(0, width, 0).
			SetRows(0, height, 0).
			AddItem(p, 1, 1, 1, 1, 0, 0, true)
	}

	loadingPopup = popup(loadingTextBox, 40, 10)
	pages.AddPage("loading", loadingPopup, true, false)

	app.SetInputCapture(handleKeys)
}

func main() {
	speaker.Init(sr, sr.N(time.Second/10))

	initView()

	if _, err := os.Stat(databaseFile); err != nil {
		pages.SwitchToPage("loading")
		go loadDB()
	} else {
		pages.SwitchToPage("main")
		loadMain()
	}

	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}

func loadDB() {
	createDatabase()

	fmt.Fprint(loadingTextBox, "Loading...")
	app.Draw()

	loadDatabase()

	pages.SwitchToPage("main")
	loadMain()
	app.Draw()
}

func loadMain() {
	listArtists()

	main, secondary := artistList.GetItemText(0)
	listAlbums(0, main, secondary, 0)

	main, secondary = albumList.GetItemText(0)
	listTracks(0, main, secondary, 0)
}
