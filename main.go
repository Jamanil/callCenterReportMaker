package main

import (
	"callCenterReportMaker/controller"
	"callCenterReportMaker/repository/database"
	"callCenterReportMaker/service"
	"callCenterReportMaker/tgBot"
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"log"
	"regexp"
	"strconv"
	"time"
)

var (
	citiesAndLines          = make(map[string]*regexp.Regexp)
	operators               = make([]string, 0, 7)
	motivationMap           = make(map[float64]float64)
	orderFee                float64
	personalConversionGrade float64
	serverHost              string
	serverPort              string
	dbHost                  string
	dbPort                  string
	dbName                  string
	dbUser                  string
	dbPassword              string
	telegramToken           string
	telegramChatId          int64
	weekdayReportTime       string
	weekendReportTime       string
)

func init() {
	viper.SetConfigFile("config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	orderFee = viper.GetFloat64("salary.orderFee")
	personalConversionGrade = viper.GetFloat64("salary.personalConversionGrade")
	operators = viper.GetStringSlice("salary.operators")
	serverHost = viper.GetString("server.host")
	serverPort = viper.GetString("server.port")
	dbHost = viper.GetString("database.host")
	dbPort = viper.GetString("database.port")
	dbName = viper.GetString("database.database")
	dbUser = viper.GetString("database.user")
	dbPassword = viper.GetString("database.password")
	telegramToken = viper.GetString("telegram.token")
	telegramChatId = viper.GetInt64("telegram.chatId")
	weekdayReportTime = viper.GetString("report.weekdayReportTime")
	weekendReportTime = viper.GetString("report.weekendReportTime")

	for city, rExp := range viper.GetStringMapString("citiesAndLinesRegexpMap") {
		citiesAndLines[cases.Title(language.Russian).String(city)] = regexp.MustCompile(rExp)
	}

	for gradeStr, bonusStr := range viper.GetStringMapString("salary.motivationMap") {
		grade, err := strconv.ParseFloat(gradeStr, 64)
		if err != nil {
			log.Fatal(err)
		}
		bonus, err := strconv.ParseFloat(bonusStr, 64)
		if err != nil {
			log.Fatal(err)
		}
		motivationMap[grade] = bonus
	}
}

func main() {

	srv := service.New(citiesAndLines, operators, motivationMap, orderFee, personalConversionGrade)
	db := database.New(dbHost, dbPort, dbName, dbUser, dbPassword, operators)
	bot := tgBot.New(telegramToken, telegramChatId)
	ctrl := controller.New(srv, db, bot, weekdayReportTime, weekendReportTime)

	go ctrl.StartDailyReportSending()

	<-make(chan error)

	report, err := ctrl.MakeReport(time.Date(2023, 6, 12, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 6, 18, 0, 0, 0, 0, time.UTC))
	if err != nil {
		log.Println(err)
	}

	err = report.SaveAsXlsx("data/result.xlsx")
	if err != nil {
		log.Println(err)
	}

	//report := entity.WeeklyReport{
	//	OperatorReports: []entity.OperatorReport{
	//		{Name: "Алексей Поминов",
	//			Salary:         6720,
	//			Bonus:          1600,
	//			SummaryPayment: 8320,
	//			OrdersCount:    84,
	//			PricePerOrder:  99.05,
	//			UniqCalls:      244,
	//			Conversion:     0.3443},
	//		{Name: "Алена Миронова",
	//			Salary:         3440,
	//			Bonus:          800,
	//			SummaryPayment: 4240,
	//			OrdersCount:    43,
	//			PricePerOrder:  98.60,
	//			UniqCalls:      95,
	//			Conversion:     0.4526},
	//		{Name: "Игорь Манойло",
	//			Salary:         2960,
	//			Bonus:          0,
	//			SummaryPayment: 2960,
	//			OrdersCount:    37,
	//			PricePerOrder:  80.00,
	//			UniqCalls:      139,
	//			Conversion:     0.2662},
	//		{Name: "Кирилл Соколовский",
	//			Salary:         5520,
	//			Bonus:          1300,
	//			SummaryPayment: 6820,
	//			OrdersCount:    69,
	//			PricePerOrder:  98.84,
	//			UniqCalls:      219,
	//			Conversion:     0.3151},
	//		{Name: "Маргарита Долгушина",
	//			Salary:         4640,
	//			Bonus:          1100,
	//			SummaryPayment: 5740,
	//			OrdersCount:    58,
	//			PricePerOrder:  98.97,
	//			UniqCalls:      182,
	//			Conversion:     0.3187},
	//		{Name: "Суханов Антон",
	//			Salary:         3280,
	//			Bonus:          1500,
	//			SummaryPayment: 4780,
	//			OrdersCount:    41,
	//			PricePerOrder:  116.59,
	//			UniqCalls:      141,
	//			Conversion:     0.2908},
	//		{Name: "Татьяна Старцева",
	//			Salary:         880,
	//			Bonus:          0,
	//			SummaryPayment: 880,
	//			OrdersCount:    11,
	//			PricePerOrder:  80,
	//			UniqCalls:      25,
	//			Conversion:     0.4400},
	//		{Name: "Виктор",
	//			Salary:         7480,
	//			Bonus:          1180,
	//			SummaryPayment: 8660},
	//	},
	//	DepartmentPayment:       42400,
	//	DepartmentPricePerOrder: 113.37,
	//	TelephonyPayment:        15697.68,
	//	SmsPayment:              5000,
	//	TotalExpenses:           63097.68,
	//	TotalOrdersCount:        374,
	//	TotalPricePerOrder:      168.71,
	//	CityStatistics: []entity.CityStatistic{
	//		{City: "Москва",
	//			UniqCallsTotal:    790,
	//			UniqCallsReceived: 771,
	//			UniqCallsMissed:   19,
	//			OrdersCount:       264,
	//			Conversion:        0.3342},
	//		{City: "Спб",
	//			UniqCallsTotal:    288,
	//			UniqCallsReceived: 281,
	//			UniqCallsMissed:   7,
	//			OrdersCount:       110,
	//			Conversion:        0.3819},
	//		{City: "Курск",
	//			UniqCallsTotal:    20,
	//			UniqCallsReceived: 18,
	//			UniqCallsMissed:   2,
	//			OrdersCount:       0,
	//			Conversion:        0},
	//		{City: "Итого",
	//			UniqCallsTotal:    1098,
	//			UniqCallsReceived: 1070,
	//			UniqCallsMissed:   28,
	//			OrdersCount:       374,
	//			Conversion:        0.3406},
	//	},
	//	SummaryDepartmentSalary: 34920,
	//	SummaryDepartmentBonus:  7480,
	//	SumToPay:                66215.50,
	//}

	//err = report.SaveAsXlsx("data/result.xlsx")
	//if err != nil {
	//	log.Fatal(err)
	//}

}
