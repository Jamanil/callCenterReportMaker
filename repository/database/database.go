package database

import (
	"callCenterReportMaker/entity"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"slices"
	"sort"
	"strings"
	"time"
)

const (
	dbDateLayout = "2006-01-02"
)

type Database interface {
	GetHistory(frameWidthInDays int) ([]entity.HistoryRecord, error)
	GetOrders(dateFrom, dateTo time.Time) ([]entity.Orders, error)
	GetUniqCallsByOperators(dateFrom, dateTo time.Time) ([]entity.DatabaseStatistic, error)
}

type database struct {
	db        *sql.DB
	operators []string
}

func New(host, port, dbname, user, password string, operatorsNames []string) Database {
	cfg := mysql.Config{
		User:                 user,
		Passwd:               password,
		Net:                  "tcp",
		Addr:                 host + ":" + port,
		DBName:               dbname,
		AllowNativePasswords: true,
	}
	var err error
	var db database
	db.db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	err = db.db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	db.operators = operatorsNames

	return db
}

func (d database) GetHistory(frameWidthInDays int) ([]entity.HistoryRecord, error) {
	//goland:noinspection SpellCheckingInspection
	rows, err := d.db.Query(
		`SELECT data_postupil_vkompan, tel_kto_zvonil, komu_zvonil, kuda_zvonil FROM mango_history
					WHERE
    			data_postupil_vkompan BETWEEN ? AND ?
    			AND (gruppa LIKE '7 Операторы%' OR gruppa LIKE '%Курск первоначальные обращения' OR gruppa LIKE '04 Курск')
    			AND napravlenie LIKE 'Входящий%';`,
		time.Now().AddDate(0, 0, -frameWidthInDays),
		time.Now(),
	)
	if err != nil {
		return nil, err
	}

	historyRecords := make([]entity.HistoryRecord, 0, 10000)

	for rows.Next() {
		var dateStr, abonent, operator, lineNumber string
		_ = rows.Scan(&dateStr, &abonent, &operator, &lineNumber)

		historyRecords = append(historyRecords, entity.HistoryRecord{
			Date:       d.parseTime(dateStr),
			Abonent:    abonent,
			Operator:   d.normalizeOperatorName(operator),
			LineNumber: lineNumber,
		})
	}

	return historyRecords, err
}

func (d database) GetOrders(dateFrom, dateTo time.Time) ([]entity.Orders, error) {
	//goland:noinspection SpellCheckingInspection
	rows, err := d.db.Query(`SELECT orders.id, orders.date_add_, cities.name, users.fio  FROM orders
    JOIN cities on cities.city_id = orders.city_id
    LEFT JOIN users on users.id = orders.id_operator
WHERE date_add_ BETWEEN ? AND ?;`, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	orders := make([]entity.Orders, 0, 500)
	for rows.Next() {
		var dateStr, city, operator string
		var id uint
		_ = rows.Scan(&id, &dateStr, &city, &operator)

		orders = append(orders, entity.Orders{
			Id:       id,
			Date:     d.parseTime(dateStr),
			City:     city,
			Operator: d.normalizeOperatorName(operator),
		})
	}

	return orders, nil
}

func (d database) GetUniqCallsByOperators(dateFrom, dateTo time.Time) ([]entity.DatabaseStatistic, error) {
	//goland:noinspection SpellCheckingInspection
	rows, err := d.db.Query(
		`SELECT komu_zvonil, napravlenie mango_history FROM mango_history
				WHERE
				data_postupil_vkompan BETWEEN ? AND ?
				AND unik = 1
				AND (napravlenie = 'Входящий внешний вызов' OR napravlenie = 'Исходящий внешний вызов')
				AND (gruppa LIKE '7 Операторы%');`, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	statMap := make(map[string]entity.DatabaseStatistic)
	for _, operator := range d.operators {
		statMap[operator] = entity.DatabaseStatistic{Operator: operator}
	}

	for rows.Next() {
		var operator, direction string
		_ = rows.Scan(&operator, &direction)

		if slices.Contains(d.operators, operator) {
			if stat, ok := statMap[operator]; ok {
				switch direction {
				case "Исходящий внешний вызов":
					stat.AddOutgoingCalls(1)
				case "Входящий внешний вызов":
					stat.AddIncomingCalls(1)
				}
				statMap[operator] = stat
			}
		}
	}

	statistics := make([]entity.DatabaseStatistic, 0, len(statMap))
	for _, statistic := range statMap {
		statistics = append(statistics, statistic)
	}

	sort.Slice(statistics, func(i, j int) bool { return statistics[i].Operator < statistics[j].Operator })

	return statistics, nil
}

func (d database) parseTime(dateStr string) time.Time {
	date, _ := time.Parse(dbDateLayout, dateStr)
	return date
}

func (d database) normalizeOperatorName(name string) string {
	for _, normalizedName := range d.operators {
		if strings.Contains(strings.ToLower(name), strings.ToLower(normalizedName)) {
			return normalizedName
		}
	}
	return name
}
