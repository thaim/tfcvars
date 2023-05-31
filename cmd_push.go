package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type PushOption struct {
	varFile       string
	variableKey   string
	variableValue string
	delete        bool
	autoApprove   bool
	in            io.Reader
	out           io.Writer
}

func NewPushOption(c *cli.Context) *PushOption {
	var opt = &PushOption{}
	opt.varFile = c.String("var-file")

	variable := c.String("variable")
	if variable != "" {
		splitVariable := strings.SplitN(variable, "=", 2)
		opt.variableKey = splitVariable[0]
		opt.variableValue = splitVariable[1]
	}

	opt.delete = c.Bool("delete")
	opt.autoApprove = c.Bool("auto-approve")

	opt.in = os.Stdin
	opt.out = os.Stdout

	return opt
}

func Push(c *cli.Context) error {
	ctx := context.Background()
	log.Debug().Msg("push command")

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

	pushOpt := NewPushOption(c)
	log.Debug().Msgf("pushOption: %+v", pushOpt)

	var vars *tfe.VariableList
	if pushOpt.variableKey == "" {
		vars, err = variableFile(pushOpt.varFile, true)
	} else {
		vars = BuildVariableList(pushOpt.variableKey, pushOpt.variableValue)
	}
	if err != nil {
		return err
	}

	return push(ctx, w.ID, tfeClient.Variables, pushOpt, vars)
}

func push(ctx context.Context, workspaceId string, tfeVariables tfe.Variables, pushOpt *PushOption, vars *tfe.VariableList) error {
	previousVars, err := tfeVariables.List(ctx, workspaceId, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to list variables")
		return err
	}

	if !pushOpt.autoApprove {
		diff(ctx, workspaceId, tfeVariables, nil, nil, &DiffOption{varFile: pushOpt.varFile}, pushOpt.out)

		fmt.Printf("confirm?")
		res, err := confirm(pushOpt.in)
		if err != nil {
			return err
		}
		if !res {
			return nil
		}
	}

	countUpdate := 0
	countCreate := 0
	countDelete := 0

	for _, variable := range vars.Items {
		pushed := false

		for _, targetVar := range previousVars.Items {
			if targetVar.Key == variable.Key {
				updateOpt := tfe.VariableUpdateOptions{
					Key:         tfe.String(variable.Key),
					Value:       tfe.String(variable.Value),
					Description: tfe.String(targetVar.Description),
					Category:    tfe.Category(targetVar.Category),
					HCL:         tfe.Bool(targetVar.HCL),
					Sensitive:   tfe.Bool(targetVar.Sensitive),
				}
				if !variableEqual(updateOpt, targetVar) {
					tfeVariables.Update(ctx, workspaceId, targetVar.ID, updateOpt)
					fmt.Printf("update: %s\n", variable.Key)
					countUpdate++
				}
				pushed = true
				break
			}
		}

		if !pushed {
			createOpt := tfe.VariableCreateOptions{
				Key:       tfe.String(variable.Key),
				Value:     tfe.String(variable.Value),
				Category:  tfe.Category(tfe.CategoryTerraform),
				HCL:       tfe.Bool(false),
				Sensitive: tfe.Bool(false),
			}
			tfeVariables.Create(ctx, workspaceId, createOpt)
			fmt.Printf("create: %s\n", variable.Key)
			countCreate++
		}
	}

	if pushOpt.delete {
		for _, targetVar := range previousVars.Items {
			for _, localVar := range vars.Items {
				if targetVar.Key == localVar.Key {
					continue
				}

				// variable that are defined in remote but not in local
				tfeVariables.Delete(ctx, workspaceId, targetVar.ID)
				fmt.Printf("delete: %s\n", targetVar.Key)
				countDelete++
			}
		}
	}

	log.Info().Msgf("create: %d, update: %d, delete: %d", countCreate, countUpdate, countDelete)

	return nil
}

func variableFile(varfile string, required bool) (*tfe.VariableList, error) {
	vars := &tfe.VariableList{}

	if _, err := os.Stat(varfile); os.IsNotExist(err) {
		if required {
			return nil, err
		}
		vars.Items = []*tfe.Variable{{}}
		return vars, nil
	}

	p := hclparse.NewParser()
	file, diags := p.ParseHCLFile(varfile)
	if diags.HasErrors() {
		return nil, errors.New(diags.Error())
	}
	attrs, _ := file.Body.JustAttributes()
	for attrKey, attrValue := range attrs {
		val, _ := attrValue.Expr.Value(nil)

		vars.Items = append(vars.Items, &tfe.Variable{
			Key:   attrKey,
			Value: String(val),
			HCL:   !IsPrimitive(val),
		})
	}

	return vars, nil
}

func variableEqual(updateOpt tfe.VariableUpdateOptions, targetVariable *tfe.Variable) bool {
	if *updateOpt.Key != targetVariable.Key ||
		*updateOpt.Value != targetVariable.Value ||
		*updateOpt.Description != targetVariable.Description ||
		*updateOpt.Category != targetVariable.Category ||
		*updateOpt.HCL != targetVariable.HCL ||
		*updateOpt.Sensitive != targetVariable.Sensitive {
		return false
	}

	return true
}

func confirm(in io.Reader) (bool, error) {
	r := bufio.NewReader(in)

	input, err := r.ReadString('\n')
	if err != nil {
		return false, err
	}

	switch strings.ToLower(strings.TrimRight(input, "\n")) {
	case "y", "yes":
		return true, nil
	}

	return false, nil
}
