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
)

var (
	citiesAndLines          = make(map[string]*regexp.Regexp)
	operators               = make([]string, 0, 7)
	motivationMap           = make(map[float64]float64)
	orderFee                float64
	personalConversionGrade float64
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
	ctrl := controller.New(srv, db)
	bot := tgBot.New(ctrl, telegramToken, telegramChatId, weekdayReportTime, weekendReportTime)

	go bot.StartBot()

	<-make(chan error)
}
