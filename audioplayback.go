package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dhowden/tag"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

var sr = beep.SampleRate(44100)
var playerCtrl *CtrlVolume

var currentTrack Track

var ticker *time.Ticker
var killTicker = make(chan bool, 1)

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
	Disc     int    `json:"discNumber"`
	AlbumID  string `json:"albumId"`
	ArtistID string `json:"artistId"`
}

var volume = effects.Volume{
	Base:   2.0,
	Silent: false,
}

func playTrack(trackIndex int, _ string, trackID string, _ rune) {
	downloadMutex.RLock()
	defer downloadMutex.RUnlock()

	// track is scheduled for download - play it as soon as it downloads
	if _, ok := downloadMap[trackIndex]; ok {
		// TODO: notification for saying "this will play next when downloaded"
		playNextMutex.Lock()
		playNext = trackIndex
		playNextMutex.Unlock()
		return
	}

	fileName := getTrackPath(trackID)

	stream := getStream(fileName)
	tags := getTags(fileName)

	track, _ := tags.Track()
	currentTrack = Track{
		stream: stream,
		ID:     trackID,
		Title:  tags.Title(),
		Album:  tags.Album(),
		Artist: tags.Artist(),
		Year:   tags.Year(),
		Track:  track,
	}

	speaker.Clear()
	playerCtrl.Streamer = currentTrack.stream
	speaker.Play(playerCtrl)

	setQueuePosition(trackIndex)

	scrobble(toInt(currentTrack.ID), "false")

	go trackTime()
}

func togglePlay() {
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
	if currentTrack.stream == nil {
		return
	}

	speaker.Clear()
	currentTrack = Track{stream: nil}
	if !playerCtrl.Paused {
		killTicker <- true
	} else {
		updateCurrentTrackText()
	}
	setQueuePosition(-1)
}

func trackTime() {
	updateCurrentTrackText()
	ticker = time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			// prevents freak situation where rapidly switching songs
			// causes track stream to be accessed before its loaded
			if currentTrack.stream == nil {
				continue
			}
			if currentTrack.stream.Position() >= currentTrack.stream.Len()/2 {
				scrobble(toInt(currentTrack.ID), "true")
			}
			if currentTrack.stream.Position() == currentTrack.stream.Len() {
				nextTrack()
			}
			updateCurrentTrackText()
		case <-killTicker:
			ticker.Stop()
			updateCurrentTrackText()
			app.Draw()
			return
		}
	}
}

func nextTrack() {
	if queuePosition+1 == queueList.GetItemCount() {
		stopTrack()
		return
	}

	nextTrackName, nextTrackID := queueList.GetItemText(queuePosition + 1)
	playTrack(queuePosition+1, nextTrackName, nextTrackID, 0)
}

func previousTrack() {
	if queuePosition-1 < 0 {
		return
	}

	nextTrackName, nextTrackID := queueList.GetItemText(queuePosition - 1)
	playTrack(queuePosition-1, nextTrackName, nextTrackID, 0)
}

func changeVolume(step float64) {
	playerCtrl.Volume += step
	volumePercent += int(step * 10)
	downloadProgressText.Clear()
	fmt.Fprintf(downloadProgressText, "%d%%", volumePercent)
}

func toggleMute() {
	playerCtrl.Silent = !playerCtrl.Silent
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

func getDownloadProgress(done chan bool, filePath string, fileSize int, trackIndex int) {
	pattern := `\[::b\]-> \(\d+%\)\s`
	re, _ := regexp.Compile(pattern)
	var originalTrackName, originalTrackID string

	for {
		select {
		case <-done:
			downloadProgressText.Clear()
			fmt.Fprintf(downloadProgressText, "%d%%", volumePercent)

			// remove placeholder
			originalTrackName = strings.Replace(originalTrackName, trackNotDownloadedMarker, "", 1)
			queueList.SetItemText(trackIndex, originalTrackName, originalTrackID)
			app.Draw()

			playIfNext(originalTrackID, trackIndex)
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

			// add progress to track
			progress := fmt.Sprintf("[::b]-> (%.0f%%) ", downloadPercent)

			trackName, trackID := queueList.GetItemText(trackIndex)
			found := re.FindString(trackName)
			if found == "" {
				originalTrackName, originalTrackID = trackName, trackID
				trackName = progress + trackName
			} else {
				trackName = strings.Replace(trackName, found, progress, 1)
			}

			queueList.SetItemText(trackIndex, trackName, trackID)
			app.Draw()
		}
		time.Sleep(time.Second)
	}
}
