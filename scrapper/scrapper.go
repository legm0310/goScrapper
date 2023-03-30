package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)
type extractedJob struct {
	id string
	title string
	detail string
}


func Scrape(term string) {
	var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?&searchword="+term
	var jobs []extractedJob
	c := make(chan []extractedJob)
	totalPages := getPages(baseURL)
	for i:=1 ; i <= totalPages; i++ {
		go getPage(i, baseURL, c)
	}

	for i:=1 ; i <= totalPages; i++ {
		extractedJobs:= <-c
		jobs = append(jobs, extractedJobs...)
	}

	writeJobs(jobs)
	fmt.Println("Done, extracted", len(jobs))
}

func getPage(page int, url string, mainChan chan<- []extractedJob) {
	var jobs []extractedJob
	c := make(chan extractedJob)
	pageURL := url + "&recruitPage=" + strconv.Itoa(page) + "&recruitPageCount=40"
	fmt.Println("Requesting", pageURL)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err:= goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".item_recruit")

	searchCards.Each(func(i int, card *goquery.Selection){
		go extractJob(card, c)
	})

	for i:=0 ; i < searchCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}
	mainChan <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob){
	id, _ := card.Attr("value")
	title := card.Find(".job_tit>a>span").Text()
	detail := CleanString(card.Find(".job_condition").Text())
	c <- extractedJob{id: id,
		title: title,
		detail: detail}
}

func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), "")
}

func getPages(url string) int {
	pages := 0
	res, err := http.Get(url)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err:= goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection){
		pages = s.Find("a").Length()
	})

	return pages
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"Link", "Title", "Detail"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		jobSlice := []string{"https://www.saramin.co.kr/zf_user/jobs/relay/view?isMypage=no&rec_idx="+job.id, job.title, job.detail}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(222)
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200{
		log.Fatalln("Request failed with Status:", res)
	}
}
