package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/faiface/beep/speaker"
	_ "github.com/mattn/go-sqlite3"
)

var cacheDirectory, lyricsDirectory, coversDirectory, configDirectory, playlistDirectory, databaseFile, configFile, initScriptFile, readmeFile string
var reloadDatabaseFlag *bool

func init() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile)

	homeDirectory, _ := os.UserHomeDir()
	configDirectory = homeDirectory + "/.config/music-player-go/"
	playlistDirectory = configDirectory + "playlists/"

	databaseFile = configDirectory + "database.db"
	configFile = configDirectory + "config"
	initScriptFile = configDirectory + "init.sql"
	readmeFile = "README.md"

	cacheDirectory = homeDirectory + "/.cache/music-player-go/tracks/"
	lyricsDirectory = homeDirectory + "/.cache/music-player-go/lyrics/"
	coversDirectory = homeDirectory + "/.cache/music-player-go/covers/"

	if _, err := os.Stat(cacheDirectory); err != nil {
		os.MkdirAll(cacheDirectory, 0755)
	}

	if _, err := os.Stat(lyricsDirectory); err != nil {
		os.Mkdir(lyricsDirectory, 0755)
	}

	if _, err := os.Stat(coversDirectory); err != nil {
		os.Mkdir(coversDirectory, 0755)
	}

	if _, err := os.Stat(configDirectory); err != nil {
		os.Mkdir(configDirectory, 0755)
		makeInitScript()
	}

	if _, err := os.Stat(initScriptFile); err != nil {
		makeInitScript()
	}

	if _, err := os.Stat(playlistDirectory); err != nil {
		os.Mkdir(playlistDirectory, 0755)
	}

	go downloadWorker()
	go playerWorker()
}

// the only way you should kill the app. ensures required work is done before it's stopped
func stopApp() {
	removeUnfinishedDownloads()
	app.Stop()
}

func main() {
	reloadDatabaseFlag = flag.Bool("r", false, "Reload library on startup")

	flag.Parse()

	speaker.Init(sr, sr.N(time.Second/10))
	currentTrack = Track{stream: nil}

	playerCtrl = &CtrlVolume{Streamer: nil, Paused: false, Silent: false, Base: 2.0, Volume: 0.0}

	initView()
	go trackTime()

	if !validConfig() {
		pages.SwitchToPage("login")
	} else if !ping() {
		stopApp()
	} else if _, err := os.Stat(databaseFile); err != nil || *reloadDatabaseFlag {
		// TODO: after initial load, playlist page isn't loaded
		gotoLoadingPage()
	} else {
		gotoLibraryPage()
		initPlaylistPage()
	}

	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}
