# music-player-go

## requirements
* tested on go version `go1.21.2 linux/amd64`
* requires `alsa` lib: `sudo apt install libasound2-dev`

## features
a music player for [Subsonic](subsonic.org) music servers, specifically my own at [music.lazic.xyz](https://music.lazic.xyz), written in Go and tview

modeled after [ncmpcpp](https://github.com/ncmpcpp/ncmpcpp)

current features:
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
* ``-/=`` - decrease/increase volume
* `m` - mute
* `i` - show track info

