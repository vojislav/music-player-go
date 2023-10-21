package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var searchList *tview.List
var searchInput *tview.InputField
var searchIndexes []int
var searchCurrentIndex int

func nextSearchResult() {
	if len(searchIndexes) != 0 {
		currentHighlightedItemIndex := searchList.GetCurrentItem()
		searchCurrentIndex = upper_bound(searchIndexes, currentHighlightedItemIndex) % len(searchIndexes)
		searchList.SetCurrentItem(searchIndexes[searchCurrentIndex])
	}
}

func previousSearchResult() {
	if len(searchIndexes) != 0 {
		currentHighlightedItemIndex := searchList.GetCurrentItem()
		arrayLen := len(searchIndexes)
		searchCurrentIndex = upper_bound_reverse(searchIndexes, currentHighlightedItemIndex)
		searchCurrentIndex = (searchCurrentIndex + arrayLen) % arrayLen
		searchList.SetCurrentItem(searchIndexes[searchCurrentIndex])
	}
}

func searchInputHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		bottomPanel.SwitchToPage("current track info")
		app.SetFocus(searchList)
		searchInput.SetText("") // search input shouldn't persist for next search
		return nil

	case tcell.KeyEnter:
		searchString := searchInput.GetText()
		if len(searchString) == 0 {
			searchIndexes = nil
			go searchStatus("Search cleared", "")
		} else {
			searchIndexes = searchList.FindItems(searchString, "-", false, true)
			if len(searchIndexes) == 0 {
				go searchStatus("No results found!", "")
			} else {
				searchList.SetCurrentItem(searchIndexes[0])
				go searchStatus("Searching: ", searchString)
			}
		}

		bottomPanel.SwitchToPage("current track info")
		app.SetFocus(searchList)
		searchInput.SetText("")

		return nil
	}

	return event
}

func searchStatus(message, searchString string) {
	currentTrackText.Clear()
	fmt.Fprint(currentTrackText, message, searchString)
	time.Sleep(2 * time.Second)
	updateCurrentTrackText()
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
