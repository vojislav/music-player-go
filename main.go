package main

import (
	"flag"
	"os"
	"path"
	"time"

	"github.com/faiface/beep/speaker"
	_ "github.com/mattn/go-sqlite3"
)

var tracksDirectory, lyricsDirectory, coversDirectory, configDirectory, playlistDirectory, databaseFile, configFile, initScriptFile string
var reloadDatabaseFlag *bool

func init() {
	homeDirectory, _ := os.UserHomeDir()
	configDirectory = path.Join(homeDirectory, ".config", "music-player-go")
	playlistDirectory = path.Join(configDirectory, "playlists")

	databaseFile = path.Join(configDirectory, "database.db")
	configFile = path.Join(configDirectory, "config")
	initScriptFile = path.Join(configDirectory, "init.sql")

	tracksDirectory = path.Join(homeDirectory, ".cache", "music-player-go", "tracks")
	lyricsDirectory = path.Join(homeDirectory, ".cache", "music-player-go", "lyrics")
	coversDirectory = path.Join(homeDirectory, ".cache", "music-player-go", "covers")
	// configDirectory = homeDirectory + "/.config/music-player-go/"
	// playlistDirectory = configDirectory + "playlists/"

	// databaseFile = configDirectory + "database.db"
	// configFile = configDirectory + "config"
	// initScriptFile = configDirectory + "init.sql"

	// cacheDirectory = homeDirectory + "/.cache/music-player-go/tracks/"
	// lyricsDirectory = homeDirectory + "/.cache/music-player-go/lyrics/"
	// coversDirectory = homeDirectory + "/.cache/music-player-go/covers/"

	if _, err := os.Stat(tracksDirectory); err != nil {
		os.Mkdir(tracksDirectory, 0755)
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

	if !validConfig() {
		pages.SwitchToPage("login")
	} else if !ping() {
		stopApp()
	} else if _, err := os.Stat(databaseFile); err != nil || *reloadDatabaseFlag {
		gotoLoadingPage()
	} else {
		gotoLibraryPage()
		initPlaylistPage()
	}

	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}
