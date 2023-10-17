package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"os"
)

func displayNowPlaying() {
	f, _ := os.Open("cover.png")
	reader := bufio.NewReader(f)
	content, _ := ioutil.ReadAll(reader)

	encoded := base64.StdEncoding.EncodeToString(content)

	b, _ := base64.StdEncoding.DecodeString(encoded)
	photo, _ := jpeg.Decode(bytes.NewReader(b))
	// resizedPhoto := resize.Resize(100, 0, photo, resize.Lanczos3)
	nowPlayingCover.SetImage(photo)

	nowPlayingTrackTextBox.Clear()
	fmt.Fprint(nowPlayingTrackTextBox, "Buc Kesidi - Kafe aparat")

	nowPlayingTimeTextBox.Clear()
	fmt.Fprint(nowPlayingTimeTextBox, "02:54")
}
