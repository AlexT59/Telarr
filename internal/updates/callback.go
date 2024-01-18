package updates

import (
	"sort"
	"strings"
	"telarr/configuration"
	"telarr/internal/radarr"
	"telarr/internal/sonarr"

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
	usersAction map[int]string
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
	if strings.Contains(strings.ToLower(rcvCallback.Data), "movie") {
		log.Trace().Str("username", rcvCallback.From.Username).Msg("getting movies list")
		films, err = radarr.GetFilmsList(cb.radarrConfig)
		if err != nil {
			log.Err(err).Msg("error when getting movies messages")
			editSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, "An error occurred while getting the movies list.\nPlease contact the administrator.")
			return
		}
		messages = printMoviesList(films)
	}
	if strings.Contains(strings.ToLower(rcvCallback.Data), "serie") {
		log.Trace().Str("username", rcvCallback.From.Username).Msg("getting series list")
		series, err = sonarr.GetSeriesList(cb.sonarrConfig)
		if err != nil {
			log.Err(err).Msg("error when getting movies messages")
			editSimpleMessage(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, "An error occurred while getting the series list.\nPlease contact the administrator.")
			return
		}
		messages = printSeriesList(series)
	}

	switch rcvCallback.Data {
	/* Movies */
	// get the next page of the movies list
	case "nextMovie":
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing next page of movies list")

		// get the current page and the total number of pages
		pageNb, totalPages, err := getPageStatus(rcvCallback.Message.Text)
		if err != nil {
			log.Err(err).Msg("error when getting page status")
			return
		}

		keyboard := getMediaListKeyboard(pageNb+1, totalPages, mediaTypeMovies)
		editMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, messages[pageNb+1-1]+printPageNum(pageNb+1, len(messages)), &keyboard)
	// get the previous page of the movies list
	case "previousMovie":
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing previous page of movies list")

		// get the current page and the total number of pages
		pageNb, totalPages, err := getPageStatus(rcvCallback.Message.Text)
		if err != nil {
			log.Err(err).Msg("error when getting page status")
			return
		}

		keyboard := getMediaListKeyboard(pageNb-1, totalPages, mediaTypeMovies)
		editMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, messages[pageNb-1-1]+printPageNum(pageNb-1, len(messages)), &keyboard)
	// get the first page of the movies list
	case "firstMovie":
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing first page of movies list")

		keyboard := getMediaListKeyboard(1, len(messages), mediaTypeMovies)
		editMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, messages[0]+printPageNum(1, len(messages)), &keyboard)
	// get the last page of the movies list
	case "lastMovie":
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing last page of movies list")

		keyboard := getMediaListKeyboard(len(messages), len(messages), mediaTypeMovies)
		editMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, messages[len(messages)-1]+printPageNum(len(messages), len(messages)), &keyboard)
	case "movieDetails":
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
			cb.usersAction[rcvCallback.From.ID] = "movieDetails"
		}
	case "backToMoviesList":
		log.Trace().Str("username", rcvCallback.From.Username).Msg("back to movies list")

		// remove the two last messages
		cb.bot.DeleteMessage(rcvCallback.Message.Chat.ID, rcvCallback.Message.ID)
		cb.bot.DeleteMessage(rcvCallback.Message.Chat.ID, rcvCallback.Message.ID-1)

		// show the movies list
		sendMoviesList(cb.bot, rcvCallback.Message, cb.radarrConfig)

		/* Series */
	// get the next page of the series list
	case "nextSerie":
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing next page of series list")

		// get the current page and the total number of pages
		pageNb, totalPages, err := getPageStatus(rcvCallback.Message.Text)
		if err != nil {
			log.Err(err).Msg("error when getting page status")
			return
		}

		keyboard := getMediaListKeyboard(pageNb+1, totalPages, mediaTypeSeries)
		editMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, messages[pageNb+1-1]+printPageNum(pageNb+1, len(messages)), &keyboard)
	// get the previous page of the series list
	case "previousSerie":
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing previous page of series list")

		// get the current page and the total number of pages
		pageNb, totalPages, err := getPageStatus(rcvCallback.Message.Text)
		if err != nil {
			log.Err(err).Msg("error when getting page status")
			return
		}

		keyboard := getMediaListKeyboard(pageNb-1, totalPages, mediaTypeSeries)
		editMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, messages[pageNb-1-1]+printPageNum(pageNb-1, len(messages)), &keyboard)
	// get the first page of the series list
	case "firstSerie":
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing first page of series list")

		keyboard := getMediaListKeyboard(1, len(messages), mediaTypeSeries)
		editMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, messages[0]+printPageNum(1, len(messages)), &keyboard)
	// get the last page of the series list
	case "lastSerie":
		log.Trace().Str("username", rcvCallback.From.Username).Msg("showing last page of series list")

		keyboard := getMediaListKeyboard(len(messages), len(messages), mediaTypeSeries)
		editMessageWithKeyboard(cb.bot, rcvCallback.Message.Chat.ID, rcvCallback.Message.ID, messages[len(messages)-1]+printPageNum(len(messages), len(messages)), &keyboard)
	case "serieDetails":
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
			cb.usersAction[rcvCallback.From.ID] = "serieDetails"
		}
	case "backToSeriesList":
		log.Trace().Str("username", rcvCallback.From.Username).Msg("back to series list")

		// remove the two last messages
		cb.bot.DeleteMessage(rcvCallback.Message.Chat.ID, rcvCallback.Message.ID)
		cb.bot.DeleteMessage(rcvCallback.Message.Chat.ID, rcvCallback.Message.ID-1)

		// show the series list
		sendSeriesList(cb.bot, rcvCallback.Message, cb.sonarrConfig)

	default:
		log.Warn().Str("username", rcvCallback.From.Username).Str("callback", rcvCallback.Data).Msg("unknown callback")
	}
}
