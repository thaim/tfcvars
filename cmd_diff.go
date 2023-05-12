package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/rs/zerolog/log"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/urfave/cli/v2"
)

type DiffOption struct {
	varFile            string
	includeEnv         bool
	includeVariableSet bool
}

func NewDiffOption(c *cli.Context) *DiffOption {
	opt := &DiffOption{}

	opt.varFile = c.String("var-file")
	opt.includeEnv = c.Bool("include-env")
	opt.includeVariableSet = c.Bool("include-variable-set")

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

	s, err := tfeClient.VariableSets.ListForWorkspace(ctx, w.ID, nil)
	if err != nil {
		log.Error().Err(err).Msgf("failed to list variable set in workspace %s", w.ID)
		return err
	}
	variableSetIds := make([]string, 0)
	for _, set := range s.Items {
		variableSetIds = append(variableSetIds, set.ID)
	}

	return diff(ctx, w.ID, tfeClient.Variables, variableSetIds, tfeClient.VariableSetVariables, diffOpt, os.Stdout)
}

func diff(ctx context.Context, workspaceId string, tfeVariables tfe.Variables, variableSetIds []string, tfeVariableSetVariables tfe.VariableSetVariables, diffOpt *DiffOption, w io.Writer) error {
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
	if diffOpt.includeVariableSet {
		varSetVariables := []*tfe.Variable{}
		for _, variableSetId := range variableSetIds {
			variableList, err := tfeVariableSetVariables.List(ctx, variableSetId, nil)
			if err != nil {
				log.Error().Err(err).Msgf("failed to list variables in Variable Set %s", variableSetId)
				return err
			}

			for _, variableSetVariable := range variableList.Items {
				variable := &tfe.Variable{}
				variable.ID = variableSetVariable.ID
				variable.Key = variableSetVariable.Key
				variable.Value = variableSetVariable.Value
				variable.Description = variableSetVariable.Description
				variable.Category = variableSetVariable.Category
				variable.HCL = variableSetVariable.HCL
				variable.Sensitive = variableSetVariable.Sensitive

				varSetVariables = append(varSetVariables, variable)
			}
		}

		varsSrc.Items = append(varsSrc.Items, varSetVariables...)
	}
	vfSrc := NewTfvarsVariable(varsSrc.Items)

	vfDest, err := NewTfvarsFile(diffOpt.varFile)
	if err != nil {
		return err
	}

	dmp := diffmatchpatch.New()
	a, b, c := dmp.DiffLinesToChars(vfSrc.BuildHCLFileString(), vfDest.BuildHCLFileString())
	diffs := dmp.DiffMain(a, b, false)
	diffs = dmp.DiffCharsToLines(diffs, c)
	var buf strings.Builder
	includeDiff := false
	for _, diff := range diffs {
		if diff.Type == diffmatchpatch.DiffDelete {
			buf.WriteString("- " + diff.Text)
			includeDiff = true
		} else if diff.Type == diffmatchpatch.DiffInsert {
			buf.WriteString("+ " + diff.Text)
			includeDiff = true
		} else {
			lines := strings.Split(diff.Text, "\n")
			for i, line := range lines {
				if i == len(lines)-1 {
					break
				}
				buf.WriteString("  " + line + "\n")
			}
		}
	}
	if includeDiff {
		fmt.Fprint(w, buf.String())
	}

	return nil
}
