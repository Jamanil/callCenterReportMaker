package service

import (
	"callCenterReportMaker/entity"
	"fmt"
	"io"
	"log"
	"math"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

const debug = false

type Service interface {
	GetUniqTotalCallsCountPerCity(historyRecords []entity.HistoryRecord, dateFrom, dateTo time.Time) map[string]int
	GetUniqReceivedCallsCountPerCity(historyRecords []entity.HistoryRecord, dateFrom, dateTo time.Time) map[string]int
	GetUniqReceivedCallsByOperator(historyRecords []entity.HistoryRecord, dateFrom, dateTo time.Time) map[time.Time]map[string]int
	GetDatabaseStatistic(callsByOperators []entity.DatabaseStatistic, orders []entity.Orders) []entity.DatabaseStatistic
	GetWeeklyReport(callsByOperators []entity.DatabaseStatistic,
		orders []entity.Orders,
		callHistory []entity.HistoryRecord,
		dateFrom, dateTo time.Time,
		readWriter io.ReadWriter) entity.WeeklyReport
}

func New(citiesLineMap map[string]*regexp.Regexp, operatorsList []string, bonusMap map[float64]float64, orderCost, personalConversionGrade float64) Service {
	return &service{
		citiesAndLines:     citiesLineMap,
		operators:          operatorsList,
		motivationMap:      bonusMap,
		orderFee:           orderCost,
		minConversionGrade: personalConversionGrade,
	}
}

type service struct {
	citiesAndLines     map[string]*regexp.Regexp
	operators          []string
	motivationMap      map[float64]float64
	orderFee           float64
	minConversionGrade float64
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

func (s *service) GetWeeklyReport(callsByOperators []entity.DatabaseStatistic, orders []entity.Orders,
	callHistory []entity.HistoryRecord, dateFrom, dateTo time.Time, readWriter io.ReadWriter) entity.WeeklyReport {

	databaseStatistics := s.GetDatabaseStatistic(callsByOperators, orders)
	totalOrdersCount := databaseStatistics[len(databaseStatistics)-1].OrdersCount
	departmentBonus, personalBonusPerOrder := s.calculateBonus(databaseStatistics, readWriter)
	operatorReports := s.calculateOperatorsReport(databaseStatistics, departmentBonus, personalBonusPerOrder)
	departmentPayment := s.calculateDepartmentPayment(operatorReports)
	departmentPricePerOrder := s.calculateDepartmentPricePerOrder(totalOrdersCount, departmentPayment)
	var telephonyPayment, smsPayment float64
	if !debug {
		telephonyPayment = s.getFloat64FromIO(readWriter, "Сколько заплатили за телефонию?")
		smsPayment = s.getFloat64FromIO(readWriter, "Сколько заплатили за СМС?")
	} else {
		telephonyPayment = 15698.67
		smsPayment = 5000
	}

	totalExpenses := s.calculateTotalExpenses(departmentPayment, telephonyPayment, smsPayment)
	totalPricePerOrder := s.calculateTotalPricePerOrder(totalOrdersCount, totalExpenses)
	cityStatistics := s.calculateCityStatistics(orders, callHistory, dateFrom, dateTo)

	return entity.WeeklyReport{
		OperatorReports:         operatorReports,
		DepartmentPayment:       departmentPayment,
		DepartmentPricePerOrder: departmentPricePerOrder,
		TelephonyPayment:        telephonyPayment,
		SmsPayment:              smsPayment,
		TotalExpenses:           totalExpenses,
		TotalOrdersCount:        totalOrdersCount,
		TotalPricePerOrder:      totalPricePerOrder,
		CityStatistics:          cityStatistics,
		SummaryDepartmentSalary: departmentPayment - departmentBonus,
		SummaryDepartmentBonus:  departmentBonus,
		SumToPay:                departmentPayment,
		DateFrom:                dateFrom,
		DateTo:                  dateTo,
	}
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
func (s *service) calculateBonus(databaseStatistics []entity.DatabaseStatistic, readWriter io.ReadWriter) (totalBonus, personalBonusPerOrder float64) {
	totalDepartmentStatistics := databaseStatistics[len(databaseStatistics)-1]
	generalBonusPerOrder := s.calculateGeneralBonusPerOrder(totalDepartmentStatistics.Conversion)
	if generalBonusPerOrder >= 0 {
		totalBonus = math.RoundToEven(generalBonusPerOrder * float64(totalDepartmentStatistics.OrdersCount))
		if !debug {
			personalBonusPerOrder = s.setPersonalBonusPerOrder(databaseStatistics, totalDepartmentStatistics.Conversion, generalBonusPerOrder, totalBonus, readWriter)
		} else {
			personalBonusPerOrder = 19
		}

	}
	return totalBonus, personalBonusPerOrder
}
func (s *service) calculateGeneralBonusPerOrder(totalConversion float64) (generalBonusPerOrder float64) {
	motivationGrades := make([]float64, 0, len(s.motivationMap))
	for grade := range s.motivationMap {
		motivationGrades = append(motivationGrades, grade)
	}
	sort.Slice(motivationGrades, func(i, j int) bool { return motivationGrades[i] > motivationGrades[j] })

	for _, grade := range motivationGrades {
		if totalConversion*100 > grade {
			generalBonusPerOrder = s.motivationMap[grade]
			break
		}
	}
	return generalBonusPerOrder
}
func (s *service) setPersonalBonusPerOrder(databaseStatistics []entity.DatabaseStatistic,
	totalDepartmentConversion, generalBonusPerOrder, totalBonus float64, readWriter io.ReadWriter) (personalBonusPerOrder float64) {
	if generalBonusPerOrder <= 0 {
		_, _ = readWriter.Write([]byte(fmt.Sprintf("Мои соболезнования, на премию не заработали. Конверсия составила %g\n",
			totalDepartmentConversion*100)))
	} else {
		_, _ = readWriter.Write([]byte(fmt.Sprintf("Поздравляю, конверсии хватило на премию. Общая премия за заказ %g руб., суммарная премия %g руб. Сколько раздать на брата?\n",
			generalBonusPerOrder, totalBonus)))
		for {
			var err error
			var floatFromReader float64

			str := getSrtFromReader(readWriter)

			if floatFromReader, err = strconv.ParseFloat(str, 64); err == nil {
				personalBonusPerOrder = floatFromReader
			} else {
				_, _ = readWriter.Write([]byte(fmt.Sprintf("Размер персональной премии %g руб.\n", personalBonusPerOrder)))
				break
			}
			var summaryOperatorsBonus float64

			var answerString strings.Builder

			for i := 0; i < len(databaseStatistics)-2; i++ {
				operatorBonus := s.calculatePersonalBonus(databaseStatistics[i].Conversion, databaseStatistics[i].OrdersCount, personalBonusPerOrder)
				summaryOperatorsBonus += operatorBonus

				answerString.WriteString(fmt.Sprintln(databaseStatistics[i].Operator, operatorBonus))
			}
			answerString.WriteString(fmt.Sprintln("Виктор", totalBonus-summaryOperatorsBonus))
			answerString.WriteString(fmt.Sprintln("Годится или переиграть?"))
			_, _ = readWriter.Write([]byte(answerString.String()))
		}
	}
	return personalBonusPerOrder
}
func (s *service) calculatePersonalBonus(conversion float64, ordersCount int, personalBonusPerOrder float64) (personalBonus float64) {
	if conversion > s.minConversionGrade {
		personalBonus = float64(int(math.RoundToEven(personalBonusPerOrder*float64(ordersCount))) / 100 * 100)
	}
	return personalBonus
}
func (s *service) calculateOperatorsReport(databaseStatistics []entity.DatabaseStatistic,
	totalBonus, personalBonusPerOrder float64) []entity.OperatorReport {
	operatorsReport := make([]entity.OperatorReport, 0, len(databaseStatistics)-2)
	var summaryOperatorsBonus float64

	for i := 0; i < len(databaseStatistics)-2; i++ {
		currentOperatorSalary := float64(databaseStatistics[i].OrdersCount) * s.orderFee
		currentOperatorBonus := s.calculatePersonalBonus(databaseStatistics[i].Conversion, databaseStatistics[i].OrdersCount, personalBonusPerOrder)
		currentOperatorSummaryPay := currentOperatorSalary + currentOperatorBonus
		currentOperatorPricePerOrder := currentOperatorSummaryPay / float64(databaseStatistics[i].OrdersCount)
		currentOperatorUniqCalls := databaseStatistics[i].UniqIncomingCalls + databaseStatistics[i].UniqOutgoingCalls
		summaryOperatorsBonus += currentOperatorBonus
		operatorsReport = append(operatorsReport, entity.OperatorReport{
			Name:           databaseStatistics[i].Operator,
			Salary:         currentOperatorSalary,
			Bonus:          currentOperatorBonus,
			SummaryPayment: currentOperatorSummaryPay,
			OrdersCount:    databaseStatistics[i].OrdersCount,
			PricePerOrder:  currentOperatorPricePerOrder,
			UniqCalls:      currentOperatorUniqCalls,
			Conversion:     databaseStatistics[i].Conversion,
		})
	}
	bossSalary := float64(databaseStatistics[len(databaseStatistics)-1].OrdersCount) * 20
	bossBonus := totalBonus - summaryOperatorsBonus
	operatorsReport = append(operatorsReport, entity.OperatorReport{
		Name:           "Виктор",
		Salary:         bossSalary,
		Bonus:          bossBonus,
		SummaryPayment: bossSalary + bossBonus,
	})

	return operatorsReport
}
func (s *service) calculateDepartmentPayment(operatorsReport []entity.OperatorReport) (departmentPayment float64) {
	for _, report := range operatorsReport {
		departmentPayment += report.SummaryPayment
	}
	return departmentPayment
}
func (s *service) calculateDepartmentPricePerOrder(totalOrdersCount int, departmentPayment float64) (departmentPricePerOrder float64) {
	if totalOrdersCount > 0 {
		departmentPricePerOrder = departmentPayment / float64(totalOrdersCount)
	}
	return departmentPricePerOrder
}
func (s *service) getFloat64FromIO(readWriter io.ReadWriter, message string) float64 {
	_, _ = readWriter.Write([]byte(message))
	var floatFromConsole float64

	floatFromConsole, err := strconv.ParseFloat(getSrtFromReader(readWriter), 64)

	if err != nil {
		log.Println(err)
		floatFromConsole = s.getFloat64FromIO(readWriter, message)
	}
	return floatFromConsole
}
func (s *service) calculateTotalExpenses(departmentPayment, telephonyPayment, smsPayment float64) float64 {
	return departmentPayment + telephonyPayment + smsPayment
}
func (s *service) calculateTotalPricePerOrder(totalOrdersCount int, totalExpenses float64) (totalPricePerOrder float64) {
	if totalOrdersCount > 0 {
		totalPricePerOrder = totalExpenses / float64(totalOrdersCount)
	}
	return totalPricePerOrder
}
func (s *service) calculateCityStatistics(orders []entity.Orders, callHistory []entity.HistoryRecord,
	dateFrom, dateTo time.Time) []entity.CityStatistic {
	citiesCount := len(s.citiesAndLines)
	citiesNames := make([]string, 0, citiesCount)
	for cityName := range s.citiesAndLines {
		citiesNames = append(citiesNames, cityName)
	}

	cityStatistics := make([]entity.CityStatistic, 0, citiesCount)

	uniqTotalCallsCountPerCity := s.GetUniqTotalCallsCountPerCity(callHistory, dateFrom, dateTo)
	uniqReceivedCallsCountPerCity := s.GetUniqReceivedCallsCountPerCity(callHistory, dateFrom, dateTo)
	ordersPerCity := s.GetOrdersPerCity(orders)

	var uniqCallsTotalGeneral, uniqCallsReceivedGeneral, uniqCallsMissedGeneral, ordersCountGeneral int
	for _, city := range citiesNames {
		uniqCallsTotal := uniqTotalCallsCountPerCity[city]
		uniqCallsTotalGeneral += uniqCallsTotal
		uniqCallsReceived := uniqReceivedCallsCountPerCity[city]
		uniqCallsReceivedGeneral += uniqCallsReceived
		uniqCallsMissed := uniqCallsTotal - uniqCallsReceived
		uniqCallsMissedGeneral += uniqCallsMissed
		ordersCount := ordersPerCity[city]
		ordersCountGeneral += ordersCount
		conversion := s.calculateConversion(uniqCallsTotal, ordersCount)
		cityStatistics = append(cityStatistics, entity.CityStatistic{
			City:              city,
			UniqCallsTotal:    uniqCallsTotal,
			UniqCallsReceived: uniqCallsReceived,
			UniqCallsMissed:   uniqCallsMissed,
			OrdersCount:       ordersCount,
			Conversion:        conversion,
		})
	}

	sort.Slice(cityStatistics, func(i, j int) bool { return cityStatistics[i].UniqCallsTotal > cityStatistics[j].UniqCallsTotal })

	cityStatistics = append(cityStatistics, entity.CityStatistic{
		City:              "Итого",
		UniqCallsTotal:    uniqCallsTotalGeneral,
		UniqCallsReceived: uniqCallsReceivedGeneral,
		UniqCallsMissed:   uniqCallsMissedGeneral,
		OrdersCount:       ordersCountGeneral,
		Conversion:        s.calculateConversion(uniqCallsTotalGeneral, ordersCountGeneral),
	})

	return cityStatistics
}
func (s *service) calculateConversion(uniqCallsCount, orders int) (conversion float64) {
	if uniqCallsCount > 0 {
		conversion = float64(orders) / float64(uniqCallsCount)
	}
	return conversion
}

func getSrtFromReader(reader io.Reader) string {
	buf := make([]byte, 1024)
	var err error
	var n int
	for {
		n, err = reader.Read(buf)
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			log.Println(err)
			continue
		}
	}
	return string(buf[:n])
}
