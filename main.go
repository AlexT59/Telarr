package main

import (
	"context"
	"os"
	"os/signal"
	"telarr/configuration"
	"telarr/internal/updates"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	ret := 0
	defer func() {
		os.Exit(ret)
	}()

	// get the log level
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	lv, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Err(err).Msg("error when parsing the log level")
		ret = 1
		return
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}).Level(lv).With().Logger()

	// getting the configuration
	log.Debug().Msg("getting the configuration")
	config, err := configuration.GetConfiguration()
	if err != nil {
		log.Err(err).Msg("error when getting the configuration")
		ret = 1
		return
	}

	// creating the context
	ctx, cancel := context.WithCancel(context.Background())

	// catching the signal interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// handle updates
	mess, err := updates.New(config)
	if err != nil {
		log.Err(err).Msg("error when creating the messages handler")
		ret = 1
		return
	}
	err = mess.Start(ctx)
	if err != nil {
		log.Err(err).Msg("error when starting the messages handler")
		ret = 1
		return
	}

	// wait for the signal interrupt
	<-c
	log.Info().Msg("signal interrupt received")
	cancel()

	// stop all running services
	err = mess.Stop()
	if err != nil {
		log.Err(err).Msg("error when stopping the bot")
		ret = 1
	}
}
