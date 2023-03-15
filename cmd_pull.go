package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/zclconf/go-cty/cty"
)

type PullOption struct {
	varFile   string
	overwrite bool
	prevVarfile []byte
}

func NewPullOption(c *cli.Context) *PullOption {
	var opt = &PullOption{}

	opt.varFile = c.String("var-file")
	opt.overwrite = c.Bool("overwrite") && !c.Bool("merge")
	opt.prevVarfile = nil

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
	if !pullOpt.overwrite {
		src, _ := os.ReadFile(pullOpt.varFile)
		pullOpt.prevVarfile = src
	}
	log.Debug().Msgf("pullOption: %+v", pullOpt)

	f, err := os.Create(pullOpt.varFile)
	if err != nil {
		log.Fatal().Err(err).Msgf("cannot open varfile: %s", pullOpt.varFile)
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
	var f *hclwrite.File
	if pullOpt.overwrite {
		f = hclwrite.NewEmptyFile()
	} else {
		var diags hcl.Diagnostics
		f, diags = hclwrite.ParseConfig(pullOpt.prevVarfile, pullOpt.varFile, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			log.Fatal().Msgf("failed to parse existing varfile: %s", diags.Error())
			return errors.New(diags.Error())
		}
	}
	rootBody := f.Body()

	for _, v := range vars.Items {
		if v.Category == tfe.CategoryEnv {
			// Terraform VariablesではなくEnvironment Variablesであれば出力しない
			// TODO: Env対応は別オプションで実装する
			continue
		}
		if v.Sensitive {
			rootBody.AppendUnstructuredTokens(generateComment(v.Key))
			continue
		}
		rootBody.SetAttributeValue(v.Key, cty.StringVal(v.Value))
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
