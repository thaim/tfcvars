package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/mocks"
)

func TestCmdPull(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockVariables := mocks.NewMockVariables(ctrl)

	cases := []struct {
		name        string
		workspaceId string
		setClient   func(*mocks.MockVariables)
		expect      string
		wantErr     bool
		expectErr   string
	}{
		{
			name:        "pull empty variable",
			workspaceId: "w-test-no-vars-workspace",
			setClient:   func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-no-vars-workspace", nil).
					Return(&tfe.VariableList{
						Items: nil,
					}, nil).
					AnyTimes()
			},
			expect:      "",
			wantErr:     false,
			expectErr:   "",
		},
		{
			name:        "pull single variable",
			workspaceId: "w-test-single-variable-workspace",
			setClient:   func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-single-variable-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								Key:         "var1",
								Value:       "value1",
								Description: "description1",
							},
						},
					}, nil).
					AnyTimes()
			},
			expect:      "var1 = \"value1\"\n",
			wantErr:     false,
			expectErr:   "",
		},
		{
			name:        "pull multiple variables",
			workspaceId: "w-test-multiple-variables-workspace",
			setClient:   func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-multiple-variables-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								Key:   "var1",
								Value: "value1",
							},
							{
								Key:   "var2",
								Value: "value2",
							},
						},
					}, nil).
					AnyTimes()
			},
			expect:      "var1 = \"value1\"\nvar2 = \"value2\"\n",
			wantErr:     false,
			expectErr:   "",
		},
		{
			name:        "pull sensitive variable",
			workspaceId: "w-test-sensitive-variable-workspace",
			setClient:   func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-sensitive-variable-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								Key:         "var1",
								Value:       "",
								Description: "sensitive",
								Sensitive:   true,
							},
							{
								Key:         "var2",
								Value:       "",
								Description: "sensitive environment",
								Sensitive:   true,
								Category:    tfe.CategoryEnv,
							},
						},
					}, nil).
					AnyTimes()
			},
			expect:      "// var1 = \"***\"\n",
			wantErr:     false,
			expectErr:   "",
		},
		{
			name:        "treat multiple variable types",
			workspaceId: "w-test-linclude-multiple-variable-types-workspace",
			setClient:   func(mc *mocks.MockVariables) {
				// test for Types
				// https://developer.hashicorp.com/terraform/cloud-docs/workspaces/variables#types
				mc.EXPECT().
					List(context.TODO(), "w-test-linclude-multiple-variable-types-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								Key:         "var1",
								Value:       "value1",
								Description: "Terraform Variables",
								Category:    tfe.CategoryTerraform,
							},
							{
								Key:         "var2",
								Value:       "value2",
								Description: "Environment Variables",
								Category:    tfe.CategoryEnv,
							},
							// TODO support policy-set
						},
					}, nil).
					AnyTimes()
			},
			expect:      "var1 = \"value1\"\n",
			wantErr:     false,
			expectErr:   "",
		},
		{
			name:        "return error if failed to access terraform cloud",
			workspaceId: "w-test-access-error",
			setClient:   func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-access-error", nil).
					Return(nil, tfe.ErrInvalidWorkspaceID)
			},
			expect:      "",
			wantErr:     true,
			expectErr:   "invalid value for workspace ID",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			pullOpt := &PullOption{}
			tt.setClient(mockVariables)
			var buf bytes.Buffer

			err := pull(ctx, tt.workspaceId, mockVariables, pullOpt, &buf)

			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.expectErr) {
					t.Errorf("expect %s error, got %s", tt.expectErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("expect no error, got error: %v", err)
			}
			if bufString := buf.String(); bufString != tt.expect {
				t.Errorf("expect '%s', got '%s'", tt.expect, buf.String())
			}
		})
	}
}
