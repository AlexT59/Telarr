package updates

import (
	"errors"
	"strconv"
	"strings"

	"gitlab.com/toby3d/telegram"
)

type mediaType string

const (
	mediaTypeMovie mediaType = "movie"
	mediaTypeSerie mediaType = "serie"
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
	keyboard := getNavigationKeyboard(pageNb, totalPages, mediaType)
	if mediaType == mediaTypeMovie {
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard,
			telegram.NewInlineKeyboardRow(telegram.NewInlineKeyboardButton("ℹ️ Show movie details", callbackMovieDetails.String())),
			telegram.NewInlineKeyboardRow(telegram.NewInlineKeyboardButton("🗑 Remove movie", callbackRemoveMovie.String())),
		)
	} else if mediaType == mediaTypeSerie {
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard,
			telegram.NewInlineKeyboardRow(telegram.NewInlineKeyboardButton("ℹ️ Show serie details", callbackSerieDetails.String())),
			telegram.NewInlineKeyboardRow(telegram.NewInlineKeyboardButton("🗑 Remove serie", callbackRemoveSerie.String())),
		)
	}
	return keyboard
}

// getNavigationKeyboard returns the navigation keyboard for the media type to navigate between pages.
func getNavigationKeyboard(pageNb int, totalPages int, mediaType mediaType) telegram.InlineKeyboardMarkup {
	var row = telegram.NewInlineKeyboardRow()

	if totalPages <= 1 {
		return telegram.InlineKeyboardMarkup{}
	}

	if mediaType == mediaTypeMovie {
		if pageNb == 1 {
			row = append(row, telegram.NewInlineKeyboardButton("Next ->", callbackNextMovie.String()))
			if totalPages > 2 {
				row = append(row, telegram.NewInlineKeyboardButton(">>", callbackLastMovie.String()))
			}
		} else if pageNb == totalPages {
			if totalPages > 2 {
				row = append(row, telegram.NewInlineKeyboardButton("<<", callbackFirstMovie.String()))
			}
			row = append(row, telegram.NewInlineKeyboardButton("<- Previous", callbackPreviousMovie.String()))
		} else {
			if totalPages > 2 && pageNb > 2 {
				row = append(row, telegram.NewInlineKeyboardButton("<<", callbackFirstMovie.String()))
			}
			row = append(row, telegram.NewInlineKeyboardButton("<- Previous", callbackPreviousMovie.String()), telegram.NewInlineKeyboardButton("Next ->", callbackNextMovie.String()))
			if totalPages > 2 && pageNb < totalPages-1 {
				row = append(row, telegram.NewInlineKeyboardButton(">>", callbackLastMovie.String()))
			}
		}
	} else if mediaType == mediaTypeSerie {
		if pageNb == 1 {
			row = append(row, telegram.NewInlineKeyboardButton("Next ->", callbackNextSerie.String()))
			if totalPages > 2 {
				row = append(row, telegram.NewInlineKeyboardButton(">>", callbackLastSerie.String()))
			}
		} else if pageNb == totalPages {
			if totalPages > 2 {
				row = append(row, telegram.NewInlineKeyboardButton("<<", callbackFirstSerie.String()))
			}
			row = append(row, telegram.NewInlineKeyboardButton("<- Previous", callbackPreviousSerie.String()))
		} else {
			if totalPages > 2 && pageNb > 2 {
				row = append(row, telegram.NewInlineKeyboardButton("<<", callbackFirstSerie.String()))
			}
			row = append(row, telegram.NewInlineKeyboardButton("<- Previous", callbackPreviousSerie.String()), telegram.NewInlineKeyboardButton("Next ->", callbackNextSerie.String()))
			if totalPages > 2 && pageNb < totalPages-1 {
				row = append(row, telegram.NewInlineKeyboardButton(">>", callbackLastSerie.String()))
			}
		}
	}

	return telegram.NewInlineKeyboardMarkup(row)
}

func getConfirmRemoveKeyboard(mediaType mediaType) telegram.InlineKeyboardMarkup {
	if mediaType == mediaTypeMovie {
		return telegram.NewInlineKeyboardMarkup(
			telegram.NewInlineKeyboardRow(telegram.NewInlineKeyboardButton("Confirm ✅", callbackConfirmRemoveMovie.String()), telegram.NewInlineKeyboardButton("Cancel ❌", callbackCancelRemoveMovie.String())),
		)
	} else if mediaType == mediaTypeSerie {
		return telegram.NewInlineKeyboardMarkup(
			telegram.NewInlineKeyboardRow(telegram.NewInlineKeyboardButton("Confirm ✅", callbackConfirmRemoveSerie.String()), telegram.NewInlineKeyboardButton("Cancel ❌", callbackCancelRemoveSerie.String())),
		)
	}

	return telegram.InlineKeyboardMarkup{}
}

func getAddMediaKeyboard(pageNb int, totalPages int, mediaType mediaType, addable bool) telegram.InlineKeyboardMarkup {
	var addRow []*telegram.InlineKeyboardButton
	if mediaType == mediaTypeMovie {
		addRow = append(addRow, telegram.NewInlineKeyboardButton("Add to Radarr 🎬", callbackNextAddMovie.String()))
	} else if mediaType == mediaTypeSerie {
		addRow = append(addRow, telegram.NewInlineKeyboardButton("Add to Sonarr 📺", callbackNextAddSerie.String()))
	}

	var editRow []*telegram.InlineKeyboardButton
	if mediaType == mediaTypeMovie {
		editRow = []*telegram.InlineKeyboardButton{
			telegram.NewInlineKeyboardButton("Edit request 🔍", callbackEditRequestAddMovie.String()),
			telegram.NewInlineKeyboardButton("Cancel ❌", callbackCancel.String()),
		}
	} else if mediaType == mediaTypeSerie {
		editRow = []*telegram.InlineKeyboardButton{
			telegram.NewInlineKeyboardButton("Edit request 🔍", callbackEditRequestAddSerie.String()),
			telegram.NewInlineKeyboardButton("Cancel ❌", callbackCancel.String()),
		}
	}

	if totalPages <= 1 {
		if addable {
			return telegram.NewInlineKeyboardMarkup(addRow, editRow)
		}
		return telegram.NewInlineKeyboardMarkup(editRow)
	}

	var navRow = telegram.NewInlineKeyboardRow()

	if mediaType == mediaTypeMovie {
		if pageNb == 1 {
			navRow = append(navRow, telegram.NewInlineKeyboardButton("Next ->", callbackNextAddMovie.String()))
		} else if pageNb == totalPages {
			navRow = append(navRow, telegram.NewInlineKeyboardButton("<- Previous", callbackPreviousAddMovie.String()))
		} else {
			navRow = append(navRow, telegram.NewInlineKeyboardButton("<- Previous", callbackPreviousAddMovie.String()), telegram.NewInlineKeyboardButton("Next ->", callbackNextAddMovie.String()))
		}
	} else if mediaType == mediaTypeSerie {
		if pageNb == 1 {
			navRow = append(navRow, telegram.NewInlineKeyboardButton("Next ->", callbackNextAddSerie.String()))
		} else if pageNb == totalPages {
			navRow = append(navRow, telegram.NewInlineKeyboardButton("<- Previous", callbackPreviousAddSerie.String()))
		} else {
			navRow = append(navRow, telegram.NewInlineKeyboardButton("<- Previous", callbackPreviousAddSerie.String()), telegram.NewInlineKeyboardButton("Next ->", callbackNextAddSerie.String()))
		}
	}

	if addable {
		return telegram.NewInlineKeyboardMarkup(navRow, addRow, editRow)
	}
	return telegram.NewInlineKeyboardMarkup(navRow, editRow)
}
