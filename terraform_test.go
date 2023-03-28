package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateTerraformCloudWorkspace(t *testing.T) {
	cases := []struct {
		name                  string
		env                   string
		tfstate               string
		defaultOrganization   string
		defaultWorkspaceName  string
		expectedOrganization  string
		expectedWorkspaceName string
	}{
		{
			name:                  "tfstate not found",
			defaultOrganization:   "test-org",
			defaultWorkspaceName:  "test-ws",
			expectedOrganization:  "test-org",
			expectedWorkspaceName: "test-ws",
		},
		{
			name:                  "local backend",
			tfstate:               `{"backend": {"type": "local"}}`,
			defaultOrganization:   "test-org",
			defaultWorkspaceName:  "test-ws",
			expectedOrganization:  "test-org",
			expectedWorkspaceName: "test-ws",
		},
		{
			name:                  "workspace with name",
			tfstate:               `{"backend": {"type": "remote", "config": {"organization": "test-organization", "workspaces": {"name": "test-workspace"}}}}`,
			defaultOrganization:   "test-org",
			defaultWorkspaceName:  "test-ws",
			expectedOrganization:  "test-organization",
			expectedWorkspaceName: "test-workspace",
		},
		{
			name:                  "workspace with prefix",
			tfstate:               `{"backend": {"type": "remote", "config": {"organization": "test-organization", "workspaces": {"name": null, "prefix": "test-workspace-"}}}}`,
			env:                   "test",
			defaultOrganization:   "test-org",
			defaultWorkspaceName:  "test-ws",
			expectedOrganization:  "test-organization",
			expectedWorkspaceName: "test-workspace-test",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.tfstate != "" || tt.env != "" {
				os.Mkdir(filepath.Join(dir, ".terraform"), 0755)
				os.WriteFile(filepath.Join(dir, ".terraform", "terraform.tfstate"), []byte(tt.tfstate), 0644)
				os.WriteFile(filepath.Join(dir, ".terraform", "environment"), []byte(tt.env), 0644)
			}

			organization, workspace := updateTerraformCloudWorkspace(tt.defaultOrganization, tt.defaultWorkspaceName, dir)

			if organization != tt.expectedOrganization {
				t.Errorf("expect %s, got %s", tt.expectedOrganization, organization)
			}
			if workspace != tt.expectedWorkspaceName {
				t.Errorf("expect %s, got %s", tt.expectedWorkspaceName, workspace)
			}
		})
	}
}
