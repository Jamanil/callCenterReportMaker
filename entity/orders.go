package entity

import "time"

type Orders struct {
	Id       uint
	Date     time.Time
	City     string
	Operator string
}
