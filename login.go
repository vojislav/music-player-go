package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"

	"github.com/rivo/tview"
)

var loginForm *tview.Form

func loginUser() {
	writeConfig()

	if ping() {
		if _, err := os.Stat(databaseFile); err == nil || *reloadDatabaseFlag {
			gotoLibraryPage()
		} else {
			gotoLoadingPage()
		}
	} else {
		deleteConfig()
		loginStatus.Clear()
		fmt.Fprintf(loginStatus, "Login failed!")
	}
}

func setToken(password string) {
	config.Salt = fmt.Sprint(rand.Int())
	token := md5.Sum([]byte(fmt.Sprint(password, config.Salt)))
	config.Token = hex.EncodeToString(token[:])
}
