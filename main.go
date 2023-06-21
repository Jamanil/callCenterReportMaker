package main

import (
	"callCenterReportMaker/repository/csvReader"
	"callCenterReportMaker/repository/database"
	"callCenterReportMaker/service"
	"fmt"
	"log"
	"regexp"
	"time"
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
	//viper.SetConfigName("config")
	//viper.SetConfigType("yaml")
	//viper.AddConfigPath(".")
	//
	//err := viper.ReadInConfig()
	//if err != nil {
	//	log.Fatalf("Error reading config file: %s", err)
	//}
	//
	//fmt.Println(viper.GetString("database.password"))

	//goland:noinspection SpellCheckingInspection
	dbReader := database.New("37.140.195.228", "3306", "sudak", "viktor", "6A2b8N2t", operators)
	//callHistory, err := dbReader.GetHistory(34)
	//if err != nil {
	//	log.Fatal(err)
	//}

	csvR := csvReader.New()
	callHistory, err := csvR.GetHistory("data/20_06_2023_21_50.csv")
	if err != nil {
		log.Fatal(err)
	}
	//
	//fmt.Println(callHistory[0])
	//fmt.Println(callHistory[len(callHistory)-1])
	//
	startDate := time.Date(2023, 06, 12, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 06, 18, 0, 0, 0, 0, time.UTC)
	//
	//service := service2.New()
	//callsPerCity := service.GetUniqTotalCallsCountPerCity(callHistory, startDate, endDate)
	//fmt.Println("Total calls", callsPerCity)
	//
	//callsPerCityReceived := service.GetUniqReceivedCallsCountPerCity(callHistory, startDate, endDate)
	//fmt.Println("Received calls", callsPerCityReceived)
	//
	//callsPerOperByDay := service.GetUniqReceivedCallsByOperator(callHistory, startDate, endDate)
	//
	//dates := make([]time.Time, 0)
	//for t, _ := range callsPerOperByDay {
	//	dates = append(dates, t)
	//}
	//
	//sort.Slice(dates, func(i, j int) bool {
	//	return dates[i].Before(dates[j])
	//})
	//
	//for _, date := range dates {
	//	fmt.Println(date, callsPerOperByDay[date])
	//}

	orders, err := dbReader.GetOrders(startDate, endDate)
	if err != nil {
		log.Fatal(err)
	}

	uniqCallsCount, err := dbReader.GetUniqCallsByOperators(startDate, endDate)
	if err != nil {
		log.Fatal(err)
	}

	srv := service.New(citiesAndLines, operators, motivationMap, orderFee)

	stat := srv.GetDatabaseStatistic(uniqCallsCount, orders)
	fmt.Printf("%-25s %-7s %-7s %-7s %4s\n", "Имя оператора", "заказы", "ун.вх.", "ун.исх.", "конв")
	//for _, statistic := range stat {
	//	fmt.Println(statistic)
	//}

	srv.GetSalaries(stat, orders, callHistory, startDate, endDate)

}
