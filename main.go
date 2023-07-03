package main

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	filesdir := os.Getenv("BOT_DIR")
	if filesdir == "" {
		log.Panic(errors.New("empty directory path"))
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	proceeding := false
	shutdown := false

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if shutdown {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Bot is not available now. Check it tommorow")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
			} else {
				proceeding = true

				if strings.HasPrefix(update.Message.Text, "/kroki") {
					// planttxt := strings.Split(update.Message.Text, "/kroki ")[1]
					// encoded, err := encode(planttxt)
					// if err != nil {
					// 	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Can not encode given plantUML")
					// 	msg.ReplyToMessageID = update.Message.MessageID
					// 	bot.Send(msg)
					// } else {
					// 	res, err := http.Get("https://kroki.io/graphviz/svg/" + encoded)
					// 	if err != nil {
					// 		fmt.Printf("error making http request: %s\n", err)
					// 		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Can not encode given plantUML")
					// 		msg.ReplyToMessageID = update.Message.MessageID
					// 		bot.Send(msg)
					// 	}
					// }
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Command is under constructions")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				} else if strings.HasPrefix(update.Message.Text, "/read") {
					filename := strings.Split(update.Message.Text, "/read ")[1]

					if _, err := os.Stat(filesdir + filename); errors.Is(err, os.ErrNotExist) {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "No such file")
						msg.ReplyToMessageID = update.Message.MessageID
						bot.Send(msg)
					} else {
						content, err := ioutil.ReadFile(filesdir + filename)

						if err != nil {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Something went wrong")
							msg.ReplyToMessageID = update.Message.MessageID
							bot.Send(msg)

						} else {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, string(content))
							msg.ReplyToMessageID = update.Message.MessageID
							bot.Send(msg)
						}
					}
				} else if strings.HasPrefix(update.Message.Text, "/exit") {
					shutdown = true
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Bot is going home")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				}
				proceeding = false
			}
			if shutdown && !proceeding {
				break
			}
		}
	}
}

// Encode takes a string and returns an encoded string in deflate + base64 format
func encode(input string) (string, error) {
	var buffer bytes.Buffer
	writer, err := zlib.NewWriterLevel(&buffer, 9)
	if err != nil {
		return "", errors.Join(err, errors.New("fail to create the writer"))
	}
	_, err = writer.Write([]byte(input))
	writer.Close()
	if err != nil {
		return "", errors.Join(err, errors.New("fail to create the payload"))
	}
	result := base64.URLEncoding.EncodeToString(buffer.Bytes())
	return result, nil
}
