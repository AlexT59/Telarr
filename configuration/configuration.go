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

	configEnv = "CONFIG_PATH"
)

type Configuration struct {
	Telegram  Telegram  `yaml:"telegram"`
	Radarr    Radarr    `yaml:"radarr"`
	Sonarr    Sonarr    `yaml:"sonarr"`
	WakeOnLan WakeOnLan `yaml:"wakeOnLan"`

	PathForDiskUsage string `yaml:"pathForDiskUsage"`
}

type Telegram struct {
	// Token token to use with the telegram API
	Token  string `yaml:"token"`
	Passwd string `yaml:"passwd"`
}

type Radarr struct {
	ApiKey   string `yaml:"apiKey"`
	Endpoint string `yaml:"endpoint"`
}

type Sonarr struct {
	ApiKey   string `yaml:"apiKey"`
	Endpoint string `yaml:"endpoint"`
}

type WakeOnLan struct {
	MacAddress string `yaml:"mac"`
	IP         string `yaml:"ip"`
	Password   string `yaml:"password"`
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

	if len(bytes) == 0 {
		return Configuration{}, fmt.Errorf("configuration file is empty")
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

	// check if the endpoints contain http or https
	if !strings.HasPrefix(config.Radarr.Endpoint, "http") {
		config.Radarr.Endpoint = "http://" + config.Radarr.Endpoint
	}
	if !strings.HasPrefix(config.Sonarr.Endpoint, "http") {
		config.Sonarr.Endpoint = "http://" + config.Sonarr.Endpoint
	}

	return config, nil
}

// getConfigPath returns the path to the configuration file
func getConfigPath() string {
	configPath := os.Getenv(configEnv)
	if len(configPath) == 0 {
		configPath = defaultConfigPath
	}
	return configPath
}
