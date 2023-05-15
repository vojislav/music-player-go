package main

import (
	"flag"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	_ "github.com/mattn/go-sqlite3"
)

var cacheDirectory, configDirectory, databaseFile, configFile string
var reloadDatabaseFlag *bool

func init() {
	homeDirectory, _ := os.UserHomeDir()

	cacheDirectory = homeDirectory + "/.cache/music-player-go/"
	if _, err := os.Stat(cacheDirectory); err != nil {
		os.Mkdir(cacheDirectory, 0644)
	}

	configDirectory = homeDirectory + "/.config/music-player-go/"
	if _, err := os.Stat(configDirectory); err != nil {
		os.Mkdir(configDirectory, 0644)
	}

	databaseFile = configDirectory + "database.db"
	configFile = configDirectory + "config"
}

func main() {
	reloadDatabaseFlag = flag.Bool("r", false, "Reload library on startup")

	flag.Parse()

	speaker.Init(sr, sr.N(time.Second/10))
	playerCtrl = &beep.Ctrl{Streamer: nil, Paused: false}

	initView()

	if !validConfig() {
		pages.SwitchToPage("login")
	} else if !ping() {
		app.Stop()
	} else if _, err := os.Stat(databaseFile); err != nil || *reloadDatabaseFlag {
		gotoLoadingPage()
	} else {
		gotoLibraryPage()
	}

	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}
