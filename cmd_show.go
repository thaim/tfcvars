package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type ShowOption struct {
	varFile string
	local   bool
}

func NewShowOption(c *cli.Context) *ShowOption {
	var opt = &ShowOption{}

	opt.varFile = c.String("var-file")
	opt.local = c.Bool("local")

	return opt
}

// Show display variable list
func Show(c *cli.Context) error {
	ctx := context.Background()
	log.Debug().Msg("show command")

	showOpt := NewShowOption(c)
	workspaceId := ""
	var tfeVariables tfe.Variables
	if requireTfcAccess(showOpt) {
		tfeClient, err := NewTfeClient(c)
		if err != nil {
			log.Error().Err(err).Msg("faile to build tfe client")
			return err
		}
		w, err := tfeClient.Workspaces.Read(ctx, organization, workspaceName)
		if err != nil {
			log.Error().Err(err).Msgf("failed to access workspace %s/%s", organization, workspaceName)
			return err
		}

		workspaceId = w.ID
		tfeVariables = tfeClient.Variables
	}

	err := show(ctx, workspaceId, tfeVariables, showOpt, os.Stdout)
	if err != nil {
		return err
	}

	return nil
}

func show(ctx context.Context, workspaceId string, tfeVariables tfe.Variables, showOpt *ShowOption, w io.Writer) error {
	var vars *tfe.VariableList
	var err error

	if showOpt.local {
		// terraform.tfvarsを読んで vars 変数に格納する
		log.Debug().Msg("local variable show command")
		vars = &tfe.VariableList{}
		vars.Items = []*tfe.Variable{}

		p := hclparse.NewParser()
		file, diags := p.ParseHCLFile(showOpt.varFile)
		if diags.HasErrors() {
			return errors.New(diags.Error())
		}
		attrs, _ := file.Body.JustAttributes()
		for attrKey, attrValue := range attrs {
			val, _ := attrValue.Expr.Value(nil)
			vars.Items = append(vars.Items, &tfe.Variable{
				Key:   attrKey,
				Value: val.AsString(),
			})
		}

	} else {
		vars, err = tfeVariables.List(ctx, workspaceId, nil)
		if err != nil {
			log.Error().Err(err).Msg("failed to list variables")
			return err
		}
	}

	for _, v := range vars.Items {
		fmt.Fprintf(w, "Key: %s\n", v.Key)
		fmt.Fprintf(w, "Value: %s\n", v.Value)
		fmt.Fprintf(w, "Description: %s\n", v.Description)
		fmt.Fprintf(w, "Sensitive: %s\n", strconv.FormatBool(v.Sensitive))
		fmt.Fprintf(w, "\n")
	}

	return nil
}

func requireTfcAccess(opt *ShowOption) bool {
	// local以外のオプションでも条件分岐が生じそうなので関数化している
	return !opt.local
}
