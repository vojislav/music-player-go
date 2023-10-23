package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"

	"github.com/nfnt/resize"
)

// used to test if cover art needs changing as it is costly
var currentAlbumID = -1

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

// returns file path (string) of album cover with id albumID
func getCoverPath(albumID int) string {
	return fmt.Sprint(coversDirectory, albumID, ".jpg")
}

// removes cover art from nowplaying page
func removeCoverArt() {
	currentAlbumID = -1
	nowPlayingCover.SetImage(nil)
}

// caches cover art to disk if it doesn't exist and displays it on nowplaying page
func displayCoverArt() {
	if currentTrack.stream == nil {
		return
	}

	albumID := getAlbumID(currentTrack.ID)
	if currentAlbumID == albumID { // skip changing cover art if possible
		return
	}
	currentAlbumID = albumID

	var coverPath string = getCoverPath(albumID)

	if !fileExists(coverPath) {
		coverArtBytes := getCoverArt(currentTrack.ID) // TODO: lazy load

		var format string
		coverArt, format, err := image.Decode(bytes.NewReader(coverArtBytes))
		if err != nil {
			log.Fatal(err)
		}

		coverFile, err := os.Create(coverPath)
		if err != nil {
			log.Fatal(err)
		}

		resizedCoverArt := resize.Resize(300, 0, coverArt, resize.Lanczos3)

		// Encode writes into file coverFile
		if format == "jpeg" {
			err = jpeg.Encode(coverFile, resizedCoverArt, nil)
		} else if format == "png" {
			err = png.Encode(coverFile, resizedCoverArt)
		}

		if err != nil {
			log.Fatal(err)
		}
	}

	coverArtBytes, err := os.ReadFile(coverPath)
	if err != nil {
		log.Fatal(err)
	}

	encodedCoverArt := base64.StdEncoding.EncodeToString(coverArtBytes)

	b, _ := base64.StdEncoding.DecodeString(encodedCoverArt)
	coverArt, _, _ := image.Decode(bytes.NewReader(b))
	nowPlayingCover.SetImage(coverArt)
}
