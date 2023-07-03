package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	cfgKrokiUrl = "https://kroki.io/graphviz"
)

var (
	cfgLevel    = os.Getenv("LOG_LEVEL")
	cfgBotToken = os.Getenv("BOT_TOKEN")
	cfgFilesDir = os.Getenv("BOT_DIR")
)

func main() {
	log.Info().Msg("Starting bot")

	if level, e := zerolog.ParseLevel(cfgLevel); e == nil {
		zerolog.SetGlobalLevel(level)
	}

	if cfgFilesDir == "" {
		log.Fatal().Msg("empty files directory path")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	retChan := make(chan error, 1)

	go func() {
		err2 := loop(ctx, cfgBotToken)
		if err2 != nil {
			retChan <- err2
		}
		close(retChan)
	}()

	// Waiting signals from OS
	go func() {
		quit := make(chan os.Signal, 10)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
		log.Warn().Msgf("Signal '%s' was caught. Exiting", <-quit)
		cancel()
	}()

	// Listening for the main loop response
	for e := range retChan {
		log.Info().Err(e).Msg("Exiting.")
	}
}

func loop(ctx context.Context, token string) error {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal().Err(err)
	}

	bot.Debug = false

	log.Info().Msgf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Bye!")
			return ctx.Err()
		case update := <-updates:
			go processUpdate(bot, update)
		}
	}
}

func processUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	var inputmsg *tgbotapi.Message

	switch {
	case update.Message != nil:
		inputmsg = update.Message
	case update.EditedMessage != nil:
		inputmsg = update.EditedMessage
	}

	if inputmsg != nil {
		log.Info().Msgf("[%s] %s", inputmsg.From.UserName, inputmsg.Text)

		switch {
		case strings.HasPrefix(inputmsg.Text, "/kroki"):
			ProcessKroki(bot, inputmsg)
		case strings.HasPrefix(inputmsg.Text, "/read"):
			ProcessRead(bot, inputmsg)
		}

	}
}
