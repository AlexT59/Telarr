package updates

import (
	"telarr/configuration"
	"telarr/internal/radarr"
	"telarr/internal/sonarr"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/toby3d/telegram"
)

const (
	// messageTimeOut is the maximum time a message can be old.
	messageTimeOut = 60
)

// messages is the struct designed to handle the messages.
type messages struct {
	// Bot is the telegram bot.
	bot *telegram.Bot

	radarrConfig configuration.Radarr
	sonarrConfig configuration.Sonarr

	// list of users actions
	usersAction map[int]string
}

func (mess *messages) handle(rcvMess *telegram.Message) {
	if rcvMess == nil {
		log.Warn().Msg("received nil message")
		return
	}

	// check if the message is too old
	if time.Now().Unix()-rcvMess.Date > messageTimeOut {
		log.Warn().Msg("message is too old")
		return
	}

	// if it's a command
	if rcvMess.IsCommand() {
		log.Debug().Str("username", rcvMess.From.Username).Str("command", rcvMess.Command()).Msg("command received")
		switch rcvMess.Command() {
		case "help":
			sendSimpleMessage(mess.bot, rcvMess.Chat.ID, printHelp())
		case "movies":
			sendMoviesList(mess.bot, rcvMess, mess.radarrConfig)
		case "addmovie":
			log.Trace().Str("username", rcvMess.From.Username).Msg("adding movie")
			mess.usersAction[rcvMess.From.ID] = "addmovie"
		case "series":
			sendSeriesList(mess.bot, rcvMess, mess.sonarrConfig)
		case "addserie":
			log.Trace().Str("username", rcvMess.From.Username).Msg("adding serie")
			mess.usersAction[rcvMess.From.ID] = "addserie"
		case "stop":
			log.Trace().Str("username", rcvMess.From.Username).Msg("canceling action")
			delete(mess.usersAction, rcvMess.From.ID)
			sendMessageWithKeyboard(mess.bot, rcvMess.Chat.ID, "Action canceled ✅", telegram.NewReplyKeyboardRemove(false))

		default:
			log.Warn().Str("username", rcvMess.From.Username).Str("command", rcvMess.Command()).Msg("unknown command")
		}
	} else {
		// if it's a message
		log.Debug().Str("username", rcvMess.From.Username).Msg("message received")
		if action, exist := mess.usersAction[rcvMess.From.ID]; exist {
			switch action {
			case "addmovie":
			case "addserie":
			case "movieDetails":
				movieName := rcvMess.Text
				log.Trace().Str("username", rcvMess.From.Username).Str("movieName", movieName).Msg("getting movie details")

				foundFilms, err := radarr.GetFilmDetails(mess.radarrConfig, movieName)
				if err != nil {
					log.Err(err).Msg("error when getting movie details")
					sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "An error occurred while getting the movie details.\nPlease contact the administrator.")
					return
				}

				if len(foundFilms) == 0 {
					sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "No movie found with this name.")
					return
				}

				// send the movie details
				film := foundFilms[0]
				log.Trace().Str("username", rcvMess.From.Username).Str("movieName", film.Title).Msg("sending movie details")
				sendImageMessage(mess.bot, rcvMess.Chat.ID, film.CoverImage, film.PrintMovieTitle())
				sendMessageWithKeyboard(mess.bot, rcvMess.Chat.ID, film.PrintMovieDetails(), telegram.NewInlineKeyboardMarkup([]*telegram.InlineKeyboardButton{telegram.NewInlineKeyboardButton("<< Back to movies list", "backToMoviesList")}))
			case "serieDetails":
				serieName := rcvMess.Text
				log.Trace().Str("username", rcvMess.From.Username).Str("serieName", serieName).Msg("getting serie details")

				foundSeries, err := sonarr.GetSerieDetails(mess.sonarrConfig, serieName)
				if err != nil {
					log.Err(err).Msg("error when getting serie details")
					sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "An error occurred while getting the serie details.\nPlease contact the administrator.")
					return
				}

				if len(foundSeries) == 0 {
					sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "No serie found with this name.")
					return
				}

				// send the serie details
				serie := foundSeries[0]
				log.Trace().Str("username", rcvMess.From.Username).Str("serieName", serie.Title).Msg("sending serie details")
				sendImageMessage(mess.bot, rcvMess.Chat.ID, serie.CoverImage, serie.PrintSerieTitle())
				sendMessageWithKeyboard(mess.bot, rcvMess.Chat.ID, serie.PrintSerieDetails(), telegram.NewInlineKeyboardMarkup([]*telegram.InlineKeyboardButton{telegram.NewInlineKeyboardButton("<< Back to series list", "backToSeriesList")}))
			default:
				log.Warn().Str("username", rcvMess.From.Username).Str("action", action).Msg("unknown action")
			}

			delete(mess.usersAction, rcvMess.From.ID)
		} else {
			log.Trace().Str("username", rcvMess.From.Username).Msg("unknown message")
			sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "I don't understand what you mean.\nPlease use /help to see the commands list.")
		}
	}
}

/* Tools */

// sendMoviesList sends the movies list to the user.
func sendMoviesList(bot *telegram.Bot, rcvMess *telegram.Message, radarrConfig configuration.Radarr) {
	log.Trace().Str("username", rcvMess.From.Username).Msg("getting movies list")
	films, err := radarr.GetFilmsList(radarrConfig)
	if err != nil {
		log.Err(err).Msg("error when getting movies messages")
		sendSimpleMessage(bot, rcvMess.Chat.ID, "An error occurred while getting the movies list.\nPlease contact the administrator.")
		return
	}
	messages := printMoviesList(films)

	// send the movies list
	log.Trace().Str("username", rcvMess.From.Username).Msg("sending movies list")
	keyboard := getMediaListKeyboard(1, len(messages), mediaTypeMovies)
	sendMessageWithKeyboard(bot, rcvMess.Chat.ID, messages[0]+printPageNum(1, len(messages)), keyboard)
}

// sendSeriesList sends the series list to the user.
func sendSeriesList(bot *telegram.Bot, rcvMess *telegram.Message, sonarrConfig configuration.Sonarr) {
	log.Trace().Str("username", rcvMess.From.Username).Msg("getting series list")
	series, err := sonarr.GetSeriesList(sonarrConfig)
	if err != nil {
		log.Err(err).Msg("error when getting series messages")
		sendSimpleMessage(bot, rcvMess.Chat.ID, "An error occurred while getting the series list.\nPlease contact the administrator.")
		return
	}
	messages := printSeriesList(series)

	// send the series list
	log.Trace().Str("username", rcvMess.From.Username).Msg("sending series list")
	keyboard := getMediaListKeyboard(1, len(messages), mediaTypeSeries)
	sendMessageWithKeyboard(bot, rcvMess.Chat.ID, messages[0]+printPageNum(1, len(messages)), keyboard)
}
