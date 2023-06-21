package service

import (
	"callCenterReportMaker/entity"
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"log"
	"math"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Service interface {
	GetUniqTotalCallsCountPerCity(historyRecords []entity.HistoryRecord, dateFrom, dateTo time.Time) map[string]int
	GetUniqReceivedCallsCountPerCity(historyRecords []entity.HistoryRecord, dateFrom, dateTo time.Time) map[string]int
	GetUniqReceivedCallsByOperator(historyRecords []entity.HistoryRecord, dateFrom, dateTo time.Time) map[time.Time]map[string]int
	GetDatabaseStatistic(callsByOperators []entity.DatabaseStatistic, orders []entity.Orders) []entity.DatabaseStatistic
	GetSalaries(databaseStatistics []entity.DatabaseStatistic, orders []entity.Orders, history []entity.HistoryRecord, dateFrom, dateTo time.Time)
}

func New(citiesLineMap map[string]*regexp.Regexp, operatorsList []string, bonusMap map[float64]float64, orderCost float64) Service {
	return &service{
		citiesAndLines: citiesLineMap,
		operators:      operatorsList,
		motivationMap:  bonusMap,
		orderFee:       orderCost,
	}
}

type service struct {
	citiesAndLines map[string]*regexp.Regexp
	operators      []string
	motivationMap  map[float64]float64
	orderFee       float64
}

func (s *service) GetUniqTotalCallsCountPerCity(historyRecords []entity.HistoryRecord, dateFrom, dateTo time.Time) map[string]int {
	return s.getUniqCallsWithFilter(historyRecords, func(record entity.HistoryRecord) bool {
		return s.isDateBetween(dateFrom, dateTo, record.Date)
	})
}

func (s *service) GetUniqReceivedCallsCountPerCity(historyRecords []entity.HistoryRecord, dateFrom, dateTo time.Time) map[string]int {
	return s.getUniqCallsWithFilter(historyRecords, func(record entity.HistoryRecord) bool {
		return s.isDateBetween(dateFrom, dateTo, record.Date) && record.Operator != ""
	})
}

func (s *service) GetUniqReceivedCallsByOperator(historyRecords []entity.HistoryRecord, dateFrom, dateTo time.Time) map[time.Time]map[string]int {
	uniqCallsMap := make(map[string]struct{})

	result := make(map[time.Time]map[string]int)

	for _, record := range historyRecords {
		if _, recordExists := uniqCallsMap[record.Abonent]; !recordExists {
			uniqCallsMap[record.Abonent] = struct{}{}
			if s.isDateBetween(dateFrom, dateTo, record.Date) && slices.Contains(s.operators, record.Operator) {
				if _, dateExists := result[record.Date]; !dateExists {
					result[record.Date] = make(map[string]int)
				}
				result[record.Date][record.Operator]++
			}
		}
	}

	return result
}

func (s *service) GetOrdersPerCity(orders []entity.Orders) map[string]int {
	cities := make([]string, 0, len(s.citiesAndLines))
	for city := range s.citiesAndLines {
		cities = append(cities, city)
	}

	ordersPerCity := make(map[string]int)

	for _, order := range orders {
		for _, city := range cities {
			if strings.Contains(strings.ToLower(order.City), strings.ToLower(city)) {
				ordersPerCity[city]++
			}
		}
	}
	return ordersPerCity
}

func (s *service) GetDatabaseStatistic(callsByOperators []entity.DatabaseStatistic, orders []entity.Orders) []entity.DatabaseStatistic {
	var totalIncomingCalls, totalOutgoingCalls, wildOrdersCount int

	statMap := make(map[string]entity.DatabaseStatistic)
	for _, operatorCalls := range callsByOperators {
		statMap[operatorCalls.Operator] = operatorCalls
	}

	for _, order := range orders {
		if order.Operator == "" {
			wildOrdersCount++
			continue
		}
		if stat, ok := statMap[order.Operator]; ok {
			stat.AddOrders(1)
			statMap[order.Operator] = stat
		}
	}

	databaseStatistics := make([]entity.DatabaseStatistic, 0, len(statMap))

	for _, statistic := range statMap {
		statistic.CalculateConversion()
		totalIncomingCalls += statistic.UniqIncomingCalls
		totalOutgoingCalls += statistic.UniqOutgoingCalls
		databaseStatistics = append(databaseStatistics, statistic)
	}
	sort.Slice(databaseStatistics, func(i, j int) bool { return databaseStatistics[i].Operator < databaseStatistics[j].Operator })
	wildOrders := entity.DatabaseStatistic{
		Operator:    "Бесхозные заказы",
		OrdersCount: wildOrdersCount}
	total := entity.DatabaseStatistic{
		Operator:          "ИТОГО",
		OrdersCount:       len(orders),
		UniqIncomingCalls: totalIncomingCalls,
		UniqOutgoingCalls: totalOutgoingCalls}
	total.CalculateConversion()
	databaseStatistics = append(databaseStatistics, wildOrders, total)
	return databaseStatistics
}

func (s *service) GetSalaries(databaseStatistics []entity.DatabaseStatistic, orders []entity.Orders, history []entity.HistoryRecord, dateFrom, dateTo time.Time) {
	totalStat := databaseStatistics[len(databaseStatistics)-1]
	//goland:noinspection SpellCheckingInspection
	report := struct {
		SalariesTitle struct {
			Operator         string
			Salary           string
			Bonus            string
			ToPay            string
			OrdersCount      string
			MoneyPerOrder    string
			SummaryUniqCalls string
			Conversion       string
		}
		Salaries []struct {
			Operator         string
			Salary           float64
			Bonus            float64
			ToPay            float64
			OrdersCount      int
			MoneyPerOrder    float64
			SummaryUniqCalls int
			Conversion       float64
		}
		OperatorsMoney struct {
			Description            string
			OperatorsMoney         float64
			OperatorsMoneyPerOrder float64
		}
		CommunicationsFee []struct {
			Description string
			Cost        float64
		}
		TotalByStaff struct {
			Description        string
			TotalMoney         float64
			TotalOrders        int
			TotalMoneyPerOrder float64
		}
		CitiesStatistic struct {
			Titles struct {
				Moscow string
				Spb    string
				Kursk  string
				Total  string
			}
			UniqCallsTotal struct {
				Description string
				Msk         int
				Spb         int
				Kursk       int
				Total       int
			}
			UniqCallsReceived struct {
				Description string
				Msk         int
				Spb         int
				Kursk       int
				Total       int
			}
			UniqCallsMissed struct {
				Description string
				Msk         int
				Spb         int
				Kursk       int
				Total       int
			}
			OrdersPerCity struct {
				Description string
				Msk         int
				Spb         int
				Kursk       int
				Total       int
			}
			ConversionPerCity struct {
				Description string
				Msk         float64
				Spb         float64
				Kursk       float64
				Total       float64
			}
		}
		SummaryStaffSalary struct {
			Description   string
			SummarySalary float64
		}
		Bonus struct {
			Description string
			Bonus       float64
		}
		AdditionalPayments []struct {
			Description string
			Money       float64
		}
		Total struct {
			Description string
			Money       float64
		}
	}{
		SalariesTitle: struct {
			Operator         string
			Salary           string
			Bonus            string
			ToPay            string
			OrdersCount      string
			MoneyPerOrder    string
			SummaryUniqCalls string
			Conversion       string
		}{
			Operator:         "ФИО",
			Salary:           "ЗП",
			Bonus:            "Премия",
			ToPay:            "ЗП + Премия",
			OrdersCount:      "Принято заказов",
			MoneyPerOrder:    "Цена за заказ",
			SummaryUniqCalls: "ун. зв.",
			Conversion:       "конв.",
		},
		OperatorsMoney: struct {
			Description            string
			OperatorsMoney         float64
			OperatorsMoneyPerOrder float64
		}{
			Description:            "Цена заказа по операторам",
			OperatorsMoney:         0,
			OperatorsMoneyPerOrder: 0,
		},
		TotalByStaff: struct {
			Description        string
			TotalMoney         float64
			TotalOrders        int
			TotalMoneyPerOrder float64
		}{
			Description:        "Итого",
			TotalMoney:         0,
			TotalOrders:        0,
			TotalMoneyPerOrder: 0,
		},
		CitiesStatistic: struct {
			Titles struct {
				Moscow string
				Spb    string
				Kursk  string
				Total  string
			}
			UniqCallsTotal struct {
				Description string
				Msk         int
				Spb         int
				Kursk       int
				Total       int
			}
			UniqCallsReceived struct {
				Description string
				Msk         int
				Spb         int
				Kursk       int
				Total       int
			}
			UniqCallsMissed struct {
				Description string
				Msk         int
				Spb         int
				Kursk       int
				Total       int
			}
			OrdersPerCity struct {
				Description string
				Msk         int
				Spb         int
				Kursk       int
				Total       int
			}
			ConversionPerCity struct {
				Description string
				Msk         float64
				Spb         float64
				Kursk       float64
				Total       float64
			}
		}{
			Titles: struct {
				Moscow string
				Spb    string
				Kursk  string
				Total  string
			}{
				Moscow: "Москва",
				Spb:    "Спб",
				Kursk:  "Курск",
				Total:  "Итого",
			},
			UniqCallsTotal: struct {
				Description string
				Msk         int
				Spb         int
				Kursk       int
				Total       int
			}{
				Description: "Звонков уникальных всего",
				Msk:         0,
				Spb:         0,
				Kursk:       0,
				Total:       0,
			},
			UniqCallsReceived: struct {
				Description string
				Msk         int
				Spb         int
				Kursk       int
				Total       int
			}{
				Description: "Звонков уникальных успешных",
				Msk:         0,
				Spb:         0,
				Kursk:       0,
				Total:       0,
			},
			UniqCallsMissed: struct {
				Description string
				Msk         int
				Spb         int
				Kursk       int
				Total       int
			}{
				Description: "Звонков уникальных пропущено",
				Msk:         0,
				Spb:         0,
				Kursk:       0,
				Total:       0,
			},
			OrdersPerCity: struct {
				Description string
				Msk         int
				Spb         int
				Kursk       int
				Total       int
			}{
				Description: "Заказов принято",
				Msk:         0,
				Spb:         0,
				Kursk:       0,
				Total:       0,
			},
			ConversionPerCity: struct {
				Description string
				Msk         float64
				Spb         float64
				Kursk       float64
				Total       float64
			}{
				Description: "Конверсия",
				Msk:         0,
				Spb:         0,
				Kursk:       0,
				Total:       0,
			},
		},
		SummaryStaffSalary: struct {
			Description   string
			SummarySalary float64
		}{
			Description:   "ЗП операторы, общая сумма",
			SummarySalary: 0,
		},
		Bonus: struct {
			Description string
			Bonus       float64
		}{
			Description: "Премия операторы, общая сумма",
			Bonus:       0,
		},
		AdditionalPayments: nil,
		Total: struct {
			Description string
			Money       float64
		}{
			Description: "Итого за неделю",
			Money:       0,
		},
	}

	for _, statistic := range databaseStatistics {
		if slices.Contains(s.operators, statistic.Operator) {
			report.Salaries = append(report.Salaries, struct {
				Operator         string
				Salary           float64
				Bonus            float64
				ToPay            float64
				OrdersCount      int
				MoneyPerOrder    float64
				SummaryUniqCalls int
				Conversion       float64
			}{Operator: statistic.Operator,
				Salary:           float64(statistic.OrdersCount) * s.orderFee,
				Bonus:            0,
				ToPay:            0,
				OrdersCount:      statistic.OrdersCount,
				MoneyPerOrder:    0,
				SummaryUniqCalls: statistic.UniqIncomingCalls + statistic.UniqOutgoingCalls,
				Conversion:       statistic.Conversion * 100})
		}
	}

	motivationGrades := make([]float64, 0, len(s.motivationMap))
	for grade, _ := range s.motivationMap {
		motivationGrades = append(motivationGrades, grade)
	}
	sort.Slice(motivationGrades, func(i, j int) bool { return motivationGrades[i] > motivationGrades[j] })

	var bonusPerOrder, personalBonusPerOrder float64
	for _, grade := range motivationGrades {
		if totalStat.Conversion*100 > grade {
			bonusPerOrder = s.motivationMap[grade]
			break
		}
	}
	var summaryOperatorsBonus float64
	if bonusPerOrder > 0 {
		report.Bonus.Bonus = math.RoundToEven(bonusPerOrder * float64(totalStat.OrdersCount))
		fmt.Printf("Поздравляю, конверсии хватило на премию. Премия за заказ %.3g, суммарная премия %f. Сколько раздать на брата?\n", bonusPerOrder, report.Bonus.Bonus)
		for {
			_, err := fmt.Scan(&personalBonusPerOrder)
			if err != nil {
				fmt.Printf("Размер персональной премии %3g\n", personalBonusPerOrder)
				break
			}

			for i := 0; i < len(databaseStatistics)-2; i++ {
				if databaseStatistics[i].Conversion > 0.28 {
					report.Salaries[i].Bonus = float64(int(math.RoundToEven(personalBonusPerOrder*float64(databaseStatistics[i].OrdersCount))) / 100 * 100)
				}
				fmt.Println(databaseStatistics[i].Operator, report.Salaries[i].Bonus)
				summaryOperatorsBonus += report.Salaries[i].Bonus
			}
			fmt.Println("Виктор", report.Bonus.Bonus-summaryOperatorsBonus)
			fmt.Println("Годится или переиграть?")
		}
	}

	for i, salary := range report.Salaries {
		toPay := salary.Bonus + salary.Salary
		report.Salaries[i].ToPay = toPay
		report.Salaries[i].MoneyPerOrder = toPay / float64(salary.OrdersCount)
		report.SummaryStaffSalary.SummarySalary += salary.Salary
	}
	chefSalary := float64(totalStat.OrdersCount * 20)
	chefBonus := report.Bonus.Bonus - summaryOperatorsBonus
	report.SummaryStaffSalary.SummarySalary += chefSalary

	report.Salaries = append(report.Salaries, struct {
		Operator         string
		Salary           float64
		Bonus            float64
		ToPay            float64
		OrdersCount      int
		MoneyPerOrder    float64
		SummaryUniqCalls int
		Conversion       float64
	}{Operator: "Виктор",
		Salary:      chefSalary,
		Bonus:       chefBonus,
		ToPay:       chefSalary + chefBonus,
		OrdersCount: 0, MoneyPerOrder: 0, SummaryUniqCalls: 0, Conversion: 0})

	report.OperatorsMoney.OperatorsMoney = report.SummaryStaffSalary.SummarySalary + report.Bonus.Bonus
	report.OperatorsMoney.OperatorsMoneyPerOrder = report.OperatorsMoney.OperatorsMoney / float64(totalStat.OrdersCount)

	report.CommunicationsFee = make([]struct {
		Description string
		Cost        float64
	},
		2)

	report.CommunicationsFee[0].Description = "Манго"
	report.CommunicationsFee[1].Description = "Оплата СМС сервиса"

	fmt.Println("Сколько заплатили за телефонию?")
	_, _ = fmt.Scan(&report.CommunicationsFee[0].Cost)
	fmt.Println("Сколько заплатили за СМС?")
	_, _ = fmt.Scan(&report.CommunicationsFee[1].Cost)

	report.TotalByStaff.TotalMoney = report.OperatorsMoney.OperatorsMoney + report.CommunicationsFee[0].Cost + report.CommunicationsFee[1].Cost
	report.TotalByStaff.TotalOrders = totalStat.OrdersCount
	report.TotalByStaff.TotalMoneyPerOrder = report.TotalByStaff.TotalMoney / float64(totalStat.OrdersCount)

	report.Total.Money = report.SummaryStaffSalary.SummarySalary + report.Bonus.Bonus

	totalCallsPerCity := s.GetUniqTotalCallsCountPerCity(history, dateFrom, dateTo)
	report.CitiesStatistic.UniqCallsTotal.Msk = totalCallsPerCity["Москва"]
	report.CitiesStatistic.UniqCallsTotal.Spb = totalCallsPerCity["Спб"]
	report.CitiesStatistic.UniqCallsTotal.Kursk = totalCallsPerCity["Курск"]
	report.CitiesStatistic.UniqCallsTotal.Total = totalCallsPerCity["Москва"] + totalCallsPerCity["Спб"] + totalCallsPerCity["Курск"]

	receivedCallsPerCity := s.GetUniqReceivedCallsCountPerCity(history, dateFrom, dateTo)
	report.CitiesStatistic.UniqCallsReceived.Msk = receivedCallsPerCity["Москва"]
	report.CitiesStatistic.UniqCallsReceived.Spb = receivedCallsPerCity["Спб"]
	report.CitiesStatistic.UniqCallsReceived.Kursk = receivedCallsPerCity["Курск"]
	report.CitiesStatistic.UniqCallsReceived.Total = receivedCallsPerCity["Москва"] + receivedCallsPerCity["Спб"] + receivedCallsPerCity["Курск"]

	report.CitiesStatistic.UniqCallsMissed.Msk = report.CitiesStatistic.UniqCallsTotal.Msk - report.CitiesStatistic.UniqCallsReceived.Msk
	report.CitiesStatistic.UniqCallsMissed.Spb = report.CitiesStatistic.UniqCallsTotal.Spb - report.CitiesStatistic.UniqCallsReceived.Spb
	report.CitiesStatistic.UniqCallsMissed.Kursk = report.CitiesStatistic.UniqCallsTotal.Kursk - report.CitiesStatistic.UniqCallsReceived.Kursk
	report.CitiesStatistic.UniqCallsMissed.Total = report.CitiesStatistic.UniqCallsTotal.Total - report.CitiesStatistic.UniqCallsReceived.Total

	ordersPerCity := s.GetOrdersPerCity(orders)

	report.CitiesStatistic.OrdersPerCity.Msk = ordersPerCity["Москва"]
	report.CitiesStatistic.OrdersPerCity.Spb = ordersPerCity["Спб"]
	report.CitiesStatistic.OrdersPerCity.Kursk = ordersPerCity["Курск"]
	report.CitiesStatistic.OrdersPerCity.Total = ordersPerCity["Москва"] + ordersPerCity["Спб"] + ordersPerCity["Курск"]

	getConv := func(totalCalls int, orders int) float64 {
		if totalCalls == 0 {
			return 0
		} else {
			return float64(orders) / float64(totalCalls) * 100
		}
	}

	report.CitiesStatistic.ConversionPerCity.Msk = getConv(report.CitiesStatistic.UniqCallsTotal.Msk, report.CitiesStatistic.OrdersPerCity.Msk)
	report.CitiesStatistic.ConversionPerCity.Spb = getConv(report.CitiesStatistic.UniqCallsTotal.Spb, report.CitiesStatistic.OrdersPerCity.Spb)
	report.CitiesStatistic.ConversionPerCity.Kursk = getConv(report.CitiesStatistic.UniqCallsTotal.Kursk, report.CitiesStatistic.OrdersPerCity.Kursk)
	report.CitiesStatistic.ConversionPerCity.Total = getConv(report.CitiesStatistic.UniqCallsTotal.Total, report.CitiesStatistic.OrdersPerCity.Total)

	file, err := os.Create("data/result.csv")
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	encWriter := transform.NewWriter(file, charmap.Windows1251.NewEncoder().Transformer)

	encWriter.Write([]byte(strings.Join([]string{
		report.SalariesTitle.Operator,
		report.SalariesTitle.Salary,
		report.SalariesTitle.Bonus,
		report.SalariesTitle.ToPay,
		report.SalariesTitle.OrdersCount,
		report.SalariesTitle.MoneyPerOrder,
		report.SalariesTitle.SummaryUniqCalls,
		report.SalariesTitle.Conversion,
		"\n"}, ";")))

	for _, salary := range report.Salaries {
		encWriter.Write([]byte(strings.Join([]string{
			salary.Operator,
			strconv.FormatFloat(salary.Salary, 'f', 2, 64),
			strconv.FormatFloat(salary.Bonus, 'f', 2, 64),
			strconv.FormatFloat(salary.ToPay, 'f', 2, 64),
			strconv.Itoa(salary.OrdersCount),
			strconv.FormatFloat(salary.MoneyPerOrder, 'f', 2, 64),
			strconv.Itoa(salary.SummaryUniqCalls),
			strconv.FormatFloat(salary.Conversion, 'f', 2, 64),
			"\n"},
			";")))
	}
	encWriter.Write([]byte(strings.Join([]string{
		report.OperatorsMoney.Description,
		strconv.FormatFloat(report.OperatorsMoney.OperatorsMoney, 'f', 2, 64),
		"", "", "",
		strconv.FormatFloat(report.OperatorsMoney.OperatorsMoneyPerOrder, 'f', 2, 64),
		"\n"}, ";")))

	for _, fee := range report.CommunicationsFee {
		encWriter.Write([]byte(strings.Join([]string{
			fee.Description,
			strconv.FormatFloat(fee.Cost, 'f', 2, 64),
			"\n"}, ";")))
	}

	encWriter.Write([]byte(strings.Join([]string{
		report.TotalByStaff.Description,
		strconv.FormatFloat(report.TotalByStaff.TotalMoney, 'f', 2, 64),
		"", "",
		strconv.Itoa(report.TotalByStaff.TotalOrders),
		strconv.FormatFloat(report.TotalByStaff.TotalMoneyPerOrder, 'f', 2, 64),
		"\n"}, ";")))

	encWriter.Write([]byte(strings.Join([]string{
		"",
		report.CitiesStatistic.Titles.Moscow,
		report.CitiesStatistic.Titles.Spb,
		report.CitiesStatistic.Titles.Kursk,
		report.CitiesStatistic.Titles.Total,
		"\n"}, ";")))
	encWriter.Write([]byte(strings.Join([]string{
		report.CitiesStatistic.UniqCallsTotal.Description,
		strconv.Itoa(report.CitiesStatistic.UniqCallsTotal.Msk),
		strconv.Itoa(report.CitiesStatistic.UniqCallsTotal.Spb),
		strconv.Itoa(report.CitiesStatistic.UniqCallsTotal.Kursk),
		strconv.Itoa(report.CitiesStatistic.UniqCallsTotal.Total),
		"\n"}, ";")))
	encWriter.Write([]byte(strings.Join([]string{
		report.CitiesStatistic.UniqCallsReceived.Description,
		strconv.Itoa(report.CitiesStatistic.UniqCallsReceived.Msk),
		strconv.Itoa(report.CitiesStatistic.UniqCallsReceived.Spb),
		strconv.Itoa(report.CitiesStatistic.UniqCallsReceived.Kursk),
		strconv.Itoa(report.CitiesStatistic.UniqCallsReceived.Total),
		"\n"}, ";")))
	encWriter.Write([]byte(strings.Join([]string{
		report.CitiesStatistic.UniqCallsMissed.Description,
		strconv.Itoa(report.CitiesStatistic.UniqCallsMissed.Msk),
		strconv.Itoa(report.CitiesStatistic.UniqCallsMissed.Spb),
		strconv.Itoa(report.CitiesStatistic.UniqCallsMissed.Kursk),
		strconv.Itoa(report.CitiesStatistic.UniqCallsMissed.Total),
		"\n"}, ";")))
	encWriter.Write([]byte(strings.Join([]string{
		report.CitiesStatistic.OrdersPerCity.Description,
		strconv.Itoa(report.CitiesStatistic.OrdersPerCity.Msk),
		strconv.Itoa(report.CitiesStatistic.OrdersPerCity.Spb),
		strconv.Itoa(report.CitiesStatistic.OrdersPerCity.Kursk),
		strconv.Itoa(report.CitiesStatistic.OrdersPerCity.Total),
		"\n"}, ";")))
	encWriter.Write([]byte(strings.Join([]string{
		report.CitiesStatistic.ConversionPerCity.Description,
		strconv.FormatFloat(report.CitiesStatistic.ConversionPerCity.Msk, 'f', 2, 64) + "%",
		strconv.FormatFloat(report.CitiesStatistic.ConversionPerCity.Spb, 'f', 2, 64) + "%",
		strconv.FormatFloat(report.CitiesStatistic.ConversionPerCity.Kursk, 'f', 2, 64) + "%",
		strconv.FormatFloat(report.CitiesStatistic.ConversionPerCity.Total, 'f', 2, 64) + "%",
		"\n"}, ";")))
	encWriter.Write([]byte(";\n"))
	encWriter.Write([]byte(strings.Join([]string{
		report.SummaryStaffSalary.Description,
		strconv.FormatFloat(report.SummaryStaffSalary.SummarySalary, 'f', 2, 64) + " ₽;\n",
		"\n"}, ";")))
	encWriter.Write([]byte(strings.Join([]string{
		report.Bonus.Description,
		strconv.FormatFloat(report.Bonus.Bonus, 'f', 2, 64) + " ₽;\n",
		"\n"}, ";")))
	encWriter.Write([]byte(strings.Join([]string{
		report.Total.Description,
		strconv.FormatFloat(report.Total.Money, 'f', 2, 64) + " ₽;\n",
		"\n"}, ";")))

}

func (s *service) isDateBetween(dateFrom, dateTo, date time.Time) bool {
	return date == dateFrom || date.After(dateFrom) && date.Before(dateTo) || date == dateTo
}

func (s *service) getUniqCallsWithFilter(historyRecords []entity.HistoryRecord, filter func(record entity.HistoryRecord) bool) map[string]int {
	uniqCallsMap := make(map[string]struct{})

	result := make(map[string]int)

	for _, record := range historyRecords {
		if _, ok := uniqCallsMap[record.Abonent]; !ok {
			uniqCallsMap[record.Abonent] = struct{}{}
			if filter(record) {
				for city, r := range s.citiesAndLines {
					if r.MatchString(record.LineNumber) {
						result[city]++
						break
					}
				}
			}
		}
	}

	return result
}
