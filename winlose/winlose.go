package winlose

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	//"log"
	//"math"
	//"net/http"
	"container/list"
	"strconv"
	"strings"
	"time"

	. "../utils"
)

var g_crown_data *SortedLinkedList
var g_ticai_data *SortedLinkedList

type SumInfor_t struct {
	count int
	max   float64
}

type GameWinLose_t struct {
}

type InputTiCaiWinLose_t struct {
	Num      int
	Host     string
	Guest    string
	Start    time.Time
	Close    time.Time
	Odds1    [3]float64
	Odds2    [3]float64
	Handicap int
}

func (input *InputTiCaiWinLose_t) ToString() (str string) {
	str = fmt.Sprintf("%v: %s-%s, %v, %v %v %v",
		input.Num,
		input.Host,
		input.Guest,
		input.Close,
		input.Odds1,
		input.Odds2,
		input.Handicap)
	return
}

type InputCrownWinLose_t struct {
	Num  int
	Odds [3]float64
}

func (input *InputCrownWinLose_t) ToString() (str string) {
	str = fmt.Sprintf("%v: %v", input.Num, input.Odds)
	return
}

func NewInputCrownInfo(n string, o1 string, o2 string, o3 string) InputCrownWinLose_t {
	nn, _ := strconv.Atoi(n)
	oo1, _ := strconv.ParseFloat(o1, 64)
	oo2, _ := strconv.ParseFloat(o2, 64)
	oo3, _ := strconv.ParseFloat(o3, 64)
	return InputCrownWinLose_t{nn, [3]float64{oo1, oo2, oo3}}
}

type DecisionWinLose_t struct {
	Bet1       [3]float64
	Bet2       [3]float64
	BetCrown   [3]float64
	Delta      float64
	Benefit    float64
	InputCrown InputCrownWinLose_t
	InputTiCai InputTiCaiWinLose_t
}

var g_sum_infor SumInfor_t

func ClearSumInfor() {
	g_sum_infor.count = 0
	g_sum_infor.max = 0.0

}

func BetTypeString(t int, index int) string {
	var bet_type string

	switch t {
	case -1:
		bet_type = "主"
	case 3:
		bet_type = "内"
	case 1:
		bet_type = "客"
	case 2:
		bet_type = "外"
	}

	switch index {
	case 0:
		bet_type += "胜"
	case 1:
		bet_type += "平"
	case 2:
		bet_type += "负"
	}
	return bet_type
}

func PrintDecisionWinLose(dec DecisionWinLose_t) {
	input_ticai := dec.InputTiCai
	input_crown := dec.InputCrown

	bet_ticai := 0.0
	bet_crown := 0.0

	for i := 0; i < 3; i++ {
		/*		dec.Bet1[i] *= dec.Delta
				dec.Bet2[i] *= dec.Delta
				dec.BetCrown[i] *= dec.Delta
		*/
		bet_ticai += dec.Bet1[i] + dec.Bet2[i]
		bet_crown += dec.BetCrown[i]
	}
	/*
		if bet_ticai == 0.0 {
			for i := 0; i < 3; i++ {
				dec.BetCrown[i] /= dec.Delta
			}
			bet_crown /= dec.Delta
		}
	*/
	allow_total := bet_ticai*ALLOWANCE_TICAI + bet_crown*ALLOWANCE_CROWN

	bet_total := bet_ticai + bet_crown
	benefit := 0.0

	MailBufferClean()
	MailBufferWrite("%v\n", dec.InputTiCai.ToString())
	MailBufferWrite("%v\n", dec.InputCrown.ToString())
	for i := 0; i < 3; i++ {
		if 0 != dec.Bet1[i] {
			if benefit == 0 {
				benefit = input_ticai.Odds1[i]*dec.Bet1[i]*(1+BONUS_TICAI) + allow_total
				if benefit < BENCHMARCK {
					return
				}
				MailBufferWrite("注种       赔率  投注资金  投注收益  投注返水  额外收入  投注奖金    总收益\n")
			}
			MailBufferWrite("%s   %8.2f  %8.2f  %8.2f  %8.2f  %8.2f  %8.2f  %8.2f\n",
				BetTypeString(3, i),
				input_ticai.Odds1[i],
				dec.Bet1[i],
				input_ticai.Odds1[i]*dec.Bet1[i],
				allow_total,
				input_ticai.Odds1[i]*dec.Bet1[i]*BONUS_TICAI,
				input_ticai.Odds1[i]*dec.Bet1[i]+allow_total,
				input_ticai.Odds1[i]*dec.Bet1[i]*(1+BONUS_TICAI)+allow_total)
		}
		if 0 != dec.Bet2[i] {
			if benefit == 0 {
				benefit = input_ticai.Odds2[i]*dec.Bet2[i]*(1+BONUS_TICAI) + allow_total
				if benefit < BENCHMARCK {
					return
				}
				MailBufferWrite("注种       赔率  投注资金  投注收益  投注返水  额外收入  投注奖金    总收益\n")
			}
			MailBufferWrite("%s   %8.2f  %8.2f  %8.2f  %8.2f  %8.2f  %8.2f  %8.2f\n",
				BetTypeString(input_ticai.Handicap, i),
				input_ticai.Odds2[i],
				dec.Bet2[i],
				input_ticai.Odds2[i]*dec.Bet2[i],
				allow_total,
				input_ticai.Odds2[i]*dec.Bet2[i]*BONUS_TICAI,
				input_ticai.Odds2[i]*dec.Bet2[i]+allow_total,
				input_ticai.Odds2[i]*dec.Bet2[i]*(1+BONUS_TICAI)+allow_total)
		}
		if 0 != dec.BetCrown[i] {
			bet_benefit := input_crown.Odds[i] * dec.BetCrown[i]
			allow_crown := allow_total - ALLOWANCE_CROWN*dec.BetCrown[i] + ALLOWANCE_CROWN*dec.BetCrown[i]*(input_crown.Odds[i]-1)
			benefit_crown := bet_benefit + allow_crown
			if benefit == 0 {
				benefit = benefit_crown
				if benefit < BENCHMARCK {
					return
				}
				MailBufferWrite("注种       赔率  投注资金  投注收益  投注返水  额外收入  投注奖金    总收益\n")
			}
			MailBufferWrite("%s   %8.2f  %8.2f  %8.2f  %8.2f  %8.2f  %8.2f  %8.2f\n",
				BetTypeString(2, i),
				input_crown.Odds[i],
				dec.BetCrown[i],
				benefit_crown,
				allow_crown,
				0.0,
				benefit_crown,
				benefit_crown)
		}
	}

	MailBufferWrite("总投资额: %8.2f, 利润：%8.2f， 利润率: %8.4f, Delta:%8.2f\n\n", bet_total, benefit-bet_total, benefit/bet_total-1, dec.Delta)
	/*
		if benefit/bet_total-1 <= 0 {
			return
		}
	*/
	MailBufferDump()
}

func MakeDecisionK1(index int, input_cat InputTiCaiWinLose_t, input_dog InputCrownWinLose_t) DecisionWinLose_t {
	var dec DecisionWinLose_t
	odds := 0.0
	//WriteMailBody("Make K1\n")
	/*
		b.	主让， 用<1 代替 =0和<0,记为K1
		* if Q2 > P1*P2/(P1 + P2)
		0.  T		K1
		1. 	C		K1
	*/
	x_total := 1 / (input_cat.Odds2[2] + input_cat.Odds2[2]*BONUS_TICAI)

	ticai_total := 0.0
	switch index {
	case 1:
		odds = input_dog.Odds[0]
		x_total += 1 / (odds + ALLOWANCE_CROWN*odds - 2*ALLOWANCE_CROWN)

		dec.BetCrown[0] = TOTAL_BET / x_total / (odds + ALLOWANCE_CROWN*odds - 2*ALLOWANCE_CROWN)
	case 0:
		odds = input_cat.Odds1[0]
		x_total += 1 / (odds + odds*BONUS_TICAI)
		dec.Bet1[0] = TOTAL_BET / x_total / (odds + odds*BONUS_TICAI)
		ticai_total += dec.Bet1[0]
	default:
		fmt.Println("!!!!FATAL ERROR ", index)
	}

	dec.Bet2[2] = TOTAL_BET / x_total / (input_cat.Odds2[2] + input_cat.Odds2[2]*BONUS_TICAI)
	ticai_total += dec.Bet2[2]
	crown_total := dec.BetCrown[0]
	dec.Delta = x_total * (1 + BONUS_TICAI)

	MailBufferWrite("Benefit = %v\n", TOTAL_BET/dec.Delta+ALLOWANCE_TICAI*ticai_total+ALLOWANCE_CROWN*crown_total)
	return dec
}
func MakeDecisionK2(index int, input_cat InputTiCaiWinLose_t, input_dog InputCrownWinLose_t) DecisionWinLose_t {
	var dec DecisionWinLose_t
	var odds [3]float64
	//	WriteMailBody("Make K2\n")
	/*
		c. 主让, 用=1，>1组合代替>0,记为K2
		* if  Q0*Q1/(Q0 + Q1) > P0
		0.  K2		T		T
		2.  K2		C		T
		4.  K2		T		C
		6.  K2		C		C
	*/

	x_total := 1/(input_cat.Odds2[0]+input_cat.Odds2[0]*BONUS_TICAI) + 1/(input_cat.Odds2[1]+input_cat.Odds2[1]*BONUS_TICAI)

	for i := uint(1); i < 3; i++ {
		if index&(1<<i) != 0 {
			odds[i] = input_dog.Odds[i]
			x_total += 1 / (odds[i] + ALLOWANCE_CROWN*odds[i] - 2*ALLOWANCE_CROWN)
		} else {
			odds[i] = input_cat.Odds1[i]
			x_total += 1 / (odds[i] + odds[i]*BONUS_TICAI)
		}
	}

	dec.Bet2[0] = TOTAL_BET / (input_cat.Odds2[0] + input_cat.Odds2[0]*BONUS_TICAI) / x_total
	dec.Bet2[1] = TOTAL_BET / (input_cat.Odds2[1] + input_cat.Odds2[1]*BONUS_TICAI) / x_total

	for i := uint(1); i < 3; i++ {
		if index&(1<<i) != 0 {
			dec.BetCrown[i] = TOTAL_BET / (odds[i] + ALLOWANCE_CROWN*odds[i] - 2*ALLOWANCE_CROWN) / x_total
		} else {
			dec.Bet1[i] = TOTAL_BET / (odds[i] + BONUS_TICAI*odds[i]) / x_total
		}
	}

	dec.Delta = (1 + BONUS_TICAI) * x_total

	return dec
}

func MakeDecisionK1K2(index int, input_cat InputTiCaiWinLose_t, input_dog InputCrownWinLose_t) DecisionWinLose_t {
	var dec DecisionWinLose_t
	//	WriteMailBody("Make K1K2\n")

	x_total := 0.0
	for i := 0; i < 3; i++ {
		x_total += 1 / (input_cat.Odds2[i] + input_cat.Odds2[i]*BONUS_TICAI)
	}

	for i := 0; i < 3; i++ {
		dec.Bet2[i] = TOTAL_BET / (input_cat.Odds2[i] + input_cat.Odds2[i]*BONUS_TICAI) / x_total
	}
	dec.Delta = (1 + BONUS_TICAI) * x_total

	return dec
}

func MakeDecisionG1(index int, input_cat InputTiCaiWinLose_t, input_dog InputCrownWinLose_t) DecisionWinLose_t {
	var dec DecisionWinLose_t
	odds := 0.0
	//WriteMailBody("Make G1\n")
	/*
		d.	客让G1， 用 >-1 代替 >0和=0
		* if Q0 > P0*P1/(P0 + P1)
		0.	G1				T
		4.  G1				C
	*/

	x_total := 1 / (input_cat.Odds2[0] + input_cat.Odds2[0]*BONUS_TICAI)

	switch index {
	case 4:
		odds = input_dog.Odds[2]
		x_total += 1 / (odds + ALLOWANCE_CROWN*odds - 2*ALLOWANCE_CROWN)

		dec.BetCrown[2] = TOTAL_BET / x_total / (odds + ALLOWANCE_CROWN*odds - 2*ALLOWANCE_CROWN)
	case 0:
		odds = input_cat.Odds1[2]
		x_total += 1 / (odds + odds*BONUS_TICAI)
		dec.Bet1[2] = TOTAL_BET / x_total / (odds + odds*BONUS_TICAI)
	default:
		fmt.Println("!!!!FATAL ERROR ", index)
	}

	dec.Bet2[0] = TOTAL_BET / x_total / (input_cat.Odds2[0] + input_cat.Odds2[0]*BONUS_TICAI)
	dec.Delta = x_total * (1 + BONUS_TICAI)
	return dec
}
func MakeDecisionG2(index int, input_cat InputTiCaiWinLose_t, input_dog InputCrownWinLose_t) DecisionWinLose_t {
	var dec DecisionWinLose_t
	var odds [3]float64
	/*
		e.	客让G2， 用=-1和<-1代替<0
		if  Q1*Q2/(Q1 + Q2) > P1
		0.  T		T		G2
		1.  C		T		G2
		2.	T		C		G2
		3.  C		C		G2
	*/

	//	WriteMailBody("Make G2\n")

	x_total := 1/(input_cat.Odds2[1]+input_cat.Odds2[1]*BONUS_TICAI) + 1/(input_cat.Odds2[2]+input_cat.Odds2[2]*BONUS_TICAI)

	for i := uint(0); i < 2; i++ {
		if index&(1<<i) != 0 {
			odds[i] = input_dog.Odds[i]
			x_total += 1 / (odds[i] + ALLOWANCE_CROWN*odds[i] - 2*ALLOWANCE_CROWN)
		} else {
			odds[i] = input_cat.Odds1[i]
			x_total += 1 / (odds[i] + odds[i]*BONUS_TICAI)
		}
	}

	dec.Bet2[1] = TOTAL_BET / (input_cat.Odds2[1] + input_cat.Odds2[1]*BONUS_TICAI) / x_total
	dec.Bet2[2] = TOTAL_BET / (input_cat.Odds2[2] + input_cat.Odds2[2]*BONUS_TICAI) / x_total

	for i := uint(0); i < 2; i++ {
		if index&(1<<i) != 0 {
			dec.BetCrown[i] = TOTAL_BET / (odds[i] + ALLOWANCE_CROWN*odds[i] - 2*ALLOWANCE_CROWN) / x_total
		} else {
			dec.Bet1[i] = TOTAL_BET / (odds[i] + BONUS_TICAI*odds[i]) / x_total
		}
	}

	dec.Delta = (1 + BONUS_TICAI) * x_total

	return dec
}

func MakeDecisionG1G2(index int, input_cat InputTiCaiWinLose_t, input_dog InputCrownWinLose_t) DecisionWinLose_t {
	//	WriteMailBody("Make G1G2\n")
	return MakeDecisionK1K2(index, input_cat, input_dog)
}

func MakeDecisionNormal(index int, input_cat InputTiCaiWinLose_t, input_dog InputCrownWinLose_t) DecisionWinLose_t {
	var odds [3]float64
	//is_crown := [3]bool{false, false, false}

	//var total_ticai float64
	//var total_crow float64
	var dec DecisionWinLose_t
	x_total := 0.0
	for i := uint(0); i < 3; i++ {
		if index&(1<<i) != 0 {
			odds[i] = input_dog.Odds[i]
			x_total += 1 / (odds[i] + ALLOWANCE_CROWN*odds[i] - 2*ALLOWANCE_CROWN)
		} else {
			odds[i] = input_cat.Odds1[i]
			x_total += 1 / (odds[i] + odds[i]*BONUS_TICAI)
		}
	}

	for i := uint(0); i < 3; i++ {
		if index&(1<<i) != 0 {
			dec.BetCrown[i] = TOTAL_BET / (odds[i] + ALLOWANCE_CROWN*odds[i] - 2*ALLOWANCE_CROWN) / x_total
		} else {
			dec.Bet1[i] = TOTAL_BET / (odds[i] + BONUS_TICAI*odds[i]) / x_total
		}
	}

	dec.Delta = x_total * (1 + BONUS_TICAI)

	return dec
}

func MakeDecision(input_cat InputTiCaiWinLose_t, input_dog InputCrownWinLose_t) DecisionWinLose_t {
	dec := DecisionWinLose_t{Delta: 100}

	//TODO: 只考虑 K1， G1
	switch input_cat.Handicap {
	case -1:
		dec = MakeDecisionK1(1, input_cat, input_dog)
	case 1:
		dec = MakeDecisionG1(4, input_cat, input_dog)

	}
	/*
		isK1 := ToApplyK1(input_cat, input_dog)
		isK2 := ToApplyK2(input_cat, input_dog)
		isG1 := ToApplyG1(input_cat, input_dog)
		isG2 := ToApplyG2(input_cat, input_dog)

		GetDecisionMethod := func(i int) (method func(index int, input_cat InputTiCaiWinLose_t, input_dog InputCrownWinLose_t) DecisionWinLose_t) {

			method = MakeDecisionNormal

			if i%2 == 0 && isK2 {
				method = MakeDecisionK2
			}

			if i < 2 && isK1 {
				method = MakeDecisionK1
			}

			if i == 0 && isK1 && isK2 {
				method = MakeDecisionK1K2
			}

			if i < 4 && isG2 {
				method = MakeDecisionG2
			}

			if i%4 == 0 && isG1 {
				method = MakeDecisionG1
			}

			if i == 0 && isG1 && isG2 {
				method = MakeDecisionG1G2
			}

			return
		}
		new_dec := DecisionWinLose_t{}
		for i := 0; i < 8; i++ {
			m1 := GetDecisionMethod(i)
			new_dec = m1(i, input_cat, input_dog)
			new_dec.InputCrown = input_dog
			new_dec.InputTiCai = input_cat
			PrintDecisionWinLose(new_dec)
			/*				count1 := 0
							count2 := 0

							//TODO: 只考虑两注对冲
							for j := 0; j < 3; j++ {
								if new_dec.Bet1[j] > 0 || new_dec.Bet2[j] > 0 {
									count1++
								}

								if new_dec.BetCrown[j] > 0 {
									count2++
								}

							}

							if count1 > 1 || count2 > 1 {
								continue
							}
			if dec.Delta > new_dec.Delta {
				dec = new_dec
			}
		}
	*/

	dec.InputCrown = input_dog
	dec.InputTiCai = input_cat
	/*
		对冲情况

		a. 无让球， C投皇冠， T投体彩
			>0,		=0,		<0
		0.  T		T		T
		1.  C		T		T
		2.	T		C		T
		3.  C		C		T
		4.  T		T		C
		5.  C		T		C
		6.  T		C		C
		7.  C		C		C

		b.	主让， 用<1 代替 =0和<0,记为K1
		* if Q2 > P1*P2/(P1 + P2)
		0.  T		K1
		1. 	C		K1

		c. 主让, 用=1，>1组合代替>0,记为K2
		* if  Q0*Q1/(Q0 + Q1) > P0
		0.  K2		T		T
		2.  K2		C		T
		4.  K2		T		C
		6.  K2		C		C

		d.	客让G1， 用 >-1 代替 >0和=0
		* if Q0 > P0*P1/(P0 + P1)
		0.	G1				T
		4.  G1				C

		e.	客让G2， 用=-1和<-1代替<0
		* if  Q1*Q2/(Q1 + Q2) > P1
		0.  T		T		G2
		1.  C		T		G2
		2.	T		C		G2
		3.  C		C		G2

	*/

	return dec
}

func CampareCrownInfo(old, new interface{}) bool {
	if new.(InputCrownWinLose_t).Num < old.(InputCrownWinLose_t).Num {
		return true
	}
	return false
}

func FindCrownInfo(old, key interface{}) bool {
	return old.(InputCrownWinLose_t).Num == key.(int)
}

func FetchCrownData(url string) {
	url += fmt.Sprintf("%d000", time.Now().Unix())
	doc := FetchURL(url)
	g_crown_data = NewSortedLinkedList(1000, CampareCrownInfo, FindCrownInfo)
	doc.Find("odds i").Each(func(i int, s *goquery.Selection) {
		g := strings.Split(s.Text(), ",")
		if g[0] == "3" {
			g_crown_data.PutOnTop(NewInputCrownInfo(g[1], g[6], g[7], g[8]))
		}
	})

}

func ParseTiCaiData(s *goquery.Selection) (input_ti_cai InputTiCaiWinLose_t) {
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
		case 7:
			input_ti_cai.Guest = elem.Find("a").Text()
		case 12:
			elem.Find("td").Each(func(j int, se *goquery.Selection) {
				switch j {
				case 1:
					input_ti_cai.Odds1[0], _ = strconv.ParseFloat(se.Text(), 64)
				case 2:
					input_ti_cai.Odds1[1], _ = strconv.ParseFloat(se.Text(), 64)
				case 3:
					input_ti_cai.Odds1[2], _ = strconv.ParseFloat(se.Text(), 64)
				case 5:
					input_ti_cai.Handicap, _ = strconv.Atoi(se.Text())
				case 6:
					input_ti_cai.Odds2[0], _ = strconv.ParseFloat(se.Text(), 64)
				case 7:
					input_ti_cai.Odds2[1], _ = strconv.ParseFloat(se.Text(), 64)
				case 8:
					input_ti_cai.Odds2[2], _ = strconv.ParseFloat(se.Text(), 64)
				}
			})

		}
	})

	return
}

func CompareTiCaiData(old, new interface{}) bool {
	return new.(InputTiCaiWinLose_t).Num < old.(InputTiCaiWinLose_t).Num
}

func FindTiCaiData(old, key interface{}) bool {
	return old.(InputTiCaiWinLose_t).Num == key.(int)
}

func FetchTiCaiData(url string) {
	doc := FetchURL(url)

	// Find the urls
	elem := doc.Find(".td_div tbody tr")
	if elem.Length() == 0 {
		return
	}

	g_ticai_data = NewSortedLinkedList(1000, CompareTiCaiData, FindTiCaiData)
	elem.Each(func(i int, s *goquery.Selection) {
		val, exists := s.Attr("class")
		if exists && (val == "ni" || val == "ni2") {
			ticai_data := ParseTiCaiData(s)

			if time.Now().After(ticai_data.Close) {
				return
			}

			//TODO: 暂时只处理盘口-1， 1
			if ticai_data.Handicap != -1 && ticai_data.Handicap != 1 {
				return
			}

			g_ticai_data.PutOnTop(ticai_data)
		}
	})
}

func CompareDecision(old, new interface{}) bool {
	return new.(DecisionWinLose_t).InputTiCai.Num < old.(DecisionWinLose_t).InputTiCai.Num
}

func FindDecision(old, key interface{}) bool {
	return old.(DecisionWinLose_t).InputTiCai.Num == key.(int)
}

var g_dec_list *SortedLinkedList

func WorkOutDecisions() {
	g_dec_list = NewSortedLinkedList(1000, CompareDecision, FindDecision)
	var dec DecisionWinLose_t
	input_cat := g_ticai_data.Front()
	input_dog := g_crown_data.FindElementWithKey(input_cat.Value.(InputTiCaiWinLose_t).Num)

	for {
		//WriteMailBody("%+v\n%+v\n", input_cat.Value.(InputTiCaiWinLose_t), input_dog.Value.(InputCrownWinLose_t))
		dec = MakeDecision(input_cat.Value.(InputTiCaiWinLose_t), input_dog.Value.(InputCrownWinLose_t))
		g_dec_list.PutOnTop(dec)

		//PrintDecisionWinLose(dec)

		input_cat = input_cat.Next()
		input_dog = input_dog.Next()

		if input_cat == nil {
			break
		}
		for input_dog != nil && input_dog.Value.(InputCrownWinLose_t).Num != input_cat.Value.(InputTiCaiWinLose_t).Num {
			input_dog = input_dog.Next()
		}
		if input_dog == nil {
			break
		}
	}

	d1 := g_dec_list.Front()
	d2 := NextDecisionWithEnoughInterval(d1, d1)
	d3 := NextDecisionWithEnoughInterval(d2, d2)

	for {

		if d1 == nil || d2 == nil || d3 == nil {
			break
		}

		for {

			if d2 == nil || d3 == nil {
				break
			}
			for {

				if d3 == nil {
					break
				}
				BuildDecsions(d1.Value.(DecisionWinLose_t), d2.Value.(DecisionWinLose_t), d3.Value.(DecisionWinLose_t))
				CalculateFinalDecision()
				d3 = NextDecisionWithEnoughInterval(d3, d2)
			}
			d2 = NextDecisionWithEnoughInterval(d2, d1)
			d3 = NextDecisionWithEnoughInterval(d2, d2)
		}
		d1 = d1.Next()

		d2 = NextDecisionWithEnoughInterval(d1, d1)
		d3 = NextDecisionWithEnoughInterval(d2, d2)
	}
}

func CompareTicaiGameCloseTime(d1 DecisionWinLose_t, d2 DecisionWinLose_t) time.Duration {
	ret := d1.InputTiCai.Close.Sub(d2.InputTiCai.Close)
	return ret
}

func NextDecisionWithEnoughInterval(dec *list.Element, base *list.Element) *list.Element {
	if dec == nil || base == nil {
		return nil
	}
	for {
		dec = dec.Next()
		if dec == nil || CompareTicaiGameCloseTime(dec.Value.(DecisionWinLose_t), base.Value.(DecisionWinLose_t)) >= ONE_FOR_THREE_GAME_INTERVAL {
			break
		}
	}

	return dec
}

var g_dec_array [3]DecisionWinLose_t

func InitDecision() {
	for i := 0; i < 3; i++ {
		g_dec_array[i].Delta = 100
	}
}

func BuildDecsions(d1 DecisionWinLose_t, d2 DecisionWinLose_t, d3 DecisionWinLose_t) {
	g_dec_array[0] = d1
	g_dec_array[1] = d2
	g_dec_array[2] = d3
}

func AddDecision(dec DecisionWinLose_t) {
	if g_dec_array[0].Delta < dec.Delta {
		return
	}
	i := 0
	for i = 2; i >= 0; i-- {
		if dec.Delta < g_dec_array[i].Delta {
			break
		}
	}
	if i < 0 {
		return
	}

	for j := 0; j < i; j++ {
		g_dec_array[j] = g_dec_array[j+1]
	}
	g_dec_array[i] = dec
}

var dec_final [3]DecisionFinal_t

func PrintFinalDecisions() {
	for i := 0; i < 3; i++ {
		PrintFinalDecision(dec_final[i])
	}
}
func AddFinalDecision(f DecisionFinal_t) {
	var i int
	for i = 0; i < 3; i++ {
		if dec_final[i].benefit > f.benefit {
			break
		}
	}

	if i == 0 {
		return
	}

	for j := 0; j < i-1; j++ {
		dec_final[j] = dec_final[j+1]
	}
	dec_final[i-1] = f
}

type DecisionFinal_t struct {
	bet_crown        [3]float64
	allow_crown      [3]float64
	allow_ticai      float64
	bet_ticai        float64
	odds_crown       [3]float64
	odds_ticai       [3]float64
	bet_crown_string [3]string
	bet_ticai_string [3]string
	reserve          [3]float64
	benefit          float64
	odds_ticai_final float64
	dec_array        [3]DecisionWinLose_t
}

func CalculateFinalDecision() (f DecisionFinal_t) {
	g_sum_infor.count++

	f.odds_ticai_final = 1.0
	var x [3]float64
	var y [3]float64

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if g_dec_array[i].Bet1[j] != 0 && f.odds_ticai[i] != 0 {
				fmt.Println("ERROR: unexpeted decision, GOT two TICAI odds")
			}
			if g_dec_array[i].Bet1[j] != 0 && f.odds_ticai[i] == 0 {
				f.odds_ticai[i] = g_dec_array[i].InputTiCai.Odds1[j]
				f.bet_ticai_string[i] = BetTypeString(3, j)
			}

			if g_dec_array[i].Bet2[j] != 0 && f.odds_ticai[i] != 0 {
				fmt.Println("ERROR: unexpeted decision, GOT two TICAI odds")
			}
			if g_dec_array[i].Bet2[j] != 0 && f.odds_ticai[i] == 0 {
				f.odds_ticai[i] = g_dec_array[i].InputTiCai.Odds2[j]
				f.bet_ticai_string[i] = BetTypeString(g_dec_array[i].InputTiCai.Handicap, j)
			}

			if g_dec_array[i].BetCrown[j] != 0 && f.odds_crown[i] != 0 {
				fmt.Println("ERROR: unexpeted decision, GOT two TICAI odds")
			}
			if g_dec_array[i].BetCrown[j] != 0 && f.odds_crown[i] == 0 {
				f.odds_crown[i] = g_dec_array[i].InputCrown.Odds[j]
				f.bet_crown_string[i] = BetTypeString(2, j)
			}
		}

		f.odds_ticai_final *= f.odds_ticai[i]
		x[i] = (1+ALLOWANCE_CROWN)*f.odds_crown[i] - 2*ALLOWANCE_CROWN
	}

	y_ticai := (1+BONUS_TICAI)*f.odds_ticai_final - ALLOW_TICAI_TWO

	y[2] = x[2]
	y[1] = y[2] * x[1] / (x[2]*(y_ticai+ALLOW_TICAI_TWO-ALLOW_TICAI_ONE)/y_ticai - 1 + ALLOWANCE_CROWN)
	y[0] = y[1] * x[0] / (x[1] - 1 + ALLOWANCE_CROWN)
	x_total := 1/y_ticai + 1/y[0] + 1/y[1] + 1/y[2]

	f.bet_ticai = TOTAL_BET / x_total / y_ticai
	f.bet_crown[0] = TOTAL_BET / x_total / y[0]
	f.bet_crown[1] = TOTAL_BET / x_total / y[1]
	f.bet_crown[2] = TOTAL_BET / x_total / y[2]

	f.allow_ticai = ALLOWANCE_TICAI*f.bet_ticai + ALLOWANCE_CROWN*(f.bet_crown[0]+f.bet_crown[1]+f.bet_crown[2])
	f.allow_crown[0] = (ALLOWANCE_TICAI+ALLOW_TICAI_ONE)*f.bet_ticai + ALLOWANCE_CROWN*(f.odds_crown[0]-1)*f.bet_crown[0]
	f.allow_crown[1] = (ALLOWANCE_TICAI+ALLOW_TICAI_ONE)*f.bet_ticai + ALLOWANCE_CROWN*(f.odds_crown[1]-1)*f.bet_crown[1] + ALLOWANCE_CROWN*f.bet_crown[0]
	f.allow_crown[2] = (ALLOWANCE_TICAI+ALLOW_TICAI_TWO)*f.bet_ticai + ALLOWANCE_CROWN*(f.odds_crown[2]-1)*f.bet_crown[2] + ALLOWANCE_CROWN*(f.bet_crown[0]+f.bet_crown[1])

	f.benefit = f.bet_ticai*f.odds_ticai_final*BONUS_TICAI + f.bet_ticai*f.odds_ticai_final + f.allow_ticai

	f.dec_array = g_dec_array
	AddFinalDecision(f)

	if g_sum_infor.max < f.benefit {
		g_sum_infor.max = f.benefit
	}

	return
}
func PrintFinalDecision(f DecisionFinal_t) {
	MailBufferClean()

	MailBufferWrite("\n三场比赛信息：\n")
	for i := 0; i < 3; i++ {
		MailBufferWrite("%v\n", f.dec_array[i].InputTiCai.ToString())
		MailBufferWrite("%v\n", f.dec_array[i].InputCrown.ToString())
	}

	MailBufferWrite("\n场次   注种       赔率  投注资金  投注收益  投注返水  额外收入  投注奖金    总收益\n")

	f.reserve[0] = f.bet_crown[1] + f.bet_crown[2]
	f.reserve[1] = f.bet_crown[2]
	f.reserve[2] = 0
	for i := 0; i < 3; i++ {
		MailBufferWrite("第%d场  %s   %8.2f  %8.2f  %8.2f  %8.2f  %8.2f  %8.2f  %8.2f\n",
			i,
			f.bet_crown_string[i],
			f.odds_crown[i],
			f.bet_crown[i],
			f.bet_crown[i]*f.odds_crown[i],
			f.allow_crown[i],
			0.0,
			f.bet_crown[i]*f.odds_crown[i]+f.allow_crown[i],
			f.bet_crown[i]*f.odds_crown[i]+f.allow_crown[i]+f.reserve[i])
	}

	for i := 0; i < 3; i++ {
		MailBufferWrite("第%d场  %s   %8.2f\n",
			i,
			f.bet_ticai_string[i],
			f.odds_ticai[i])
	}
	MailBufferWrite("%s   %8.2f  %8.2f  %8.2f  %8.2f  %8.2f  %8.2f  %8.2f\n",
		"体彩三串一 ",
		f.odds_ticai_final,
		f.bet_ticai,
		f.bet_ticai*f.odds_ticai_final,
		f.allow_ticai,
		f.bet_ticai*f.odds_ticai_final*BONUS_TICAI,
		f.bet_ticai*f.odds_ticai_final+f.allow_ticai,
		f.benefit)

	MailBufferDump()

}

func WorkOutSolution() {

	FetchCrownData(URL_CROWN_NORMAL)
	FetchTiCaiData(URL_TICAI_NORMAL)

	InitDecision()

	WorkOutDecisions()

	PrintFinalDecisions()

}

func NewGame() *GameWinLose_t {
	return &GameWinLose_t{}
}
func (game *GameWinLose_t) RunOnce() {

	PrepareMail()
	ClearSumInfor()

	WriteMailBody("Find Match on %v\n", time.Unix(time.Now().Unix(), 0))

	WorkOutSolution()

	WriteMailBody("Done on %v\n", time.Unix(time.Now().Unix(), 0))
	WriteMailBody("Max = %v, g_count = %v\n", g_sum_infor.max, g_sum_infor.count)

	title := fmt.Sprintf("胜平负： max=%.2f, count=%v", g_sum_infor.max, g_sum_infor.count)
	SendMail(title)

}

func RunCase(input_cat InputTiCaiWinLose_t, input_dog InputCrownWinLose_t) {
	MakeDecision(input_cat, input_dog)
}

func Case1() (InputTiCaiWinLose_t, InputCrownWinLose_t) {
	WriteMailBody("\n\nCASE 1: K1, K2\n")
	return InputTiCaiWinLose_t{Odds1: [3]float64{1.0, 2.0, 4.0}, Odds2: [3]float64{2.0, 4.0, 2}, Handicap: -1}, InputCrownWinLose_t{Odds: [3]float64{2.54, 3.67, 2.59}}
}

func Case2() (InputTiCaiWinLose_t, InputCrownWinLose_t) {
	WriteMailBody("\n\nCASE 2: G1, G2\n")
	return InputTiCaiWinLose_t{Odds1: [3]float64{2, 4, 0.5}, Odds2: [3]float64{1.5, 3.40, 2.46}, Handicap: 1}, InputCrownWinLose_t{Odds: [3]float64{2.54, 3.67, 2.59}}
}

func Case3() (InputTiCaiWinLose_t, InputCrownWinLose_t) {
	WriteMailBody("\n\nCASE 3: test\n")
	return InputTiCaiWinLose_t{Odds1: [3]float64{3.80, 3.40, 1.75}, Odds2: [3]float64{1.85, 3.50, 3.30}, Handicap: 1}, InputCrownWinLose_t{Odds: [3]float64{3.97, 3.70, 1.92}}
}

func Case4() (InputTiCaiWinLose_t, InputCrownWinLose_t) {
	return InputTiCaiWinLose_t{Odds1: [3]float64{1.73, 3.45, 3.85}, Odds2: [3]float64{3.22, 3.55, 1.86}, Handicap: -1}, InputCrownWinLose_t{Odds: [3]float64{1.95, 3.5, 3.3}}
}

func Case5() (InputTiCaiWinLose_t, InputCrownWinLose_t) {
	return InputTiCaiWinLose_t{Odds1: [3]float64{1.47, 3.95, 5.1}, Odds2: [3]float64{2.42, 3.45, 2.35}, Handicap: -1}, InputCrownWinLose_t{Odds: [3]float64{1.64, 3.85, 4.3}}
}

func Case6() (InputTiCaiWinLose_t, InputCrownWinLose_t) {
	return InputTiCaiWinLose_t{Odds1: [3]float64{4.5, 3.6, 1.59}, Odds2: [3]float64{2.05, 3.35, 2.92}, Handicap: -1}, InputCrownWinLose_t{Odds: [3]float64{4.1, 3.95, 1.81}}
}

func (game *GameWinLose_t) TestLoop() {
	RunCase(Case1())
	RunCase(Case2())
	RunCase(Case3())

	SendMail("")
}

func ToApplyK1(input_cat InputTiCaiWinLose_t, intput_dog InputCrownWinLose_t) bool {
	//K1: 主让， 用<1 代替 =0和<0
	if input_cat.Handicap != -1 {
		return false
	}
	OddsK1 := input_cat.Odds1[1] * input_cat.Odds1[2] / (input_cat.Odds1[1] + input_cat.Odds1[2])

	return input_cat.Odds2[2] > OddsK1
}

func ToApplyK2(input_cat InputTiCaiWinLose_t, intput_dog InputCrownWinLose_t) bool {
	//K2: 主让, 用=1，>1组合代替>0
	if input_cat.Handicap != -1 {
		return false
	}
	OddsK2 := input_cat.Odds2[0] * input_cat.Odds2[1] / (input_cat.Odds2[0] + input_cat.Odds2[1])

	return OddsK2 > input_cat.Odds1[0]
}

func ToApplyG1(input_cat InputTiCaiWinLose_t, intput_dog InputCrownWinLose_t) bool {
	//G1: 客让G1， 用 >-1 代替 >0和=0
	if input_cat.Handicap != 1 {
		return false
	}
	Odds := input_cat.Odds1[0] * input_cat.Odds1[1] / (input_cat.Odds1[0] + input_cat.Odds1[1])

	return input_cat.Odds2[0] > Odds
}

func ToApplyG2(input_cat InputTiCaiWinLose_t, intput_dog InputCrownWinLose_t) bool {
	//G2： 客让G2， 用=-1和<-1代替<0
	if input_cat.Handicap != 1 {
		return false
	}
	Odds := input_cat.Odds2[1] * input_cat.Odds2[2] / (input_cat.Odds2[1] + input_cat.Odds2[2])

	return Odds > input_cat.Odds1[2]
}
func (game *GameWinLose_t) TryRun() {

}
