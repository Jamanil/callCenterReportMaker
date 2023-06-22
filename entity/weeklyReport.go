package entity

import (
	"github.com/plandem/xlsx"
	"github.com/plandem/xlsx/format/styles"
	"github.com/plandem/xlsx/types/options/column"
)

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
	xl := xlsx.New()

	sheet := xl.AddSheet("Weekly Report")

	sheet.Col(0).SetOptions(options.New(options.Width(35)))
	sheet.Col(1).SetOptions(options.New(options.Width(14)))
	sheet.Col(2).SetOptions(options.New(options.Width(14)))
	sheet.Col(3).SetOptions(options.New(options.Width(14)))
	sheet.Col(4).SetOptions(options.New(options.Width(17)))
	sheet.Col(5).SetOptions(options.New(options.Width(14)))
	sheet.Col(6).SetOptions(options.New(options.Width(8)))
	sheet.Col(7).SetOptions(options.New(options.Width(8)))

	var rowIndex int

	sheet.Cell(0, rowIndex).SetValue("ФИО")
	sheet.Cell(1, rowIndex).SetValue("ЗП")
	sheet.Cell(2, rowIndex).SetValue("Премия")
	sheet.Cell(3, rowIndex).SetValue("ЗП + Премия")
	sheet.Cell(4, rowIndex).SetValue("Принято заказов")
	sheet.Cell(5, rowIndex).SetValue("Цена за заказ")
	sheet.Cell(6, rowIndex).SetValue("ун. зв.")
	sheet.Cell(7, rowIndex).SetValue("конв.")

	currencyEvenStyle := styles.New(styles.NumberFormatID(6))
	currencyDivStyle := styles.New(styles.NumberFormatID(7))
	percentStyle := styles.New(styles.NumberFormatID(10))

	for _, report := range r.OperatorReports {
		rowIndex++
		_ = sheet.Cell(0, rowIndex).SetText(report.Name)

		sheet.Cell(1, rowIndex).SetValue(report.Salary)
		sheet.Cell(1, rowIndex).SetStyles(currencyEvenStyle)

		sheet.Cell(2, rowIndex).SetValue(report.Bonus)
		sheet.Cell(2, rowIndex).SetStyles(currencyEvenStyle)

		sheet.Cell(3, rowIndex).SetValue(report.SummaryPayment)
		sheet.Cell(3, rowIndex).SetStyles(currencyEvenStyle)

		sheet.Cell(4, rowIndex).SetValue(report.OrdersCount)

		sheet.Cell(5, rowIndex).SetValue(report.PricePerOrder)
		sheet.Cell(5, rowIndex).SetStyles(currencyDivStyle)

		sheet.Cell(6, rowIndex).SetValue(report.UniqCalls)

		sheet.Cell(7, rowIndex).SetValue(report.Conversion)
		sheet.Cell(7, rowIndex).SetStyles(percentStyle)
	}
	//delete boss's call stats
	sheet.Range(4, rowIndex, 7, rowIndex).Clear()

	//yellow row
	rowIndex++
	style := styles.New(styles.Fill.Color("FFFF00"), styles.Fill.Type(styles.PatternTypeSolid))
	for i := 0; i < 6; i++ {
		sheet.Cell(i, rowIndex).SetStyles(style)
	}
	sheet.Cell(0, rowIndex).SetValue("Цена заказа по операторам")
	sheet.Cell(1, rowIndex).SetValue(r.DepartmentPayment)
	style.Set(styles.NumberFormatID(6))
	sheet.Cell(1, rowIndex).SetStyles(style)
	sheet.Cell(5, rowIndex).SetValue(r.DepartmentPricePerOrder)
	style.Set(styles.NumberFormatID(7))
	sheet.Cell(5, rowIndex).SetStyles(style)

	rowIndex++
	sheet.Cell(0, rowIndex).SetValue("Манго")
	sheet.Cell(1, rowIndex).SetValue(r.TelephonyPayment)
	sheet.Cell(1, rowIndex).SetStyles(currencyEvenStyle)

	rowIndex++
	sheet.Cell(0, rowIndex).SetValue("Оплата СМС сервиса")
	sheet.Cell(1, rowIndex).SetValue(r.SmsPayment)
	sheet.Cell(1, rowIndex).SetStyles(currencyEvenStyle)

	//green bold row
	rowIndex++
	style = styles.New(styles.Fill.Type(styles.PatternTypeSolid), styles.Fill.Color("92D050"), styles.Font.Bold)
	for i := 0; i < 6; i++ {
		sheet.Cell(i, rowIndex).SetStyles(style)
	}
	sheet.Cell(0, rowIndex).SetValue("Итого")
	sheet.Cell(1, rowIndex).SetValue(r.TotalExpenses)
	style.Set(styles.NumberFormatID(6))
	sheet.Cell(1, rowIndex).SetStyles(style)
	sheet.Cell(4, rowIndex).SetValue(r.TotalOrdersCount)
	sheet.Cell(5, rowIndex).SetValue(r.TotalPricePerOrder)
	style.Set(styles.NumberFormatID(7))
	sheet.Cell(5, rowIndex).SetStyles(style)

	for i := 0; i < 6; i++ {
		rowIndex++
		switch i {
		case 1:
			sheet.Cell(0, rowIndex).SetValue("Звонков уникальных всего")
		case 2:
			sheet.Cell(0, rowIndex).SetValue("Звонков уникальных успешных")
		case 3:
			sheet.Cell(0, rowIndex).SetValue("Звонков уникальных пропущено")
		case 4:
			sheet.Cell(0, rowIndex).SetValue("Заказов принято")
		case 5:
			sheet.Cell(0, rowIndex).SetValue("Конверсия")
		}
		for j := 0; j < len(r.CityStatistics); j++ {
			switch i {
			case 0:
				sheet.Cell(1+j, rowIndex).SetValue(r.CityStatistics[j].City)
			case 1:
				sheet.Cell(1+j, rowIndex).SetValue(r.CityStatistics[j].UniqCallsTotal)
			case 2:
				sheet.Cell(1+j, rowIndex).SetValue(r.CityStatistics[j].UniqCallsReceived)
			case 3:
				sheet.Cell(1+j, rowIndex).SetValue(r.CityStatistics[j].UniqCallsMissed)
			case 4:
				sheet.Cell(1+j, rowIndex).SetValue(r.CityStatistics[j].OrdersCount)
			case 5:
				sheet.Cell(1+j, rowIndex).SetValue(r.CityStatistics[j].Conversion)
				sheet.Cell(1+j, rowIndex).SetStyles(percentStyle)
			}
		}

	}
	//add empty string
	rowIndex++

	rowIndex++
	sheet.Cell(0, rowIndex).SetValue("ЗП операторы, общая сумма")
	sheet.Cell(1, rowIndex).SetValue(r.SummaryDepartmentSalary)
	sheet.Cell(1, rowIndex).SetStyles(currencyEvenStyle)

	rowIndex++
	sheet.Cell(0, rowIndex).SetValue("Премия операторы, общая сумма")
	sheet.Cell(1, rowIndex).SetValue(r.SummaryDepartmentBonus)
	sheet.Cell(1, rowIndex).SetStyles(currencyEvenStyle)

	rowIndex++
	sheet.Cell(0, rowIndex).SetValue("Итого за неделю")
	sheet.Cell(1, rowIndex).SetValue(r.SumToPay)
	sheet.Cell(1, rowIndex).SetStyles(currencyEvenStyle)

	return xl.SaveAs(path)
}
