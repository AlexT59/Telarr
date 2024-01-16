package messages

import (
	"context"
	"strconv"
	"sync"
	"telarr/configuration"
	"telarr/internal/authentication"
	"telarr/internal/radarr"
	"telarr/internal/sonarr"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/toby3d/telegram"
)

// Message is the struct that will handle the messages.
type Message struct {
	// config is the configuration.
	config configuration.Configuration

	// Bot is the telegram bot.
	bot *telegram.Bot
	// updateChan is the channel to receive the updates.
	updateChan telegram.UpdatesChannel

	// wg to wait for the goroutines to finish.
	wg *sync.WaitGroup
}

func New(config configuration.Configuration) (*Message, error) {
	// creating the telegram bot
	log.Debug().Msg("creating the telegram bot")
	bot, err := telegram.New(config.Telegram.Token)
	if err != nil {
		log.Err(err).Msg("error when creating the telegram bot")
		return nil, err
	}
	log.Info().Str("botName", bot.FullName()).Msg("telegram bot created")

	// getting the updates
	updatesChan := bot.NewLongPollingChannel(
		&telegram.GetUpdates{
			Offset:  0,
			Limit:   1,
			Timeout: 60,
		},
	)

	return &Message{
		config:     config,
		bot:        bot,
		updateChan: updatesChan,
		wg:         &sync.WaitGroup{},
	}, nil
}

func (mess *Message) Start(ctx context.Context) error {
	// create the auth struct
	auth, err := authentication.New(mess.config)
	if err != nil {
		log.Err(err).Msg("error when creating the auth struct")
		return err
	}

	// list of users waiting for the password
	waitingForPassword := make(map[int]chan string)
	var waitingForPasswordMu sync.Mutex

	// list of users actions
	usersActions := make(map[int]string)

	mess.wg.Add(1)
	go func() {
		defer mess.wg.Done()

		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("messages handler stopped")
				return

			case update := <-mess.updateChan:
				// check if the update is nil or if it is older than 60 seconds
				if update.Message == nil {
					continue
				}

				log.Trace().
					Int("fromID", update.Message.From.ID).
					Str("fromUsername", update.Message.From.Username).
					Str("text", update.Message.Text).
					Msg("new message")

				// check if the message is too old
				if time.Now().Unix()-update.Message.Date > 60 {
					log.Warn().Msg("message is too old")
					continue
				}

				// check if the user is waiting for the password
				waitingForPasswordMu.Lock()
				if c, exist := waitingForPassword[update.Message.From.ID]; exist {
					c <- update.Message.Text
					waitingForPasswordMu.Unlock()
					continue
				}
				waitingForPasswordMu.Unlock()

				// check authorization
				authorized := auth.CheckAutorized(update.Message.From.Username)
				switch authorized {
				case authentication.AuthStatusBlackListed:
					log.Warn().Str("username", update.Message.From.Username).Msg("user is blacklisted")
					mess.SendMessage(update.Message.Chat.ID, "You are blacklisted!\nPlease contact the administrator to remove you from the blacklist.")
					continue
				case authentication.AuthStatusError:
					log.Err(err).Msg("error when checking authorization")
					mess.SendMessage(update.Message.Chat.ID, "An error occurred while checking your authorization.\nPlease contact the administrator.")
					continue
				case authentication.AuthStatusNewUser:
					log.Info().Str("username", update.Message.From.Username).Msg("new user")
					mess.SendMessage(update.Message.Chat.ID, "Welcome to the group "+update.Message.From.FirstName+"!\nPlease enter the password 🔑:")

					// create the channel for the password
					textChan := make(chan string)
					// add the user to the waiting list
					waitingForPassword[update.Message.From.ID] = textChan
					// wait for the user to be autorized
					go func() {
						log.Info().Str("username", update.Message.From.Username).Msg("waiting for authorization")
						auth.WaitForAutorization(update.Message.From.Username, mess.bot, textChan, update.Message.Chat.ID)
						// remove the user from the waiting list
						waitingForPasswordMu.Lock()
						delete(waitingForPassword, update.Message.From.ID)
						waitingForPasswordMu.Unlock()
					}()
					continue
				case authentication.AuthStatusAutorized:
					// if it's a command
					if update.Message.IsCommand() {
						log.Debug().Str("username", update.Message.From.Username).Str("command", update.Message.Command()).Msg("command received")
						switch update.Message.Command() {
						case "help":
							mess.SendMessage(update.Message.Chat.ID, printHelp())
							continue
						case "movies":
							log.Trace().Str("username", update.Message.From.Username).Msg("getting movies list")
							films, err := radarr.GetFilmsList(mess.config.Radarr)
							if err != nil {
								log.Err(err).Msg("error when getting movies list")
								mess.SendMessage(update.Message.Chat.ID, "An error occurred while getting the movies list.\nPlease contact the administrator.")
								continue
							}

							mess.SendMessage(update.Message.Chat.ID, printMoviesList(films))
							continue
						case "addmovie":
							log.Trace().Str("username", update.Message.From.Username).Msg("adding movie")
							usersActions[update.Message.From.ID] = "addmovie"
						case "series":
							log.Trace().Str("username", update.Message.From.Username).Msg("getting series list")
							series, err := sonarr.GetSeriesList(mess.config.Sonarr)
							if err != nil {
								log.Err(err).Msg("error when getting series list")
								mess.SendMessage(update.Message.Chat.ID, "An error occurred while getting the series list.\nPlease contact the administrator.")
								continue
							}

							mess.SendMessage(update.Message.Chat.ID, printSeriesList(series))
							continue
						case "addserie":
							log.Trace().Str("username", update.Message.From.Username).Msg("adding serie")
							usersActions[update.Message.From.ID] = "addserie"
						case "stop":
							log.Trace().Str("username", update.Message.From.Username).Msg("canceling action")
							delete(usersActions, update.Message.From.ID)
							mess.SendMessage(update.Message.Chat.ID, "Action canceled ✅")
						}
					} else {
						// if it's a message
						if action, exist := usersActions[update.Message.From.ID]; exist {
							switch action {
							case "addmovie":
							case "addserie":
							default:
								mess.SendMessage(update.Message.Chat.ID, "I don't understand what you mean.\nPlease use /help to see the commands list.")
							}
						}
					}
				}
			}
		}
	}()

	return nil
}

func (mess *Message) Stop() error {
	mess.wg.Wait()
	return nil
}

/* Internal */

func (mess *Message) SendMessage(chatID int64, text string) {
	_, err := mess.bot.SendMessage(telegram.SendMessage{
		ChatID:    chatID,
		Text:      text,
		ParseMode: telegram.ParseModeMarkdown,
	})
	if err != nil {
		log.Err(err).Msg("error when sending message")
	}
}

/* Tools */

func printHelp() string {
	str := "/help - 📋 Show commands list\n"

	// movies
	str += "\n🎬 *Movies*\n"
	str += "/movies - Show the movies list\n"
	str += "/addmovie - Add a movie\n"

	// series
	str += "\n📺 *Series*\n"
	str += "/series - Show the series list\n"
	str += "/addserie - Add a serie\n"

	// commands
	str += "\n🔧 *Commands*\n"
	str += "/stop - 🛑 Cancel the current action"

	return str
}

func printMoviesList(list []radarr.Film) string {
	str := "🎬 *Movies*\n"
	for _, film := range list {
		str += film.Title + " (" + strconv.Itoa(film.Year) + ")\n"
	}

	return str
}

func printSeriesList(list []sonarr.Serie) string {
	str := "📺 *Series*\n"
	for _, serie := range list {
		str += serie.Title + " (" + strconv.Itoa(serie.Year) + ")\n"
	}

	return str
}
