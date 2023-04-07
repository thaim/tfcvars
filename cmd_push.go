package main

import (
	"context"
	"errors"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type PushOption struct {
	varFile string
}

func NewPushOption(c *cli.Context) *PushOption {
	var opt = &PushOption{}
	opt.varFile = c.String("var-file")

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
	vars := &tfe.VariableList{}
	p := hclparse.NewParser()
	file, diags := p.ParseHCLFile(pushOpt.varFile)
	if diags.HasErrors() {
		return errors.New(diags.Error())
	}
	attrs, _ := file.Body.JustAttributes()
	for attrKey, attrValue := range attrs {
		val, _ := attrValue.Expr.Value(nil)

		vars.Items = append(vars.Items, &tfe.Variable{
			Key:   attrKey,
			Value: String(val),
		})
	}

	return push(ctx, w.ID, tfeClient.Variables, pushOpt, vars)
}

func push(ctx context.Context, workspaceId string, tfeVariables tfe.Variables, pushOpt *PushOption, vars *tfe.VariableList) error {
	previousVars, err := tfeVariables.List(ctx, workspaceId, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to list variables")
		return err
	}

	countUpdate := 0
	countCreate := 0

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
				tfeVariables.Update(ctx, workspaceId, targetVar.ID, updateOpt)
				pushed = true
				countUpdate++
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
			countCreate++
		}
	}
	log.Info().Msgf("create: %d, update: %d, delete: 0", countCreate, countUpdate)

	return nil
}
