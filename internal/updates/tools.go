package updates

import (
	"errors"
	"strconv"
	"strings"
	"unicode"

	"gitlab.com/toby3d/telegram"
)

type mediaType string

const (
	mediaTypeMovies mediaType = "movie"
	mediaTypeSeries mediaType = "serie"
)

// getPageStatus returns the current page and the total number of pages.
func getPageStatus(message string) (int, int, error) {
	// get "page x/y" from the message and get only the x and y
	pageStr, found := strings.CutPrefix(message[strings.LastIndex(message, "\n")+1:], "page ")
	if !found {
		return 0, 0, errors.New("page not found")
	}

	pageNbStr := strings.Split(pageStr, "/")[0]
	pageNb, err := strconv.Atoi(pageNbStr)
	if err != nil {
		return 0, 0, errors.Join(err, errors.New("error when converting page number (pageNbStr: "+pageNbStr+")"))
	}

	totalPagesStr := strings.Split(pageStr, "/")[1]
	totalPages, err := strconv.Atoi(totalPagesStr)
	if err != nil {
		return 0, 0, errors.Join(err, errors.New("error when converting total pages number (totalPagesStr: "+totalPagesStr+")"))
	}

	return pageNb, totalPages, nil
}

// getMediaListKeyboard returns the keyboard for the media type (movie or serie) to navigate between pages and show the details of a media.
func getMediaListKeyboard(pageNb int, totalPages int, mediaType mediaType) telegram.InlineKeyboardMarkup {
	mediaStr := string(mediaType)

	r := []rune(mediaStr)
	capitalizedMediaStr := string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...))

	keyboard := getNavigationKeyboard(pageNb, totalPages, mediaType)
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard,
		telegram.NewInlineKeyboardRow(telegram.NewInlineKeyboardButton("ℹ️ Show "+mediaStr+" details", mediaStr+"Details")),
		telegram.NewInlineKeyboardRow(telegram.NewInlineKeyboardButton("🗑 Remove "+mediaStr, "remove"+capitalizedMediaStr)),
	)
	return keyboard
}

// getNavigationKeyboard returns the navigation keyboard for the media type to navigate between pages.
func getNavigationKeyboard(pageNb int, totalPages int, mediaType mediaType) telegram.InlineKeyboardMarkup {
	var row = telegram.NewInlineKeyboardRow()

	if totalPages <= 1 {
		return telegram.InlineKeyboardMarkup{}
	}

	r := []rune(string(mediaType))
	mediaStr := string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...))

	if pageNb == 1 {
		row = append(row, telegram.NewInlineKeyboardButton("Next ->", "next"+mediaStr))
		if totalPages > 2 {
			row = append(row, telegram.NewInlineKeyboardButton(">>", "last"+mediaStr))
		}
	} else if pageNb == totalPages {
		if totalPages > 2 {
			row = append(row, telegram.NewInlineKeyboardButton("<<", "first"+mediaStr))
		}
		row = append(row, telegram.NewInlineKeyboardButton("<- Previous", "previous"+mediaStr))
	} else {
		if totalPages > 2 && pageNb > 2 {
			row = append(row, telegram.NewInlineKeyboardButton("<<", "first"+mediaStr))
		}
		row = append(row, telegram.NewInlineKeyboardButton("<- Previous", "previous"+mediaStr), telegram.NewInlineKeyboardButton("Next ->", "next"+mediaStr))
		if totalPages > 2 && pageNb < totalPages-1 {
			row = append(row, telegram.NewInlineKeyboardButton(">>", "last"+mediaStr))
		}
	}

	return telegram.NewInlineKeyboardMarkup(row)
}

func getConfirmRemoveKeyboard(mediaType mediaType) telegram.InlineKeyboardMarkup {
	r := []rune(string(mediaType))
	mediaStr := string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...))

	return telegram.NewInlineKeyboardMarkup(
		telegram.NewInlineKeyboardRow(telegram.NewInlineKeyboardButton("Confirm ✅", "confirmRemove"+mediaStr), telegram.NewInlineKeyboardButton("Cancel ❌", "cancelRemove"+mediaStr)),
	)
}
