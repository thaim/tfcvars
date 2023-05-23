package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/urfave/cli/v2"
)

var (
	organization  string
	workspaceName string
	version       = ""
	revision      = ""
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
				Usage: "Show this help",
			},
			{
				Name:   "show",
				Action: Show,
				Flags:  showFlags(),
				Usage: "Show variables on Terraform Cloud",
			},
			{
				Name:   "diff",
				Action: Diff,
				Flags:  diffFlags(),
				Usage: "Show difference of variables between local tfvars and Terraform Cloud variables",
			},
			{
				Name:   "pull",
				Action: Pull,
				Flags:  pullFlags(),
				Usage: "update local tfvars with Terraform Cloud variables",
			},
			{
				Name:   "push",
				Action: Push,
				Flags:  pushFlags(),
				Usage: "update Terraform Cloud variables with local tfvars",
			},
		},
		Version: versionFormatter(getVersion(), getRevision()),
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	os.Exit(0)
}

func showFlags() []cli.Flag {
	return []cli.Flag{
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
		&cli.StringFlag{
			Name:  "variable",
			Usage: "Show specified variable",
		},
		&cli.BoolFlag{
			Name:  "include-env",
			Usage: "include env Category variables",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "include-variable-set",
			Usage: "include Variable Set variables",
			Value: false,
		},
		&cli.GenericFlag{
			Name:  "format",
			Usage: "format to display variables",
			Value: &FormatType{
				Enum:    []string{"detail", "tfvars", "table"},
				Default: "detail",
			},
		},
	}
}

func pullFlags() []cli.Flag {
	return []cli.Flag{
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
		&cli.BoolFlag{
			Name:  "merge",
			Usage: "merge variables into existing vars file",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "include-env",
			Usage: "include env Category variables",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "include-variable-set",
			Usage: "include Variable Set variables",
			Value: false,
		},
	}
}

func pushFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "var-file",
			Usage: "Input filename to push variables",
			Value: "terraform.tfvars",
		},
		&cli.StringFlag{
			Name:  "variable",
			Usage: "Crate or Update Specified variable",
		},
		&cli.BoolFlag{
			Name:  "delete",
			Usage: "delete variables not defined in local",
			Value: false,
		},
	}
}

func diffFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "var-file",
			Usage: "Input filename to push variables",
			Value: "terraform.tfvars",
		},
		&cli.BoolFlag{
			Name:  "include-env",
			Usage: "include env Category variables",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "include-variable-set",
			Usage: "include Variable Set variables",
			Value: false,
		},
	}
}

func versionFormatter(version string, revision string) string {
	if version == "" {
		version = "devel"
	}

	if revision == "" {
		return version
	}
	return fmt.Sprintf("%s (rev: %s)", version, revision)
}

func getVersion() string {
	if version != "" {
		return version
	}
	i, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	return i.Main.Version
}

func getRevision() string {
	if revision != "" {
		return revision
	}
	i, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	for _, s := range i.Settings {
		if s.Key == "vcs.revision" {
			return s.Value
		}
	}

	return ""
}
