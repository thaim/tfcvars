package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/google/go-cmp/cmp"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type DiffOption struct {
	varFile    string
	includeEnv bool
}

func NewDiffOption(c *cli.Context) *DiffOption {
	opt := &DiffOption{}

	opt.varFile = c.String("var-file")
	opt.includeEnv = c.Bool("include-env")

	return opt
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
	varsSrc, err := tfeVariables.List(ctx, workspaceId, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to list variables")
		return err
	}
	if !diffOpt.includeEnv {
		filteredVars := []*tfe.Variable{}
		for _, v := range varsSrc.Items {
			if v.Category != tfe.CategoryEnv {
				filteredVars = append(filteredVars, v)
			}
		}
		varsSrc.Items = filteredVars
	}
	vfSrc := NewTfvarsVariable(varsSrc.Items)

	vfDest := NewTfvarsFile(diffOpt.varFile)

	fmt.Fprint(w, cmp.Diff(vfSrc.BuildHCLFileString(), vfDest.BuildHCLFileString()))

	return nil
}
