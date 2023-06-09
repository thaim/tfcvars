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

func Remove(c *cli.Context) error {
	return nil
}

func remove(ctx context.Context, workspaceId string, tfeVariables tfe.Variables, rmOpt *RemoveOption, vars *tfe.VariableList) error {
	return nil
}
