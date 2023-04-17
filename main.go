package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

var sr = beep.SampleRate(44100)

func main() {
	speaker.Init(sr, sr.N(time.Second/10))

	var queue Queue
	speaker.Play(&queue)

	getArtists()
	for k, v := range artists {
		fmt.Printf("%d\t%s\n", k, v.name)
	}

	var artistID int

	fmt.Println("Enter artist ID")
	fmt.Scanln(&artistID)

	getAlbums(artistID)
	for k, v := range artists[artistID].albums {
		fmt.Printf("%d\t%s\n", k, v.name)
	}

	var albumID int
	fmt.Println("Enter album ID")
	fmt.Scanln(&albumID)

	getTracks(albumID)
	for k, v := range artists[artistID].albums[albumID].tracks {
		fmt.Printf("%d\t%s\n", k, v.title)
	}

	var trackID int
	fmt.Println("Enter track ID")
	fmt.Scanln(&trackID)

	stream(trackID)

	speaker.Lock()
	queue.Add(strconv.FormatInt(int64(trackID), 10) + ".mp3")
	speaker.Unlock()

	fmt.Scanln()
}

func getStream(path string) beep.StreamSeekCloser {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	streamer, _, err := mp3.Decode(f)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return streamer
}
