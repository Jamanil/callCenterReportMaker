package entity

import "time"

type HistoryRecord struct {
	Date       time.Time
	Abonent    string
	Operator   string
	LineNumber string
}
