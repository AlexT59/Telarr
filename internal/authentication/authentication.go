package authentication

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"telarr/configuration"

	"github.com/rs/zerolog/log"
	"gitlab.com/toby3d/telegram"
)

const (
	// maxAttempts is the maximum number of attempts to use the bot.
	maxAttempts = 3

	// blacklistFile is the name of the file that contains the blacklist.
	blacklistFile = "blacklist.json"
	// autorizedFile is the name of the file that contains the autorized.
	autorizedFile = "autorized.json"
)

var (
	// authPath is the path to the auth files.
	authPath = "/opt/telarr/auth"
)

type Auth struct {
	// Blacklist is a list of users that are not allowed to use the bot.
	Blacklist []int
	// Autorized is a list of users that are allowed to use the bot.
	Autorized []int

	// Attempts is a map of users and the number of attempts to use the bot.
	Attempts map[int]int

	conf configuration.Configuration
}

type AuthStatus int

const (
	// AuthStatusAutorized is returned when the user is autorized.
	AuthStatusAutorized AuthStatus = iota
	// AuthStatusBlackListed is returned when the user is not autorized.
	AuthStatusBlackListed

	// AuthStatusNewUser is returned when the user is a new user.
	AuthStatusNewUser

	// AuthStatusWrongPassword is returned when the user has entered the wrong password.
	AuthStatusWrongPassword
	// AuthStatusMaxAttempts is returned when the user has reached the maximum number of attempts.
	AuthStatusMaxAttempts

	// AuthStatusError is returned when there is an error when autorizing the user.
	AuthStatusError
)

func New(conf configuration.Configuration) (*Auth, error) {
	// create the auth directory if it does not exist
	err := os.MkdirAll(authPath, 0755)
	if err != nil {
		return nil, err
	}
	// create the blacklist file if it does not exist
	_, err = os.Stat(authPath + "/" + blacklistFile)
	if os.IsNotExist(err) {
		_, err = os.Create(authPath + "/" + blacklistFile)
		if err != nil {
			return nil, err
		}
	}
	// create the autorized file if it does not exist
	_, err = os.Stat(authPath + "/" + autorizedFile)
	if os.IsNotExist(err) {
		_, err = os.Create(authPath + "/" + autorizedFile)
		if err != nil {
			return nil, err
		}
	}

	// create the auth struct
	auth := &Auth{
		Attempts: make(map[int]int),
		conf:     conf,
	}

	// read the blacklist from the database
	bytes, err := os.ReadFile(authPath + "/" + blacklistFile)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(bytes, &auth.Blacklist)

	// read the autorized from the database
	bytes, err = os.ReadFile(authPath + "/" + autorizedFile)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(bytes, &auth.Autorized)

	return auth, nil
}

// CheckAutorized checks if the user is autorized.
func (a *Auth) CheckAutorized(userId int) AuthStatus {
	// check blacklist
	for _, u := range a.Blacklist {
		if u == userId {
			return AuthStatusBlackListed
		}
	}

	// check autorized
	for _, u := range a.Autorized {
		if u == userId {
			return AuthStatusAutorized
		}
	}

	return AuthStatusNewUser
}

// AutorizeNewUser autorizes the user if the password is correct.
// Add the user to the blacklist if the maximum number of attempts has been reached.
func (a *Auth) AutorizeNewUser(userId int, password string) (AuthStatus, int) {
	// check if the user is autorized
	status := a.CheckAutorized(userId)
	if status == AuthStatusAutorized {
		return AuthStatusAutorized, -1
	}

	a.Attempts[userId]++

	// check if the user has reached the maximum number of attempts
	if a.Attempts[userId] >= maxAttempts {
		// add user to the blacklist
		err := a.saveBlacklist(userId)
		if err != nil {
			return AuthStatusError, -1
		}

		return AuthStatusMaxAttempts, 0
	}

	// check if the password is correct
	if password != a.conf.Telegram.Passwd {
		return AuthStatusWrongPassword, maxAttempts - a.Attempts[userId]
	}

	// add the user to the autorized list
	log.Debug().Int("username", userId).Msg("saving autorized user")
	err := a.saveAutorized(userId)
	if err != nil {
		return AuthStatusError, -1
	}
	delete(a.Attempts, userId)

	return AuthStatusAutorized, -1
}

func (a *Auth) WaitForAutorization(ctx context.Context, userId int, bot *telegram.Bot, textChan chan string, chatId int64) {
	// wait for the user to be autorized
	for {
		select {
		case <-ctx.Done():
			return
		case text := <-textChan:
			status, attemps := a.AutorizeNewUser(userId, text)
			switch status {
			case AuthStatusAutorized:
				log.Debug().Int("username", userId).Msg("user is now authorized")

				_, err := bot.SendMessage(telegram.SendMessage{
					ChatID: chatId,
					Text:   "You are now authorized! 🎉",
				})
				if err != nil {
					log.Err(err).Msg("error when sending message")
				}

				return
			case AuthStatusWrongPassword:
				log.Debug().Int("username", userId).Msg("wrong password")

				_, err := bot.SendMessage(telegram.SendMessage{
					ChatID: chatId,
					Text:   "Wrong password ❌\nYou have " + strconv.Itoa(attemps) + " attempts left.",
				})
				if err != nil {
					log.Err(err).Msg("error when sending message")
				}
			case AuthStatusMaxAttempts:
				log.Debug().Int("username", userId).Msg("maximum number of attempts reached")

				_, err := bot.SendMessage(telegram.SendMessage{
					ChatID: chatId,
					Text:   "You have reached the maximum number of attempts.\nYou are now blacklisted!",
				})
				if err != nil {
					log.Err(err).Msg("error when sending message")
				}

				return
			case AuthStatusError:
				log.Debug().Int("username", userId).Msg("error when autorizing user")

				_, err := bot.SendMessage(telegram.SendMessage{
					ChatID: chatId,
					Text:   "An error occurred while checking your authorization.\nPlease contact the administrator.",
				})
				if err != nil {
					log.Err(err).Msg("error when sending message")
				}

				return
			}
		}
	}
}

/* Internal */

func (a *Auth) saveBlacklist(userId int) error {
	// check if the user is already in the blacklist
	for _, u := range a.Blacklist {
		if u == userId {
			return nil
		}
	}

	// add the user to the blacklist
	a.Blacklist = append(a.Blacklist, userId)

	// save the blacklist to the database
	bytes, err := json.Marshal(a.Blacklist)
	if err != nil {
		log.Err(err).Msg("error when marshaling the blacklist")
		return err
	}
	err = os.WriteFile(authPath+"/"+blacklistFile, bytes, 0644)
	if err != nil {
		log.Err(err).Msg("error when saving the blacklist to the database")
		return err
	}

	return nil
}

func (a *Auth) saveAutorized(userId int) error {
	// check if the user is already in the autorized list
	for _, u := range a.Autorized {
		if u == userId {
			return nil
		}
	}

	// add the user to the autorized list
	a.Autorized = append(a.Autorized, userId)

	// save the autorized list to the database
	bytes, err := json.Marshal(a.Autorized)
	if err != nil {
		log.Err(err).Msg("error when marshaling the autorized list")
		return err
	}
	err = os.WriteFile(authPath+"/"+autorizedFile, bytes, 0644)
	if err != nil {
		log.Err(err).Msg("error when saving the autorized list to the database")
		return err
	}

	return nil
}
