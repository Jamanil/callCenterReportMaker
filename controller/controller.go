package controller

import (
	"callCenterReportMaker/entity"
	"callCenterReportMaker/repository/database"
	"callCenterReportMaker/service"
	"callCenterReportMaker/tgBot"
	"fmt"
	"github.com/jasonlvhit/gocron"
	"log"
	"strings"
	"time"
)

type controller struct {
	srv                                  service.Service
	db                                   database.Database
	bot                                  tgBot.TelegramMessageSender
	weekdayReportTime, weekendReportTime string
}

type Controller interface {
	StartDailyReportSending()
	MakeReport(dateFrom, dateTo time.Time) (entity.WeeklyReport, error)
}

func New(srv service.Service, db database.Database, bot tgBot.TelegramMessageSender, weekdayReportTime, weekendReportTime string) Controller {
	return controller{
		srv:               srv,
		db:                db,
		bot:               bot,
		weekdayReportTime: weekdayReportTime,
		weekendReportTime: weekendReportTime,
	}
}

func (c controller) StartDailyReportSending() {

	if err := gocron.Every(1).Monday().At(c.weekdayReportTime).Do(c.sendStat); err != nil {
		log.Println(err)
	}
	if err := gocron.Every(1).Tuesday().At(c.weekdayReportTime).Do(c.sendStat); err != nil {
		log.Println(err)
	}
	if err := gocron.Every(1).Wednesday().At(c.weekdayReportTime).Do(c.sendStat); err != nil {
		log.Println(err)
	}
	if err := gocron.Every(1).Thursday().At(c.weekdayReportTime).Do(c.sendStat); err != nil {
		log.Println(err)
	}
	if err := gocron.Every(1).Friday().At(c.weekdayReportTime).Do(c.sendStat); err != nil {
		log.Println(err)
	}
	if err := gocron.Every(1).Saturday().At(c.weekendReportTime).Do(c.sendStat); err != nil {
		log.Println(err)
	}
	if err := gocron.Every(1).Sunday().At(c.weekendReportTime).Do(c.sendStat); err != nil {
		log.Println(err)
	}

	<-gocron.Start()
}

func (c controller) MakeReport(dateFrom, dateTo time.Time) (entity.WeeklyReport, error) {
	uniqCallsByOperators, err := c.db.GetUniqCallsByOperators(dateFrom, dateTo)
	if err != nil {
		return entity.WeeklyReport{}, err
	}
	orders, err := c.db.GetOrders(dateFrom, dateTo)
	if err != nil {
		return entity.WeeklyReport{}, err
	}

	callHistory, err := c.db.GetHistory(90)
	if err != nil {
		return entity.WeeklyReport{}, err
	}

	weeklyReport := c.srv.GetWeeklyReport(uniqCallsByOperators, orders, callHistory, dateFrom, dateTo)
	return weeklyReport, err
}
func (c controller) sendStat() {
	year, month, day := time.Now().Date()

	dateTo := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	dateFrom := getFirstDayOfCurrentWeek(dateTo)

	callsByOperator, err := c.db.GetUniqCallsByOperators(dateFrom, dateTo)
	if err != nil {
		log.Println(err)
	}

	orders, err := c.db.GetOrders(dateFrom, dateTo)
	if err != nil {
		log.Println(err)
	}

	dbStats := c.srv.GetDatabaseStatistic(callsByOperator, orders)
	message := dbStatPrettyString(dbStats, dateFrom, dateTo)
	err = c.bot.SendPreformattedMessage(message)
	if err != nil {
		log.Println(err)
	}
}

func dbStatPrettyString(dbStats []entity.DatabaseStatistic, dateFrom, dateTo time.Time) string {
	strBuilder := strings.Builder{}
	dateLayout := "02.01.2006"

	strBuilder.WriteString(fmt.Sprintf("Отчет по операторам за период с %s по %s\n",
		dateFrom.Format(dateLayout), dateTo.Format(dateLayout)))
	strBuilder.WriteString(fmt.Sprintf("%-20s %-7s %-7s %-7s %s\n",
		"ФИО", "заказы", "ун.вх.", "ун.исх.", "конв."))

	for i := 0; i < len(dbStats); i++ {
		if i != len(dbStats)-2 {
			strBuilder.WriteString(dbStats[i].String() + "\n")
		} else {
			strBuilder.WriteString(fmt.Sprintf("%-20s %-7d\n", dbStats[i].Operator, dbStats[i].OrdersCount))
		}
	}

	str := strBuilder.String()
	return str
}
func getFirstDayOfCurrentWeek(startDate time.Time) time.Time {
	switch startDate.Weekday() {
	case time.Tuesday:
		return startDate.AddDate(0, 0, -1)
	case time.Wednesday:
		return startDate.AddDate(0, 0, -2)
	case time.Thursday:
		return startDate.AddDate(0, 0, -3)
	case time.Friday:
		return startDate.AddDate(0, 0, -4)
	case time.Saturday:
		return startDate.AddDate(0, 0, -5)
	case time.Sunday:
		return startDate.AddDate(0, 0, -6)
	default:
		return startDate
	}
}
