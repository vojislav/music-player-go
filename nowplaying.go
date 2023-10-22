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
	"path"

	"github.com/nfnt/resize"
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

	var coverPath string

	if _, err := os.Stat(coverPath); err != nil {
		coverArtBytes := getCoverArt(currentTrack.ID) // TODO: lazy load

		var format string
		coverArt, format, err := image.Decode(bytes.NewReader(coverArtBytes))
		if err != nil {
			log.Fatal(err)
		}

		// coverPath = fmt.Sprint(coversDirectory, albumID, ".", format)
		coverFileName := fmt.Sprint(albumID, ".", format)
		coverPath = path.Join(coversDirectory, coverFileName)

		f, err := os.Create(coverPath)
		if err != nil {
			log.Fatal(err)
		}

		resizedCoverArt := resize.Resize(300, 0, coverArt, resize.Lanczos3)

		if format == "jpeg" {
			err = jpeg.Encode(f, resizedCoverArt, nil)
		} else if format == "png" {
			err = png.Encode(f, resizedCoverArt)
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
