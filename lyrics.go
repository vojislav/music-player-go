package main

import (
	"fmt"
	"os"
	"path"

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
	var id, title, album, artist, genre, suffix, albumID, artistID string
	var track, year, size, duration, bitrate int
	queryTrackInfo(toInt(trackID)).Scan(&id, &title, &album, &artist, &track, &year, &genre, &size, &suffix, &duration, &bitrate, &albumID, &artistID)

	// lyricsPath := fmt.Sprint(lyricsDirectory, trackID, ".txt")
	lyricsPath := path.Join(lyricsDirectory, trackID+".txt")
	if _, err := os.Stat(lyricsPath); err == nil {
		lyrics, err := os.ReadFile(lyricsPath)
		if err != nil {
			return "Error reading lyrics file"
		}

		lyricsTextBox.SetTitle(fmt.Sprintf(" %s - %s ", artist, title))

		return string(lyrics)
	}

	currentTrackText.Clear()
	fmt.Fprintf(currentTrackText, "Downloading lyrics for %s - %s...", artist, title)
	app.Draw()

	l := lyrics.New()
	lyrics, err := l.Search(artist, title)

	if err != nil {
		return fmt.Sprintf("Lyrics for %s - %s were not found", artist, title)
	}

	err = os.WriteFile(lyricsPath, []byte(lyrics), 0644)
	if err != nil {
		return "Error writing lyrics to file"
	}

	currentTrackText.Clear()
	app.Draw()

	lyricsTextBox.SetTitle(fmt.Sprintf(" %s - %s ", artist, title))

	return lyrics
}
