package utils

import (
	"container/list"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	ALLOWANCE_TICAI             float64       = 0.075
	ALLOWANCE_CROWN             float64       = 0.02
	BONUS_TICAI                 float64       = 0.0
	ALLOW_TICAI_ONE             float64       = 0.1
	ALLOW_TICAI_TWO             float64       = 0.2
	TOTAL_BET                   float64       = 10000.0
	ONE_FOR_THREE_GAME_INTERVAL time.Duration = 1 * 3600 * time.Second
	URL_TICAI_NORMAL            string        = "http://www.310win.com/buy/jingcai.aspx?typeID=105&oddstype=2"
	URL_CROWN_NORMAL            string        = "http://www.310win.com/data/op101.xml?"
	URL_TICAI_OVERUNDER         string        = "http://www.310win.com/buy/jingcai.aspx?typeID=103&oddstype=2"
	URL_CROWN_OVERUNDER         string        = "http://www.310win.com/data/dx101.xml?"
	RUN_INTERVAL_NORMAL         int           = 3600 //second
	RUN_INTERVAL_HIGH_PEAK      int           = 300
	TIME_HIGH_PEAK_START_HOUR   int           = 8
	TIME_HIGH_PEAK_START_MIN    int           = 35
	TIME_HIGH_PEAK_END_HOUR     int           = 9
	TIME_HIGH_PEAK_END_MIN      int           = 35
	MAIL_TO                     string        = "77264952@qq.com"
	MAIL_FROM                   string        = "jonathan_tester<77264952@qq.com>"
)

var BENCHMARCK float64 = 10000.0
var MAIL_RECV_LIST = []string{"77264952@qq.com"}

var g_fetch_result = FETCH_OK_AFTER_OK
var g_fetch_sleep_count int
var g_fetch_sleep_time time.Duration

var g_mail_body string
var g_mail_title string

type SortedLinkedList struct {
	*list.List
	Limit       int
	compareFunc func(old, cur interface{}) bool
	findFunc    func(old, key interface{}) bool
}

const (
	TYPE_GAME_GOALS = iota
	TYPE_GAME_WINLOSE
	TYPE_GAME_NUM
)

var flag_is_debug bool
var FlagMode int
var FlagGameType int

func ParseFlag() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	flag_benchmark := flag.Float64("b", 1.0, "ratio of benchmark by total bet for benefit")
	flag_delivery_way := flag.Int("d", 1, "1: debug version, print. 0: release version, send mail")

	flag.IntVar(&FlagGameType, "t", -1, "-1: run all. 0: goals. 1: winlose")
	flag.IntVar(&FlagMode, "m", 1, "1: run once. 2: run forever. 3. run test. 4: try run some test")

	flag.Parse()
	flag_is_debug = (*flag_delivery_way == 1)

	BENCHMARCK = TOTAL_BET * *flag_benchmark
}

func SleepSleep(weak_up *int) {
	interval := 0
	if *weak_up != 0 {
		log.Printf("Sleep for %v\n", time.Second*time.Duration(*weak_up-int(time.Now().Unix())))
		time.Sleep(time.Second * time.Duration(*weak_up-int(time.Now().Unix())))
		log.Println("Weakup")
	}

	time_now := time.Now()
	if (time_now.Hour() == TIME_HIGH_PEAK_START_HOUR && time_now.Minute() >= TIME_HIGH_PEAK_START_MIN) || (time_now.Hour() == TIME_HIGH_PEAK_END_HOUR && time_now.Minute() <= TIME_HIGH_PEAK_END_MIN) {
		interval = RUN_INTERVAL_HIGH_PEAK
	} else if *weak_up == 0 {
		if time_now.Minute() < TIME_HIGH_PEAK_START_MIN {
			interval = 60 * (TIME_HIGH_PEAK_START_MIN - time_now.Minute())
		} else {
			interval = RUN_INTERVAL_NORMAL - 60*(time_now.Minute()-TIME_HIGH_PEAK_START_MIN)
		}
	} else {
		interval = RUN_INTERVAL_NORMAL
	}

	*weak_up = int(time_now.Unix()) + interval
}

func NewSortedLinkedList(limit int, compare func(old, new interface{}) bool, find func(old, new interface{}) bool) *SortedLinkedList {
	return &SortedLinkedList{list.New(), limit, compare, find}
}

func (this SortedLinkedList) FindElement(value interface{}) *list.Element {
	for element := this.Front(); element != nil; element = element.Next() {
		tempValue := element.Value
		if this.compareFunc(tempValue, value) {
			return element
		}
	}
	return nil
}

func (this SortedLinkedList) FindElementWithKey(key interface{}) *list.Element {
	for element := this.Front(); element != nil; element = element.Next() {
		tempValue := element.Value
		if this.findFunc(tempValue, key) {
			return element
		}
	}
	return nil

}
func (this SortedLinkedList) PutOnTop(value interface{}) {
	if this.List.Len() == 0 {
		this.PushFront(value)
		return
	}
	if this.List.Len() < this.Limit && this.compareFunc(value, this.Back().Value) {
		this.PushBack(value)
		return
	}
	if this.compareFunc(this.List.Front().Value, value) {
		this.PushFront(value)
	} else if this.compareFunc(this.List.Back().Value, value) && this.compareFunc(value, this.Front().Value) {
		element := this.FindElement(value)
		if element != nil {
			this.InsertBefore(value, element)
		}
	}
	if this.Len() > this.Limit {
		this.Remove(this.Back())
	}
}

func (this SortedLinkedList) PrintAll() {
	for ele := this.Front(); ele != nil; ele = ele.Next() {
		WriteMailBody("%+v\n", ele)
	}
}

func StringToFloat(elem *goquery.Selection) float64 {
	str := elem.Text()
	res, _ := strconv.ParseFloat(str, 64)
	return res
}

const (
	FETCH_OK_AFTER_OK = iota
	FETCH_OK_AFTER_FAIL
	FETCH_FAIL_AFTER_OK
	FETCH_FAIL_AFTER_FAIL
)

func SleepBeforeFetch(fetch_result bool) {
	switch g_fetch_result {
	case FETCH_OK_AFTER_OK, FETCH_OK_AFTER_FAIL:
		if fetch_result {
			g_fetch_result = FETCH_OK_AFTER_OK
		} else {
			g_fetch_result = FETCH_FAIL_AFTER_OK
		}
	case FETCH_FAIL_AFTER_OK, FETCH_FAIL_AFTER_FAIL:
		if fetch_result {
			g_fetch_result = FETCH_OK_AFTER_FAIL
		} else {
			g_fetch_result = FETCH_FAIL_AFTER_FAIL
		}
	}

	time_to_sleep := time.Second

	switch g_fetch_result {
	case FETCH_OK_AFTER_OK:
		time_to_sleep /= 10
	case FETCH_OK_AFTER_FAIL:
		time_to_sleep *= 3
		log.Printf("Succeeded to fetch after a failure, wait for %v, count=%v, total_sleep=%v\n", time_to_sleep, g_fetch_sleep_count, g_fetch_sleep_time)
	case FETCH_FAIL_AFTER_OK:
		time_to_sleep *= 3
		log.Printf("Failed to fetch the page, wait for %v, count=%v, total_sleep=%v\n", time_to_sleep, g_fetch_sleep_count, g_fetch_sleep_time)
	case FETCH_FAIL_AFTER_FAIL:
		time_to_sleep *= 5
		log.Printf("Failed to fetch the page, wait for %v, count=%v, total_sleep=%v\n", time_to_sleep, g_fetch_sleep_count, g_fetch_sleep_time)
	}

	g_fetch_sleep_count++
	g_fetch_sleep_time += time_to_sleep
	time.Sleep(time_to_sleep)
}

func PrepareMail() {
	g_mail_body = ""
}

func WriteMailBody(format string, args ...interface{}) {
	g_mail_body += fmt.Sprintf(format, args...)
}

var g_mail_buffer string

func MailBufferClean() {
	g_mail_buffer = ""
}
func MailBufferWrite(format string, args ...interface{}) {
	g_mail_buffer += fmt.Sprintf(format, args...)
}

func MailBufferDump() {
	WriteMailBody(g_mail_buffer)
	g_mail_buffer = ""
}

func WriteMailTitle(title string) {
	//g_mail_title = fmt.Sprintf("有效场次(%d)_最大(%.2f)", g_sum_infor.count, g_sum_infor.max)

}
func SendMail(title string) {

	g_mail_title = title

	if flag_is_debug {
		fmt.Printf("%s\n", g_mail_body)
		fmt.Printf("Sumarry: %s\n", g_mail_title)
		return
	}

	g_auth := smtp.PlainAuth(
		"",
		"77264952@qq.com",
		"ddyynbvevqqtbicg",
		"smtp.qq.com",
	)
	smtp.SendMail(
		"smtp.qq.com:587",
		g_auth,
		"77264952@qq.com",
		MAIL_RECV_LIST,
		[]byte("TO: "+MAIL_TO+",\r\nFrom: "+MAIL_FROM+",\r\nsubject: "+g_mail_title+"\r\n\r\n"+g_mail_body),
		//[]byte(sub+content),
	)
}

func FetchURL(url string) *goquery.Document {

	for {
		old := time.Now()
		WriteMailBody("Start to scrape url[%s], %v\n", url, time.Unix(old.Unix(), 0))
		log.Printf("Start to scrape url[%s], %v\n", url, time.Unix(old.Unix(), 0))

		doc, err := goquery.NewDocument(url)
		WriteMailBody("End to scrape, cost %v\n\n", time.Now().Sub(old))
		log.Printf("End to scrape, cost %v\n\n", time.Now().Sub(old))

		if err != nil {
			SleepBeforeFetch(false)
			continue
		}

		return doc
	}

	//return nil
}

func FetchWithCookie(url string, cookie string) *goquery.Document {
	baseUrl := url
	client := &http.Client{}
	req, _ := http.NewRequest("GET", baseUrl, nil)
	//req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")
	//req.Header.Add("Referer", baseUrl)
	//req.Header.Add("Cookie", "oddsID_101=o_3") // 也可以通过req.Cookie()的方式来设置cookie
	req.Header.Add("Cookie", cookie)
	res, err := client.Do(req)
	if err != nil {
		WriteMailBody("Cannot connect url[%s], err[%s]\n", url, err)
		log.Fatal(err)
	}
	defer res.Body.Close()
	//最后直接把res传给goquery就可以来解析网页了
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		WriteMailBody("Cannot connect url[%s], err[%s]\n", url, err)
		log.Fatal(err)
	}

	return doc
}

func ParseGameDate(game_date string) time.Time {
	arr1 := strings.Split(game_date, "：")
	arr2 := strings.Fields(arr1[1])

	arr3 := strings.Split(arr2[0], "-")
	arr4 := strings.Split(arr2[1], ":")

	year, _ := strconv.Atoi(arr3[0])
	month, _ := strconv.Atoi(arr3[1])
	day, _ := strconv.Atoi(arr3[2])
	hour, _ := strconv.Atoi(arr4[0])
	min, _ := strconv.Atoi(arr4[1])
	return time.Date(year,
		time.Month(month),
		day,
		hour,
		min,
		0,
		0,
		time.Now().Location())
}
