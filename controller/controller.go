package controller

import (
	"callCenterReportMaker/entity"
	"callCenterReportMaker/repository/database"
	"callCenterReportMaker/service"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
)

type controller struct {
	srv service.Service
	db  database.Database
}
type Controller interface {
	MakeReport(dateFrom, dateTo time.Time, readWriter io.ReadWriter) (entity.WeeklyReport, error)
	MakeWeeklyConversionStatistics() string
}

func New(srv service.Service, db database.Database) Controller {
	return controller{
		srv: srv,
		db:  db,
	}
}

func (c controller) MakeReport(dateFrom, dateTo time.Time, readWriter io.ReadWriter) (entity.WeeklyReport, error) {
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

	weeklyReport := c.srv.GetWeeklyReport(uniqCallsByOperators, orders, callHistory, dateFrom, dateTo, readWriter)
	return weeklyReport, err
}
func (c controller) MakeWeeklyConversionStatistics() string {
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
	return message
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
			if dbStats[i].OrdersCount != 0 || dbStats[i].UniqOutgoingCalls != 0 || dbStats[i].UniqIncomingCalls != 0 {
				strBuilder.WriteString(dbStats[i].String() + "\n")
			}
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
