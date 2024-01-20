package updates

import (
	"sort"
	"strconv"
	"strings"
	"telarr/configuration"
	"telarr/internal/radarr"
	"telarr/internal/sonarr"
	"telarr/internal/types"

	"github.com/rs/zerolog/log"
	"gitlab.com/toby3d/telegram"
)

// callbacks is the struct designed to handle the callbacks.
type callbacks struct {
	// Bot is the telegram bot.
	bot *telegram.Bot

	radarrConfig configuration.Radarr
	sonarrConfig configuration.Sonarr

	// list of users actions
	usersAction map[int]types.Action
	// list of data to navigate between pages
	usersData     map[int]interface{}
	usersCurrPage map[int]int
}

func (cb *callbacks) handle(rcvCallback *telegram.CallbackQuery) {
	if rcvCallback == nil {
		return
	}

	log.Debug().Str("username", rcvCallback.From.Username).Str("callback", rcvCallback.Data).Msg("callback received")

	var err error

	// if the callback is about the movies list
	var films []radarr.Film
	var series []sonarr.Serie
	var messages []string
	if strings.Contains(strings.ToLower(rcvCallback.Data), string(mediaTypeMovie)) {
		log.Trace().Str("username", rcvCallback.From.Username).Msg("getting movies list")
		films, err = radarr.GetFilmsList(cb.radarrConfig)
		if err != nil {
			log.Err(err).Msg("error when getting movies messages")
			editSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, "An error occurred while getting the movies list.\nPlease contact the administrator.")
			return
		}
		messages = printMoviesList(films)
	}
	if strings.Contains(strings.ToLower(rcvCallback.Data), string(mediaTypeSerie)) {
		log.Trace().Str("username", rcvCallback.From.Username).Msg("getting series list")
		series, err = sonarr.GetSeriesList(cb.sonarrConfig)
		if err != nil {
			log.Err(err).Msg("error when getting movies messages")
			editSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, "An error occurred while getting the series list.\nPlease contact the administrator.")
			return
		}
		messages = printSeriesList(series)
	}

	callback := types.CallbackAction(rcvCallback.Data)
	switch callback {
	/* Movies */
	// get the next page of the movies list
	case types.CallbackNextMovie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing next page of movies list")

		// get the current page and the total number of pages
		pageNb, totalPages, err := getPageStatus(rcvCallback.Message.Text)
		if err != nil {
			log.Err(err).Msg("error when getting page status")
			return
		}

		keyboard := getMediaListKeyboard(pageNb+1, totalPages, mediaTypeMovie)
		editMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, messages[pageNb+1-1]+printPageNum(pageNb+1, len(messages)), &keyboard)
	// get the previous page of the movies list
	case types.CallbackPreviousMovie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing previous page of movies list")

		// get the current page and the total number of pages
		pageNb, totalPages, err := getPageStatus(rcvCallback.Message.Text)
		if err != nil {
			log.Err(err).Msg("error when getting page status")
			return
		}

		keyboard := getMediaListKeyboard(pageNb-1, totalPages, mediaTypeMovie)
		editMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, messages[pageNb-1-1]+printPageNum(pageNb-1, len(messages)), &keyboard)
	// get the first page of the movies list
	case types.CallbackFirstMovie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing first page of movies list")

		keyboard := getMediaListKeyboard(1, len(messages), mediaTypeMovie)
		editMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, messages[0]+printPageNum(1, len(messages)), &keyboard)
	// get the last page of the movies list
	case types.CallbackLastMovie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing last page of movies list")

		keyboard := getMediaListKeyboard(len(messages), len(messages), mediaTypeMovie)
		editMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, messages[len(messages)-1]+printPageNum(len(messages), len(messages)), &keyboard)
	case types.CallbackMovieDetails:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing movie details")

		// show the movies list into a keyboard
		var buttons [][]*telegram.KeyboardButton
		sort.Slice(films, func(i, j int) bool {
			return films[i].Title < films[j].Title
		})
		for i := 0; i < len(films); i += 2 {
			butRow := []*telegram.KeyboardButton{{Text: films[i].Title}}
			if i+1 < len(films) {
				butRow = append(butRow, &telegram.KeyboardButton{Text: films[i+1].Title})
			}
			buttons = append(buttons, butRow)
		}
		keyboard := telegram.ReplyKeyboardMarkup{
			OneTimeKeyboard: true,
			ResizeKeyboard:  true,
			Keyboard:        buttons,
		}
		sent := sendMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, "Select the movie or write his name it", keyboard)
		if sent {
			cb.usersAction[rcvCallback.From.ID] = types.CallbackMovieDetails
		}
	case types.CallbackBackToMoviesList:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("back to movies list")

		// remove the two last messages
		cb.bot.DeleteMessage(rcvCallback.Message.Chat.ID, rcvCallback.Message.ID)
		cb.bot.DeleteMessage(rcvCallback.Message.Chat.ID, rcvCallback.Message.ID-1)

		// show the movies list
		sendMoviesList(cb.bot, rcvCallback.Message, cb.radarrConfig)
	case types.CallbackRemoveMovie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("show movies list to remove")

		// show the movies list into a keyboard
		var buttons [][]*telegram.KeyboardButton
		sort.Slice(films, func(i, j int) bool {
			return films[i].Title < films[j].Title
		})
		for i := 0; i < len(films); i += 2 {
			butRow := []*telegram.KeyboardButton{{Text: films[i].Title}}
			if i+1 < len(films) {
				butRow = append(butRow, &telegram.KeyboardButton{Text: films[i+1].Title})
			}
			buttons = append(buttons, butRow)
		}
		keyboard := telegram.ReplyKeyboardMarkup{
			OneTimeKeyboard: true,
			ResizeKeyboard:  true,
			Keyboard:        buttons,
		}
		sent := sendMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, "Select the movie or write his name it", keyboard)
		if sent {
			cb.usersAction[rcvCallback.From.ID] = types.CallbackRemoveMovie
		}
	case types.CallbackConfirmRemoveMovie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("confirm remove movie")

		// remove the last message
		cb.bot.DeleteMessage(rcvCallback.Message.Chat.ID, rcvCallback.Message.ID)

		// get the TMDb ID of the movie
		movieIdStr, found := strings.CutPrefix(strings.Split(rcvCallback.Message.Text, "\n")[1], "MovieId: ") // get the second line and remove the "MovieId: " prefix
		if !found {
			log.Warn().Str("username", rcvCallback.From.Username).Msg("TMDb ID not found")
			sendSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, "An error occurred while removing the movie.\nPlease contact the administrator.")
			return
		}
		movieId, err := strconv.Atoi(movieIdStr)
		if err != nil {
			log.Err(err).Msg("error when converting TMDb ID")
			sendSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, "An error occurred while removing the movie.\nPlease contact the administrator.")
			return
		}

		// get the movie name
		movieName, err := radarr.GetMovieName(cb.radarrConfig, movieId)
		if err != nil {
			log.Err(err).Msg("error when getting movie name")
			sendSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, "An error occurred while removing the movie.\nPlease contact the administrator.")
			return
		}

		// remove the movie
		err = radarr.RemoveFilm(cb.radarrConfig, movieId)
		if err != nil {
			log.Err(err).Msg("error when removing movie")
			sendSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, "An error occurred while removing the movie.\nPlease contact the administrator.")
			return
		}

		log.Debug().Str("movieName", movieName).Str("username", rcvCallback.From.Username).Msg("movie removed successfully")

		sendSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, "Movie *"+movieName+"* removed successfully! ✅")
	case types.CallbackCancelRemoveMovie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("cancel remove movie")

		// remove the last message
		cb.bot.DeleteMessage(rcvCallback.Message.Chat.ID, rcvCallback.Message.ID)

		sendSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, "Movie not removed! ✅")
	case types.CallbackNextAddMovie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing next page of add media")

		ok := checkUserAction(cb, rcvCallback.From, rcvCallback.Message)
		if !ok {
			return
		}

		cb.usersCurrPage[rcvCallback.From.ID]++
		pageNb := cb.usersCurrPage[rcvCallback.From.ID]
		films := (cb.usersData[rcvCallback.From.ID].([]radarr.Film))
		film := films[pageNb-1]
		keyboard := getAddMediaKeyboard(pageNb, len(films), mediaTypeMovie, !film.IsInLibrary)
		str := film.PrintMovieTitle()
		if film.IsInLibrary {
			str += "\n\nAlready in library ✅"
		}
		editImageMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, film.CoverImage, str, &keyboard)
	case types.CallbackPreviousAddMovie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing previous page of add media")

		ok := checkUserAction(cb, rcvCallback.From, rcvCallback.Message)
		if !ok {
			return
		}

		cb.usersCurrPage[rcvCallback.From.ID]--
		pageNb := cb.usersCurrPage[rcvCallback.From.ID]
		films := (cb.usersData[rcvCallback.From.ID].([]radarr.Film))
		film := films[pageNb-1]
		keyboard := getAddMediaKeyboard(pageNb, len(films), mediaTypeMovie, !film.IsInLibrary)
		str := film.PrintMovieTitle()
		if film.IsInLibrary {
			str += "\n\nAlready in library ✅"
		}
		editImageMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, film.CoverImage, str, &keyboard)
	case types.CallbackEditRequestAddMovie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("edit request movie")

		// remove the last message
		cb.bot.DeleteMessage(rcvCallback.Message.Chat.ID, rcvCallback.Message.ID)
		cb.usersAction[rcvCallback.From.ID] = types.UserActionLookMovieToAdd
		delete(cb.usersData, rcvCallback.From.ID)
		delete(cb.usersCurrPage, rcvCallback.From.ID)

		sendSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, "Please enter the name of the movie you want to add:")

		/* Series */
	// get the next page of the series list
	case types.CallbackNextSerie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing next page of series list")

		// get the current page and the total number of pages
		pageNb, totalPages, err := getPageStatus(rcvCallback.Message.Text)
		if err != nil {
			log.Err(err).Msg("error when getting page status")
			return
		}

		keyboard := getMediaListKeyboard(pageNb+1, totalPages, mediaTypeSerie)
		editMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, messages[pageNb+1-1]+printPageNum(pageNb+1, len(messages)), &keyboard)
	// get the previous page of the series list
	case types.CallbackPreviousSerie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing previous page of series list")

		// get the current page and the total number of pages
		pageNb, totalPages, err := getPageStatus(rcvCallback.Message.Text)
		if err != nil {
			log.Err(err).Msg("error when getting page status")
			return
		}

		keyboard := getMediaListKeyboard(pageNb-1, totalPages, mediaTypeSerie)
		editMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, messages[pageNb-1-1]+printPageNum(pageNb-1, len(messages)), &keyboard)
	// get the first page of the series list
	case types.CallbackFirstSerie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing first page of series list")

		keyboard := getMediaListKeyboard(1, len(messages), mediaTypeSerie)
		editMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, messages[0]+printPageNum(1, len(messages)), &keyboard)
	// get the last page of the series list
	case types.CallbackLastSerie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing last page of series list")

		keyboard := getMediaListKeyboard(len(messages), len(messages), mediaTypeSerie)
		editMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, messages[len(messages)-1]+printPageNum(len(messages), len(messages)), &keyboard)
	case types.CallbackSerieDetails:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing serie details")

		// show the series list into a keyboard
		var buttons [][]*telegram.KeyboardButton
		sort.Slice(series, func(i, j int) bool {
			return series[i].Title < series[j].Title
		})
		for i := 0; i < len(series); i += 2 {
			butRow := []*telegram.KeyboardButton{{Text: series[i].Title}}
			if i+1 < len(series) {
				butRow = append(butRow, &telegram.KeyboardButton{Text: series[i+1].Title})
			}
			buttons = append(buttons, butRow)
		}
		keyboard := telegram.ReplyKeyboardMarkup{
			OneTimeKeyboard: true,
			ResizeKeyboard:  true,
			Keyboard:        buttons,
		}
		sent := sendMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, "Select the serie or write his name it", keyboard)
		if sent {
			cb.usersAction[rcvCallback.From.ID] = types.CallbackSerieDetails
		}
	case types.CallbackBackToSeriesList:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("back to series list")

		// remove the two last messages
		cb.bot.DeleteMessage(rcvCallback.Message.Chat.ID, rcvCallback.Message.ID)
		cb.bot.DeleteMessage(rcvCallback.Message.Chat.ID, rcvCallback.Message.ID-1)

		// show the series list
		sendSeriesList(cb.bot, rcvCallback.Message, cb.sonarrConfig)
	case types.CallbackRemoveSerie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("show series list to remove")

		// show the series list into a keyboard
		var buttons [][]*telegram.KeyboardButton
		sort.Slice(series, func(i, j int) bool {
			return series[i].Title < series[j].Title
		})
		for i := 0; i < len(series); i += 2 {
			butRow := []*telegram.KeyboardButton{{Text: series[i].Title}}
			if i+1 < len(series) {
				butRow = append(butRow, &telegram.KeyboardButton{Text: series[i+1].Title})
			}
			buttons = append(buttons, butRow)
		}
		keyboard := telegram.ReplyKeyboardMarkup{
			OneTimeKeyboard: true,
			ResizeKeyboard:  true,
			Keyboard:        buttons,
		}
		sent := sendMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, "Select the serie or write his name it", keyboard)
		if sent {
			cb.usersAction[rcvCallback.From.ID] = types.CallbackRemoveSerie
		}
	case types.CallbackConfirmRemoveSerie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("confirm remove serie")

		// remove the last message
		cb.bot.DeleteMessage(rcvCallback.Message.Chat.ID, rcvCallback.Message.ID)

		// get the serie ID
		serieIdStr, found := strings.CutPrefix(strings.Split(rcvCallback.Message.Text, "\n")[1], "SerieId: ") // get the second line and remove the "SerieId: " prefix
		if !found {
			log.Warn().Str("username", rcvCallback.From.Username).Msg("serie ID not found")
			sendSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, "An error occurred while removing the serie.\nPlease contact the administrator.")
			return
		}
		serieId, err := strconv.Atoi(serieIdStr)
		if err != nil {
			log.Err(err).Msg("error when converting serie ID")
			sendSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, "An error occurred while removing the serie.\nPlease contact the administrator.")
			return
		}

		// get the serie name
		serieName, err := sonarr.GetSerieName(cb.sonarrConfig, serieId)
		if err != nil {
			log.Err(err).Msg("error when getting serie name")
			sendSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, "An error occurred while removing the serie.\nPlease contact the administrator.")
			return
		}

		// remove the serie
		err = sonarr.RemoveSerie(cb.sonarrConfig, serieId)
		if err != nil {
			log.Err(err).Msg("error when removing serie")
			sendSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, "An error occurred while removing the serie.\nPlease contact the administrator.")
			return
		}

		log.Debug().Str("serieName", serieName).Str("username", rcvCallback.From.Username).Msg("serie removed successfully")

		sendSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, "Serie *"+serieName+"* removed successfully! ✅")
	case types.CallbackCancelRemoveSerie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("cancel remove serie")

		// remove the last message
		cb.bot.DeleteMessage(rcvCallback.Message.Chat.ID, rcvCallback.Message.ID)

		sendSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, "Serie not removed! ✅")
	case types.CallbackNextAddSerie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing next page of add media")

		ok := checkUserAction(cb, rcvCallback.From, rcvCallback.Message)
		if !ok {
			return
		}

		cb.usersCurrPage[rcvCallback.From.ID]++
		pageNb := cb.usersCurrPage[rcvCallback.From.ID]
		series := (cb.usersData[rcvCallback.From.ID].([]sonarr.Serie))
		serie := series[pageNb-1]
		keyboard := getAddMediaKeyboard(pageNb, len(series), mediaTypeSerie, !serie.IsInLibrary)
		str := serie.PrintSerieTitle()
		if serie.IsInLibrary {
			str += "\n\nAlready in library ✅"
		}
		editImageMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, serie.CoverImage, str, &keyboard)
	case types.CallbackPreviousAddSerie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing previous page of add media")

		ok := checkUserAction(cb, rcvCallback.From, rcvCallback.Message)
		if !ok {
			return
		}

		cb.usersCurrPage[rcvCallback.From.ID]--
		pageNb := cb.usersCurrPage[rcvCallback.From.ID]
		series := (cb.usersData[rcvCallback.From.ID].([]sonarr.Serie))
		serie := series[pageNb-1]
		keyboard := getAddMediaKeyboard(pageNb, len(series), mediaTypeSerie, !serie.IsInLibrary)
		str := serie.PrintSerieTitle()
		if serie.IsInLibrary {
			str += "\n\nAlready in library ✅"
		}
		editImageMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, serie.CoverImage, str, &keyboard)
	case types.CallbackEditRequestAddSerie:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("edit request series")

		// remove the last message
		cb.bot.DeleteMessage(rcvCallback.Message.Chat.ID, rcvCallback.Message.ID)
		cb.usersAction[rcvCallback.From.ID] = types.UserActionLookSerieToAdd
		delete(cb.usersData, rcvCallback.From.ID)
		delete(cb.usersCurrPage, rcvCallback.From.ID)

		sendSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, "Please enter the name of the serie you want to add:")

	/* Common */
	case types.CallbackCancel:
		log.Trace().Str("username", rcvCallback.From.Username).Msg("canceling action")

		// remove the last message
		cb.bot.DeleteMessage(rcvCallback.Message.Chat.ID, rcvCallback.Message.ID)

		sendMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, "Action canceled ✅", telegram.NewReplyKeyboardRemove(false))

	default:
		log.Warn().Str("username", rcvCallback.From.Username).Str("callback", rcvCallback.Data).Msg("unknown callback")
	}
}

// checkUserAction checks if the user has an action in progress.
// Return true if the user has an action in progress, false otherwise.
func checkUserAction(cb *callbacks, user *telegram.User, currentMsg *telegram.Message) bool {
	if cb.usersData[user.ID] == nil {
		log.Warn().Str("username", user.Username).Msg("no data found")
		cb.bot.DeleteMessage(currentMsg.Chat.ID, currentMsg.ID)
		sendSimpleMessage(cb.bot, currentMsg.Chat.ID, "Request timed out.\nPlease try again.")
		return false
	}

	return true
}
