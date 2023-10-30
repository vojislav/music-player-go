package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// if search is cancelled with Esc, user is returned to searchStartContext index of currently searched list
var searchStartContext = -1
var searchList *tview.List
var searchInput *tview.InputField
var searchIndexes []int

func nextSearchResult() {
	if len(searchIndexes) == 0 {
		return
	}

	currentHighlightedItemIndex := searchList.GetCurrentItem()
	searchCurrentIndex := upper_bound(searchIndexes, currentHighlightedItemIndex) % len(searchIndexes)
	searchList.SetCurrentItem(searchIndexes[searchCurrentIndex])
}

func previousSearchResult() {
	if len(searchIndexes) == 0 {
		return
	}

	currentHighlightedItemIndex := searchList.GetCurrentItem()
	arrayLen := len(searchIndexes)
	searchCurrentIndex := upper_bound_reverse(searchIndexes, currentHighlightedItemIndex)
	searchCurrentIndex = (searchCurrentIndex + arrayLen) % arrayLen
	searchList.SetCurrentItem(searchIndexes[searchCurrentIndex])
}

// called every time on searchInput change
func searchIncremental(text string) {
	if len(text) == 0 {
		return
	}

	searchIndexes = searchList.FindItems(text, "", true, true)
	if len(searchIndexes) == 0 {
		go searchStatus("No results found!", "")
		return
	}

	// if current highlighted item matches, we don't need to go further. else: next search result
	currentIndex := searchList.GetCurrentItem()
	found, _ := binary_search(searchIndexes, currentIndex)
	if !found {
		nextSearchResult()
	}
}

// restores context prior to search.
// search results are cleared, 'n'/'N' aren't available until next search
func cancelSearch() {
	searchIndexes = nil
	searchList.SetCurrentItem(searchStartContext)
	go searchStatus("Search cleared", "")

	closeSearch()
}

// closes search bar, returning focus to list where search was initiated.
// search results are persisted, user can still use 'n'/'N'
func closeSearch() {
	searchStartContext = -1
	bottomPage.SwitchToPage("current track info")
	restoreFocus()
	searchInput.SetText("")
}

func searchInputHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		cancelSearch()
		return nil

	case tcell.KeyEnter:
		closeSearch()
		return nil
	}

	return event
}

func searchStatus(message, searchString string) {
	currentTrackText.Clear()
	fmt.Fprint(currentTrackText, message, searchString)
	time.Sleep(2 * time.Second)
	startUpdate <- true
	app.Draw()
}

func upper_bound(array []int, target int) int {
	low, high, mid := 0, len(array)-1, 0

	for low <= high {
		mid = (low + high) / 2
		if array[mid] > target {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return low
}

func upper_bound_reverse(array []int, target int) int {
	low, high, mid := 0, len(array)-1, 0
	result := -1

	for low <= high {
		mid = (low + high) / 2
		if array[mid] < target {
			low = mid + 1
			result = mid
		} else {
			high = mid - 1
		}
	}
	return result
}
