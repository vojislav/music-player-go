package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/jpeg"
)

func displayNowPlaying() {
	// TODO cache image
	if currentTrack.stream == nil {
		nowPlayingTrackTextBox.Clear()
		fmt.Fprint(nowPlayingTrackTextBox, "No currently playing track.")
		return
	}

	encoded := base64.StdEncoding.EncodeToString(coverArt)

	b, _ := base64.StdEncoding.DecodeString(encoded)
	photo, _ := jpeg.Decode(bytes.NewReader(b))
	// resizedPhoto := resize.Resize(100, 0, photo, resize.Lanczos3)
	nowPlayingCover.SetImage(photo)
}
