package main

type Queue struct {
	tracks []Track
}

func (q *Queue) Add(path string) {
	stream := getStream(path)
	newTrack := Track{stream: stream}
	q.tracks = append(q.tracks, newTrack)
	// q.streamers = append(q.streamers, streamers...)
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
