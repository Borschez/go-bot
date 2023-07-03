package main

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

func ProcessRead(bot *tgbotapi.BotAPI, inputmsg *tgbotapi.Message) {
	_, cmd, _ := strings.Cut(inputmsg.Text, "/read ")
	_, filename, _ := strings.Cut(cmd, "@"+bot.Self.UserName+" ")
	fullpath := cfgFilesDir + "/" + filename

	if _, err := os.Stat(fullpath); errors.Is(err, os.ErrNotExist) {
		msg := tgbotapi.NewMessage(inputmsg.Chat.ID, "No such file")
		msg.ReplyToMessageID = inputmsg.MessageID
		bot.Send(msg)
	} else {
		content, err := ioutil.ReadFile(fullpath)

		if err != nil {
			msg := tgbotapi.NewMessage(inputmsg.Chat.ID, "Something went wrong")
			msg.ReplyToMessageID = inputmsg.MessageID
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(inputmsg.Chat.ID, string(content))
			msg.ReplyToMessageID = inputmsg.MessageID
			bot.Send(msg)
		}
	}
}

func ProcessKroki(bot *tgbotapi.BotAPI, inputmsg *tgbotapi.Message) {
	_, cmd, _ := strings.Cut(inputmsg.Text, "/kroki ")
	_, planttxt, _ := strings.Cut(cmd, "@"+bot.Self.UserName+" ")
	encoded, err := encode(planttxt)
	if err != nil {
		msg := tgbotapi.NewMessage(inputmsg.Chat.ID, "Can not encode given plantUML")
		msg.ReplyToMessageID = inputmsg.MessageID
		bot.Send(msg)
	} else {
		url := cfgKrokiUrl + "/png/" + encoded
		res, err := http.Get(url)
		if err != nil {
			log.Error().Err(err)
		}

		data, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Error().Err(err)
		}

		photo := tgbotapi.NewPhoto(inputmsg.Chat.ID, tgbotapi.FileBytes{Name: "figure.png", Bytes: data})
		photo.ReplyToMessageID = inputmsg.MessageID
		bot.Send(photo)
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
