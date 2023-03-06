package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var (
	organization  string
	workspaceName string
)

func main() {
	app := &cli.App{
		Name:  "tfcvars",
		Usage: "synchronize terraform cloud variables",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "tfetoken",
				Usage:   "The token used to authenticate with Terraform Cloud",
				EnvVars: []string{"TFE_TOKEN"},
			},
			&cli.StringFlag{
				Name:        "organization",
				Aliases:     []string{"o"},
				Usage:       "Terraform Cloud organization name to deal with",
				EnvVars:     []string{"TFCVARS_ORGANIZATION"},
				Destination: &organization,
			},
			&cli.StringFlag{
				Name:        "workspace",
				Aliases:     []string{"w"},
				Usage:       "Terraform Cloud workspace name to deal with",
				EnvVars:     []string{"TFCVARS_WORKSPACE"},
				Destination: &workspaceName,
			},
		},
		Commands: []*cli.Command{
			{
				Name: "help",
			},
			{
				Name:   "show",
				Action: Show,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "local",
						Usage: "show local variables",
						Value: false,
					},
					&cli.StringFlag{
						Name:  "var-file",
						Usage: "Input filename to read for local variable",
						Value: "terraform.tfvars",
					},
				},
			},
			{
				Name: "diff",
			},
			{
				Name:   "pull",
				Action: Pull,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "var-file",
						Usage: "Output filename to write var-file",
						Value: "terraform.tfvars",
					},
					&cli.BoolFlag{
						Name:  "overwrite",
						Usage: "overwrite existing vars file",
						Value: false,
					},
				},
			},
			{
				Name: "push",
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
