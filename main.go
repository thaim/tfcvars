package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "terravars",
		Usage: "synchronize terraform cloud variables",
		Commands: []*cli.Command{
			{
				Name: "help",
			},
			{
				Name: "diff",
			},
			{
				Name: "pull",
			},
			{
				Name: "push",
			},
			{
				Name: "login",
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	os.Exit(0)
}
