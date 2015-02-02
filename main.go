package main

import (
	"fmt"
	"os"
	"strconv"
	"text/template"

	"github.com/codegangsta/cli"
	"github.com/davewalk/pitchfork/pitchfork"
)

func main() {
	app := cli.NewApp()
	app.Name = "Pitchfork"
	app.Usage = "A Pitchfork.com reader in your shell"
	app.Author = "Dave Walk (@ddw17)"
	app.Email = "daviddwalk@gmail.com"
	app.Version = "0.2.2"
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
		if len(c.Args()) == 0 {
			reviews, err := pitchfork.GetReviews(c.String("d"))
			if err != nil {
				fmt.Println("Hmm, seems like I couldn't get all of the reviews:", err)
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
				news, err := pitchfork.GetNews(c.String("n"))
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
