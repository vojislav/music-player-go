package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	_ "github.com/mattn/go-sqlite3"
)

var sr = beep.SampleRate(44100)

func main() {
	loadDatabase()
	// speaker.Init(sr, sr.N(time.Second/10))

	// var queue Queue
	// speaker.Play(&queue)

	// getArtists()
	// for k, v := range artists {
	// 	fmt.Printf("%d\t%s\n", k, v.name)
	// }

	// var artistID int

	// fmt.Println("Enter artist ID")
	// fmt.Scanln(&artistID)

	// getAlbums(artistID)
	// for k, v := range artists[artistID].albums {
	// 	fmt.Printf("%d\t%s\n", k, v.name)
	// }

	// var albumID int
	// fmt.Println("Enter album ID")
	// fmt.Scanln(&albumID)

	// getTracks(albumID)
	// for k, v := range artists[artistID].albums[albumID].tracks {
	// 	fmt.Printf("%d\t%s\n", k, v.title)
	// }

	// var trackID int
	// fmt.Println("Enter track ID")
	// fmt.Scanln(&trackID)

	// stream(trackID)

	// speaker.Lock()
	// queue.Add(strconv.FormatInt(int64(trackID), 10) + ".mp3")
	// speaker.Unlock()

	// fmt.Scanln()
}

func loadDatabase() {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	artistQuery, err := db.Prepare("INSERT INTO artists(id, name) VALUES(?,?)")
	if err != nil {
		fmt.Println(err)
		return
	}
	albumQuery, err := db.Prepare("INSERT INTO albums(id, artistID, name, year) VALUES(?,?,?,?)")
	if err != nil {
		fmt.Println(err)
		return
	}

	trackQuery, err := db.Prepare("INSERT INTO tracks(id, title, albumID, artistID, track, duration) VALUES(?,?,?,?,?,?)")
	if err != nil {
		fmt.Println(err)
		return
	}

	getArtists()
	for k, v := range artists {
		artistID := k
		_, err := artistQuery.Exec(k, v.name)
		if err != nil {
			fmt.Println(err)
			return
		}
		getAlbums(artistID)
		for k, v := range artists[artistID].albums {
			albumID := k
			_, err := albumQuery.Exec(k, v.artistID, v.name, v.year)
			if err != nil {
				fmt.Println(err)
				return
			}
			getTracks(albumID)
			for k, v := range artists[artistID].albums[albumID].tracks {
				_, err := trackQuery.Exec(k, v.title, v.albumID, v.artistID, v.track, v.duration)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
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
