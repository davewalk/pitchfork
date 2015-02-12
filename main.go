package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/codegangsta/cli"
	"github.com/davewalk/pitchfork/pitchfork"
)

func validateArgs(c *cli.Context) (err error) {
	var minScore float64
	minScore, err = strconv.ParseFloat(c.String("s"), 64)
	if err != nil {
		err = errors.New("You didn't pass a valid score")
	}
	if minScore < 0.0 || minScore > 10.0 {
		err = errors.New("The minimum score must be between 0 and 10")
		return err
	}

	var days int
	if days, err = strconv.Atoi(c.String("d")); err != nil {
		err = errors.New("Not a valid -days value")
	}
	days = days + 1
	if days > 5 {
		err = errors.New("Sorry, I can only get reviews from the last five days")
	}
	return err
}

// A responder takes a request from the user and returns data.
type responder interface {
	displayData()
}

// A defaultResponder returns the latest reviews
type defaultResponder struct {
	ctx *cli.Context
	reviewResponder
}

func (d defaultResponder) getData() (reviews []pitchfork.Review, err error) {
	reviews, err = pitchfork.GetReviews(d.ctx.String("d"))
	return reviews, err
}

func (d defaultResponder) displayData() {
	var err error

	reviews, err := d.getData()
	if err != nil {
		fmt.Println(err)
		return
	}

	d.displayReviews(reviews, d.ctx)
}

// A newsResponder returns the latest news
type newsResponder struct {
	ctx *cli.Context
}

func (n newsResponder) getData() (news []pitchfork.NewsArticle, err error) {
	news, err = pitchfork.GetNews(n.ctx.String("n"))
	return news, err
}

func (n newsResponder) displayData() {
	news, err := n.getData()
	if err != nil {
		fmt.Println(err)
		return
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

// A searchResponder returns reviews for a search query
type searchResponder struct {
	ctx *cli.Context
	reviewResponder
}

func (s searchResponder) getData() (reviews []pitchfork.Review, err error) {
	args := s.ctx.Args()
	artist := args.Tail()
	artistStr := strings.Join(artist, "+")
	if len(args) < 2 {
		errStr := "Include the name of the artist that you want reviews for."
		err = errors.New(errStr)
		return
	}
	reviews, err = pitchfork.SearchReviews(artistStr)
	return reviews, err
}

func (s searchResponder) displayData() {
	reviews, _ := s.getData()
	s.displayReviews(reviews, s.ctx)
}

type reviewResponder struct {
}

func (r reviewResponder) displayReviews(reviews []pitchfork.Review, ctx *cli.Context) {
	var minScore float64
	minScore, _ = strconv.ParseFloat(ctx.String("s"), 64)

	for _, review := range reviews {
		if review.Score >= minScore {
			var tmplStr string = ctx.String("t") + "\n"
			t, err := template.New("review").Parse(tmplStr)
			err = t.Execute(os.Stdout, review)
			if err != nil {
				fmt.Println("There was an error with the template you passed:", err)
			}
		}
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "Pitchfork"
	app.Usage = "A Pitchfork.com reader in your shell"
	app.Author = "Dave Walk (@ddw17)"
	app.Email = "daviddwalk@gmail.com"
	app.Version = "0.3.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "t",
			Value: "{{.Artist}}: {{.Album}} [{{.Score}}] ({{.Url}})",
			Usage: "A template for how you want the reviews displayed (in quotes)",
		},
		cli.StringFlag{
			Name:  "days, d",
			Value: "0",
			Usage: "Days since the last set of reviews to return",
		},
		cli.StringFlag{
			Name:  "score, s",
			Value: "0.0",
			Usage: "Minimum score for reviews to return",
		},
		cli.StringFlag{
			Name:  "num, n",
			Value: "5",
			Usage: "Number of news articles to return (max of 10)",
		},
	}
	app.Action = func(c *cli.Context) {
		err := validateArgs(c)
		if err != nil {
			fmt.Println(err)
			return
		}

		var (
			r   responder
			cmd string
		)

		if len(c.Args()) == 0 {
			cmd = "default"
		} else {
			cmd = c.Args()[0]
		}

		switch cmd {
		case "default":
			r = defaultResponder{c, reviewResponder{}}
		case "news":
			r = newsResponder{c}
		case "search":
			r = searchResponder{c, reviewResponder{}}
		default:
			fmt.Println("Hmm, that doesn't seem to be a valid command...")
			return
		}
		r.displayData()
	}

	app.Run(os.Args)
}
