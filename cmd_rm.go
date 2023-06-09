package main

import (
	"context"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type RemoveOption struct {
	variableKey string
	autoApprove bool
}

func NewRemoveOption(c *cli.Context) *RemoveOption {
	var opt = &RemoveOption{}
	opt.variableKey = c.String("variable")
	opt.autoApprove = c.Bool("auto-approve")

	return opt
}

func Remove(c *cli.Context) error {
	ctx := context.Background()
	log.Debug().Msg("rm command")

	tfeClient, err := NewTfeClient(c)
	if err != nil {
		log.Error().Err(err).Msg("failed to build tfe client")
		return err
	}

	organization, workspaceName := updateTerraformCloudWorkspace(organization, workspaceName, ".")
	workspace, err := tfeClient.Workspaces.Read(ctx, organization, workspaceName)
	if err != nil {
		log.Error().Err(err).Msgf("failed to access workspace %s/%s", organization, workspaceName)
		return err
	}

	rmOpt := NewRemoveOption(c)
	log.Debug().Msgf("rmOpt: %+v", rmOpt)

	return remove(ctx, workspace.ID, tfeClient.Variables, rmOpt)
}

func remove(ctx context.Context, workspaceId string, tfeVariables tfe.Variables, rmOpt *RemoveOption) error {
	return nil
}
