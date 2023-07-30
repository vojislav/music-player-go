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

func queryAlbumTracks(albumID int) *sql.Rows {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM tracks WHERE albumID=? ORDER BY track", albumID)
	return rows
}

func queryTrackInfo(trackID int) *sql.Row {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM tracks WHERE id=?", trackID)
	return row
}

func queryPlaylistTracks(playlistID int) *sql.Rows {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	rows, _ := db.Query("SELECT t.id, a.name, t.title FROM tracks t JOIN artists a ON t.artistID=a.id WHERE t.playlistID=?", playlistID)
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

	trackQuery, err := db.Prepare("INSERT OR IGNORE INTO tracks(id, title, album, artist, track, year, genre, size, suffix, duration, bitrate, albumID, artistID) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		fmt.Println(err)
		return
	}

	playlistQuery, err := db.Prepare("INSERT OR IGNORE INTO playlists(id, name, comment, owner, public, songCount, duration, created, changed, coverArt) VALUES(?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		fmt.Println(err)
		return
	}

	// _, err = db.Prepare("UPDATE tracks SET playlistID=? WHERE id=?")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

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
				_, err := trackQuery.Exec(k, v.Title, v.Album, v.Artist, v.Track, v.Year, v.Genre, v.Size, v.Suffix, v.Duration, v.BitRate, v.AlbumID, v.ArtistID)
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

	playlists := getPlaylists()
	for _, playlist := range playlists {
		_, err := playlistQuery.Exec(playlist.ID, playlist.Name, playlist.Comment, playlist.Owner,
			playlist.Public, playlist.SongCount, playlist.Duration, playlist.Created, playlist.Changed, playlist.CoverArt)
		if err != nil {
			fmt.Println(err)
			return
		}
		playlistTracks := getPlaylistTracks(toInt(playlist.ID))

		os.WriteFile(playlistDirectory+playlist.Name+".json", playlistTracks, 0755)
		// }
		// playlistTracks := getPlaylistTracks(playlist.id)
		// for _, trackID := range playlistTracks {
		// 	_, err := playlistTracksQuery.Exec(playlist.id, trackID)
		// 	if err != nil {
		// 		fmt.Println(err)
		// 		return
		// 	}
		// }
	}
}

func makeInitScript() {
	initScript := `DROP TABLE IF EXISTS artists;
DROP TABLE IF EXISTS albums;
DROP TABLE IF EXISTS tracks;
DROP TABLE IF EXISTS playlists;

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
	album TEXT,
	artist TEXT,
	track INTEGER,
	year INTEGER,
	genre TEXT,
	size INTEGER,
	suffix TEXT,
	duration INTEGER,
	bitrate INTEGER,
	albumID INTEGER,
	artistID INTEGER,
	FOREIGN KEY (artistID) REFERENCES artists(id),
	FOREIGN KEY (albumID) REFERENCES albums(id)
);

CREATE TABLE playlists (
	id INTEGER PRIMARY KEY,
	name TEXT,
	comment TEXT,
	owner TEXT,
	public INTERGER,
	songCount INTERGER,
	duration INTERGER,
	created TEXT,
	changed TEXT,
	coverArt TEXT
);`

	f, err := os.Create(initScriptFile)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err = f.WriteString(initScript)
	if err != nil {
		log.Fatal(err)
	}
}

func getArtistName(artistID int) string {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	var artistName string
	db.QueryRow("SELECT name FROM artists WHERE id=?", artistID).Scan(&artistName)

	return artistName
}
