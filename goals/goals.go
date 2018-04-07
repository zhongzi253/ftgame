package goals

import (
	. "../utils"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

type GameGoals_t struct {
}

type InputCrown struct {
	Num          int
	OverInit     float64
	UnderInit    float64
	HandicapInit float64
	OverNow      float64
	UnderNow     float64
	HandicapNow  float64
	Update       string
}

func (input *InputCrown) ToString() (str string) {
	str = fmt.Sprintf("%v: %v", input.Num, input.OverNow, input.HandicapNow, input.UnderNow)
	return
}

type InputTiCai struct {
	Num   int
	Host  string
	Guest string
	Start time.Time
	Close time.Time
	Odds  [8]float64
}

func (input *InputTiCai) ToString() (str string) {
	str = fmt.Sprintf("%v: %s-%s, %v, %v",
		input.Num,
		input.Host,
		input.Guest,
		input.Close,
		input.Odds)
	return
}

type Decision_t struct {
	Type           int
	BetCrown       float64
	BenefitCrown   float64
	NumTiCai       int
	BetTicai       [8]float64
	BenefitTicai   [8]float64
	AllowanceTicai [8]float64
}

type SumInfor_t struct {
	count int
	max   float64
}

const (
	HANDICAP_1 = iota
	HANDICAP_2 // 1.25, 2.25
	HANDICAP_3 // 1.5, 2.5
	HANDICAP_4
)

var g_sum_infor SumInfor_t

func ClearSumInfor() {
	g_sum_infor.count = 0
	g_sum_infor.max = 0.0

}

func ParseHandicap(handicap string) float64 {
	array := strings.Split(handicap, "/")

	switch len(array) {
	case 1:
		hand, _ := strconv.ParseFloat(array[0], 64)
		return hand
	case 2:
		low, _ := strconv.ParseFloat(array[0], 64)
		high, _ := strconv.ParseFloat(array[1], 64)

		return (low + high) / 2
	default:
		log.Fatalln("Failed to parse handicap.")
		return 0
	}
}

func GetTiCaiNumber(handi float64) int {
	res := math.Floor(handi)
	return int(res)
}

func PrintDecision(dec Decision_t, input_ti_cai InputTiCai, input_crown InputCrown, is_over bool) {
	var string_game_type string
	var low int
	var high int
	var odds float64
	allow_ticai_normal := 0.0
	switch is_over {
	case true:
		string_game_type = "大球"
		allow_ticai_normal = dec.AllowanceTicai[0]
		low = 0
		high = dec.NumTiCai
		odds = input_crown.OverNow
	case false:
		string_game_type = "小球"
		allow_ticai_normal = dec.AllowanceTicai[7]
		low = dec.NumTiCai
		high = 7
		odds = input_crown.UnderNow
	}

	allow_ticai_normal -= ALLOWANCE_CROWN * dec.BetCrown
	total := (odds+1)*dec.BetCrown + ALLOWANCE_CROWN*odds*dec.BetCrown + allow_ticai_normal

	if total < BENCHMARCK {
		return
	}

	g_sum_infor.count++
	if total > g_sum_infor.max {
		g_sum_infor.max = total
	}

	WriteMailBody("%v\n", input_ti_cai.ToString())
	WriteMailBody("%v\n", input_crown.ToString())

	WriteMailBody("盘口(%3.2f)  赔率  投注资金  投注收益  投注返水  额外收入  投注奖金    总收益\n", input_crown.HandicapNow)
	WriteMailBody("%s:   %8.2f  %8.2f  %8.2f  %8.2f  %8.2f  %8.2f  %8.2f\n",
		string_game_type,
		odds,
		dec.BetCrown,
		(odds+1)*dec.BetCrown,
		ALLOWANCE_CROWN*odds*dec.BetCrown+allow_ticai_normal,
		0.0,
		total,
		total)

	bet_total := dec.BetCrown
	for i := low; i <= high; i++ {
		bet_total += dec.BetTicai[i]
		WriteMailBody("%4d:   %8.2f  %8.2f  %8.2f  %8.2f  %8.2f  %8.2f  %8.2f\n",
			i,
			input_ti_cai.Odds[i],
			dec.BetTicai[i],
			dec.BenefitTicai[i],
			dec.AllowanceTicai[i],
			BONUS_TICAI*input_ti_cai.Odds[i]*dec.BetTicai[i],
			dec.BenefitTicai[i]+dec.AllowanceTicai[i],
			dec.BenefitTicai[i]+dec.AllowanceTicai[i]+BONUS_TICAI*input_ti_cai.Odds[i]*dec.BetTicai[i])
	}

	WriteMailBody("投注总额:%8.2f, 竞彩总反水:%8.2f\n\n", bet_total, allow_ticai_normal)
}

func MakeDecisionOver4(input_ti_cai InputTiCai, input_crown InputCrown) Decision_t {
	var dec Decision_t
	dec.NumTiCai = GetTiCaiNumber(input_crown.HandicapNow) + 1

	odds_crown := (1+ALLOWANCE_CROWN)*input_crown.OverNow + 1 - ALLOWANCE_CROWN
	odds_crown2 := ((1+ALLOWANCE_CROWN)*input_crown.OverNow + 1) / 2

	i := 0
	x_total := 1.0
	for i = 0; i < dec.NumTiCai; i++ {
		x_total += odds_crown / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
	}
	x_total += odds_crown2 / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))

	dec.BetCrown = TOTAL_BET / x_total
	dec.BenefitCrown = dec.BetCrown * (input_crown.OverNow + 1)

	allow_ticai_total := 0.0
	for i = 0; i < dec.NumTiCai; i++ {
		dec.BetTicai[i] = dec.BetCrown * odds_crown / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
		dec.BenefitTicai[i] = dec.BetTicai[i] * (input_ti_cai.Odds[i])
		allow_ticai_total += ALLOWANCE_TICAI * dec.BetTicai[i]
	}
	dec.BetTicai[i] = dec.BetCrown * odds_crown2 / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
	dec.BenefitTicai[i] = dec.BetTicai[i]*(input_ti_cai.Odds[i]) + dec.BenefitCrown/2
	allow_ticai_total += ALLOWANCE_TICAI * dec.BetTicai[i]

	for i = 0; i < dec.NumTiCai; i++ {
		dec.AllowanceTicai[i] = allow_ticai_total + ALLOWANCE_CROWN*dec.BetCrown
	}
	dec.AllowanceTicai[i] = allow_ticai_total + ALLOWANCE_CROWN*input_crown.OverNow*dec.BetCrown/2

	return dec
}

func MakeDecisionUnder4(input_ti_cai InputTiCai, input_crown InputCrown) Decision_t {
	var dec Decision_t
	dec.NumTiCai = GetTiCaiNumber(input_crown.HandicapNow) + 1

	odds_crown := (1+ALLOWANCE_CROWN)*input_crown.UnderNow + 1 - ALLOWANCE_CROWN
	odds_crown2 := (1+ALLOWANCE_CROWN)*input_crown.UnderNow + (1-ALLOWANCE_CROWN)/2

	i := 0
	x_total := 1.0
	for i = 7; i > dec.NumTiCai; i-- {
		x_total += odds_crown / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
	}
	x_total += odds_crown2 / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))

	dec.BetCrown = TOTAL_BET / x_total
	dec.BenefitCrown = dec.BetCrown * (input_crown.UnderNow + 1)

	allow_ticai_total := 0.0
	for i = 7; i > dec.NumTiCai; i-- {
		dec.BetTicai[i] = dec.BetCrown * odds_crown / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
		dec.BenefitTicai[i] = dec.BetTicai[i] * (input_ti_cai.Odds[i])
		allow_ticai_total += ALLOWANCE_TICAI * dec.BetTicai[i]
	}
	dec.BetTicai[i] = dec.BetCrown * odds_crown2 / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
	dec.BenefitTicai[i] = dec.BetTicai[i]*(input_ti_cai.Odds[i]) + dec.BetCrown/2
	allow_ticai_total += ALLOWANCE_TICAI * dec.BetTicai[i]

	for i = 7; i > dec.NumTiCai; i-- {
		dec.AllowanceTicai[i] = allow_ticai_total + ALLOWANCE_CROWN*dec.BetCrown
	}
	dec.AllowanceTicai[i] = allow_ticai_total + ALLOWANCE_CROWN*dec.BetCrown/2
	return dec
}

func MakeDecisionUnder3(input_ti_cai InputTiCai, input_crown InputCrown) Decision_t {
	var dec Decision_t

	total := TOTAL_BET
	dec.NumTiCai = GetTiCaiNumber(input_crown.HandicapNow) + 1
	i := 0

	odds_crown := ((1+ALLOWANCE_CROWN)*input_crown.UnderNow + 1 - ALLOWANCE_CROWN)
	x_total := 1.0 / odds_crown
	for i = dec.NumTiCai; i <= 7; i++ {
		x_total += 1.0 / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
	}

	dec.BetCrown = total / (odds_crown * x_total)
	dec.BenefitCrown = dec.BetCrown * (input_crown.UnderNow + 1)

	allow_ticai_total := 0.0
	for i = dec.NumTiCai; i <= 7; i++ {
		dec.BetTicai[i] = total / ((input_ti_cai.Odds[i] * (1 + BONUS_TICAI)) * x_total)
		dec.BenefitTicai[i] = dec.BetTicai[i] * (input_ti_cai.Odds[i])
		allow_ticai_total += ALLOWANCE_TICAI * dec.BetTicai[i]
	}

	for i = dec.NumTiCai; i <= 7; i++ {
		dec.AllowanceTicai[i] = allow_ticai_total + ALLOWANCE_CROWN*dec.BetCrown
	}

	return dec
}

func MakeDecisionOver3(input_ti_cai InputTiCai, input_crown InputCrown) Decision_t {
	var dec Decision_t

	total := TOTAL_BET
	dec.NumTiCai = GetTiCaiNumber(input_crown.HandicapNow)
	i := 0

	odds_crown := ((1+ALLOWANCE_CROWN)*input_crown.OverNow + 1 - ALLOWANCE_CROWN)
	x_total := 1.0 / odds_crown
	for i = 0; i <= dec.NumTiCai; i++ {
		x_total += 1.0 / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
	}

	dec.BetCrown = total / (odds_crown * x_total)
	dec.BenefitCrown = dec.BetCrown * (input_crown.OverNow + 1)

	allow_ticai_total := 0.0
	for i = 0; i <= dec.NumTiCai; i++ {
		dec.BetTicai[i] = total / ((input_ti_cai.Odds[i] * (1 + BONUS_TICAI)) * x_total)
		dec.BenefitTicai[i] = dec.BetTicai[i] * (input_ti_cai.Odds[i])
		allow_ticai_total += ALLOWANCE_TICAI * dec.BetTicai[i]
	}

	for i = 0; i <= dec.NumTiCai; i++ {
		dec.AllowanceTicai[i] = allow_ticai_total + ALLOWANCE_CROWN*dec.BetCrown
	}

	return dec
}
func MakeDecisionOver2(input_ti_cai InputTiCai, input_crown InputCrown) Decision_t {
	var dec Decision_t
	dec.NumTiCai = GetTiCaiNumber(input_crown.HandicapNow)

	odds_crown := (1+ALLOWANCE_CROWN)*input_crown.OverNow + 1 - ALLOWANCE_CROWN
	odds_crown2 := (1+ALLOWANCE_CROWN)*input_crown.OverNow + (1-ALLOWANCE_CROWN)/2

	i := 0
	x_total := 1.0
	for i = 0; i < dec.NumTiCai; i++ {
		x_total += odds_crown / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
	}
	x_total += odds_crown2 / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))

	dec.BetCrown = TOTAL_BET / x_total
	dec.BenefitCrown = dec.BetCrown * (input_crown.OverNow + 1)

	allow_ticai_total := 0.0
	for i = 0; i < dec.NumTiCai; i++ {
		dec.BetTicai[i] = dec.BetCrown * odds_crown / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
		dec.BenefitTicai[i] = dec.BetTicai[i] * (input_ti_cai.Odds[i])
		allow_ticai_total += ALLOWANCE_TICAI * dec.BetTicai[i]
	}
	dec.BetTicai[i] = dec.BetCrown * odds_crown2 / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
	dec.BenefitTicai[i] = dec.BetTicai[i]*(input_ti_cai.Odds[i]) + dec.BetCrown/2
	allow_ticai_total += ALLOWANCE_TICAI * dec.BetTicai[i]

	for i = 0; i < dec.NumTiCai; i++ {
		dec.AllowanceTicai[i] = allow_ticai_total + ALLOWANCE_CROWN*dec.BetCrown
	}
	dec.AllowanceTicai[i] = allow_ticai_total + ALLOWANCE_CROWN*dec.BetCrown/2

	return dec
}

func MakeDecisionUnder2(input_ti_cai InputTiCai, input_crown InputCrown) Decision_t {
	var dec Decision_t
	dec.NumTiCai = GetTiCaiNumber(input_crown.HandicapNow)

	odds_crown := (1+ALLOWANCE_CROWN)*input_crown.UnderNow + 1 - ALLOWANCE_CROWN
	odds_crown2 := ((1+ALLOWANCE_CROWN)*input_crown.UnderNow + 1) / 2

	i := 0
	x_total := 1.0
	for i = 7; i > dec.NumTiCai; i-- {
		x_total += odds_crown / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
	}
	x_total += odds_crown2 / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))

	dec.BetCrown = TOTAL_BET / x_total
	dec.BenefitCrown = dec.BetCrown * (input_crown.UnderNow + 1)

	allow_ticai_total := 0.0
	for i = 7; i > dec.NumTiCai; i-- {
		dec.BetTicai[i] = dec.BetCrown * odds_crown / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
		dec.BenefitTicai[i] = dec.BetTicai[i] * (input_ti_cai.Odds[i])
		allow_ticai_total += ALLOWANCE_TICAI * dec.BetTicai[i]
	}
	dec.BetTicai[i] = dec.BetCrown * odds_crown2 / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
	dec.BenefitTicai[i] = dec.BetTicai[i]*(input_ti_cai.Odds[i]) + dec.BenefitCrown/2
	allow_ticai_total += ALLOWANCE_TICAI * dec.BetTicai[i]

	for i = 7; i > dec.NumTiCai; i-- {
		dec.AllowanceTicai[i] = allow_ticai_total + ALLOWANCE_CROWN*dec.BetCrown
	}
	dec.AllowanceTicai[i] = allow_ticai_total + ALLOWANCE_CROWN*input_crown.UnderNow*dec.BetCrown/2
	return dec
}

func MakeDecisionOver1(input_ti_cai InputTiCai, input_crown InputCrown) Decision_t {
	var dec Decision_t
	dec.NumTiCai = GetTiCaiNumber(input_crown.HandicapNow)

	odds_crown := (1+ALLOWANCE_CROWN)*input_crown.OverNow + 1 - ALLOWANCE_CROWN
	odds_crown2 := (1 + ALLOWANCE_CROWN) * input_crown.OverNow

	i := 0
	x_total := 1.0
	for i = 0; i < dec.NumTiCai; i++ {
		x_total += odds_crown / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
	}
	x_total += odds_crown2 / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))

	dec.BetCrown = TOTAL_BET / x_total
	dec.BenefitCrown = dec.BetCrown * (input_crown.OverNow + 1)

	allow_ticai_total := 0.0
	for i = 0; i < dec.NumTiCai; i++ {
		dec.BetTicai[i] = dec.BetCrown * odds_crown / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
		dec.BenefitTicai[i] = dec.BetTicai[i] * (input_ti_cai.Odds[i])
		allow_ticai_total += ALLOWANCE_TICAI * dec.BetTicai[i]
	}
	dec.BetTicai[i] = dec.BetCrown * odds_crown2 / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
	dec.BenefitTicai[i] = dec.BetTicai[i]*(input_ti_cai.Odds[i]) + dec.BetCrown
	allow_ticai_total += ALLOWANCE_TICAI * dec.BetTicai[i]

	for i = 0; i < dec.NumTiCai; i++ {
		dec.AllowanceTicai[i] = allow_ticai_total + ALLOWANCE_CROWN*dec.BetCrown
	}
	dec.AllowanceTicai[i] = allow_ticai_total
	return dec
}

func MakeDecisionUnder1(input_ti_cai InputTiCai, input_crown InputCrown) Decision_t {
	var dec Decision_t
	dec.NumTiCai = GetTiCaiNumber(input_crown.HandicapNow)

	odds_crown := (1+ALLOWANCE_CROWN)*input_crown.UnderNow + 1 - ALLOWANCE_CROWN
	odds_crown2 := (1 + ALLOWANCE_CROWN) * input_crown.UnderNow

	i := 0
	x_total := 1.0
	for i = 7; i > dec.NumTiCai; i-- {
		x_total += odds_crown / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
	}
	x_total += odds_crown2 / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))

	dec.BetCrown = TOTAL_BET / x_total
	dec.BenefitCrown = dec.BetCrown * (input_crown.UnderNow + 1)

	allow_ticai_total := 0.0
	for i = 7; i > dec.NumTiCai; i-- {
		dec.BetTicai[i] = dec.BetCrown * odds_crown / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
		dec.BenefitTicai[i] = dec.BetTicai[i] * (input_ti_cai.Odds[i])
		allow_ticai_total += ALLOWANCE_TICAI * dec.BetTicai[i]
	}
	dec.BetTicai[i] = dec.BetCrown * odds_crown2 / (input_ti_cai.Odds[i] * (1 + BONUS_TICAI))
	dec.BenefitTicai[i] = dec.BetTicai[i]*(input_ti_cai.Odds[i]) + dec.BetCrown
	allow_ticai_total += ALLOWANCE_TICAI * dec.BetTicai[i]

	for i = 7; i > dec.NumTiCai; i-- {
		dec.AllowanceTicai[i] = allow_ticai_total + ALLOWANCE_CROWN*dec.BetCrown
	}
	dec.AllowanceTicai[i] = allow_ticai_total
	return dec
}

func MakeDecisionOver(input_ti_cai InputTiCai, input_crown InputCrown) Decision_t {
	var dec Decision_t
	switch (input_crown.HandicapNow - math.Floor(input_crown.HandicapNow)) / 0.25 {
	case 0.0:
		dec = MakeDecisionOver1(input_ti_cai, input_crown)
	case 1.0:
		dec = MakeDecisionOver2(input_ti_cai, input_crown)
	case 2.0:
		dec = MakeDecisionOver3(input_ti_cai, input_crown)
	case 3.0:
		dec = MakeDecisionOver4(input_ti_cai, input_crown)
	}
	return dec
}

func MakeDecisionUnder(input_ti_cai InputTiCai, input_crown InputCrown) Decision_t {
	var dec Decision_t
	switch (input_crown.HandicapNow - math.Floor(input_crown.HandicapNow)) / 0.25 {
	case 0.0:
		dec = MakeDecisionUnder1(input_ti_cai, input_crown)
	case 1.0:
		dec = MakeDecisionUnder2(input_ti_cai, input_crown)
	case 2.0:
		dec = MakeDecisionUnder3(input_ti_cai, input_crown)
	case 3.0:
		dec = MakeDecisionUnder4(input_ti_cai, input_crown)
	}

	return dec
}

func FindDecision(input_cat InputTiCai, input_dog InputCrown) {
	dec := MakeDecisionOver(input_cat, input_dog)
	PrintDecision(dec, input_cat, input_dog, true)

	dec = MakeDecisionUnder(input_cat, input_dog)
	PrintDecision(dec, input_cat, input_dog, false)

}

func CampareCrownInfo(old, new interface{}) bool {
	if new.(InputCrown).Num < old.(InputCrown).Num {
		return true
	}
	return false
}

func FindCrownInfo(old, key interface{}) bool {
	return old.(InputCrown).Num == key.(int)
}

var g_crown_data *SortedLinkedList
var g_ticai_data *SortedLinkedList

func NewInputCrownInfo(n string, o1 string, o2 string, o3 string) (input_dog InputCrown) {
	input_dog.Num, _ = strconv.Atoi(n)
	input_dog.OverNow, _ = strconv.ParseFloat(o1, 64)
	input_dog.HandicapNow, _ = strconv.ParseFloat(o2, 64)
	input_dog.UnderNow, _ = strconv.ParseFloat(o3, 64)
	return
}

func FetchCrownData(url string) {
	url += fmt.Sprintf("%d000", time.Now().Unix())
	doc := FetchURL(url)
	g_crown_data = NewSortedLinkedList(1000, CampareCrownInfo, FindCrownInfo)
	doc.Find("odds i").Each(func(i int, s *goquery.Selection) {
		g := strings.Split(s.Text(), ",")
		if g[0] == "3" {
			g_crown_data.PutOnTop(NewInputCrownInfo(g[1], g[3], g[4], g[5]))
		}
	})
}

func HandleOneGame(s *goquery.Selection) {
	var input_ti_cai InputTiCai

	val, _ := s.Attr("id")
	spl := strings.Split(val, "_")
	input_ti_cai.Num, _ = strconv.Atoi(spl[1])

	s.Find("td").Each(func(i int, elem *goquery.Selection) {
		switch i {
		case 2:
			start_string, _ := elem.Attr("title")
			input_ti_cai.Start = ParseGameDate(start_string)
		case 3:
			close_string, _ := elem.Attr("title")
			input_ti_cai.Close = ParseGameDate(close_string)
		case 4:
			input_ti_cai.Host = elem.Find("a").Text()
		case 6:
			input_ti_cai.Guest = elem.Find("a").Text()
		case 9:
			input_ti_cai.Odds[0] = StringToFloat(elem)
		case 10:
			input_ti_cai.Odds[1] = StringToFloat(elem)
		case 11:
			input_ti_cai.Odds[2] = StringToFloat(elem)
		case 12:
			input_ti_cai.Odds[3] = StringToFloat(elem)
		case 13:
			input_ti_cai.Odds[4] = StringToFloat(elem)
		case 14:
			input_ti_cai.Odds[5] = StringToFloat(elem)
		case 15:
			input_ti_cai.Odds[6] = StringToFloat(elem)
		case 16:
			input_ti_cai.Odds[7] = StringToFloat(elem)
		}
	})

	if time.Now().After(input_ti_cai.Close) {
		return
	}

	input_dog := g_crown_data.FindElementWithKey(input_ti_cai.Num)
	if input_dog != nil {
		FindDecision(input_ti_cai, input_dog.Value.(InputCrown))
	} else {
		WriteMailBody("Cannot find crown data for %v\n%v\n", input_ti_cai.Num, input_ti_cai)
	}
}

func FetchTiCaiData(url string) bool {
	doc := FetchURL(url)

	// Find the urls
	elem := doc.Find(".td_div tbody tr")
	if elem.Length() == 0 {
		return false
	}

	elem.Each(func(i int, s *goquery.Selection) {
		val, exists := s.Attr("class")
		if exists && (val == "nii" || val == "nii2") {
			HandleOneGame(s)
		}
	})

	return true
}

func Case1() (InputTiCai, InputCrown) {
	WriteMailBody("\n\nCASE 1:\n")
	return InputTiCai{Host: "team1", Guest: "team2", Odds: [8]float64{1, 2, 5, 10, 20, 3, 4, 30}},
		InputCrown{OverNow: 0.9, HandicapNow: 2.5, UnderNow: 1.1}
}

func Case2() (InputTiCai, InputCrown) {
	WriteMailBody("\n\nCASE 2:\n")
	return InputTiCai{Host: "team1", Guest: "team2", Odds: [8]float64{1, 2, 5, 10, 20, 3, 4, 30}},
		InputCrown{OverNow: 0.9, HandicapNow: 2, UnderNow: 1.1}
}

func Case3() (InputTiCai, InputCrown) {
	WriteMailBody("\n\nCASE 3:\n")
	return InputTiCai{Host: "team1", Guest: "team2", Odds: [8]float64{1, 2, 5, 10, 20, 3, 4, 30}},
		InputCrown{OverNow: 0.9, HandicapNow: 2.25, UnderNow: 1.1}
}

func Case4() (InputTiCai, InputCrown) {
	WriteMailBody("\n\nCASE 4:\n")
	return InputTiCai{Host: "team1", Guest: "team2", Odds: [8]float64{1, 2, 5, 10, 20, 3, 4, 30}},
		InputCrown{OverNow: 0.9, HandicapNow: 2.75, UnderNow: 1.1}
}

func RunCase(input_cat InputTiCai, input_dog InputCrown) {
	FindDecision(input_cat, input_dog)
}

func NewGame() *GameGoals_t {
	return &GameGoals_t{}
}

func (game *GameGoals_t) RunOnce() {
	PrepareMail()
	ClearSumInfor()

	WriteMailBody("Find Match on %v\n", time.Unix(time.Now().Unix(), 0))
	FetchCrownData(URL_CROWN_OVERUNDER)
	FetchTiCaiData(URL_TICAI_OVERUNDER)
	WriteMailBody("Done on %v\n", time.Unix(time.Now().Unix(), 0))

	title := fmt.Sprintf("总进球： max=%.2f, count=%v", g_sum_infor.max, g_sum_infor.count)
	SendMail(title)
}

func (game *GameGoals_t) TestLoop() {
	RunCase(Case1())
	RunCase(Case2())
	RunCase(Case3())
	RunCase(Case4())

	SendMail("")
}

func (game *GameGoals_t) TryRun() {

}
