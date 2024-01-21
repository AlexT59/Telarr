package updates

import (
	"context"
	"strconv"
	"sync"
	"telarr/configuration"
	"telarr/internal/authentication"
	"telarr/internal/radarr"
	"telarr/internal/sonarr"
	"telarr/internal/types"

	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"gitlab.com/toby3d/telegram"
)

const (
	// MaxMessageLength is the maximum length of a message. It is divided by 4 due to unknown problem of telegram for formating long messages.
	MaxMessageLength = 4096 / 4
)

// Updates is the struct that will handle the messages.
type Updates struct {
	// config is the configuration.
	config configuration.Configuration

	// Bot is the telegram bot.
	bot *telegram.Bot
	// updateChan is the channel to receive the updates.
	updateChan telegram.UpdatesChannel

	// wg to wait for the goroutines to finish.
	wg *sync.WaitGroup

	// mess is the struct designed to handle the messages.
	mess *messages
	// cb is the struct designed to handle the callbacks.
	cb *callbacks

	usersAction map[int]types.Action
}

func New(config configuration.Configuration) (*Updates, error) {
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

	usersAction := make(map[int]types.Action)
	usersData := make(map[int]interface{})
	usersCurrPage := make(map[int]int)

	return &Updates{
		config:      config,
		bot:         bot,
		updateChan:  updatesChan,
		wg:          &sync.WaitGroup{},
		usersAction: usersAction,
		mess: &messages{
			bot:           bot,
			radarrConfig:  config.Radarr,
			sonarrConfig:  config.Sonarr,
			usersAction:   usersAction,
			usersData:     usersData,
			usersCurrPage: usersCurrPage,
		},
		cb: &callbacks{
			bot:           bot,
			radarrConfig:  config.Radarr,
			sonarrConfig:  config.Sonarr,
			usersAction:   usersAction,
			usersData:     usersData,
			usersCurrPage: usersCurrPage,
		},
	}, nil
}

func (upd *Updates) Start(ctx context.Context) error {
	// create the auth struct
	auth, err := authentication.New(upd.config)
	if err != nil {
		log.Err(err).Msg("error when creating the auth struct")
		return err
	}

	// list of users waiting for the password
	waitingForPassword := make(map[int]chan string)
	var waitingForPasswordMu sync.Mutex

	upd.wg.Add(1)
	go func() {
		defer upd.wg.Done()

		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("messages handler stopped")
				return

			case rcvUpdate := <-upd.updateChan:
				log.Trace().
					Msg("new update")

				// get the username
				var userId int
				if rcvUpdate.IsMessage() {
					userId = rcvUpdate.Message.From.ID
				} else if rcvUpdate.IsCallbackQuery() {
					userId = rcvUpdate.CallbackQuery.From.ID
				}

				// get the chat ID
				var chatID int64
				if rcvUpdate.IsMessage() {
					chatID = rcvUpdate.Message.Chat.ID
				} else if rcvUpdate.IsCallbackQuery() {
					chatID = rcvUpdate.CallbackQuery.Message.Chat.ID
				}

				// check if the user is waiting for the password
				waitingForPasswordMu.Lock()
				if c, exist := waitingForPassword[userId]; exist {
					c <- rcvUpdate.Message.Text
					waitingForPasswordMu.Unlock()
					continue
				}
				waitingForPasswordMu.Unlock()

				// check authorization
				authorized := auth.CheckAutorized(userId)
				switch authorized {
				// if the user is not authorized
				case authentication.AuthStatusBlackListed:
					log.Warn().Int("userId", userId).Msg("user is blacklisted")
					sendSimpleMessage(upd.bot, chatID, "You are blacklisted!\nPlease contact the administrator to remove you from the blacklist.")
					continue
				// if authorization failed
				case authentication.AuthStatusError:
					log.Err(err).Msg("error when checking authorization")
					sendSimpleMessage(upd.bot, chatID, "An error occurred while checking your authorization.\nPlease contact the administrator.")
					continue
				// if the user is new
				case authentication.AuthStatusNewUser:
					log.Info().Int("userId", userId).Msg("new user")
					sendSimpleMessage(upd.bot, chatID, "Welcome to the group "+rcvUpdate.Message.From.FirstName+"!\nPlease enter the password 🔑:")

					// create the channel for the password
					textChan := make(chan string)
					// add the user to the waiting list
					waitingForPassword[userId] = textChan
					// wait for the user to be autorized
					go func() {
						log.Info().Int("userId", userId).Msg("waiting for authorization")
						auth.WaitForAutorization(userId, upd.bot, textChan, chatID)
						// remove the user from the waiting list
						waitingForPasswordMu.Lock()
						delete(waitingForPassword, userId)
						waitingForPasswordMu.Unlock()
					}()
					continue
				// if the user is authorized
				case authentication.AuthStatusAutorized:
					if rcvUpdate.IsMessage() {
						// if it's a message
						log.Trace().
							Int("fromID", userId).
							Str("fromUsername", rcvUpdate.Message.From.Username).
							Str("text", rcvUpdate.Message.Text).
							Msg("new message")
						upd.mess.handle(rcvUpdate.Message)
					} else if rcvUpdate.IsCallbackQuery() {
						// if it's a callback query
						log.Trace().
							Int("fromID", rcvUpdate.CallbackQuery.From.ID).
							Str("fromUsername", rcvUpdate.CallbackQuery.From.Username).
							Str("data", rcvUpdate.CallbackQuery.Data).
							Msg("new callback query")
						upd.cb.handle(rcvUpdate.CallbackQuery)
					}
				}
			}
		}
	}()

	return nil
}

func (upd *Updates) Stop() error {
	upd.wg.Wait()
	return nil
}

/* Tools */

// Sening messages

// sendMessage sends a message to the chat and returns true if the message was sent successfully.
func sendMessage(bot *telegram.Bot, p telegram.SendMessage) int {
	if p.ParseMode == "" {
		p.ParseMode = telegram.ParseModeMarkdown
	}
	p.DisableWebPagePreview = true

	m, err := bot.SendMessage(p)
	if err != nil {
		log.Err(err).Msg("error when sending message")
		return -1
	}
	return m.ID
}

func sendSimpleMessage(bot *telegram.Bot, chatID int64, text string) int {
	return sendMessage(bot, telegram.SendMessage{
		ChatID: chatID,
		Text:   text,
	})
}

func sendMessageWithKeyboard(bot *telegram.Bot, chatID int64, text string, keyboard telegram.ReplyMarkup) int {
	return sendMessage(bot, telegram.SendMessage{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: keyboard,
	})
}

func sendImageMessage(bot *telegram.Bot, chatID int64, imageUrl string, caption string) bool {
	u := fasthttp.AcquireURI()
	defer fasthttp.ReleaseURI(u)
	u.Update(imageUrl)

	if len(caption) > 200 {
		caption = caption[:200-3] + "..."
	}

	_, err := bot.SendPhoto(telegram.SendPhoto{
		ChatID:      chatID,
		Photo:       &telegram.InputFile{URI: u},
		Caption:     caption,
		ParseMode:   telegram.ParseModeMarkdown,
		ReplyMarkup: telegram.NewReplyKeyboardRemove(false),
	})
	if err != nil {
		log.Err(err).Msg("error when sending image message")
		return false
	}
	return true
}

func sendImageMessageWithKeyboard(bot *telegram.Bot, chatID int64, imageUrl string, caption string, keyboard telegram.ReplyMarkup) bool {
	u := fasthttp.AcquireURI()
	defer fasthttp.ReleaseURI(u)
	u.Update(imageUrl)

	if len(caption) > 200 {
		caption = caption[:200-3] + "..."
	}

	_, err := bot.SendPhoto(telegram.SendPhoto{
		ChatID:      chatID,
		Photo:       &telegram.InputFile{URI: u},
		Caption:     caption,
		ParseMode:   telegram.ParseModeMarkdown,
		ReplyMarkup: keyboard,
	})
	if err != nil {
		log.Err(err).Msg("error when sending image message")
		return false
	}
	return true
}

// Editing messages

// editMessage edits a message and returns true if the message was edited successfully.
func editMessage(bot *telegram.Bot, p *telegram.EditMessageText) bool {
	if p.ParseMode == "" {
		p.ParseMode = telegram.ParseModeMarkdown
	}
	p.DisableWebPagePreview = true

	_, err := bot.EditMessageText(p)
	if err != nil {
		log.Err(err).Msg("error when updating message")
		return false
	}
	return true
}

func editSimpleMessage(bot *telegram.Bot, chatID int64, messageID int, text string) bool {
	return editMessage(bot, &telegram.EditMessageText{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      text,
	})
}

func editMessageWithKeyboard(bot *telegram.Bot, chatID int64, messageID int, text string, keyboard *telegram.InlineKeyboardMarkup) bool {
	return editMessage(bot, &telegram.EditMessageText{
		ChatID:      chatID,
		MessageID:   messageID,
		Text:        text,
		ReplyMarkup: keyboard,
	})
}

func editImageMessageWithKeyboard(bot *telegram.Bot, chatID int64, messageID int, imageUrl string, caption string, keyboard *telegram.InlineKeyboardMarkup) bool {
	u := fasthttp.AcquireURI()
	defer fasthttp.ReleaseURI(u)
	u.Update(imageUrl)

	if len(caption) > 200 {
		caption = caption[:200-3] + "..."
	}

	_, err := bot.EditMessageMedia(telegram.EditMessageMedia{
		ChatID:    chatID,
		MessageID: messageID,
		Media: &inputMediaPhotoCustom{
			Media:     u.String(),
			Caption:   caption,
			ParseMode: telegram.ParseModeMarkdown,
			Type:      telegram.TypePhoto,
		},
		ReplyMarkup: keyboard,
	})
	if err != nil {
		log.Err(err).Msg("error when updating image message")
		return false
	}
	return true
}

// Printing messages

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
	str += "/stop - 🛑 Cancel the current action\n"
	str += "/status - 📊 Show the status of the server\n"

	return str
}

func printMoviesList(list []radarr.Film) []string {
	var messages []string

	str := "🎬 *" + strconv.Itoa(len(list)) + " Movies*\n"
	for _, film := range list {
		sStr := "- *" + film.Title + "* (_" + strconv.Itoa(film.Year) + "_)\n"

		// check if the message is too long
		if len(str+sStr) > MaxMessageLength/2 {
			messages = append(messages, str)
			str = ""
		}
		str += sStr
	}
	messages = append(messages, str)

	return messages
}

func printSeriesList(list []sonarr.Serie) []string {
	var messages []string

	str := "📺 *" + strconv.Itoa(len(list)) + " Series*\n"
	for _, serie := range list {
		sStr := "- *" + serie.Title + "* (_" + strconv.Itoa(serie.Year) + "_)\n"
		for _, season := range serie.Seasons {
			if season.SeasonNumber == 0 {
				sStr += "\t\t- _Specials "
			} else {
				sStr += "\t\t- _Season " + strconv.Itoa(season.SeasonNumber)
			}
			sStr += " (" + strconv.Itoa(season.DownloadedEpisodes) + "/" + strconv.Itoa(season.TotalEpisodes) + ")_\n"
		}

		// check if the message is too long
		if len(str+sStr) > MaxMessageLength/2 {
			messages = append(messages, str)
			str = ""
		}
		str += sStr
	}
	messages = append(messages, str)

	return messages
}

func printPageNum(pageNb, totalPages int) string {
	return "\npage " + strconv.Itoa(pageNb) + "/" + strconv.Itoa(totalPages)
}

type inputMediaPhotoCustom struct {
	Type      string `json:"type"`
	Media     string `json:"media"`
	Caption   string `json:"caption"`
	ParseMode string `json:"parse_mode"`
}

func (i *inputMediaPhotoCustom) GetMedia() *telegram.InputFile {
	return nil
}
