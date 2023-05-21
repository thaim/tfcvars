package main

import (
	"context"
	"fmt"
	"io"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type PullOption struct {
	varFile            string
	overwrite          bool
	prevVarfile        []byte
	includeEnv         bool
	includeVariableSet bool
}

func NewPullOption(c *cli.Context) *PullOption {
	var opt = &PullOption{}

	opt.varFile = c.String("var-file")
	opt.overwrite = c.Bool("overwrite") && !c.Bool("merge")
	opt.prevVarfile = nil
	opt.includeEnv = c.Bool("include-env")
	opt.includeVariableSet = c.Bool("include-variable-set")

	return opt
}

func Pull(c *cli.Context) error {
	ctx := context.Background()
	log.Debug().Msg("pull command")

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
	pullOpt := NewPullOption(c)
	if !pullOpt.overwrite {
		src, _ := os.ReadFile(pullOpt.varFile)
		pullOpt.prevVarfile = src
	}
	log.Debug().Msgf("pullOption: %+v", pullOpt)

	f, err := os.Create(pullOpt.varFile)
	if err != nil {
		log.Error().Err(err).Msgf("cannot open varfile: %s", pullOpt.varFile)
		return err
	}
	defer f.Close()

	return pull(ctx, w.ID, tfeClient.Variables, tfeClient.VariableSets, tfeClient.VariableSetVariables, pullOpt, f)
}

func pull(ctx context.Context, workspaceId string, tfeVariables tfe.Variables, tfeVariableSets tfe.VariableSets, tfeVariableSetVariables tfe.VariableSetVariables, pullOpt *PullOption, w io.Writer) error {
	vars, err := tfeVariables.List(ctx, workspaceId, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to list variables")
		return err
	}
	if pullOpt.includeVariableSet {
		variableSetVariables, err := listVariableSetVariables(ctx, workspaceId, tfeVariableSets, tfeVariableSetVariables)
		if err != nil {
			log.Error().Err(err).Msg("failed to list VariableSetVariables")
			return err
		}
		vars.Items = append(vars.Items, variableSetVariables...)
	}
	if !pullOpt.includeEnv {
		vars.Items = FilterEnv(vars.Items)
	}

	var base []byte
	if !pullOpt.overwrite {
		base = pullOpt.prevVarfile
	}

	f, err := BuildHCLFile(vars.Items, base, pullOpt.varFile)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "%s", f.Bytes())

	return nil
}

func generateComment(key string) hclwrite.Tokens {
	tokens := hclwrite.Tokens{
		{
			Type:  hclsyntax.TokenSlash,
			Bytes: []byte("//"),
		},
		{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte(key),
		},
		{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte("="),
		},
		{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte("\"***\""),
		},
		{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte("\n"),
		},
	}

	return tokens
}
