package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dhowden/tag"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
)

type Queue struct {
	tracks []Track
}

func (q *Queue) Add(path string) {
	stream := getStream(path)
	tags := getTags(path)

	trackID, err := strconv.ParseInt(strings.Split(path, ".")[0], 10, 32)
	if err != nil {
		fmt.Println(err)
		return
	}

	track, _ := tags.Track()

	newTrack := Track{
		stream: stream,
		id:     int(trackID),
		title:  tags.Title(),
		album:  tags.Album(),
		artist: tags.Artist(),
		track:  track,
	}

	q.tracks = append(q.tracks, newTrack)
	if len(q.tracks) == 1 {
		updateCurrentTrack()
	}
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

func (q *Queue) Stream(samples [][2]float64) (n int, ok bool) {
	filled := 0
	for filled < len(samples) {
		if len(q.tracks) == 0 {
			for i := range samples[filled:] {
				samples[i][0] = 0
				samples[i][1] = 0
			}
			break
		}

		n, ok := q.tracks[0].stream.Stream(samples[filled:])
		if !ok {
			q.tracks = q.tracks[1:]
			updateCurrentTrack()
		}
		filled += n
	}

	return len(samples), true
}

func (q *Queue) Err() error {
	return nil
}

func (q *Queue) Show() {
	// for i, track := range q.tracks {
	// 	fmt.Printf("#%d\t%s - %s\n", i, track.tags.Artist(), track.tags.Title())
	// }
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
