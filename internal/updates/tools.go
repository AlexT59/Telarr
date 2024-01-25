package updates

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"telarr/internal/types"

	"gitlab.com/toby3d/telegram"
)

type mediaType string

const (
	mediaTypeMovie mediaType = "movie"
	mediaTypeSerie mediaType = "serie"
)

const pathForDiskUsage = "."

// getMsgPageInfo returns the current page and the total number of pages.
func getMsgPageInfo(message string) (int, int, error) {
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
			telegram.NewInlineKeyboardRow(telegram.NewInlineKeyboardButton("ℹ️ Show movie details", types.CallbackMovieDetails.String())),
			telegram.NewInlineKeyboardRow(telegram.NewInlineKeyboardButton("🗑 Remove movie", types.CallbackRemoveMovie.String())),
		)
	} else if mediaType == mediaTypeSerie {
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard,
			telegram.NewInlineKeyboardRow(telegram.NewInlineKeyboardButton("ℹ️ Show serie details", types.CallbackSerieDetails.String())),
			telegram.NewInlineKeyboardRow(telegram.NewInlineKeyboardButton("🗑 Remove serie", types.CallbackRemoveSerie.String())),
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
			row = append(row, telegram.NewInlineKeyboardButton("Next ->", types.CallbackNextMovie.String()))
			if totalPages > 2 {
				row = append(row, telegram.NewInlineKeyboardButton(">>", types.CallbackLastMovie.String()))
			}
		} else if pageNb == totalPages {
			if totalPages > 2 {
				row = append(row, telegram.NewInlineKeyboardButton("<<", types.CallbackFirstMovie.String()))
			}
			row = append(row, telegram.NewInlineKeyboardButton("<- Previous", types.CallbackPreviousMovie.String()))
		} else {
			if totalPages > 2 && pageNb > 2 {
				row = append(row, telegram.NewInlineKeyboardButton("<<", types.CallbackFirstMovie.String()))
			}
			row = append(row, telegram.NewInlineKeyboardButton("<- Previous", types.CallbackPreviousMovie.String()), telegram.NewInlineKeyboardButton("Next ->", types.CallbackNextMovie.String()))
			if totalPages > 2 && pageNb < totalPages-1 {
				row = append(row, telegram.NewInlineKeyboardButton(">>", types.CallbackLastMovie.String()))
			}
		}
	} else if mediaType == mediaTypeSerie {
		if pageNb == 1 {
			row = append(row, telegram.NewInlineKeyboardButton("Next ->", types.CallbackNextSerie.String()))
			if totalPages > 2 {
				row = append(row, telegram.NewInlineKeyboardButton(">>", types.CallbackLastSerie.String()))
			}
		} else if pageNb == totalPages {
			if totalPages > 2 {
				row = append(row, telegram.NewInlineKeyboardButton("<<", types.CallbackFirstSerie.String()))
			}
			row = append(row, telegram.NewInlineKeyboardButton("<- Previous", types.CallbackPreviousSerie.String()))
		} else {
			if totalPages > 2 && pageNb > 2 {
				row = append(row, telegram.NewInlineKeyboardButton("<<", types.CallbackFirstSerie.String()))
			}
			row = append(row, telegram.NewInlineKeyboardButton("<- Previous", types.CallbackPreviousSerie.String()), telegram.NewInlineKeyboardButton("Next ->", types.CallbackNextSerie.String()))
			if totalPages > 2 && pageNb < totalPages-1 {
				row = append(row, telegram.NewInlineKeyboardButton(">>", types.CallbackLastSerie.String()))
			}
		}
	}

	return telegram.NewInlineKeyboardMarkup(row)
}

func getConfirmRemoveKeyboard(mediaType mediaType) telegram.InlineKeyboardMarkup {
	if mediaType == mediaTypeMovie {
		return telegram.NewInlineKeyboardMarkup(
			telegram.NewInlineKeyboardRow(telegram.NewInlineKeyboardButton("Confirm ✅", types.CallbackConfirmRemoveMovie.String()), telegram.NewInlineKeyboardButton("Cancel ❌", types.CallbackCancelRemoveMovie.String())),
		)
	} else if mediaType == mediaTypeSerie {
		return telegram.NewInlineKeyboardMarkup(
			telegram.NewInlineKeyboardRow(telegram.NewInlineKeyboardButton("Confirm ✅", types.CallbackConfirmRemoveSerie.String()), telegram.NewInlineKeyboardButton("Cancel ❌", types.CallbackCancelRemoveSerie.String())),
		)
	}

	return telegram.InlineKeyboardMarkup{}
}

func getAddMediaKeyboard(pageNb int, totalPages int, mediaType mediaType, addable bool) telegram.InlineKeyboardMarkup {
	var addRow []*telegram.InlineKeyboardButton
	if mediaType == mediaTypeMovie {
		addRow = append(addRow, telegram.NewInlineKeyboardButton("Add to Radarr 🎬", types.CallbackAddMovie.String()))
	} else if mediaType == mediaTypeSerie {
		addRow = append(addRow, telegram.NewInlineKeyboardButton("Add to Sonarr 📺", types.CallbackAddSerie.String()))
	}

	var editRow []*telegram.InlineKeyboardButton
	if mediaType == mediaTypeMovie {
		editRow = []*telegram.InlineKeyboardButton{
			telegram.NewInlineKeyboardButton("Edit request 🔍", types.CallbackEditRequestAddMovie.String()),
			telegram.NewInlineKeyboardButton("Cancel ❌", types.CallbackCancel.String()),
		}
	} else if mediaType == mediaTypeSerie {
		editRow = []*telegram.InlineKeyboardButton{
			telegram.NewInlineKeyboardButton("Edit request 🔍", types.CallbackEditRequestAddSerie.String()),
			telegram.NewInlineKeyboardButton("Cancel ❌", types.CallbackCancel.String()),
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
			navRow = append(navRow, telegram.NewInlineKeyboardButton("Next ->", types.CallbackNextAddMovie.String()))
		} else if pageNb == totalPages {
			navRow = append(navRow, telegram.NewInlineKeyboardButton("<- Previous", types.CallbackPreviousAddMovie.String()))
		} else {
			navRow = append(navRow, telegram.NewInlineKeyboardButton("<- Previous", types.CallbackPreviousAddMovie.String()), telegram.NewInlineKeyboardButton("Next ->", types.CallbackNextAddMovie.String()))
		}
	} else if mediaType == mediaTypeSerie {
		if pageNb == 1 {
			navRow = append(navRow, telegram.NewInlineKeyboardButton("Next ->", types.CallbackNextAddSerie.String()))
		} else if pageNb == totalPages {
			navRow = append(navRow, telegram.NewInlineKeyboardButton("<- Previous", types.CallbackPreviousAddSerie.String()))
		} else {
			navRow = append(navRow, telegram.NewInlineKeyboardButton("<- Previous", types.CallbackPreviousAddSerie.String()), telegram.NewInlineKeyboardButton("Next ->", types.CallbackNextAddSerie.String()))
		}
	}

	if addable {
		return telegram.NewInlineKeyboardMarkup(navRow, addRow, editRow)
	}
	return telegram.NewInlineKeyboardMarkup(navRow, editRow)
}

func getQualityProfileKeyboard(profiles []types.QualityProfile) telegram.ReplyKeyboardMarkup {
	var buttons [][]*telegram.KeyboardButton
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})
	for i := 0; i < len(profiles); i += 2 {
		butRow := []*telegram.KeyboardButton{{Text: profiles[i].Name}}
		if i+1 < len(profiles) {
			butRow = append(butRow, &telegram.KeyboardButton{Text: profiles[i+1].Name})
		}
		buttons = append(buttons, butRow)
	}
	keyboard := telegram.ReplyKeyboardMarkup{
		OneTimeKeyboard: true,
		ResizeKeyboard:  true,
		Keyboard:        buttons,
	}

	return keyboard
}

func getFollowDownloadingStatusKeyboard(followButtonInstedOfStopRefresh bool) telegram.InlineKeyboardMarkup {
	kRow := []*telegram.InlineKeyboardButton{
		telegram.NewInlineKeyboardButton("Refresh now 🔄", types.CallbackRefreshDownloadingStatusMovie.String()),
	}

	if followButtonInstedOfStopRefresh {
		kRow = append(kRow, telegram.NewInlineKeyboardButton("Follow downloading status 📡", types.CallbackFollowDownloadingStatusMovie.String()))
	} else {
		kRow = append(kRow, telegram.NewInlineKeyboardButton("Stop refreshing ⏹", types.CallbackCancelFollowDownloadingStatusMovie.String()))
	}

	return telegram.NewInlineKeyboardMarkup(kRow)
}

func getDiskUsage() (types.DiskStatus, error) {
	var disk types.DiskStatus

	fs := syscall.Statfs_t{}
	err := syscall.Statfs(pathForDiskUsage, &fs)
	if err != nil {
		return types.DiskStatus{}, err
	}

	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free

	return disk, nil
}
