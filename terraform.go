package main

import (
	"context"
	"os"
	"path/filepath"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

func updateTerraformCloudWorkspace(organization string, workspaceName string, workdir string) (string, string) {
	srcByte, err := os.ReadFile(filepath.Join(workdir, ".terraform/terraform.tfstate"))
	if err != nil {
		log.Error().Err(err).Msg("cannot open tfstate file")
		return organization, workspaceName
	}
	src := string(srcByte)

	backendType := gjson.Get(src, "backend.type").String()
	if backendType != "remote" {
		log.Warn().Msg("backend for this workspace is not a remote")
		return organization, workspaceName
	}

	organization = gjson.Get(src, "backend.config.organization").String()
	nameGJson := gjson.Get(src, "backend.config.workspaces.name")
	if nameGJson.Type == gjson.String {
		workspaceName = nameGJson.String()
	} else if nameGJson.Type == gjson.Null {
		env, _ := os.ReadFile(filepath.Join(workdir, ".terraform/environment"))
		workspaceName = gjson.Get(src, "backend.config.workspaces.prefix").String() + string(env)
	}

	log.Debug().Msgf("retrive from tfstate: org=%s workspace=%s", organization, workspaceName)
	return organization, workspaceName
}

func listVariableSetVariables(ctx context.Context, workspaceId string, VariableSets tfe.VariableSets, VariableSetVariables tfe.VariableSetVariables) ([]*tfe.Variable, error) {
	variables := make([]*tfe.Variable, 0)
	s, err := VariableSets.ListForWorkspace(ctx, workspaceId, nil)
	if err != nil {
		log.Error().Err(err).Msgf("failed to list variable set in workspace %s", workspaceId)
		return nil, err
	}

	for setIndex := range s.Items {
		variableList, err := VariableSetVariables.List(ctx, s.Items[setIndex].ID, nil)
		if err != nil {
			log.Error().Err(err).Msgf("failed to list VariableSetVariables ID: %s", s.Items[setIndex].ID)
			return nil, err
		}

		for variableListIndex := range variableList.Items {
			variableSetVariable := variableList.Items[variableListIndex]
			variable := &tfe.Variable{}

			variable.ID = variableSetVariable.ID
			variable.Key = variableSetVariable.Key
			variable.Value = variableSetVariable.Value
			variable.Description = variableSetVariable.Description
			variable.Category = variableSetVariable.Category
			variable.HCL = variableSetVariable.HCL
			variable.Sensitive = variableSetVariable.Sensitive

			variables = append(variables, variable)
		}
	}

	return variables, nil
}
