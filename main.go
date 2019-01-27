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
	appName         string = "urdudict"
	appUsage        string = "Urdu dict in your terminal"
	appVersion             = "0.3.0"
	rekhta                 = "https://www.rekhta.org/urdudictionary/?lang=1&keyword="
	resultsTemplate string = `{{ if .Meanings }}Found meaning
~~~~~~~~~~~~~
{{range $key, $value := .Meanings }}{{ $value.Word }} - {{ $value.Meaning }} {{end}}{{ else }}No meanings found{{ end }}{{ if .WordSuggestions }}

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

	doc, err := goquery.NewDocument(rekhta + queryWord)
	if err != nil {
		log.Fatal(err)
	}

	res := Results{}

	// Found Meanings
	// only single word meaning in v0.3
	meaning := doc.Find(".dicSrchWrdSyno").Text()
	if meaning != "" {
		res.Meanings = []MeaningPairs{
			MeaningPairs{
				Word:    fmt.Sprintf("%s (%s)", strings.TrimSpace(doc.Find(".dicSrchWord").Text()), doc.Find(".dicSrchMnngUrdu").Text()),
				Meaning: meaning,
			}}
	}

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
