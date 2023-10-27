package main

type TrackRequestCode int64

const (
	Play TrackRequestCode = iota
	Pause
	Stop
	TogglePlay
	Next
	Previous
	ChangeVolume
	Mute
)

type TrackRequest struct {
	request TrackRequestCode
	args    any
}

type PlayRequest struct {
	trackIndex int
	trackID    string
}

var trackRequestChan = make(chan TrackRequest, 10)

func requestPlayTrack(trackIndex int, _ string, trackID string, _ rune) {
	trackRequestChan <- TrackRequest{
		Play,
		PlayRequest{trackIndex, trackID},
	}
}

func requestStopTrack() {
	trackRequestChan <- TrackRequest{Stop, nil}
}

func requestTogglePlay() {
	trackRequestChan <- TrackRequest{TogglePlay, nil}
}

func requestNextTrack() {
	trackRequestChan <- TrackRequest{Next, nil}
}

func requestPreviousTrack() {
	trackRequestChan <- TrackRequest{Previous, nil}
}

func requestChangeVolume(step float64) {
	trackRequestChan <- TrackRequest{ChangeVolume, step}
}

func requestMute() {
	trackRequestChan <- TrackRequest{Mute, nil}
}

func playerWorker() {
	for {
		message := <-trackRequestChan
		request := message.request
		switch request {
		case Play:
			args := message.args.(PlayRequest)
			playTrack(args.trackIndex, "", args.trackID, 0)
		case Stop:
			stopTrack()
		case TogglePlay:
			togglePlay()
		case Next:
			nextTrack()
		case Previous:
			previousTrack()
		case ChangeVolume:
			step := message.args.(float64)
			changeVolume(step)
		case Mute:
			toggleMute()
		}
	}
}
