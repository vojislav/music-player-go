package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/itchyny/gojq"
)

// var config.ServerURL = "https://music.lazic.xyz/rest/"
// var config.Username = "voja"
// var config.Salt = "eYEy8Yue"
// var config.Token = "ee5d78b9d676fd5ab119a68860db3c59"
// var config.Version = "1.16.1"
// var client_name = "music-player-go"

type Artist struct {
	id     int
	name   string
	albums map[int]*Album
}

type Album struct {
	id         int
	artistID   int
	artistName string
	name       string
	year       int
	tracks     map[int]*Track
}

type Playlist struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Comment   string    `json:"comment"`
	Owner     string    `json:"owner"`
	Public    bool      `json:"public"`
	SongCount int       `json:"songCount"`
	Duration  int       `json:"duration"`
	Created   time.Time `json:"created"`
	Changed   time.Time `json:"changed"`
	CoverArt  string    `json:"coverArt"`
}

var artists = make(map[int]*Artist)
var downloadPercent float64

func ping() bool {
	req, err := http.NewRequest("GET", config.ServerURL+"ping", nil)
	if err != nil {
		log.Fatal(err)
	}

	params := req.URL.Query()
	params.Add("u", config.Username)
	params.Add("t", config.Token)
	params.Add("s", config.Salt)
	params.Add("v", config.Version)
	params.Add("c", client_name)
	params.Add("f", "json")
	req.URL.RawQuery = params.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var resJSON map[string]interface{}
	json.Unmarshal(body, &resJSON)

	if resJSON["subsonic-response"].(map[string]interface{})["status"] == "ok" {
		return true
	} else {
		return false
	}
}

func getArtists() bool {
	req, err := http.NewRequest("GET", config.ServerURL+"getArtists", nil)
	if err != nil {
		log.Fatal(err)
	}

	params := req.URL.Query()
	params.Add("u", config.Username)
	params.Add("t", config.Token)
	params.Add("s", config.Salt)
	params.Add("v", config.Version)
	params.Add("c", client_name)
	params.Add("f", "json")
	req.URL.RawQuery = params.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	query, err := gojq.Parse(`."subsonic-response".artists.index[].artist[] | .id + "\t" + .name`)
	if err != nil {
		log.Fatal(err)
	}

	var resJSON map[string]interface{}
	json.Unmarshal(body, &resJSON)

	iter := query.Run(resJSON)
	for v, ok := iter.Next(); ok; v, ok = iter.Next() {
		split := strings.Split(v.(string), "\t")
		id := toInt(split[0])
		name := split[1]
		albums := make(map[int]*Album)
		artists[int(id)] = &Artist{id: int(id), name: name, albums: albums}
	}

	return true
}

func getAlbums(artistID int) bool {
	req, err := http.NewRequest("GET", config.ServerURL+"getArtist", nil)
	if err != nil {
		log.Fatal(err)
	}

	params := req.URL.Query()
	params.Add("u", config.Username)
	params.Add("t", config.Token)
	params.Add("s", config.Salt)
	params.Add("v", config.Version)
	params.Add("c", client_name)
	params.Add("f", "json")
	params.Add("id", strconv.FormatInt(int64(artistID), 10))
	req.URL.RawQuery = params.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	query, err := gojq.Parse(`."subsonic-response".artist.album[] | .id + "\t" + .artist + "\t" + .name + "\t" + (.year|tostring)`)
	if err != nil {
		log.Fatal(err)
	}

	var resJSON map[string]interface{}
	json.Unmarshal(body, &resJSON)

	iter := query.Run(resJSON)
	for v, ok := iter.Next(); ok; v, ok = iter.Next() {
		split := strings.Split(v.(string), "\t")
		id := toInt(split[0])
		artistName := split[1]
		name := split[2]
		year := toInt(split[3])
		tracks := make(map[int]*Track)
		artists[artistID].albums[int(id)] =
			&Album{
				id:         int(id),
				artistID:   artistID,
				artistName: artistName,
				name:       name,
				year:       int(year),
				tracks:     tracks}
	}

	return true
}

func getTracks(albumID int) bool {
	req, err := http.NewRequest("GET", config.ServerURL+"getAlbum", nil)
	if err != nil {
		log.Fatal(err)
	}

	params := req.URL.Query()
	params.Add("u", config.Username)
	params.Add("t", config.Token)
	params.Add("s", config.Salt)
	params.Add("v", config.Version)
	params.Add("c", client_name)
	params.Add("f", "json")
	params.Add("id", fmt.Sprint(albumID))
	req.URL.RawQuery = params.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// ."subsonic-response".album.song[]

	var resJSON map[string]interface{}
	json.Unmarshal(body, &resJSON)

	query, err := gojq.Parse(`."subsonic-response".album.artistId`)
	if err != nil {
		log.Fatal(err)
	}

	var artistID string

	iter := query.Run(resJSON)
	if artistIDAny, ok := iter.Next(); ok {
		artistID = artistIDAny.(string)
	}

	query, err = gojq.Parse(`."subsonic-response".album.song[]`)
	if err != nil {
		log.Fatal(err)
	}

	iter = query.Run(resJSON)
	for track, ok := iter.Next(); ok; track, ok = iter.Next() {
		trackMap := track.(map[string]any)

		var track float64
		if track, ok = trackMap["track"].(float64); !ok {
			track = 0
		}

		newTrack := Track{
			ID:       trackMap["id"].(string),
			Title:    trackMap["title"].(string),
			Album:    trackMap["album"].(string),
			AlbumID:  fmt.Sprint(albumID),
			Artist:   trackMap["artist"].(string),
			ArtistID: artistID,
			Track:    int(track),
			Duration: int(trackMap["duration"].(float64))}

		artists[toInt(artistID)].albums[albumID].tracks[toInt(newTrack.ID)] = &newTrack
	}

	return true
}

func downloadCallback(trackIDString string, callback func(int, string, string, rune)) {
	_ = download(trackIDString)
	callback(0, "", trackIDString, 0)
	app.Draw()
}

func download(trackIDString string) string {
	filePath := fmt.Sprint(cacheDirectory, trackIDString, ".mp3")

	if _, err := os.Stat(filePath); err != nil {
		trackID := toInt(trackIDString)

		req, err := http.NewRequest("GET", config.ServerURL+"stream", nil)
		if err != nil {
			log.Fatal(err)
		}

		params := req.URL.Query()
		params.Add("u", config.Username)
		params.Add("t", config.Token)
		params.Add("s", config.Salt)
		params.Add("v", config.Version)
		params.Add("c", client_name)
		params.Add("f", "json")
		params.Add("id", fmt.Sprint(trackID))
		req.URL.RawQuery = params.Encode()

		headResp, err := http.Head(req.URL.String())
		if err != nil {
			log.Fatal(err)
		}

		fileSize, err := strconv.Atoi(headResp.Header.Get("Content-Length"))
		if err != nil {
			log.Fatal(err)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()

		downloadDone := make(chan bool)
		go getDownloadProgress(downloadDone, filePath, fileSize)

		file, err := os.Create(filePath)
		if err != nil {
			log.Fatal(err)
		}

		_, err = io.Copy(file, res.Body)
		if err != nil {
			log.Fatal(err)
		}

		downloadDone <- true
	}

	return filePath
}

func scrobble(trackID int, submission string) bool {
	req, err := http.NewRequest("GET", config.ServerURL+"scrobble", nil)
	if err != nil {
		log.Fatal(err)
	}

	elapsedTime := currentTrack.stream.Position() / sr.N(time.Second)
	time := int(time.Now().UnixMilli()) - elapsedTime*1000

	params := req.URL.Query()
	params.Add("u", config.Username)
	params.Add("t", config.Token)
	params.Add("s", config.Salt)
	params.Add("v", config.Version)
	params.Add("c", client_name)
	params.Add("f", "json")
	params.Add("id", fmt.Sprint(trackID))
	params.Add("submission", submission)
	params.Add("time", fmt.Sprint(time))
	req.URL.RawQuery = params.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var resJSON map[string]interface{}
	json.Unmarshal(body, &resJSON)

	if resJSON["subsonic-response"].(map[string]interface{})["status"] == "ok" {
		return true
	} else {
		return false
	}
}

func toInt(s string) int {
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return -1
	}
	return int(n)
}

func getPlaylists() []Playlist {
	req, err := http.NewRequest("GET", config.ServerURL+"getPlaylists", nil)
	if err != nil {
		log.Fatal(err)
	}

	params := req.URL.Query()
	params.Add("u", config.Username)
	params.Add("t", config.Token)
	params.Add("s", config.Salt)
	params.Add("v", config.Version)
	params.Add("c", client_name)
	params.Add("f", "json")
	req.URL.RawQuery = params.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	query, err := gojq.Parse(`."subsonic-response".playlists.playlist[]`)
	if err != nil {
		log.Fatal(err)
	}

	var resJSON map[string]interface{}
	json.Unmarshal(body, &resJSON)

	iter := query.Run(resJSON)

	var playlists []Playlist

	for playlistMap, ok := iter.Next(); ok; playlistMap, ok = iter.Next() {
		// id := playlist.(map[string]any)["id"].(string)
		// name := playlist.(map[string]any)["name"].(string)
		// playlists = append(playlists, Playlist{ID: id, Name: name})
		newPlaylist := Playlist{}
		playlistJSON, _ := json.Marshal(playlistMap)
		json.Unmarshal(playlistJSON, &newPlaylist)
		playlists = append(playlists, newPlaylist)
	}

	return playlists
}

func getPlaylistTracks(playlistID int) []byte {
	req, err := http.NewRequest("GET", config.ServerURL+"getPlaylist", nil)
	if err != nil {
		log.Fatal(err)
	}

	params := req.URL.Query()
	params.Add("u", config.Username)
	params.Add("t", config.Token)
	params.Add("s", config.Salt)
	params.Add("v", config.Version)
	params.Add("c", client_name)
	params.Add("f", "json")
	params.Add("id", fmt.Sprint(playlistID))
	req.URL.RawQuery = params.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	tracksJSON, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// query, err := gojq.Parse(`."subsonic-response".playlist.entry`)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// var resJSON map[string]interface{}
	// json.Unmarshal(body, &resJSON)

	// var tracksJSON []byte

	// iter := query.Run(resJSON)
	// if trackMap, ok := iter.Next(); ok {
	// 	tracksJSON, _ = json.Marshal(trackMap)
	// }

	return tracksJSON
}
