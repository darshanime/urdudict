package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli"
)

const (
	app_name    string = "urdudict"
	app_usage   string = "Urdu dict in your terminal"
	app_version        = "0.0.1"
)

type TooManyArgsError struct {
	c *cli.Context
}

func (t TooManyArgsError) Error() string {
	return fmt.Sprintf("Enter only 1 word. Received %d (%v)", len(t.c.Args()), strings.Join(t.c.Args(), ", "))
}

type Result struct {
	query string
}

func run(c *cli.Context) error {
	if c.NArg() > 1 {
		return &TooManyArgsError{c}
	}

	fmt.Printf("Looking up the word: %+v\n", c.Args().Get(0))
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
