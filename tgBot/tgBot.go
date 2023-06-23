package tgBot

import (
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
)

const (
	msgLayout = "```\n%s```"
	parseMode = "MarkdownV2"
)

type tgBot struct {
	token  string
	chatId int64
}

type TelegramMessageSender interface {
	SendPreformattedMessage(message string) error
}

func New(token string, chatId int64) TelegramMessageSender {
	return tgBot{
		token:  token,
		chatId: chatId,
	}
}

func (t tgBot) SendPreformattedMessage(message string) error {
	bot, err := tgbotapi.NewBotAPI(t.token)
	if err != nil {
		return err
	}
	// ТЕСТ			-887162800
	// ОПЕРАТОРЫ 	-1001531046100
	msg := tgbotapi.NewMessage(t.chatId, fmt.Sprintf(msgLayout, message))
	msg.ParseMode = parseMode
	_, err = bot.Send(msg)
	return err
}
