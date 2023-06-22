package main

import (
	"callCenterReportMaker/entity"
	"log"
	"regexp"
)

var (
	citiesAndLines = map[string]*regexp.Regexp{
		"Москва": regexp.MustCompile("^7[499|495|800]{3}\\d{7}$|^sip:[[:alnum:]]+@[[:alnum:]]+[.][[:alnum:]]+"),
		"Спб":    regexp.MustCompile("^7812\\d{7}$"),
		"Курск":  regexp.MustCompile("^7471\\d{7}$")}
	operators     = []string{"Алексей Поминов", "Алена Миронова", "Игорь Манойло", "Кирилл Соколовский", "Маргарита Долгушина", "Суханов Антон", "Татьяна Старцева"}
	motivationMap = map[float64]float64{50: 42.5, 49: 41, 48: 39.5, 47: 38, 46: 36.5, 45: 35, 44: 33.5, 43: 32, 42: 30.5, 41: 29, 40: 27.5, 39: 26, 38: 24.5, 37: 23, 36: 21.5, 35: 20, 0: 0}
	orderFee      = 80.0
)

func main() {

	report := entity.WeeklyReport{
		OperatorReports: []entity.OperatorReport{
			{Name: "Алексей Поминов",
				Salary:         6720,
				Bonus:          1600,
				SummaryPayment: 8320,
				OrdersCount:    84,
				PricePerOrder:  99.05,
				UniqCalls:      244,
				Conversion:     0.3443},
			{Name: "Алена Миронова",
				Salary:         3440,
				Bonus:          800,
				SummaryPayment: 4240,
				OrdersCount:    43,
				PricePerOrder:  98.60,
				UniqCalls:      95,
				Conversion:     0.4526},
			{Name: "Игорь Манойло",
				Salary:         2960,
				Bonus:          0,
				SummaryPayment: 2960,
				OrdersCount:    37,
				PricePerOrder:  80.00,
				UniqCalls:      139,
				Conversion:     0.2662},
			{Name: "Кирилл Соколовский",
				Salary:         5520,
				Bonus:          1300,
				SummaryPayment: 6820,
				OrdersCount:    69,
				PricePerOrder:  98.84,
				UniqCalls:      219,
				Conversion:     0.3151},
			{Name: "Маргарита Долгушина",
				Salary:         4640,
				Bonus:          1100,
				SummaryPayment: 5740,
				OrdersCount:    58,
				PricePerOrder:  98.97,
				UniqCalls:      182,
				Conversion:     0.3187},
			{Name: "Суханов Антон",
				Salary:         3280,
				Bonus:          1500,
				SummaryPayment: 4780,
				OrdersCount:    41,
				PricePerOrder:  116.59,
				UniqCalls:      141,
				Conversion:     0.2908},
			{Name: "Татьяна Старцева",
				Salary:         880,
				Bonus:          0,
				SummaryPayment: 880,
				OrdersCount:    11,
				PricePerOrder:  80,
				UniqCalls:      25,
				Conversion:     0.4400},
			{Name: "Виктор",
				Salary:         7480,
				Bonus:          1180,
				SummaryPayment: 8660},
		},
		DepartmentPayment:       42400,
		DepartmentPricePerOrder: 113.37,
		TelephonyPayment:        15697.68,
		SmsPayment:              5000,
		TotalExpenses:           63097.68,
		TotalOrdersCount:        374,
		TotalPricePerOrder:      168.71,
		CityStatistics: []entity.CityStatistic{
			{City: "Москва",
				UniqCallsTotal:    790,
				UniqCallsReceived: 771,
				UniqCallsMissed:   19,
				OrdersCount:       264,
				Conversion:        0.3342},
			{City: "Спб",
				UniqCallsTotal:    288,
				UniqCallsReceived: 281,
				UniqCallsMissed:   7,
				OrdersCount:       110,
				Conversion:        0.3819},
			{City: "Курск",
				UniqCallsTotal:    20,
				UniqCallsReceived: 18,
				UniqCallsMissed:   2,
				OrdersCount:       0,
				Conversion:        0},
			{City: "Итого",
				UniqCallsTotal:    1098,
				UniqCallsReceived: 1070,
				UniqCallsMissed:   28,
				OrdersCount:       374,
				Conversion:        0.3406},
		},
		SummaryDepartmentSalary: 34920,
		SummaryDepartmentBonus:  7480,
		SumToPay:                66215.50,
	}

	err := report.SaveAsXlsx("data/result.xlsx")
	if err != nil {
		log.Fatal(err)
	}

}
