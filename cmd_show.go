package main

import (
	"context"
	"fmt"
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

	return show(ctx, tfeClient, showOpt)
}

func show(ctx context.Context, tfeClient *tfe.Client, showOpt *ShowOption) error {
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
		fmt.Println("Key: " + v.Key)
		fmt.Println("Value: " + v.Value)
		fmt.Println("Description: " + v.Description)
		fmt.Println("Sensitive: " + strconv.FormatBool(v.Sensitive))
		fmt.Println()
	}

	return nil
}
