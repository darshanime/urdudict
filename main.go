package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/urfave/cli"
)

const (
	appName         string = "urdudict"
	appUsage        string = "Urdu dict in your terminal"
	appVersion             = "v0.5.0"
	rekhta                 = "https://www.rekhta.org/urdudictionary/?lang=1&keyword="
	resultsTemplate string = `{{ if .Meanings }}Found meaning
~~~~~~~~~~~~~{{range $key, $value := .Meanings }}
{{ $value.Word }} - {{ $value.Meaning }} {{end}}{{ else }}No meanings found{{ end }}{{ if .WordSuggestions }}

Did you mean
~~~~~~~~~~~~
{{range $value := .WordSuggestions }}{{ $value }} {{end}}

Source: rekhta.org{{ end }}
`
)

type MeaningPairs struct {
	Word    string
	Meaning string
}

type Results struct {
	Meanings        []MeaningPairs
	WordSuggestions []string
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

	queryWord := c.Args().First()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", rekhta+queryWord, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:66.0) Gecko/20100101 Firefox/66.0")
	response, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()
	if response.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", response.StatusCode, response.Status)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)

	if err != nil {
		log.Fatal(err)
	}

	res := Results{}

	doc.Find(".rekhtaDicSrchWord").Each(
		func(i int, s *goquery.Selection) {
			word := s.Find("h4").Text()
			hindi := s.Find(".dicSrchMnngUrdu").Text()
			meanings := []string{}
			s.Find(".dicSrchWrdSyno").Each(
				func(i int, s *goquery.Selection) {
					meanings = append(meanings, s.Text())
				})
			res.Meanings = append(res.Meanings, MeaningPairs{
				Word:    fmt.Sprintf("%s (%s)", strings.TrimSpace(word), hindi),
				Meaning: strings.TrimSpace(strings.Join(meanings, " | ")),
			})
		})

	// Did you mean
	doc.Find("a.didUMeanWrd").Each(
		func(i int, s *goquery.Selection) {
			didYouMean := s.Find("span").Text()
			res.WordSuggestions = append(res.WordSuggestions, didYouMean)
		})

	resultsTmpl, err := template.New("Meanings").Parse(resultsTemplate)

	if err != nil {
		panic(err)
	}

	err = resultsTmpl.Execute(os.Stdout, res)

	if err != nil {
		panic(err)
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = appUsage
	app.Version = appVersion
	app.Action = run
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
