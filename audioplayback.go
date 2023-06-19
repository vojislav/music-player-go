package main

import (
	"fmt"
	"log"
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
	ID       string `json:"id"`
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
	AlbumID  string `json:"albumId"`
	ArtistID string `json:"artistId"`
}

func playTrack(trackIndex int, _ string, trackID string, _ rune) {
	fileName := download(trackID)

	stream := getStream(fileName)
	tags := getTags(fileName)

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

	scrobble(toInt(currentTrack.ID), "false")

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
				scrobble(toInt(currentTrack.ID), "true")
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

func getDownloadProgress(done chan bool, filePath string, fileSize int) {
	for {
		select {
		case <-done:
			downloadProgressText.Clear()
			return
		default:
			file, err := os.Open(filePath)
			if err != nil {
				continue
			}

			fi, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}

			size := fi.Size()

			if size == 0 {
				size = 1
			}

			downloadPercent = float64(size) / float64(fileSize) * 100
			downloadProgressText.Clear()
			fmt.Fprintf(downloadProgressText, "%.0f%%", downloadPercent)

			app.Draw()
		}
		time.Sleep(time.Second)
	}
}
