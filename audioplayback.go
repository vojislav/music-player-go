package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dhowden/tag"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

var sr = beep.SampleRate(44100)
var playerCtrl *beep.Ctrl

var currentTrack Track

var ticker *time.Ticker
var killTicker = make(chan bool)

type Track struct {
	stream   beep.StreamSeekCloser
	id       int
	title    string
	album    string
	albumID  int
	artist   string
	artistID int
	track    int
	duration int
}

func playTrack(_ int, trackName string, trackIDString string, _ rune) {
	fileName := download(trackIDString)

	stream := getStream(fileName)
	tags := getTags(fileName)

	trackID := toInt(trackIDString)

	track, _ := tags.Track()
	currentTrack = Track{
		stream: stream,
		id:     trackID,
		title:  tags.Title(),
		album:  tags.Album(),
		artist: tags.Artist(),
		track:  track,
	}

	speaker.Clear()
	playerCtrl = &beep.Ctrl{Streamer: currentTrack.stream, Paused: false}
	speaker.Play(playerCtrl)

	scrobble(currentTrack.id, "false")

	go trackTime()
}

func trackTime() {
	updateCurrentTrackText()
	ticker = time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			if currentTrack.stream.Position() >= currentTrack.stream.Len()/2 {
				scrobble(currentTrack.id, "true")
			}
			updateCurrentTrackText()
		case <-killTicker:
			ticker.Stop()
			updateCurrentTrackText()
			return
		}
	}
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

func getTags(path string) tag.Metadata {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	tags, err := tag.ReadFrom(f)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return tags
}
