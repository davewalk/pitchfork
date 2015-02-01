package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	"github.com/codegangsta/cli"
)

// Review includes the details of the Pitchfork review of an album
type Review struct {
	Artist     string
	Album      string
	Label      string
	Year       int
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

func getNews(count string) (articles []NewsArticle, err error) {
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

func getReviews(daysStr string) (reviews []Review, err error) {
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
				return
			}

			reviews = append(reviews, response.review)
			wg.Done()
		}
	}()

	wg.Wait()
	close(reschan)

	return reviews, err
}

func getReviewDetails(q query) {
	url := baseurl + q.path
	doc, err := goquery.NewDocument(url)

	artist := doc.Find(".info h1 a").First().Text()

	album := doc.Find(".info h2").First().Text()

	albummeta := doc.Find(".info h3").First().Text()
	albummeta = strings.Trim(albummeta, " ")
	label := splitMetadata(albummeta)[0]
	yearStr := splitMetadata(albummeta)[1]
	yearStr = strings.Trim(yearStr, " ")
	var year int
	year, err = strconv.Atoi(yearStr)
	if err != nil {
		return
	}

	reviewmeta := doc.Find(".info h4").First().Text()
	reviewmeta = strings.Trim(reviewmeta, " ")
	author := splitMetadata(reviewmeta)[0]
	author = strings.Trim(author, "By ")
	reviewdate := splitMetadata(reviewmeta)[1]
	reviewdate = strings.Trim(reviewdate, " ")

	score := doc.Find(".score").Text()
	score = strings.Trim(score, " ")
	var scoreNum float64
	scoreNum, err = strconv.ParseFloat(score, 64)

	review, err := doc.Find(".object-detail .editorial").First().Html()
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

func main() {
	app := cli.NewApp()
	app.Name = "Pitchfork"
	app.Usage = "A Pitchfork.com reader in your shell"
	app.Author = "Dave Walk (@ddw17)"
	app.Email = "daviddwalk@gmail.com"
	app.Version = "0.2.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "t",
			Value: "{{.Artist}}: {{.Album}} [{{.Score}}] ({{.Url}})",
			Usage: "A template for how you want the reviews displayed (in quotes)",
		},
		cli.StringFlag{
			Name:  "d",
			Value: "0",
			Usage: "Days since the last set of reviews to return",
		},
		cli.StringFlag{
			Name:  "s",
			Value: "0.0",
			Usage: "Minimum score for reviews to return",
		},
		cli.StringFlag{
			Name:  "n",
			Value: "5",
			Usage: "Number of news articles to return (max of 10)",
		},
	}
	app.Action = func(c *cli.Context) {
		if len(c.Args()) == 0 {
			reviews, err := getReviews(c.String("d"))
			if err != nil {
				fmt.Println(err)
			}

			var minScore float64
			minScore, err = strconv.ParseFloat(c.String("s"), 64)
			if err != nil {
				fmt.Println("You didn't pass a valid score")
				return
			}

			for _, review := range reviews {
				if review.Score >= minScore {
					var t *template.Template
					var tmplStr string = c.String("t") + "\n"
					t, err = template.New("review").Parse(tmplStr)
					err = t.Execute(os.Stdout, review)
					if err != nil {
						fmt.Println("There was an error with the template you passed:", err)
					}
				}
			}
		}

		if len(c.Args()) > 0 {
			if c.Args()[0] == "news" {
				news, err := getNews(c.String("n"))
				if err != nil {
					fmt.Println(err)
				}

				for _, article := range news {
					var t *template.Template
					t, err = template.New("news").Parse("{{.Title}} ({{.Url}})\n")
					err = t.Execute(os.Stdout, article)
					if err != nil {
						fmt.Println("There was an error with the template you passed:", err)
					}
				}
			}
		}
	}

	app.Run(os.Args)
}
