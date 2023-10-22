package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"log"
	"os"
)

func displayNowPlaying() {
	if currentTrack.stream == nil {
		nowPlayingTrackTextBox.Clear()
		fmt.Fprint(nowPlayingTrackTextBox, "No currently playing track.")
		return
	}

	nowPlayingTrackTextBox.Clear()
	fmt.Fprintf(nowPlayingTrackTextBox, "%s - %s", currentTrack.Artist, currentTrack.Title)

	displayCoverArt()
}

func displayCoverArt() {
	if currentTrack.stream == nil {
		return
	}

	albumID := getAlbumID(currentTrack.ID)

	coverPath := fmt.Sprint(cacheDirectory, albumID, ".png")
	if _, err := os.Stat(coverPath); err != nil {
		coverArt := getCoverArt(currentTrack.ID) // TODO: lazy load
		f, err := os.Create(coverPath)
		if err != nil {
			log.Fatal(err)
		}

		_, err = f.Write(coverArt)
		if err != nil {
			log.Fatal(err)
		}
	}

	coverArt, err := os.ReadFile(coverPath)
	if err != nil {
		log.Fatal(err)
	}

	encoded := base64.StdEncoding.EncodeToString(coverArt)

	b, _ := base64.StdEncoding.DecodeString(encoded)
	photo, _ := jpeg.Decode(bytes.NewReader(b))
	// resizedPhoto := resize.Resize(100, 0, photo, resize.Lanczos3)
	nowPlayingCover.SetImage(photo)
}
