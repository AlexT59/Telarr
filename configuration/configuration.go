package configuration

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	// configPath is the path to the configuration file
	defaultConfigPath = "./configuration"
	// configFileName is the name of the configuration file
	configFileName = "config.yaml"
)

type Configuration struct {
	Telegram *Telegram `yaml:"telegram"`
}

type Telegram struct {
	// Token token to use with the telegram API
	Token string `yaml:"token"`
	Passwd string `yaml:"passwd"`
}

// GetConfiguration returns the configuration
func GetConfiguration() (Configuration, error) {
	filePath := path.Join(getConfigPath(), configFileName)

	// open the configuration file
	log.Trace().Str("filePath", filePath).Msg("oppening configuration file")
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return Configuration{}, err
	}

	// parse the configuration file
	log.Trace().Msg("parsing the configuration file")
	var config Configuration
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return Configuration{}, err
	}

	// TelegramToken must not be empty
	if len(strings.Trim(config.Telegram.Token, string(' '))) == 0 {
		return Configuration{}, fmt.Errorf("TelegramToken is empty")
	}

	return config, nil
}

// getConfigPath returns the path to the configuration file
func getConfigPath() string {
	configPath := os.Getenv("CONFIG_PATH")
	if len(configPath) == 0 {
		configPath = defaultConfigPath
	}
	return configPath
}
