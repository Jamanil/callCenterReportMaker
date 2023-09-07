package tgBot

import (
	"callCenterReportMaker/controller"
	"errors"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	msgLayout   = "```\n%s```"
	parseMode   = "MarkdownV2"
	adminChatId = 738984490
	reportPath  = "data/report.xlsx"
)

type tgBot struct {
	token                                      string
	chatId                                     int64
	controller                                 controller.Controller
	weekdayReportingTime, weekendReportingTime string
	tgApi                                      *tgbotapi.BotAPI
	updates                                    tgbotapi.UpdatesChannel
}

type TelegramStatisticsBot interface {
	SendPreformattedMessage(message string) error
	StartDailyReportSending()
	StartBot()
}

func New(controller controller.Controller, token string, chatId int64, weekdayReportingTime, weekendReportingTime string) TelegramStatisticsBot {
	bot, _ := tgbotapi.NewBotAPI(token)

	return tgBot{
		token:                token,
		controller:           controller,
		chatId:               chatId,
		weekdayReportingTime: weekdayReportingTime,
		weekendReportingTime: weekendReportingTime,
		tgApi:                bot,
	}
}

func (t tgBot) SendPreformattedMessage(message string) error {
	msg := tgbotapi.NewMessage(t.chatId, fmt.Sprintf(msgLayout, message))
	msg.ParseMode = parseMode
	_, err := t.tgApi.Send(msg)
	return err
}

func (t tgBot) StartDailyReportSending() {
	_ = gocron.Every(1).Monday().At(t.weekdayReportingTime).Do(t.MakeWeeklyConversionStatisticsAndSend)
	_ = gocron.Every(1).Tuesday().At(t.weekdayReportingTime).Do(t.MakeWeeklyConversionStatisticsAndSend)
	_ = gocron.Every(1).Wednesday().At(t.weekdayReportingTime).Do(t.MakeWeeklyConversionStatisticsAndSend)
	_ = gocron.Every(1).Thursday().At(t.weekdayReportingTime).Do(t.MakeWeeklyConversionStatisticsAndSend)
	_ = gocron.Every(1).Friday().At(t.weekdayReportingTime).Do(t.MakeWeeklyConversionStatisticsAndSend)
	_ = gocron.Every(1).Saturday().At(t.weekdayReportingTime).Do(t.MakeWeeklyConversionStatisticsAndSend)
	_ = gocron.Every(1).Sunday().At(t.weekdayReportingTime).Do(t.MakeWeeklyConversionStatisticsAndSend)

	<-gocron.Start()
}

func (t tgBot) MakeWeeklyConversionStatisticsAndSend() error {
	stats := t.controller.MakeWeeklyConversionStatistics()
	err := t.SendPreformattedMessage(stats)
	return err
}

func (t tgBot) StartBot() {
	go t.StartDailyReportSending()
	t.processMessages()
}

func (t tgBot) processMessages() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	var err error
	t.updates, err = t.tgApi.GetUpdatesChan(u)
	if err != nil {
		log.Println(err)
	}

	for update := range t.updates {

		chatId := update.Message.Chat.ID

		if !isAuthorized(chatId) {
			t.sendMsg(fmt.Sprintf("Неизвестный пользователь %s (%s %s) написал: \n%s",
				update.Message.From,
				update.Message.From.FirstName,
				update.Message.From.LastName,
				update.Message.Text))
		} else {
			selectCommand(t, update.Message.Text)

		}
	}
}

func (t tgBot) makeReport() {

	dateFrom, err := t.parseDateFromReader(t)
	if err != nil {
		t.sendMsg(err.Error())
		return
	}

	t.sendMsg("По какое число? (ДД.ММ.ГГГГ)")

	dateTo, err := t.parseDateFromReader(t)
	if err != nil {
		t.sendMsg(err.Error())
		return
	}

	t.sendMsg("Даты заданы. Считаем бонус")

	report, err := t.controller.MakeReport(dateFrom, dateTo, t)
	err = report.SaveAsXlsx(reportPath)
	if err != nil {
		t.sendMsg(err.Error())
		return
	}

	docConfig := tgbotapi.NewDocumentUpload(adminChatId, reportPath)
	_, err = t.tgApi.Send(docConfig)
	if err != nil {
		t.sendMsg(err.Error())
		return
	}

}

func selectCommand(t tgBot, usrTxt string) {
	switch usrTxt {
	case "Статистика":
		err := t.MakeWeeklyConversionStatisticsAndSend()
		if err != nil {
			t.sendMsg(err.Error())
		}
	case "Отчет":
		t.sendMsg("С какого числа? (ДД.ММ.ГГГГ)")
		t.makeReport()

	case "Пришли":
		docConfig := tgbotapi.NewDocumentUpload(adminChatId, reportPath)
		_, err := t.tgApi.Send(docConfig)
		if err != nil {
			t.sendMsg(err.Error())
		}

	default:
		t.sendMsg("Неизвестная команда")
	}
}

func isAuthorized(id int64) bool {
	return id == adminChatId
}

func (t tgBot) sendMsg(msg string) {
	_, err := t.tgApi.Send(tgbotapi.NewMessage(adminChatId, msg))
	if err != nil {
		log.Println(err)
	}
}

func (t tgBot) parseDateFromReader(reader io.Reader) (time.Time, error) {

	buf := make([]byte, 1024)
	var err error
	var n int
	for {
		n, err = reader.Read(buf)
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			log.Println(err)
			continue
		}

	}

	splitMessage := strings.Split(string(buf[:n]), ".")
	if len(splitMessage) != 3 {
		return time.Time{}, errors.New(fmt.Sprintf("Неверный формат даты, вместо 3 чисел введено %d", len(splitMessage)))
	}

	day, err := strconv.Atoi(splitMessage[0])
	if err != nil {
		return time.Time{}, err
	}

	month, err := strconv.Atoi(splitMessage[1])
	if err != nil {
		return time.Time{}, err
	}

	year, err := strconv.Atoi(splitMessage[2])
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

func (t tgBot) Read(b []byte) (n int, err error) {
	message := <-t.updates
	bytesFromTg := []byte(message.Message.Text)
	var count int
	for count = 0; count < len(bytesFromTg); count++ {
		b[count] = bytesFromTg[count]
	}
	return count, io.EOF
}

func (t tgBot) Write(b []byte) (n int, err error) {
	message := tgbotapi.NewMessage(adminChatId, string(b))
	_, err = t.tgApi.Send(message)
	return len(b), err
}
