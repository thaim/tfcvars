package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rs/zerolog/log"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/urfave/cli/v2"
	"github.com/zclconf/go-cty/cty"
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

	return diff(ctx, w.ID, tfeClient.Variables, tfeClient.VariableSets, tfeClient.VariableSetVariables, diffOpt, os.Stdout)
}

func diff(ctx context.Context, workspaceId string, tfeVariables tfe.Variables, tfeVariableSets tfe.VariableSets, tfeVariableSetVariables tfe.VariableSetVariables, diffOpt *DiffOption, w io.Writer) error {
	varsSrc, err := tfeVariables.List(ctx, workspaceId, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to list variables")
		return err
	}
	if diffOpt.includeVariableSet {
		variableSetVariables, err := listVariableSetVariables(ctx, workspaceId, tfeVariableSets, tfeVariableSetVariables)
		if err != nil {
			log.Error().Err(err).Msg("failed to list VariableSetVariables")
			return err
		}
		varsSrc.Items = append(varsSrc.Items, variableSetVariables...)
	}
	if !diffOpt.includeEnv {
		varsSrc.Items = FilterEnv(varsSrc.Items)
	}
	vfSrc := NewTfvarsVariable(varsSrc.Items)

	vfDest, err := NewTfvarsFile(diffOpt.varFile)
	if err != nil {
		return err
	}

	includeDiff, diffString := destBasedDiff(vfSrc, vfDest)
	if includeDiff {
		fmt.Fprint(w, diffString)
	}

	return nil
}

func fileDiff(srcText, destText string) (bool, string) {
	includeDiff := false

	dmp := diffmatchpatch.New()
	a, b, c := dmp.DiffLinesToChars(srcText, destText)
	diffs := dmp.DiffMain(a, b, false)
	diffs = dmp.DiffCharsToLines(diffs, c)
	var buf strings.Builder

	for _, diff := range diffs {
		prefix := "  "
		if diff.Type == diffmatchpatch.DiffDelete {
			includeDiff = true
			prefix = "- "
		} else if diff.Type == diffmatchpatch.DiffInsert {
			includeDiff = true
			prefix = "+ "
		}

		lines := strings.Split(diff.Text, "\n")
		for i, line := range lines {
			if i == len(lines)-1 {
				continue
			}
			buf.WriteString(prefix + line + "\n")
		}
	}

	return includeDiff, buf.String()
}

// destBasedDiff creates a diff based on destination file format(includeing comments and variable order)
func destBasedDiff(srcVariable *Tfvars, destText *Tfvars) (bool, string) {
	w, diag := hclwrite.ParseConfig(destText.vardata, srcVariable.filename, hcl.InitialPos)
	if diag.HasErrors() {
		log.Error().Msg("failed to parse src file")
		return false, ""
	}

	// remove attributes defined in destText but not in srcVariable
destVariableLoop:
	for _, vDest := range destText.vars {
		for _, vSrc := range srcVariable.vars {
			if vDest.Key == vSrc.Key {
				continue destVariableLoop
			}
		}

		w.Body().RemoveAttribute(vDest.Key)
	}

	// add or update attributes defined in srcVariable
	for _, v := range srcVariable.vars {
		ctyValue := CtyValue(v.Value)
		if IsPrimitive(ctyValue) {
			w.Body().SetAttributeValue(v.Key, cty.StringVal(v.Value))
		} else {
			w.Body().SetAttributeValue(v.Key, ctyValue)
		}
	}

	return fileDiff(string(w.Bytes()), destText.BuildHCLFileString())
}
