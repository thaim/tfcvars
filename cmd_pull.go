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

type PullOption struct {
}

func NewPullOption(c *cli.Context) *PullOption {
	var opt = &PullOption{}

	return opt
}

func Pull(c *cli.Context) error {
	ctx := context.Background()
	log.Debug().Msg("pull command")

	tfeClient, err := NewTfeClient(c)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to build tfe client")
		return err
	}
	w, err := tfeClient.Workspaces.Read(ctx, organization, workspaceName)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to access workspace %s/%s", organization, workspaceName)
		return err
	}
	pullOpt := NewPullOption(c)

	f, err := os.Open("terraform.tfvars")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot open terraform.tfvars")
		return err
	}
	defer f.Close()

	return pull(ctx, w.ID, tfeClient.Variables, pullOpt, f)
}

func pull(ctx context.Context, workspaceId string, tfeVariables tfe.Variables, pullOpt *PullOption, w io.Writer) error {
	vars, err := tfeVariables.List(ctx, workspaceId, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to list variables")
		return err
	}
	for _, v := range(vars.Items) {
		fmt.Fprintf(w, "Key: " + v.Key)
		fmt.Fprintf(w, "Value: " + v.Value)
		fmt.Fprintf(w, "Description: " + v.Description)
		fmt.Fprintf(w, "Sensitive: " + strconv.FormatBool(v.Sensitive))
		fmt.Fprintf(w, "\n")
	}

	return nil
}
