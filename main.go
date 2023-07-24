package main

import (
	"flag"
	"os"
	"time"

	"github.com/faiface/beep/speaker"
	_ "github.com/mattn/go-sqlite3"
)

var cacheDirectory, configDirectory, playlistDirectory, databaseFile, configFile, initScriptFile string
var reloadDatabaseFlag *bool

func init() {
	homeDirectory, _ := os.UserHomeDir()
	configDirectory = homeDirectory + "/.config/music-player-go/"
	playlistDirectory = configDirectory + "playlists/"

	databaseFile = configDirectory + "database.db"
	configFile = configDirectory + "config"
	initScriptFile = configDirectory + "init.sql"

	cacheDirectory = homeDirectory + "/.cache/music-player-go/"
	if _, err := os.Stat(cacheDirectory); err != nil {
		os.Mkdir(cacheDirectory, 0755)
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
}

func main() {
	reloadDatabaseFlag = flag.Bool("r", false, "Reload library on startup")

	flag.Parse()

	speaker.Init(sr, sr.N(time.Second/10))
	currentTrack = Track{stream: nil}

	playerCtrl = &CtrlVolume{Streamer: nil, Paused: false, Silent: false, Base: 2.0, Volume: 0.0}

	initView()

	if !validConfig() {
		pages.SwitchToPage("login")
	} else if !ping() {
		app.Stop()
	} else if _, err := os.Stat(databaseFile); err != nil || *reloadDatabaseFlag {
		gotoLoadingPage()
	} else {
		gotoLibraryPage()
		// initPlaylistPage()
	}

	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}
