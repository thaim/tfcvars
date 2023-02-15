package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type ShowOption struct {
	local bool
}

func NewShowOption(c *cli.Context) *ShowOption {
	var opt = &ShowOption{}

	opt.local = c.Bool("local")

	return opt
}

// Show display variable list
func Show(c *cli.Context) error {
	ctx := context.Background()
	log.Debug().Msg("show command")

	tfeClient, err := buildClient(c)
	if err != nil {
		log.Fatal().Err(err).Msg("faile to build tfe client")
		return err
	}
	showOpt := NewShowOption(c)

	err = show(ctx, tfeClient, showOpt, os.Stdout)
	if err != nil {
		return err
	}

	return nil
}

func show(ctx context.Context, tfeClient *tfe.Client, showOpt *ShowOption, w io.Writer) error {
	var vars *tfe.VariableList

	if showOpt.local {
		// terraform.tfvarsを読んで vars 変数に格納する
		log.Debug().Msg("local variable show command")
		vars = &tfe.VariableList{}
	} else {
		w, err := tfeClient.Workspaces.Read(ctx, organization, workspaceName)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to access workspace %s/%s", organization, workspaceName)
			return err
		}
		vars, err = tfeClient.Variables.List(ctx, w.ID, tfe.VariableListOptions{})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to list variables")
			return err
		}
	}


	for _, v := range(vars.Items) {
		fmt.Fprintf(w, "Key: %s\n", v.Key)
		fmt.Fprintf(w, "Value: %s\n", v.Value)
		fmt.Fprintf(w, "Description: %s\n", v.Description)
		fmt.Fprintf(w, "Sensitive: %s\n", strconv.FormatBool(v.Sensitive))
		fmt.Fprintf(w, "\n")
	}

	return nil
}
