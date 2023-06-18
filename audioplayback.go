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
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Album    string `json:"album"`
	Artist   string `json:"artist"`
	Track    int    `json:"track"`
	Year     int    `json:"year"`
	Genre    string `json:"genre"`
	Size     int    `json:"size"`
	Suffix   string `json:"suffix"`
	Duration int    `json:"duration"`
	BitRate  int    `json:"bitRate"`
	AlbumID  int    `json:"albumId"`
	ArtistID int    `json:"artistId"`
}

func playTrack(trackIndex int, _ string, trackIDString string, _ rune) {
	fileName := download(trackIDString)

	stream := getStream(fileName)
	tags := getTags(fileName)

	trackID := toInt(trackIDString)

	track, _ := tags.Track()
	currentTrack = Track{
		stream: stream,
		ID:     trackID,
		Title:  tags.Title(),
		Album:  tags.Album(),
		Artist: tags.Artist(),
		Track:  track,
	}

	speaker.Clear()
	playerCtrl = &beep.Ctrl{Streamer: currentTrack.stream, Paused: false}
	speaker.Play(playerCtrl)

	queuePosition = trackIndex

	scrobble(currentTrack.ID, "false")

	go trackTime()
}

func playPause() {
	if currentTrack.stream == nil {
		return
	}

	speaker.Lock()
	playerCtrl.Paused = !playerCtrl.Paused
	if playerCtrl.Paused {
		killTicker <- true
	} else {
		go trackTime()
	}
	speaker.Unlock()
}

func stopTrack() {
	speaker.Clear()
	currentTrack = Track{stream: nil}
	killTicker <- true
}

func trackTime() {
	updateCurrentTrackText()
	ticker = time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			if currentTrack.stream.Position() >= currentTrack.stream.Len()/2 {
				scrobble(currentTrack.ID, "true")
			}
			updateCurrentTrackText()
		case <-killTicker:
			ticker.Stop()
			updateCurrentTrackText()
			return
		}
	}
}

func nextTrack() {
	if queuePosition+1 == queueList.GetItemCount() {
		return
	}

	queuePosition += 1

	nextTrackName, nextTrackID := queueList.GetItemText(queuePosition)
	playTrack(queuePosition, nextTrackName, nextTrackID, 0)
}

func previousTrack() {
	if queuePosition-1 < 0 {
		return
	}

	queuePosition -= 1

	nextTrackName, nextTrackID := queueList.GetItemText(queuePosition)
	playTrack(queuePosition, nextTrackName, nextTrackID, 0)
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
