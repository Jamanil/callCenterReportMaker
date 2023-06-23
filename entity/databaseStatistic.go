package entity

import "fmt"

type DatabaseStatistic struct {
	Operator          string
	OrdersCount       int
	UniqIncomingCalls int
	UniqOutgoingCalls int
	Conversion        float64
}

func (d *DatabaseStatistic) AddIncomingCalls(n int) {
	d.UniqIncomingCalls += n
}

func (d *DatabaseStatistic) AddOutgoingCalls(n int) {
	d.UniqOutgoingCalls += n
}

func (d *DatabaseStatistic) AddOrders(n int) {
	d.OrdersCount += n
}

func (d *DatabaseStatistic) CalculateConversion() {
	totalCalls := d.UniqOutgoingCalls + d.UniqIncomingCalls
	if totalCalls == 0 {
		d.Conversion = 0
	} else {
		d.Conversion = float64(d.OrdersCount) / float64(totalCalls)
	}
}

func (d DatabaseStatistic) String() string {
	return fmt.Sprintf("%-20s %-7d %-7d %-7d %.4g%%", d.Operator, d.OrdersCount, d.UniqIncomingCalls, d.UniqOutgoingCalls, d.Conversion*100)
}
