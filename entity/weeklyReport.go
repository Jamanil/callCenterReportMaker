package entity

type WeeklyReport struct {
	OperatorReports         []OperatorReport
	DepartmentPayment       float64
	DepartmentPricePerOrder float64
	TelephonyPayment        float64
	SmsPayment              float64
	TotalExpenses           float64
	TotalOrdersCount        int
	TotalPricePerOrder      float64
	CityStatistics          []CityStatistic
	SummaryDepartmentSalary float64
	SummaryDepartmentBonus  float64
	SumToPay                float64
}

type OperatorReport struct {
	Name           string
	Salary         float64
	Bonus          float64
	SummaryPayment float64
	OrdersCount    int
	PricePerOrder  float64
	UniqCalls      int
	Conversion     float64
}

type CityStatistic struct {
	City              string
	UniqCallsTotal    int
	UniqCallsReceived int
	UniqCallsMissed   int
	OrdersCount       int
	Conversion        float64
}

func (r WeeklyReport) SaveAsXlsx(path string) error {
	return nil
}
