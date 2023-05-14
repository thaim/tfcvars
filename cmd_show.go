package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/zclconf/go-cty/cty"
)

type ShowOption struct {
	varFile            string
	variableKey        string
	local              bool
	includeEnv         bool
	includeVariableSet bool
	format             string
}

func NewShowOption(c *cli.Context) *ShowOption {
	var opt = &ShowOption{}

	opt.varFile = c.String("var-file")
	opt.variableKey = c.String("variable")
	opt.local = c.Bool("local")
	opt.includeEnv = c.Bool("include-env")
	opt.includeVariableSet = c.Bool("include-variable-set")
	opt.format = c.String("format")

	return opt
}

type FormatType struct {
	Enum     []string
	Default  string
	selected string
}

func (e *FormatType) Set(value string) error {
	for _, enum := range e.Enum {
		if enum == value {
			e.selected = value
			return nil
		}
	}

	return fmt.Errorf("allowed values are %s", strings.Join(e.Enum, ", "))
}

func (e FormatType) String() string {
	if e.selected == "" {
		return e.Default
	}
	return e.selected
}

// Show display variable list
func Show(c *cli.Context) error {
	ctx := context.Background()
	log.Debug().Msg("show command")

	showOpt := NewShowOption(c)
	workspaceId := ""
	var Variables tfe.Variables
	var VariableSets tfe.VariableSets
	var VariableSetVariables tfe.VariableSetVariables
	if requireTfcAccess(showOpt) {
		tfeClient, err := NewTfeClient(c)
		if err != nil {
			log.Error().Err(err).Msg("faile to build tfe client")
			return err
		}
		organization, workspaceName = updateTerraformCloudWorkspace(organization, workspaceName, ".")
		w, err := tfeClient.Workspaces.Read(ctx, organization, workspaceName)
		if err != nil {
			log.Error().Err(err).Msgf("failed to access workspace %s/%s", organization, workspaceName)
			return err
		}

		workspaceId = w.ID
		Variables = tfeClient.Variables
		VariableSets = tfeClient.VariableSets
		VariableSetVariables = tfeClient.VariableSetVariables
	}

	err := show(ctx, workspaceId, Variables, VariableSets, VariableSetVariables, showOpt, os.Stdout)
	if err != nil {
		return err
	}

	return nil
}

func show(ctx context.Context, workspaceId string, tfeVariables tfe.Variables, tfeVariableSets tfe.VariableSets, tfeVariableSetVariables tfe.VariableSetVariables, showOpt *ShowOption, w io.Writer) error {
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
		for _, attr := range SortAttributes(attrs) {
			val, _ := attr.Expr.Value(nil)
			tfVariable := &tfe.Variable{
				Key:   attr.Name,
				Value: String(val),
			}
			if !IsPrimitive(val) {
				tfVariable.HCL = true
			}

			vars.Items = append(vars.Items, tfVariable)
		}

	} else {
		vars, err = tfeVariables.List(ctx, workspaceId, nil)
		if err != nil {
			log.Error().Err(err).Msg("failed to list variables")
			return err
		}
		if showOpt.includeVariableSet {
			variableSetVariables, err := listVariableSetVariables(ctx, workspaceId, tfeVariableSets, tfeVariableSetVariables)
			if err != nil {
				log.Error().Err(err).Msg("failed to list VariableSetVariables")
				return err
			}
			vars.Items = append(vars.Items, variableSetVariables...)
		}
	}

	filteredVars := []*tfe.Variable{}
	for _, v := range vars.Items {
		if showOpt.variableKey != "" && showOpt.variableKey != v.Key {
			continue
		}
		if !showOpt.includeEnv && v.Category == tfe.CategoryEnv {
			continue
		}

		filteredVars = append(filteredVars, v)
	}
	printVariable(w, filteredVars, showOpt)

	return nil
}

func requireTfcAccess(opt *ShowOption) bool {
	// local以外のオプションでも条件分岐が生じそうなので関数化している
	return !opt.local
}

func printVariable(w io.Writer, variables []*tfe.Variable, opt *ShowOption) {
	switch opt.format {
	case "detail":
		for _, v := range variables {
			fmt.Fprintf(w, "Key: %s\n", v.Key)
			fmt.Fprintf(w, "Value: %s\n", v.Value)
			fmt.Fprintf(w, "Description: %s\n", v.Description)
			fmt.Fprintf(w, "Sensitive: %s\n", strconv.FormatBool(v.Sensitive))
			fmt.Fprintf(w, "\n")
		}
	case "tfvars":
		f := hclwrite.NewEmptyFile()
		rootBody := f.Body()

		for _, v := range variables {
			if v.Sensitive {
				rootBody.AppendUnstructuredTokens(generateComment(v.Key))
			} else if v.HCL {
				rootBody.SetAttributeValue(v.Key, CtyValue(v.Value))
			} else {
				rootBody.SetAttributeValue(v.Key, cty.StringVal(v.Value))
			}
		}

		fmt.Fprintf(w, "%s", f.Bytes())
	case "table":
		var data [][]string
		for _, v := range variables {
			row := []string{v.Key, v.Value, strconv.FormatBool(v.Sensitive), v.Description}
			data = append(data, row)
		}
		table := tablewriter.NewWriter(w)
		table.SetHeader([]string{"Key", "Value", "Sensitive", "Description"})
		table.AppendBulk(data)
		table.Render()
	default:
		log.Error().Msgf("unknown format %s specified", opt.format)
	}
}
