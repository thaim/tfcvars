package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/mocks"
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
			name:                  "cloud backend",
			tfstate:               `{"backend": {"type": "cloud", "config": {"organization": "test-organization", "workspaces": {"name": "test-workspace"}}}}`,
			defaultOrganization:   "test-org",
			defaultWorkspaceName:  "test-ws",
			expectedOrganization:  "test-organization",
			expectedWorkspaceName: "test-workspace",
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

func TestListVariableSetVariables(t *testing.T) {
	ctrl := gomock.NewController(t)
	VariableSets := mocks.NewMockVariableSets(ctrl)
	VariableSetVariables := mocks.NewMockVariableSetVariables(ctrl)

	cases := []struct {
		name              string
		workspaceId       string
		setClient         func(*mocks.MockVariableSets, *mocks.MockVariableSetVariables)
		expectedVariables []*tfe.Variable
		wantErr           bool
		expectErr         string
	}{
		{
			name:        "list-variable-from-single-variable-set",
			workspaceId: "w-test-single-variable-set",
			setClient: func(vs *mocks.MockVariableSets, vsv *mocks.MockVariableSetVariables) {
				vs.EXPECT().
					ListForWorkspace(context.TODO(), "w-test-single-variable-set", nil).
					Return(&tfe.VariableSetList{
						Items: []*tfe.VariableSet{
							{
								ID: "single-variable-set",
							},
						},
					}, nil).
					AnyTimes()
				vsv.EXPECT().
					List(context.TODO(), "single-variable-set", nil).
					Return(&tfe.VariableSetVariableList{
						Items: []*tfe.VariableSetVariable{
							{
								Key:   "environment",
								Value: "test",
							},
						},
					}, nil).
					AnyTimes()
			},
			expectedVariables: []*tfe.Variable{
				{
					Key:   "environment",
					Value: "test",
				},
			},
			wantErr: false,
		},
		{
			name:        "list-variables-from-single-variable-set",
			workspaceId: "w-test-multiple-variables-from-single-variable-set",
			setClient: func(vs *mocks.MockVariableSets, vsv *mocks.MockVariableSetVariables) {
				vs.EXPECT().
					ListForWorkspace(context.TODO(), "w-test-multiple-variables-from-single-variable-set", nil).
					Return(&tfe.VariableSetList{
						Items: []*tfe.VariableSet{
							{
								ID: "multiple-variable-set",
							},
						},
					}, nil).
					AnyTimes()
				vsv.EXPECT().
					List(context.TODO(), "multiple-variable-set", nil).
					Return(&tfe.VariableSetVariableList{
						Items: []*tfe.VariableSetVariable{
							{
								Key:   "environment",
								Value: "test",
							},
							{
								Key:   "terraform",
								Value: "true",
							},
						},
					}, nil).
					AnyTimes()
			},
			expectedVariables: []*tfe.Variable{
				{
					Key:   "environment",
					Value: "test",
				},
				{
					Key:   "terraform",
					Value: "true",
				},
			},
			wantErr: false,
		},
		{
			name:        "list-variable-from-multiple-variable-set",
			workspaceId: "w-test-single-variable-from-multiple-variable-set",
			setClient: func(vs *mocks.MockVariableSets, vsv *mocks.MockVariableSetVariables) {
				vs.EXPECT().
					ListForWorkspace(context.TODO(), "w-test-single-variable-from-multiple-variable-set", nil).
					Return(&tfe.VariableSetList{
						Items: []*tfe.VariableSet{
							{
								ID: "single-variable-set-first",
							},
							{
								ID: "single-variable-set-second",
							},
						},
					}, nil).
					AnyTimes()
				vsv.EXPECT().
					List(context.TODO(), "single-variable-set-first", nil).
					Return(&tfe.VariableSetVariableList{
						Items: []*tfe.VariableSetVariable{
							{
								Key:   "environment",
								Value: "test",
							},
						},
					}, nil).
					AnyTimes()
				vsv.EXPECT().
					List(context.TODO(), "single-variable-set-second", nil).
					Return(&tfe.VariableSetVariableList{
						Items: []*tfe.VariableSetVariable{
							{
								Key:   "param1",
								Value: "value1",
							},
						},
					}, nil).
					AnyTimes()
			},
			expectedVariables: []*tfe.Variable{
				{
					Key:   "environment",
					Value: "test",
				},
				{
					Key:   "param1",
					Value: "value1",
				},
			},
			wantErr: false,
		},
		{
			name:        "list-variables-from-multiple-variable-set",
			workspaceId: "w-test-multiple-variable-from-multiple-variable-set",
			setClient: func(vs *mocks.MockVariableSets, vsv *mocks.MockVariableSetVariables) {
				vs.EXPECT().
					ListForWorkspace(context.TODO(), "w-test-multiple-variable-from-multiple-variable-set", nil).
					Return(&tfe.VariableSetList{
						Items: []*tfe.VariableSet{
							{
								ID: "multiple-variable-set-first",
							},
							{
								ID: "multiple-variable-set-second",
							},
						},
					}, nil).
					AnyTimes()
				vsv.EXPECT().
					List(context.TODO(), "multiple-variable-set-first", nil).
					Return(&tfe.VariableSetVariableList{
						Items: []*tfe.VariableSetVariable{
							{
								Key:   "set1var1",
								Value: "value1",
							},
							{
								Key:   "set1var2",
								Value: "value2",
							},
						},
					}, nil).
					AnyTimes()
				vsv.EXPECT().
					List(context.TODO(), "multiple-variable-set-second", nil).
					Return(&tfe.VariableSetVariableList{
						Items: []*tfe.VariableSetVariable{
							{
								Key:   "set2var1",
								Value: "value3",
							},
							{
								Key:   "set2var2",
								Value: "value4",
							},
						},
					}, nil).
					AnyTimes()
			},
			expectedVariables: []*tfe.Variable{
				{
					Key:   "set1var1",
					Value: "value1",
				},
				{
					Key:   "set1var2",
					Value: "value2",
				},
				{
					Key:   "set2var1",
					Value: "value3",
				},
				{
					Key:   "set2var2",
					Value: "value4",
				},
			},
			wantErr: false,
		},
		{
			name:        "error-list-variable-set",
			workspaceId: "w-test-error-list-variable-set",
			setClient: func(vs *mocks.MockVariableSets, vsv *mocks.MockVariableSetVariables) {
				vs.EXPECT().
					ListForWorkspace(context.TODO(), "w-test-error-list-variable-set", nil).
					Return(nil, errors.New("failed to list variable set in workspace")).
					AnyTimes()
			},
			wantErr:   true,
			expectErr: "failed to list variable set in workspace",
		},
		{
			name:        "error-list-variable-set-variables",
			workspaceId: "w-test-error-list-variable-set-variables",
			setClient: func(vs *mocks.MockVariableSets, vsv *mocks.MockVariableSetVariables) {
				vs.EXPECT().
					ListForWorkspace(context.TODO(), "w-test-error-list-variable-set-variables", nil).
					Return(&tfe.VariableSetList{
						Items: []*tfe.VariableSet{
							{
								ID: "error-list-variable-set-variables",
							},
						},
					}, nil).
					AnyTimes()
				vsv.EXPECT().
					List(context.TODO(), "error-list-variable-set-variables", nil).
					Return(nil, errors.New("failed to list VariableSetVariables")).
					AnyTimes()
			},
			wantErr:   true,
			expectErr: "failed to list VariableSetVariables",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.setClient(VariableSets, VariableSetVariables)

			variables, err := listVariableSetVariables(context.TODO(), tt.workspaceId, VariableSets, VariableSetVariables)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expect '%s' error, got no error", tt.expectErr)
				} else if !strings.Contains(err.Error(), tt.expectErr) {
					t.Errorf("expect %s error, got %s", tt.expectErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("expect no error, got error: %v", err)
			}
			if len(variables) != len(tt.expectedVariables) {
				t.Errorf("expected variable length: %d, got %d", len(tt.expectedVariables), len(variables))
			}
			for i := range variables {
				if variables[i].Key != tt.expectedVariables[i].Key {
					t.Errorf("expected variable key '%s', got '%s'", tt.expectedVariables[i].Key, variables[i].Key)
				}
				if variables[i].Value != tt.expectedVariables[i].Value {
					t.Errorf("expected variable value '%s', got '%s'", tt.expectedVariables[i].Value, variables[i].Value)
				}
			}
		})
	}
}
