package main

import (
	"os"
	"path/filepath"

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
