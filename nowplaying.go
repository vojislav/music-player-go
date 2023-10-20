package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/jpeg"
)

var coverArt []byte

func displayNowPlaying() {
	if currentTrack.stream == nil {
		fmt.Fprint(nowPlayingTrackTextBox, "No currently playing track.")
		return
	}

	nowPlayingTrackTextBox.Clear()
	fmt.Fprintf(nowPlayingTrackTextBox, "%s - %s", currentTrack.Artist, currentTrack.Title)

	displayCoverArt()
}

func displayCoverArt() {
	// TODO cache image
	if currentTrack.stream == nil {
		return
	}

	coverArt = getCoverArt(currentTrack.ID) // TODO: lazy load
	encoded := base64.StdEncoding.EncodeToString(coverArt)

	b, _ := base64.StdEncoding.DecodeString(encoded)
	photo, _ := jpeg.Decode(bytes.NewReader(b))
	// resizedPhoto := resize.Resize(100, 0, photo, resize.Lanczos3)
	nowPlayingCover.SetImage(photo)
}
