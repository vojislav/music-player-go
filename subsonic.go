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

	"github.com/faiface/beep"
	"github.com/itchyny/gojq"
)

var serverURL = "https://music.lazic.xyz/rest/"
var username = "voja"
var salt = "eYEy8Yue"
var token = "ee5d78b9d676fd5ab119a68860db3c59"
var version = "1.16.1"
var client = "music-player-go"

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

var artists = make(map[int]*Artist)

func ping() bool {
	req, err := http.NewRequest("GET", serverURL+"ping", nil)
	if err != nil {
		log.Fatal(err)
		return false
	}

	params := req.URL.Query()
	params.Add("u", username)
	params.Add("t", token)
	params.Add("s", salt)
	params.Add("v", version)
	params.Add("c", client)
	params.Add("f", "json")
	req.URL.RawQuery = params.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
		return false
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
	req, err := http.NewRequest("GET", serverURL+"getArtists", nil)
	if err != nil {
		log.Fatal(err)
		return false
	}

	params := req.URL.Query()
	params.Add("u", username)
	params.Add("t", token)
	params.Add("s", salt)
	params.Add("v", version)
	params.Add("c", client)
	params.Add("f", "json")
	req.URL.RawQuery = params.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
		return false
	}

	query, err := gojq.Parse(`."subsonic-response".artists.index[].artist[] | .id + "\t" + .name`)
	if err != nil {
		log.Fatal(err)
		return false
	}

	var resJSON map[string]interface{}
	json.Unmarshal(body, &resJSON)

	iter := query.Run(resJSON)
	for v, ok := iter.Next(); ok; v, ok = iter.Next() {
		split := strings.Split(v.(string), "\t")
		id, _ := strconv.ParseInt(split[0], 10, 32)
		name := split[1]
		albums := make(map[int]*Album)
		artists[int(id)] = &Artist{id: int(id), name: name, albums: albums}
	}

	return true
}

func getAlbums(artistID int) bool {
	req, err := http.NewRequest("GET", serverURL+"getArtist", nil)
	if err != nil {
		log.Fatal(err)
		return false
	}

	params := req.URL.Query()
	params.Add("u", username)
	params.Add("t", token)
	params.Add("s", salt)
	params.Add("v", version)
	params.Add("c", client)
	params.Add("f", "json")
	params.Add("id", strconv.FormatInt(int64(artistID), 10))
	req.URL.RawQuery = params.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
		return false
	}

	query, err := gojq.Parse(`."subsonic-response".artist.album[] |  .id + "\t" + .artist + "\t" + .name + "\t" + (.year|tostring)`)
	if err != nil {
		log.Fatal(err)
		return false
	}

	var resJSON map[string]interface{}
	json.Unmarshal(body, &resJSON)

	iter := query.Run(resJSON)
	for v, ok := iter.Next(); ok; v, ok = iter.Next() {
		split := strings.Split(v.(string), "\t")
		id, _ := strconv.ParseInt(split[0], 10, 32)
		artistName := split[1]
		name := split[2]
		year, _ := strconv.ParseInt(split[3], 10, 32)
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
	req, err := http.NewRequest("GET", serverURL+"getAlbum", nil)
	if err != nil {
		log.Fatal(err)
		return false
	}

	params := req.URL.Query()
	params.Add("u", username)
	params.Add("t", token)
	params.Add("s", salt)
	params.Add("v", version)
	params.Add("c", client)
	params.Add("f", "json")
	params.Add("id", strconv.FormatInt(int64(albumID), 10))
	req.URL.RawQuery = params.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
		return false
	}

	// ."subsonic-response".album.song[]

	var resJSON map[string]interface{}
	json.Unmarshal(body, &resJSON)

	query, err := gojq.Parse(`."subsonic-response".album.song[]`)
	if err != nil {
		log.Fatal(err)
		return false
	}

	artistID := 0

	iter := query.Run(resJSON)
	for track, ok := iter.Next(); ok; track, ok = iter.Next() {
		trackMap := track.(map[string]any)
		if artistID == 0 {
			artistID = toInt(trackMap["artistId"].(string))
		}

		var track float64
		if track, ok = trackMap["track"].(float64); !ok {
			track = 0
		}

		newTrack := Track{
			id:       toInt(trackMap["id"].(string)),
			title:    trackMap["title"].(string),
			album:    trackMap["album"].(string),
			albumID:  albumID,
			artist:   trackMap["artist"].(string),
			artistID: toInt(trackMap["artistId"].(string)),
			track:    int(track),
			duration: int(trackMap["duration"].(float64))}

		artists[artistID].albums[albumID].tracks[newTrack.id] = &newTrack
	}

	return true
}

func stream(trackID int) bool {
	req, err := http.NewRequest("GET", serverURL+"stream", nil)
	if err != nil {
		log.Fatal(err)
		return false
	}

	params := req.URL.Query()
	params.Add("u", username)
	params.Add("t", token)
	params.Add("s", salt)
	params.Add("v", version)
	params.Add("c", client)
	params.Add("f", "json")
	params.Add("id", strconv.FormatInt(int64(trackID), 10))
	req.URL.RawQuery = params.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer res.Body.Close()

	file, err := os.Create(strconv.FormatInt(int64(trackID), 10) + ".mp3")
	if err != nil {
		log.Fatal(err)
		return false
	}

	size, err := io.Copy(file, res.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created %d.mp3 with size %d\n", trackID, size)

	return true
}

func toInt(s string) int {
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return -1
	}
	return int(n)
}
