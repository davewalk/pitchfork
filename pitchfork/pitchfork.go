package pitchfork

import (
	"errors"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type response struct {
	review Review
	err    error
}

type query struct {
	path         string
	responseChan chan response
}

const baseurl = "http://pitchfork.com"

func splitMetadata(str string) (metadata []string) {
	metadata = strings.Split(str, ";")
	return metadata
}

func parseNewsArticle(s *goquery.Selection) (newsarticle NewsArticle) {
	title := s.Find(".info h1 a").Text()
	url, _ := s.Find(".info h1 a").Attr("href")

	a := NewsArticle{
		Title: title,
		Url:   baseurl + url,
	}

	return a
}

func getReviewDetails(q query) {
	url := baseurl + q.path
	doc, err := goquery.NewDocument(url)
	if err != nil {
		q.responseChan <- response{review: Review{}, err: err}
		return
	}

	artist := doc.Find(".info h1 a").First().Text()

	album := doc.Find(".info h2").First().Text()

	albummeta := doc.Find(".info h3").First().Text()
	albummeta = strings.Trim(albummeta, " ")
	label := splitMetadata(albummeta)[0]
	var year string
	if len(splitMetadata(albummeta)) > 1 {
		year = splitMetadata(albummeta)[1]
		year = strings.Trim(year, " ")
	}

	reviewmeta := doc.Find(".info h4").First().Text()
	reviewmeta = strings.Trim(reviewmeta, " ")
	author := splitMetadata(reviewmeta)[0]
	author = strings.Trim(author, "By ")
	var reviewdate string
	if len(splitMetadata(reviewmeta)) > 1 {
		reviewdate = splitMetadata(reviewmeta)[1]
		reviewdate = strings.Trim(reviewdate, " ")
	}

	score := doc.Find(".score").First().Text()
	score = strings.Trim(score, " ")
	var scoreNum float64
	scoreNum, err = strconv.ParseFloat(score, 64)
	if err != nil {
		q.responseChan <- response{review: Review{}, err: err}
		return
	}

	review, err := doc.Find(".object-detail .editorial").First().Html()
	if err != nil {
		q.responseChan <- response{review: Review{}, err: err}
		return
	}
	review = strings.Replace(review, "</p>", "\n", 10)

	r := Review{
		Artist:     artist,
		Album:      album,
		Label:      label,
		Year:       year,
		Reviewdate: reviewdate,
		Author:     author,
		Review:     review,
		Url:        url,
		Score:      scoreNum}

	q.responseChan <- response{review: r, err: err}
}

// Review includes the details of the Pitchfork review of an album
type Review struct {
	Artist     string
	Album      string
	Label      string
	Year       string
	Reviewdate string
	Author     string
	Review     string
	Url        string
	Score      float64
}

// NewsArticle includes the details of a news article on the Pitchfork.com homepage.
type NewsArticle struct {
	Title string
	Url   string
}

// GetNews returns the latest news articles from pitchfork.com/news
func GetNews(count string) (articles []NewsArticle, err error) {
	var countNum int
	if countNum, err = strconv.Atoi(count); err != nil {
		return
	}
	if countNum > 10 {
		err = errors.New("Sorry, I can only get the last 10 news articles.")
		return
	}

	doc, err := goquery.NewDocument(baseurl + "/news")
	if err != nil {
		return
	}

	doc.Find("#main .object-list .player-target").Each(func(i int, s *goquery.Selection) {
		if i < countNum {
			var article NewsArticle
			article = parseNewsArticle(s)
			articles = append(articles, article)
		}
	})

	return articles, err
}

// GetReviews returns the five reviews for the given day from the latest weekday
// of reviews.
func GetReviews(daysStr string) (reviews []Review, err error) {
	var days int
	if days, err = strconv.Atoi(daysStr); err != nil {
		return
	}
	days = days + 1
	if days > 5 {
		err = errors.New("Sorry, I can only get reviews from the last five days")
		return
	}
	doc, err := goquery.NewDocument(baseurl)
	if err != nil {
		return
	}

	reviews = make([]Review, 0)

	reschan := make(chan response, 5)
	var wg sync.WaitGroup

	doc.Find("#review-day-" + strconv.Itoa(days) + " .review-list a").Each(func(i int, s *goquery.Selection) {
		path, _ := s.Attr("href")
		q := query{path: path, responseChan: reschan}
		wg.Add(1)
		go getReviewDetails(q)
	})

	go func() {
		for response := range reschan {
			if response.err != nil {
				err = response.err
			} else {
				reviews = append(reviews, response.review)
			}
			wg.Done()
		}
	}()

	wg.Wait()
	close(reschan)

	return reviews, err
}

// SearchReviews searches for reviews given a query string
func SearchReviews(queryStr string) (reviews []Review, err error) {
	var doc *goquery.Document
	doc, err = goquery.NewDocument(baseurl + "/search/?query=" +
		queryStr + "&filters=album_reviews")
	if err != nil {
		return
	}
	reviews = make([]Review, 0)

	reschan := make(chan response)
	var wg sync.WaitGroup

	doc.Find(".search-group").Eq(1).Find("a").Each(func(i int, s *goquery.Selection) {
		path, _ := s.Attr("href")
		q := query{path: path, responseChan: reschan}
		wg.Add(1)
		go getReviewDetails(q)
	})

	go func() {
		for response := range reschan {
			if response.err != nil {
				err = response.err
			} else {
				reviews = append(reviews, response.review)
			}
			wg.Done()
		}
	}()

	wg.Wait()
	close(reschan)

	return
}
