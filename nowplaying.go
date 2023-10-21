package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"io"
	"os"
)

func displayNowPlaying() {
	// TODO cache image
	if currentTrack.stream == nil {
		nowPlayingTrackTextBox.Clear()
		fmt.Fprint(nowPlayingTrackTextBox, "No currently playing track.")
		return
	}

	f, _ := os.Open("cover.png")
	reader := bufio.NewReader(f)
	content, _ := io.ReadAll(reader)

	encoded := base64.StdEncoding.EncodeToString(content)

	b, _ := base64.StdEncoding.DecodeString(encoded)
	photo, _ := jpeg.Decode(bytes.NewReader(b))
	// resizedPhoto := resize.Resize(100, 0, photo, resize.Lanczos3)
	nowPlayingCover.SetImage(photo)
}
