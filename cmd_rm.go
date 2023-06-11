package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type RemoveOption struct {
	variableKey string
	autoApprove bool
	in          io.Reader
	out         io.Writer
}

func NewRemoveOption(c *cli.Context) *RemoveOption {
	var opt = &RemoveOption{}
	opt.variableKey = c.String("variable")
	if opt.variableKey == "" {
		return nil
	}

	opt.autoApprove = c.Bool("auto-approve")

	opt.in = os.Stdin
	opt.out = os.Stdout

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
	if rmOpt == nil {
		log.Error().Msg("failed to parse options")
		return errors.New("failed to parse options")
	}
	log.Debug().Msgf("rmOpt: %+v", rmOpt)

	return remove(ctx, workspace.ID, tfeClient.Variables, rmOpt)
}

func remove(ctx context.Context, workspaceId string, tfeVariables tfe.Variables, rmOpt *RemoveOption) error {
	variables, err := tfeVariables.List(ctx, workspaceId, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to list variables")
		return err
	}

	var targetVariable *tfe.Variable
	for _, variable := range variables.Items {
		if variable.Key == rmOpt.variableKey {
			targetVariable = variable
			break
		}
	}

	if targetVariable == nil {
		msg := fmt.Sprintf("variable '%s' not found", rmOpt.variableKey)
		log.Error().Msg(msg)
		return fmt.Errorf(msg)
	}
	if !rmOpt.autoApprove {
		fmt.Fprintf(rmOpt.out, "delete variable %s", targetVariable.Key)
		fmt.Print(rmOpt.out, "Are you shure you want to delete variable in Terraform Cloud? [y/n]: ")
		res, err := confirm(rmOpt.in)
		if err != nil {
			return err
		}
		if !res {
			return nil
		}
	}

	err = tfeVariables.Delete(ctx, workspaceId, targetVariable.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete variable")
		return err
	}

	return nil
}
