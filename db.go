package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func createDatabase() {
	db, _ := sql.Open("sqlite3", databaseFile)
	defer db.Close()

	initScript := configDirectory + "init.sql"
	script, err := os.ReadFile(initScript)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(string(script))
	if err != nil {
		log.Fatal(err)
	}
}

func queryArtists() *sql.Rows {
	db, _ := sql.Open("sqlite3", databaseFile)
	defer db.Close()

	rows, err := db.Query("SELECT * FROM artists ORDER BY name")
	if err != nil {
		log.Fatal(err)
	}
	return rows
}

func queryAlbums(artistID int) *sql.Rows {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM albums WHERE artistID=? ORDER BY year", artistID)
	return rows
}

func queryTracks(albumID int) *sql.Rows {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM tracks WHERE albumID=? ORDER BY track", albumID)
	return rows
}

func loadDatabase() {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	artistQuery, err := db.Prepare("INSERT OR IGNORE INTO artists(id, name) VALUES(?,?)")
	if err != nil {
		fmt.Println(err)
		return
	}
	albumQuery, err := db.Prepare("INSERT OR IGNORE INTO albums(id, artistID, name, year) VALUES(?,?,?,?)")
	if err != nil {
		fmt.Println(err)
		return
	}

	trackQuery, err := db.Prepare("INSERT OR IGNORE INTO tracks(id, title, albumID, artistID, track, duration) VALUES(?,?,?,?,?,?)")
	if err != nil {
		fmt.Println(err)
		return
	}

	trackCount, err := db.Prepare("SELECT COUNT(*) FROM tracks")
	if err != nil {
		fmt.Println(err)
		return
	}

	var trackNum int

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
			_ = trackCount.QueryRow().Scan(&trackNum)
			loadingTextBox.Clear()
			fmt.Fprintf(loadingTextBox, "Loading, %d tracks\n", trackNum)
			app.Draw()
		}
	}
}

func makeInitScript() {
	initScript := `DROP TABLE IF EXISTS artists;
DROP TABLE IF EXISTS albums;
DROP TABLE IF EXISTS tracks;

CREATE TABLE artists (
    id INTEGER PRIMARY KEY,
    name TEXT
);

CREATE TABLE albums (
    id INTEGER PRIMARY KEY,
    artistID INTEGER,
    name TEXT,
    year INT,
    FOREIGN KEY (artistID) REFERENCES artists(id)
);

CREATE TABLE tracks (
    id INTEGER PRIMARY KEY,
    title TEXT,
    albumID INTEGER,
    artistID INTEGER,
    track INTEGER,
    duration INTEGER,
    FOREIGN KEY (artistID) REFERENCES artists(id),
    FOREIGN KEY (albumID) REFERENCES albums(id)
);`

	f, err := os.Create(initScriptFile)
	if err != nil {
		log.Fatal("nesto", err)
	}

	defer f.Close()

	_, err = f.WriteString(initScript)
	if err != nil {
		log.Fatal(err)
	}
}
