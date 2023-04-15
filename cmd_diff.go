package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type DiffOption struct {
}

func NewDiffOption(c *cli.Context) *DiffOption {
	return &DiffOption{}
}

func Diff(c *cli.Context) error {
	ctx := context.Background()
	log.Debug().Msg("show command")

	diffOpt := NewDiffOption(c)

	tfeClient, err := NewTfeClient(c)
	if err != nil {
		log.Error().Err(err).Msg("failed to build tfe client")
		return err
	}
	organization, workspaceName = updateTerraformCloudWorkspace(organization, workspaceName, ".")
	w, err := tfeClient.Workspaces.Read(ctx, organization, workspaceName)
	if err != nil {
		log.Error().Err(err).Msgf("failed to access workspace %s/%s", organization, workspaceName)
		return err
	}

	return diff(ctx, w.ID, tfeClient.Variables, diffOpt, os.Stdout)
}

func diff(ctx context.Context, workspaceId string, tfeVariables tfe.Variables, diffOpt *DiffOption, w io.Writer) error {
	varsDest, err := tfeVariables.List(ctx, workspaceId, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to list variables")
		return err
	}

	varsSrc, err := variableFile("terraform.tfvars", false)
	if err != nil {
		log.Error().Err(err).Msg("failed to read variable file")
		return err
	}

	opts := []cmp.Option{
		cmpopts.IgnoreFields(tfe.Variable{}, "ID", "Description", "Category", "HCL", "Workspace"),
		cmpopts.SortSlices(func(x, y *tfe.Variable) bool {
			return strings.Compare(x.Key, y.Key) < 0
		}),
	}

	fmt.Fprint(w, cmp.Diff(varsSrc.Items, varsDest.Items, opts...))

	return nil
}
