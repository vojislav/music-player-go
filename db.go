package main

import (
	_ "github.com/mattn/go-sqlite3"
)

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
