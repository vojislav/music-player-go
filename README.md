# music-player-go
a music player for [Subsonic](subsonic.org) music servers, specifically my own at [music.lazic.xyz](https://music.lazic.xyz), written in `Go` and `tview`

modeled after [ncmpcpp](https://github.com/ncmpcpp/ncmpcpp)

feel free to give it a go using my server (which is the default) with the `guest` user and password

## screenshots
<img src="https://github.com/vojislav/music-player-go/raw/main/screenshots/library.png" alt="library" width="80%" />
<img src="https://github.com/vojislav/music-player-go/raw/main/screenshots/queue.png" alt="queue" width="80%" />
<img src="https://github.com/vojislav/music-player-go/raw/main/screenshots/playlist.png" alt="playlist" width="80%" />
<img src="https://github.com/vojislav/music-player-go/raw/main/screenshots/nowplaying.png" alt="now playing" width="80%" />

## requirements
* tested on go version `go1.23.4 linux/amd64`
* requires `alsa` lib: `sudo apt install libasound2-dev`
* tested on `Subsonic 6.1.6` and `Airsonic 11.1.4`

## features
- library view
- queue
- playlists
- track info
- last.fm scrobbling

## keyboard shortcuts

* `h/j/k/l` - left/down/up/right movement
* `space` - add song to queue
* `enter` - add song to queue and play
* `p` - toggle play/pause
* `>/<` - next/previous track
* `/` - search
* `n/N` - next/previous search result
* `-/=` - decrease/increase volume
* `m` - mute
* `i` - show track info
* `.` - show lyrics
* `x` - remove current track from queue
