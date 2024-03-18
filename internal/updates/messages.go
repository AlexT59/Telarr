package updates

import (
	"strconv"
	"telarr/configuration"
	"telarr/internal/radarr"
	"telarr/internal/sonarr"
	"telarr/internal/types"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/showwin/speedtest-go/speedtest"
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

	radarrConfig     configuration.Radarr
	sonarrConfig     configuration.Sonarr
	pathForDiskUsage string

	// list of users actions
	usersAction map[int]types.Action
	// list of data to navigate between pages
	usersData     map[int]interface{}
	usersCurrPage map[int]int
}

func (mess *messages) handle(rcvMess *telegram.Message, isAdmin bool) {
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
		delete(mess.usersAction, rcvMess.From.ID)

		switch rcvMess.Command() {
		case "help":
			sendSimpleMessage(mess.bot, rcvMess.Chat.ID, printHelp())
		case "movies":
			sendMoviesList(mess.bot, rcvMess, mess.radarrConfig)
		case "addmovie":
			log.Trace().Str("username", rcvMess.From.Username).Msg("adding movie")
			mess.usersAction[rcvMess.From.ID] = types.UserActionLookMovieToAdd

			sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "Please enter the name of the movie you want to add:")
		case "series":
			sendSeriesList(mess.bot, rcvMess, mess.sonarrConfig)
		case "addserie":
			log.Trace().Str("username", rcvMess.From.Username).Msg("adding serie")
			mess.usersAction[rcvMess.From.ID] = types.UserActionLookSerieToAdd

			sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "Please enter the name of the serie you want to add:")
		case "status":
			log.Trace().Str("username", rcvMess.From.Username).Msg("getting status")

			mId := sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "Getting radarr status...")
			radarrStatus := radarr.GetStatus(mess.radarrConfig)
			mess.bot.DeleteMessage(rcvMess.Chat.ID, mId)
			mId = sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "Getting sonarr status...")
			sonarrStatus := sonarr.GetStatus(mess.sonarrConfig)
			mess.bot.DeleteMessage(rcvMess.Chat.ID, mId)
			str := radarrStatus.String() + "\n" + sonarrStatus.String() + "\n"

			// get the speedtest
			log.Trace().Msg("getting speedtest")
			mId = sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "Getting speedtest...")
			spd := speedtest.New()
			srvList, err := spd.FetchServers()
			if err != nil {
				log.Err(err).Msg("error when fetching speedtest servers")
				str += "\n\nAn error occurred while fetching the speedtest servers."
			} else {
				targets, err := srvList.FindServer([]int{})
				if err != nil {
					log.Err(err).Msg("error when finding speedtest server")
					str += "\n\nAn error occurred while finding the speedtest servers."
				} else {
					s := (*targets.Available())[0]
					err = s.PingTest(nil)
					if err != nil {
						log.Err(err).Msg("error when pinging speedtest")
					}
					err = s.DownloadTest()
					if err != nil {
						log.Err(err).Msg("error when downloading speedtest")
					}
					err = s.UploadTest()
					if err != nil {
						log.Err(err).Msg("error when uploading speedtest")
					}

					str += "\n*Speed test*:\n"
					str += "\t⏱ " + strconv.Itoa(int(s.Latency.Milliseconds())) + " ms\n"
					str += "\t⬇️ " + strconv.FormatFloat(s.DLSpeed, 'f', 2, 64) + " Mbps\n"
					str += "\t⬆️ " + strconv.FormatFloat(s.ULSpeed, 'f', 2, 64) + " Mbps\n"

					s.Context.Reset()
				}
			}
			mess.bot.DeleteMessage(rcvMess.Chat.ID, mId)

			mId = sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "Getting disk usage...")
			diskStatus, err := getDiskUsage(mess.pathForDiskUsage)
			if err != nil {
				log.Err(err).Msg("error when getting disk usage")
			}
			str += "\n*Disk usage*:\n"
			str += "\tfree: " + diskStatus.FreeOfAll() + " (" + strconv.FormatFloat(diskStatus.FreePercent(), 'f', 2, 64) + "%)\n"
			mess.bot.DeleteMessage(rcvMess.Chat.ID, mId)

			sendMessageWithKeyboard(mess.bot, rcvMess.Chat.ID, str, telegram.NewReplyKeyboardRemove(false))
		case "stop":
			log.Trace().Str("username", rcvMess.From.Username).Msg("canceling action")

			// remove the data from the user
			delete(mess.usersData, rcvMess.From.ID)
			delete(mess.usersCurrPage, rcvMess.From.ID)

			sendMessageWithKeyboard(mess.bot, rcvMess.Chat.ID, "Action canceled ✅", telegram.NewReplyKeyboardRemove(false))
		case "admin":
			if !isAdmin {
				log.Warn().Str("username", rcvMess.From.Username).Msg("user is not admin")
				sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "You are not an administrator.")
				return
			}
			sendMessageWithKeyboard(mess.bot, rcvMess.Chat.ID, "Select an action:", getAdminKeyboard())

		default:
			log.Warn().Str("username", rcvMess.From.Username).Str("command", rcvMess.Command()).Msg("unknown command")
			sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "I don't understand this command.\nPlease use /help to see the commands list.")
		}
	} else {
		// if it's a message
		log.Debug().Str("username", rcvMess.From.Username).Msg("message received")

		if action, exist := mess.usersAction[rcvMess.From.ID]; exist {
			delete(mess.usersAction, rcvMess.From.ID)

			if action.(types.UserAction) == types.UserActionRemoveSerie {
				println("remove serie")
			} else {
				println("not remove serie")
			}

			switch action.(types.UserAction) {
			/* Movies */
			case types.UserActionLookMovieToAdd:
				movieName := rcvMess.Text
				log.Trace().Str("username", rcvMess.From.Username).Str("movieName", movieName).Msg("looking for movie to add")

				foundFilms, err := radarr.LookupFilm(mess.radarrConfig, movieName)
				if err != nil {
					log.Err(err).Msg("error when looking for movie")
					sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "An error occurred while looking for the movie.\nPlease contact the administrator.")
					return
				}

				if len(foundFilms) == 0 {
					sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "No movie found with this name.")
					return
				}

				mess.usersData[rcvMess.From.ID] = foundFilms
				mess.usersCurrPage[rcvMess.From.ID] = 1

				// send the first movie found
				film := foundFilms[0]
				str := film.PrintMovieTitle()
				if film.IsInLibrary {
					str += "\n\nAlready in your library ✅"
				}
				sendImageMessageWithKeyboard(mess.bot, rcvMess.Chat.ID, film.CoverImage, str, getAddMediaKeyboard(1, len(foundFilms), mediaTypeMovie, !film.IsInLibrary))
			case types.UserActionMovieDetails:
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
			case types.UserActionRemoveMovie:
				movieName := rcvMess.Text
				log.Trace().Str("username", rcvMess.From.Username).Str("movieName", movieName).Msg("getting movie to remove")

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

				film := foundFilms[0]
				str := film.PrintMovieTitle()
				str += "\n_MovieId: " + strconv.Itoa(int(film.MovieId)) + "_"
				str += "\n\nAre you sure you want to remove this movie from your library?"
				sendMessageWithKeyboard(mess.bot, rcvMess.Chat.ID, str, getConfirmRemoveKeyboard(mediaTypeMovie))
			case types.UserActionAddMovie:
				qualityProfileName := rcvMess.Text

				pageNb := mess.usersCurrPage[rcvMess.From.ID]
				films := (mess.usersData[rcvMess.From.ID].([]radarr.Film))
				film := films[pageNb-1]

				log.Trace().Str("username", rcvMess.From.Username).Str("qualityProfileName", qualityProfileName).Str("movie", film.Title).Msg("adding movie")

				// get the quality profile id
				qualityProfileId, err := radarr.GetQualityProfileId(mess.radarrConfig, qualityProfileName)
				if err != nil {
					log.Err(err).Msg("error when getting quality profile id")
					sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "An error occurred while getting the quality profile id.\nPlease contact the administrator.")
					return
				}

				// add the movie
				newFilmId, err := radarr.AddFilm(mess.radarrConfig, film, qualityProfileId)
				if err != nil {
					log.Err(err).Msg("error when adding movie")
					sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "An error occurred while adding the movie.\nPlease contact the administrator.")
					return
				}

				// remove the data from the user
				delete(mess.usersData, rcvMess.From.ID)
				delete(mess.usersCurrPage, rcvMess.From.ID)

				// send the confirmation message
				log.Trace().Str("username", rcvMess.From.Username).Str("movie", film.Title).Msg("movie added")
				sendMessageWithKeyboard(mess.bot, rcvMess.Chat.ID, "Movie "+film.PrintMovieTitle()+" added ✅\n_movieId: "+strconv.Itoa(int(newFilmId))+"_", telegram.NewInlineKeyboardMarkup([]*telegram.InlineKeyboardButton{telegram.NewInlineKeyboardButton("Follow downloading status 📡", types.CallbackFollowDownloadingStatusMovie.String())}))

				/* Series */
			case types.UserActionLookSerieToAdd:
				serieName := rcvMess.Text
				log.Trace().Str("username", rcvMess.From.Username).Str("serieName", serieName).Msg("looking for serie to add")

				foundSeries, err := sonarr.LookupSerie(mess.sonarrConfig, serieName)
				if err != nil {
					log.Err(err).Msg("error when looking for serie")
					sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "An error occurred while looking for the serie.\nPlease contact the administrator.")
					return
				}

				if len(foundSeries) == 0 {
					sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "No serie found with this name.")
					return
				}

				mess.usersData[rcvMess.From.ID] = foundSeries
				mess.usersCurrPage[rcvMess.From.ID] = 1

				// send the first serie found
				serie := foundSeries[0]
				str := serie.PrintSerieTitle()
				if serie.IsInLibrary {
					str += "\n\nAlready in your library ✅"
				}
				sendImageMessageWithKeyboard(mess.bot, rcvMess.Chat.ID, serie.CoverImage, serie.PrintSerieTitle(), getAddMediaKeyboard(1, len(foundSeries), mediaTypeSerie, !serie.IsInLibrary))
			case types.UserActionSerieDetails:
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
			case types.UserActionRemoveSerie:
				serieName := rcvMess.Text
				log.Trace().Str("username", rcvMess.From.Username).Str("serieName", serieName).Msg("getting serie to remove")

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

				serie := foundSeries[0]
				str := serie.PrintSerieTitle()
				str += "\n_SerieId: " + strconv.Itoa(int(serie.SerieId)) + "_"
				str += "\n\nAre you sure you want to remove this serie from your library?"
				sendMessageWithKeyboard(mess.bot, rcvMess.Chat.ID, str, getConfirmRemoveKeyboard(mediaTypeSerie))
			case types.UserActionAddSerie:
				qualityProfileName := rcvMess.Text

				pageNb := mess.usersCurrPage[rcvMess.From.ID]
				series := (mess.usersData[rcvMess.From.ID].([]sonarr.Serie))
				serie := series[pageNb-1]

				log.Trace().Str("username", rcvMess.From.Username).Str("qualityProfileName", qualityProfileName).Str("serie", serie.Title).Msg("adding serie")

				// get the quality profile id
				qualityProfileId, err := sonarr.GetQualityProfileId(mess.sonarrConfig, qualityProfileName)
				if err != nil {
					log.Err(err).Msg("error when getting quality profile id")
					sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "An error occurred while getting the quality profile id.\nPlease contact the administrator.")
					return
				}

				// add the serie
				err = sonarr.AddSerie(mess.sonarrConfig, serie, qualityProfileId)
				if err != nil {
					log.Err(err).Msg("error when adding serie")
					sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "An error occurred while adding the serie.\nPlease contact the administrator.")
					return
				}

				// remove the data from the user
				delete(mess.usersData, rcvMess.From.ID)
				delete(mess.usersCurrPage, rcvMess.From.ID)

				// send the confirmation message
				log.Trace().Str("username", rcvMess.From.Username).Str("serie", serie.Title).Msg("serie added")
				sendSimpleMessage(mess.bot, rcvMess.Chat.ID, "Serie "+serie.PrintSerieTitle()+" added ✅")

			default:
				log.Warn().Str("username", rcvMess.From.Username).Str("action", action.String()).Msg("unknown action")
			}
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
	keyboard := getMediaListKeyboard(1, len(messages), mediaTypeMovie)
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
	keyboard := getMediaListKeyboard(1, len(messages), mediaTypeSerie)
	sendMessageWithKeyboard(bot, rcvMess.Chat.ID, messages[0]+printPageNum(1, len(messages)), keyboard)
}
