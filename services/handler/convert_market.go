package handler

import (
	"fmt"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/sportmarket-service/pkg/store"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
	"strings"
)

var NotConvertedError = errors.New("NotConverted")

const (
	OVER  = "over"
	UNDER = "under"
	Team1 = "h"
	Team2 = "a"
	Draw  = "d"
)

var teamSide = map[string]string{"1": Team1, "2": Team2, "Х": Draw, "X": Draw, "1Х": "h,d", "Х2": "a,d", "12": "h,a",
	"Чёт": "even", "Нечёт": "odd", "Б": OVER, "М": UNDER, "Не обе забьют": "no", "Обе забьют": "yes"}
var backLayMap = map[bool]string{false: "for", true: "against"}

var sportMap = map[string]string{
	"Футбол":       "fb",
	"Баскетбол":    "basket",
	"Теннис":       "tennis",
	"Rugby League": "rl",
	"Rugby Union":  "ru",
	"Регби":        "rl",
	"Хоккей":       "ih",
}

var moneyRe = regexp.MustCompile(`П(\d)$`)
var threeRe = regexp.MustCompile(`(^|\s)([12ХX]$)`)
var handicapRe = regexp.MustCompile(`Ф(\d)\((-?\d+,?\d{0,3})\)`)
var totalRe = regexp.MustCompile(`Т([МБ])\((-?\d+,?\d{0,3})\)`)
var teamTotalRe = regexp.MustCompile(`ИТ(\d)([МБ])\((-?\d+,?\d{0,3})\)`)
var DoubleChanceRe = regexp.MustCompile(`(1Х|Х2|12)$`)
var OddEvenRe = regexp.MustCompile(`(Чёт|Нечёт)$`)
var EuroHandRe = regexp.MustCompile(`(\d)\((\d+):(\d+)\)`)
var toScoreRe = regexp.MustCompile(`(Не обе забьют|Обе забьют)`)
var correctScoreRe = regexp.MustCompile(`\s?(\d):(\d)`)

func Convert(sb *pb.Surebet, u *store.UrlEvent) error {
	side := sb.Members[0]
	sport, ok := sportMap[sb.FortedSport]
	if !ok {
		return errors.Errorf("not_found_sport:%v", sb.FortedSport)
	}
	u.Sport = sport
	var found []string
	bookingIndex := strings.Index(side.MarketName, "ЖК")
	if bookingIndex != -1 {
		return NotConvertedError
	}

	cornIndex := strings.Index(side.MarketName, "УГЛ")
	if cornIndex != -1 {
		u.Sport += "_corn"
	}
	layIndex := strings.Index(side.MarketName, "Против")
	if layIndex != -1 {
		u.IsLay = true
	}
	u.PeriodNumber = processPeriodNumber(side.MarketName)
	switch u.PeriodNumber {
	case 0:
	case 1:
		u.Sport += "_ht"
	default:
		return errors.New("period>1")
	}

	found = moneyRe.FindStringSubmatch(side.MarketName)
	if len(found) == 2 {
		return processMoney(found, u)
	}
	found = threeRe.FindStringSubmatch(side.MarketName)
	if len(found) == 3 {
		return processThreeWay(found, u)
	}

	found = handicapRe.FindStringSubmatch(side.MarketName)
	if len(found) == 3 {
		return processHandicap(found, u)
	}

	found = totalRe.FindStringSubmatch(side.MarketName)
	if len(found) == 3 {
		return processTotal(found, u)
	}
	found = teamTotalRe.FindStringSubmatch(side.MarketName)
	if len(found) == 4 {
		return processTeamTotal(found, u)
	}
	found = DoubleChanceRe.FindStringSubmatch(side.MarketName)
	if len(found) == 2 {
		return processDoubleChance(found, u)
	}
	found = OddEvenRe.FindStringSubmatch(side.MarketName)
	if len(found) == 2 {
		return processOddEven(found, u)
	}
	found = EuroHandRe.FindStringSubmatch(side.MarketName)
	if len(found) == 4 {
		return processEuroHand(found, u)
	}
	found = toScoreRe.FindStringSubmatch(side.MarketName)
	if len(found) == 2 {
		return processToScore(found, u)
	}
	found = correctScoreRe.FindStringSubmatch(side.MarketName)
	if len(found) == 3 {
		return processCorrectScore(found, u)
	}
	return NotConvertedError
}
func processPoint(point string) (float64, error) {
	pointsStr := strings.Replace(point, ",", ".", -1)
	return strconv.ParseFloat(pointsStr, 64)
}

func processPeriodNumber(market string) int {
	var periodNumber = 0
	switch {
	case strings.Index(market, "1/2") != -1:
		periodNumber = 1
	case strings.Index(market, "1/3") != -1:
		periodNumber = 1
	case strings.Index(market, "2/3") != -1:
		periodNumber = 2
	case strings.Index(market, "3/3") != -1:
		periodNumber = 3
	case strings.Index(market, "2/2") != -1:
		periodNumber = 2
	case strings.Index(market, "1/4") != -1:
		periodNumber = 3
	case strings.Index(market, "3/2") != -1:
		periodNumber = 3
	case strings.Index(market, "2/4") != -1:
		periodNumber = 4
	case strings.Index(market, "4/2") != -1:
		periodNumber = 4
	case strings.Index(market, "3/4") != -1:
		periodNumber = 5
	case strings.Index(market, "5/2") != -1:
		periodNumber = 5
	case strings.Index(market, "4/4") != -1:
		periodNumber = 6
	case strings.Index(market, "6/2") != -1:
		periodNumber = 6
	case strings.Index(market, "7/2") != -1:
		periodNumber = 7
	case strings.Index(market, "8/2") != -1:
		periodNumber = 8
	}
	return periodNumber
}

func processDoubleChance(found []string, u *store.UrlEvent) error {
	u.BetType = fmt.Sprintf("%v,dc,%v", backLayMap[u.IsLay], teamSide[found[1]])
	return nil
}

func processMoney(found []string, u *store.UrlEvent) error {
	u.BetType = fmt.Sprintf("%v,ml,%v", backLayMap[u.IsLay], teamSide[found[1]])
	return nil
}
func processOddEven(found []string, u *store.UrlEvent) error {
	u.BetType = fmt.Sprintf("%v,%v", backLayMap[u.IsLay], teamSide[found[1]])
	return nil
}
func processThreeWay(found []string, u *store.UrlEvent) error {
	u.BetType = fmt.Sprintf("%v,%v", backLayMap[u.IsLay], teamSide[found[2]])
	return nil
}
func processHandicap(found []string, u *store.UrlEvent) error {
	point, err := processPoint(found[2])
	if err != nil {
		return err
	}
	var sign float64 = 1
	if found[1] == "2" {
		sign = -1
	}
	u.BetType = fmt.Sprintf("%v,ah,%v,%v", backLayMap[u.IsLay], teamSide[found[1]], point*4*sign)
	u.Handicap = &point
	return nil
}

func processEuroHand(found []string, u *store.UrlEvent) error {
	var sign float64 = 1
	if found[1] == "2" {
		sign = -1
	}

	hp, err := processPoint(found[2])
	if err != nil {
		return errors.New("convert_error")
	}
	ap, err := processPoint(found[3])
	if err != nil {
		return errors.New("convert_error")
	}
	var point float64
	if found[1] == "1" {
		point = hp - ap - 0.5
	} else if found[1] == "2" {
		point = ap - hp - 0.5
	}
	u.BetType = fmt.Sprintf("%v,ah,%v,%v", backLayMap[u.IsLay], teamSide[found[1]], point*4*sign)
	u.Handicap = &point

	return nil
}

func processTotal(found []string, u *store.UrlEvent) error {
	point, err := processPoint(found[2])
	if err != nil {
		return err
	}
	u.BetType = fmt.Sprintf("%v,ah%v,%v", backLayMap[u.IsLay], teamSide[found[1]], point*4)
	u.Handicap = &point
	return nil
}
func processTeamTotal(found []string, u *store.UrlEvent) error {
	point, err := processPoint(found[3])
	if err != nil {
		return err
	}
	u.BetType = fmt.Sprintf("%v,tah%v,%v,%v", backLayMap[u.IsLay], teamSide[found[2]], teamSide[found[1]], point*4)
	u.Handicap = &point
	return nil
}

func processToScore(found []string, u *store.UrlEvent) error {
	//for,score,both,no
	u.BetType = fmt.Sprintf("%v,score,both,%v", backLayMap[u.IsLay], teamSide[found[1]])
	return nil
}
func processCorrectScore(found []string, u *store.UrlEvent) error {
	//for,cs,1,1
	//against,cs,0,0
	u.BetType = fmt.Sprintf("%v,cs,%v,%v", backLayMap[u.IsLay], found[1], found[2])
	//fmt.Println(u.BetType)
	return nil
}
