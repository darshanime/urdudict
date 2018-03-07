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

	res := Results{}

	// Found Meanings
	doc.Find("div.wordMeanings.didYouMean div.dict_match ul.clearfix li div.dict_card ").Each(
		func(i int, s *goquery.Selection) {
			word := s.Find("div.dict_card_left h4").Text()
			meaning := fmt.Sprintf("%s %s", s.Find("div.dict_card_right h4").Eq(0).Text(), s.Find("div.dict_card_right h4").Eq(2).Text())
			urdu := s.Find("div.dict_card_left p.meaningUrduText").Text()
			res.Meanings = append(res.Meanings, MeaningPairs{Word: fmt.Sprintf("%s (%s)", word, urdu), Meaning: meaning})
		})

	// Dictionary meaning
	heading := doc.Find("div.meaningDetailContainer ul li div div span div2 span").Text()
	description := doc.Find("div.meaningDetailContainer ul li div div span div2 p").Text()

	res.Dictionary = DictionaryMeaning{Word: heading, Meaning: description}

	// Did you mean
	doc.Find("div.search_word ul li").Each(
		func(i int, s *goquery.Selection) {
			did_you_mean := s.Find("a").Text()
			res.Word_suggestions = append(res.Word_suggestions, did_you_mean)
		})

	results_tmpl, err := template.New("Meanings").Parse(results_template)

	if err != nil {
		panic(err)
	}

	err = results_tmpl.Execute(os.Stdout, res)

	if err != nil {
		panic(err)
	}

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
