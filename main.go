package main

import (
	"fmt"
	"os"
	"time"

	"github.com/faiface/beep/speaker"
	_ "github.com/mattn/go-sqlite3"
)

var cacheDirectory, configDirectory, databaseFile, configFile string

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
	configFile = configDirectory + "config"
}

func main() {
	speaker.Init(sr, sr.N(time.Second/10))

	initView()

	if !validConfig() {
		pages.SwitchToPage("login")
	} else if !ping() {
		app.Stop()
	} else if _, err := os.Stat(databaseFile); err != nil {
		gotoLoadingPage()
	} else {
		gotoLibraryPage()
	}

	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}

func gotoLoadingPage() {
	if _, err := os.Stat(databaseFile); err == nil {
		gotoLibraryPage()
		return
	}

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
