package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func queryArtists() *sql.Rows {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM artists ORDER BY name")
	return rows
}

func queryAlbums(artistID int) *sql.Rows {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM albums WHERE artistID=? ORDER BY year", artistID)
	return rows
}

func queryTracks(albumID int) *sql.Rows {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM tracks WHERE albumID=? ORDER BY track", albumID)
	return rows
}

// func loadDatabase() {
// 	db, err := sql.Open("sqlite3", "./database.db")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	artistQuery, err := db.Prepare("INSERT INTO artists(id, name) VALUES(?,?)")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	albumQuery, err := db.Prepare("INSERT INTO albums(id, artistID, name, year) VALUES(?,?,?,?)")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	trackQuery, err := db.Prepare("INSERT INTO tracks(id, title, albumID, artistID, track, duration) VALUES(?,?,?,?,?,?)")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	getArtists()
// 	for k, v := range artists {
// 		artistID := k
// 		_, err := artistQuery.Exec(k, v.name)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		getAlbums(artistID)
// 		for k, v := range artists[artistID].albums {
// 			albumID := k
// 			_, err := albumQuery.Exec(k, v.artistID, v.name, v.year)
// 			if err != nil {
// 				fmt.Println(err)
// 				return
// 			}
// 			getTracks(albumID)
// 			for k, v := range artists[artistID].albums[albumID].tracks {
// 				_, err := trackQuery.Exec(k, v.title, v.albumID, v.artistID, v.track, v.duration)
// 				if err != nil {
// 					fmt.Println(err)
// 					return
// 				}
// 			}
// 		}
// 	}
// }

// func main() {
// 	db, err := sql.Open("sqlite3", "./database.db")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	query, err := ioutil.ReadFile("./init.sql")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	if _, err := db.Exec(string(query)); err != nil {
// 		panic(err)
// 	}
// }
