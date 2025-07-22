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

	_, err := db.Exec(string(dbInitScript))
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

	rows, _ := db.Query("SELECT * FROM albums WHERE artistID=?  ORDER BY year, name", artistID)
	return rows
}

func queryAlbumTracks(albumID int) *sql.Rows {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM tracks WHERE albumID=? ORDER BY disc, track", albumID)
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

func queryArtistAndAlbum(trackID int) *sql.Row {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	row := db.QueryRow("SELECT artist, album FROM tracks WHERE id=?", trackID)
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

func queryDuration(trackID string) int {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	row := db.QueryRow("SELECT duration FROM tracks WHERE id=?", trackID)

	var duration int
	row.Scan(&duration)

	return duration
}

func queryArtistAndTitleAndDuration(trackID int) *sql.Row {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	row := db.QueryRow("SELECT artist, title, duration FROM tracks WHERE id=?", trackID)
	return row
}

func queryArtistAndTitle(trackID int) *sql.Row {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	row := db.QueryRow("SELECT artist, title FROM tracks WHERE id=?", trackID)
	return row
}

func getAlbumID(trackID string) int {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	var albumID int
	db.QueryRow("SELECT albumID FROM tracks WHERE id=?", trackID).Scan(&albumID)

	return albumID
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

	trackQuery, err := db.Prepare("INSERT OR IGNORE INTO tracks(id, title, album, artist, track, year, genre, size, suffix, duration, bitrate, disc, albumID, artistID) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		fmt.Println(err)
		return
	}

	playlistQuery, err := db.Prepare("INSERT OR IGNORE INTO playlists(id, name, comment, owner, public, songCount, duration, created, changed, coverArt) VALUES(?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		fmt.Println(err)
		return
	}

	trackCount, err := db.Prepare("SELECT COUNT(*) FROM tracks")
	if err != nil {
		fmt.Println(err)
		return
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
				_, err := trackQuery.Exec(k, v.Title, v.Album, v.Artist, v.Track, v.Year, v.Genre, v.Size, v.Suffix, v.Duration, v.BitRate, v.Disc, v.AlbumID, v.ArtistID)
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

var dbInitScript = `DROP TABLE IF EXISTS artists;
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
	disc INTEGER,
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
