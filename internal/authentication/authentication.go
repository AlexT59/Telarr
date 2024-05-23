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
	// adminFile is the name of the file that contains the admins.
	adminFile = "admin.json"
)

var (
	// authPath is the path to the auth files.
	authPath = "/opt/telarr/auth"
)

type User struct {
	// Id is the id of the user.
	Id int `json:"id"`
	// Username is the username of the user.
	Username string `json:"username"`
}

type Auth struct {
	// Blacklist is a list of users that are not allowed to use the bot.
	Blacklist []User
	// Autorized is a list of users that are allowed to use the bot.
	Autorized []User
	// Admins is a list of users that are allowed to use the admin commands.
	Admins []User

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
	// create the admin file if it does not exist
	_, err = os.Stat(authPath + "/" + adminFile)
	if os.IsNotExist(err) {
		_, err = os.Create(authPath + "/" + adminFile)
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

	// read the admin from the database
	bytes, err = os.ReadFile(authPath + "/" + adminFile)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(bytes, &auth.Admins)

	return auth, nil
}

// CheckAutorized checks if the user is autorized.
func (a *Auth) CheckAutorized(userId int) (AuthStatus, bool) {
	// check blacklist
	for _, u := range a.Blacklist {
		if u.Id == userId {
			return AuthStatusBlackListed, false
		}
	}

	// check autorized
	for _, u := range a.Autorized {
		if u.Id == userId {
			return AuthStatusAutorized, a.isAdmin(userId)
		}
	}

	return AuthStatusNewUser, false
}

// AutorizeNewUser autorizes the user if the password is correct.
// Add the user to the blacklist if the maximum number of attempts has been reached.
func (a *Auth) AutorizeNewUser(user User, password string) (AuthStatus, int) {
	// check if the user is autorized
	status, _ := a.CheckAutorized(user.Id)
	if status == AuthStatusAutorized {
		return AuthStatusAutorized, -1
	}

	// check if the password is correct
	if password != a.conf.Telegram.Passwd {
		a.Attempts[user.Id]++

		// check if the user has reached the maximum number of attempts
		if a.Attempts[user.Id] >= maxAttempts {
			// add user to the blacklist
			err := a.saveBlacklist(user)
			if err != nil {
				return AuthStatusError, -1
			}

			return AuthStatusMaxAttempts, 0
		}
		return AuthStatusWrongPassword, maxAttempts - a.Attempts[user.Id]
	}

	// add the user to the autorized list
	log.Debug().Str("username", user.Username).Msg("saving autorized user")
	err := a.saveAutorized(user)
	if err != nil {
		return AuthStatusError, -1
	}
	delete(a.Attempts, user.Id)

	return AuthStatusAutorized, -1
}

func (a *Auth) WaitForAutorization(ctx context.Context, user User, bot *telegram.Bot, textChan chan string, chatId int64) {
	// wait for the user to be autorized
	for {
		select {
		case <-ctx.Done():
			return
		case text := <-textChan:
			status, attemps := a.AutorizeNewUser(user, text)
			switch status {
			case AuthStatusAutorized:
				log.Info().Str("username", user.Username).Msg("user is now authorized")

				_, err := bot.SendMessage(telegram.SendMessage{
					ChatID: chatId,
					Text:   "You are now authorized! 🎉",
				})
				if err != nil {
					log.Err(err).Msg("error when sending message")
				}

				return
			case AuthStatusWrongPassword:
				log.Debug().Str("username", user.Username).Msg("wrong password")

				_, err := bot.SendMessage(telegram.SendMessage{
					ChatID: chatId,
					Text:   "Wrong password ❌\nYou have " + strconv.Itoa(attemps) + " attempts left.",
				})
				if err != nil {
					log.Err(err).Msg("error when sending message")
				}
			case AuthStatusMaxAttempts:
				log.Debug().Str("username", user.Username).Msg("maximum number of attempts reached")

				_, err := bot.SendMessage(telegram.SendMessage{
					ChatID: chatId,
					Text:   "You have reached the maximum number of attempts.\nYou are now blacklisted!",
				})
				if err != nil {
					log.Err(err).Msg("error when sending message")
				}

				return
			case AuthStatusError:
				log.Debug().Str("username", user.Username).Msg("error when autorizing user")

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

func (a *Auth) isAdmin(userId int) bool {
	for _, u := range a.Admins {
		if userId == u.Id {
			return true
		}
	}
	return false
}

func (a *Auth) saveBlacklist(user User) error {
	// check if the user is already in the blacklist
	for _, u := range a.Blacklist {
		if u.Id == user.Id {
			return nil
		}
	}

	// add the user to the blacklist
	a.Blacklist = append(a.Blacklist, user)

	// save the blacklist to the database
	bytes, err := json.MarshalIndent(a.Blacklist, "", "  ")
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

func (a *Auth) saveAutorized(user User) error {
	// check if the user is already in the autorized list
	for _, u := range a.Autorized {
		if u.Id == user.Id {
			return nil
		}
	}

	// add the user to the autorized list
	a.Autorized = append(a.Autorized, user)

	// save the autorized list to the database
	bytes, err := json.MarshalIndent(a.Autorized, "", "  ")
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
