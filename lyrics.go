package main

import (
	"fmt"
	"os"

	lyrics "github.com/rhnvrm/lyric-api-go"
	"github.com/rivo/tview"
)

func showLyrics(list *tview.List) {
	_, trackID := list.GetItemText(list.GetCurrentItem())
	lyrics := getLyrics(trackID)

	lyricsTextBox.Clear()
	fmt.Fprint(lyricsTextBox, lyrics)
	pages.ShowPage("lyrics")
	pages.SendToFront("lyrics")
	app.Draw()
}

func getLyrics(trackID string) string {
	var artist, title string
	queryArtistAndTitle(toInt(trackID)).Scan(&artist, &title)

	lyricsPath := fmt.Sprint(lyricsDirectory, trackID, ".txt")
	if _, err := os.Stat(lyricsPath); err == nil {
		lyrics, err := os.ReadFile(lyricsPath)
		if err != nil {
			return "Error reading lyrics file"
		}

		lyricsTextBox.SetTitle(fmt.Sprintf(" %s - %s ", artist, title))

		return string(lyrics)
	}

	syncRequestCustomStatus(fmt.Sprintf("Downloading lyrics for %s - %s...", artist, title), 2000)

	l := lyrics.New()
	lyrics, err := l.Search(artist, title)

	if err != nil {
		return fmt.Sprintf("Lyrics for %s - %s were not found", artist, title)
	}

	err = os.WriteFile(lyricsPath, []byte(lyrics), 0644)
	if err != nil {
		return "Error writing lyrics to file"
	}

	lyricsTextBox.SetTitle(fmt.Sprintf(" %s - %s ", artist, title))

	return lyrics
}
