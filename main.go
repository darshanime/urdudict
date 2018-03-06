package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/urfave/cli"
)

const (
	app_name    string = "urdudict"
	app_usage   string = "Urdu dict in your terminal"
	app_version        = "0.0.1"
	rekhta             = "https://www.rekhta.org/urdudictionary/?lang=1&keyword="
)

type TooManyArgsError struct {
	c *cli.Context
}

func (t TooManyArgsError) Error() string {
	return fmt.Sprintf("Enter only 1 word. Received %d (%v)", len(t.c.Args()), strings.Join(t.c.Args(), ", "))
}

func run(c *cli.Context) error {
	if c.NArg() > 1 {
		return &TooManyArgsError{c}
	}

	query_word := c.Args().First()

	doc, err := goquery.NewDocument(rekhta + query_word)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("div.wordMeanings.didYouMean div.dict_match ul.clearfix li div.dict_card ").Each(
		func(i int, s *goquery.Selection) {
			word := s.Find("div.dict_card_left h4").Text()
			meanings := s.Find("div.dict_card_right h4").Text()
			urdu := s.Find("div.dict_card_left p.meaningUrduText").Text()
			fmt.Printf("%d: %s (%s) - %s\n", i+1, word, urdu, meanings)
		})

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = app_name
	app.Usage = app_usage
	app.Version = app_version
	app.Action = run
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
