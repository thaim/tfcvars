package main

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/rs/zerolog/log"
)

func updateTerraformCloudWorkspace(organization string, workspaceName string) (string, string) {
	p := hclparse.NewParser()
	file, diags := p.ParseHCLFile("main.tf")
	if diags.HasErrors() {
		log.Warn().Msgf("failed to parse main.tf to retrive organization and workspace %s", diags.Error())
		return organization, workspaceName
	}

	fmt.Printf("%s\n", file.Bytes)

	return organization, workspaceName
}
