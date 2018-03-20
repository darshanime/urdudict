package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/urfave/cli"
)

const (
	app_name         string = "urdudict"
	app_usage        string = "Urdu dict in your terminal"
	app_version             = "0.2.0"
	rekhta                  = "https://www.rekhta.org/urdudictionary/?lang=1&keyword="
	results_template string = `{{ if .Meanings }}
Found meanings:
~~~~~~~~~~~~~~~
{{range $key, $value := .Meanings }}
{{ $value.Word }} - {{ $value.Meaning }} {{end}}
{{ if .Dictionary.Word }}
Dictionary meanings:
~~~~~~~~~~~~~~~~~~~~
{{ .Dictionary.Word }}
{{ .Dictionary.Meaning }}
{{ end }}{{ if .Word_suggestions }}
Did you mean
~~~~~~~~~~~~
{{range $value := .Word_suggestions }} {{ $value }} {{end}} {{ end }}

Source: rekhta.org
{{ else }}No meanings found{{ end }}
`
)

type MeaningPairs struct {
	Word    string
	Meaning string
}

type DictionaryMeaning struct {
	Word    string
	Meaning string
}

type Results struct {
	Meanings         []MeaningPairs
	Dictionary       DictionaryMeaning
	Word_suggestions []string
}

type InvalidArgsError struct {
	c *cli.Context
}

func (t InvalidArgsError) Error() string {
	return fmt.Sprintf("Invalid arguments: Received %d (%v)", len(t.c.Args()), strings.Join(t.c.Args(), ", "))
}

func run(c *cli.Context) error {
	if c.NArg() == 0 || c.NArg() > 2 {
		return &InvalidArgsError{c}
	}

	var show_dict bool
	// true if asked for dict meaning
	if c.BoolT("hide_dictionary") {
		show_dict = false
	} else {
		show_dict = true
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
	if show_dict == true {
		heading := doc.Find("div.meaningDetailContainer ul li div div span div2 span").Text()
		description := doc.Find("div.meaningDetailContainer ul li div div span div2 p").Text()
		res.Dictionary = DictionaryMeaning{Word: heading, Meaning: description}
	} else {
		res.Dictionary = DictionaryMeaning{}
	}

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
	app.Flags = []cli.Flag{
		cli.BoolTFlag{
			Name:  "hide_dictionary, hd",
			Usage: "set to display the dictionary meaning",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
