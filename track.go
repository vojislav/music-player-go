package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dhowden/tag"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
)

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

func changeCurrentTrack(path string) {
	stream := getStream(path)

	tags := getTags(path)

	trackID := toInt(strings.Split(path, ".")[0])

	track, _ := tags.Track()
	currentTrack = Track{
		stream: stream,
		id:     int(trackID),
		title:  tags.Title(),
		album:  tags.Album(),
		artist: tags.Artist(),
		track:  track,
	}

	play()
}
