DROP TABLE IF EXISTS artists;
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
);