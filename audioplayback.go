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
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	gomp3 "github.com/hajimehoshi/go-mp3"
)

var sr = beep.SampleRate(44100)
var playerCtrl *CtrlVolume

var currentTrack Track

// how often will download progress in queue be updated
const downloadProgressSleepTime = time.Second

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

// sets playNext to trackIndex, removes playNext indicator from old track (if such exists) and adds it to new (if such exists)
func setNext(trackIndex int) {
	if trackIndex == playNext {
		return
	}

	oldPlayNext := playNext
	// remove old playNext marker
	if oldPlayNext >= 0 && oldPlayNext < queueList.GetItemCount() {
		trackText, trackID := queueList.GetItemText(oldPlayNext)
		trackText = strings.Replace(trackText, playNextIndicator, "", 1)
		queueList.SetItemText(oldPlayNext, trackText, trackID)
	}

	// set new playNext marker
	if trackIndex >= 0 {
		trackText, trackID := queueList.GetItemText(trackIndex)
		trackText = playNextIndicator + trackText
		queueList.SetItemText(trackIndex, trackText, trackID)
	}

	playNext = trackIndex
}

func playTrack(trackIndex int, _ string, trackID string, _ rune) {
	downloadMutex.RLock()
	defer downloadMutex.RUnlock()

	// track is scheduled for download - play it as soon as it downloads
	if _, ok := downloadMap[trackIndex]; ok {
		// TODO: notification for saying "this will play next when downloaded"
		setNext(trackIndex)
		return
	} else {
		setNext(-1)
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
	playerCtrl.Paused = false
	speaker.Play(playerCtrl)

	setQueuePosition(trackIndex)

	scrobble(toInt(currentTrack.ID), "false")

	asyncRequestStatusUpdate()
}

// if next song is to be played, play it
func playIfNext(args PlayRequest) {
	if args.trackIndex == playNext {
		playTrack(args.trackIndex, "", args.trackID, 0)
	}
}

func togglePlay() {
	if currentTrack.stream == nil {
		return
	}

	speaker.Lock()
	playerCtrl.Paused = !playerCtrl.Paused
	if playerCtrl.Paused {
		asyncRequestStatusPause()
	} else {
		asyncRequestStatusUpdate()
	}
	speaker.Unlock()
}

func stopTrack() {
	if currentTrack.stream == nil {
		return
	}

	speaker.Clear()
	currentTrack = Track{stream: nil}
	asyncRequestStatusPause()
	setQueuePosition(-1)
}

func nextTrackIndex(incr int) int {
	if playNext == -1 {
		return queuePosition + incr
	} else {
		return playNext + incr
	}
}

func nextTrack() {
	nextIndex := nextTrackIndex(+1)

	if nextIndex >= queueList.GetItemCount() {
		stopTrack()
		return
	}

	nextTrackName, nextTrackID := queueList.GetItemText(nextIndex)
	playTrack(nextIndex, nextTrackName, nextTrackID, 0)
}

func previousTrack() {
	nextIndex := nextTrackIndex(-1)

	if nextIndex < 0 {
		return
	}

	nextTrackName, nextTrackID := queueList.GetItemText(nextIndex)
	playTrack(nextIndex, nextTrackName, nextTrackID, 0)
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

func seek(step int) {
	if currentTrack.stream == nil {
		return
	}

	seekTo := currentTrack.stream.Position() + step*sr.N(time.Second)
	if seekTo < 0 {
		seekTo = 0
	} else if seekTo > currentTrack.stream.Len() {
		nextTrack()
		return
	}

	speaker.Lock()
	currentTrack.stream.Seek(seekTo)
	speaker.Unlock()
	asyncRequestStatusUpdate()
}

func getStream(path string) beep.StreamSeekCloser {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// get sample rate of current stream and update the speaker with it
	decodedStream, _ := gomp3.NewDecoder(f)
	if decodedStream.SampleRate() != int(sr) {
		sr = beep.SampleRate(decodedStream.SampleRate())
		speaker.Close()
		speaker.Init(sr, sr.N(time.Second/10))
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

func trackDownloadProgress(done chan bool, filePath string, fileSize int, trackIndex int) {
	pattern := `\s\[red::b\]<- \(\d+%\)\s`
	re, _ := regexp.Compile(pattern)
	var originalTrackName, originalTrackID string

	for {
		select {
		case <-done:
			downloadProgressText.Clear()
			fmt.Fprintf(downloadProgressText, "%d%%", volumePercent)

			// remove placeholder
			originalTrackName = strings.Replace(originalTrackName, trackNotDownloadedMarker, "", 1)
			// remove playNext marker in case it exists
			originalTrackName = strings.Replace(originalTrackName, playNextIndicator, "", 1)

			queueList.SetItemText(trackIndex, originalTrackName, originalTrackID)
			app.Draw()

			requestPlayIfNext(originalTrackID, trackIndex, false)
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
			progress := fmt.Sprintf(" [red::b]<- (%.0f%%) ", downloadPercent)

			trackName, trackID := queueList.GetItemText(trackIndex)
			found := re.FindString(trackName)
			if found == "" {
				originalTrackName, originalTrackID = trackName, trackID
				trackName = trackName + progress
			} else {
				trackName = strings.Replace(trackName, found, progress, 1)
			}

			queueList.SetItemText(trackIndex, trackName, trackID)
			app.Draw()
		}
		time.Sleep(downloadProgressSleepTime)
	}
}
