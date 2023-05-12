package main

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/mocks"
	"github.com/urfave/cli/v2"
)

func TestCmdDiff(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockVariables := mocks.NewMockVariables(ctrl)
	mockVariableSetVariables := mocks.NewMockVariableSetVariables(ctrl)

	cases := []struct {
		name           string
		workspaceId    string
		variableSetIds []string
		diffOpt        *DiffOption
		setClient      func(*mocks.MockVariables, *mocks.MockVariableSetVariables)
		expect         string
		wantErr        bool
		expectErr      string
	}{
		{
			name:           "show no diff with empty local variable and empty variable list",
			workspaceId:    "w-test-no-vars-workspace",
			variableSetIds: []string{},
			diffOpt:        &DiffOption{},
			setClient: func(mc *mocks.MockVariables, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-no-vars-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{},
					}, nil).
					AnyTimes()
			},
			expect: "",
		},
		{
			name:           "show no diff with same variables",
			workspaceId:    "w-test-single-variable-workspace",
			variableSetIds: []string{},
			diffOpt:        &DiffOption{varFile: "testdata/terraform.tfvars"},
			setClient: func(mc *mocks.MockVariables, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-single-variable-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								Key:         "environment",
								Value:       "development",
								Description: "env",
							},
						},
					}, nil).
					AnyTimes()
			},
			expect: "",
		},
		{
			name:           "show no diff with mutiple variables",
			workspaceId:    "w-test-multiple-variables-workspace",
			variableSetIds: []string{},
			diffOpt:        &DiffOption{varFile: "testdata/mixedtypes.tfvars"},
			setClient: func(mc *mocks.MockVariables, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-multiple-variables-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								Key:   "environment",
								Value: "test",
							},
							{
								Key:   "port",
								Value: "3000",
							},
							{
								Key:   "terraform",
								Value: "true",
							},
							{
								Key:   "availability_zones",
								Value: `["ap-northeast-1a", "ap-northeast-1c", "ap-northeast-1d"]`,
							},
							{
								Key:   "tags",
								Value: `{reop = "github.com/thaim/tfcvars"}`,
							},
						},
					}, nil).
					AnyTimes()
			},
			expect: "",
		},
		{
			name:           "show diff with dirrerent key",
			workspaceId:    "w-test-single-variable-different-key-workspace",
			variableSetIds: []string{},
			diffOpt:        &DiffOption{varFile: "testdata/terraform.tfvars"},
			setClient: func(mc *mocks.MockVariables, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-single-variable-different-key-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								Key:         "env",
								Value:       "development",
								Description: "env",
							},
						},
					}, nil).
					AnyTimes()
			},
			expect: `- env = "development"
+ environment = "development"
`,
		},
		{
			name:           "show no diff include env category with include-env disabled",
			workspaceId:    "w-test-variable-include-env-show-diff-workspace",
			variableSetIds: []string{},
			diffOpt:        &DiffOption{varFile: "testdata/terraform.tfvars", includeEnv: true},
			setClient: func(mc *mocks.MockVariables, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-variable-include-env-show-diff-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								Key:         "environment",
								Value:       "development",
								Description: "env",
							},
							{
								Key:      "ENV",
								Value:    "TEST",
								Category: tfe.CategoryEnv,
							},
						},
					}, nil).
					AnyTimes()
			},
			expect: `- ENV         = "TEST"
`,
		},
		{
			name:           "show diff include env category with include-env enabled",
			workspaceId:    "w-test-variable-include-env-show-diff-workspace",
			variableSetIds: []string{},
			diffOpt:        &DiffOption{varFile: "testdata/terraform.tfvars", includeEnv: true},
			setClient: func(mc *mocks.MockVariables, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-variable-include-env-show-diff-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								Key:         "environment",
								Value:       "development",
								Description: "env",
							},
							{
								Key:      "ENV",
								Value:    "TEST",
								Category: tfe.CategoryEnv,
							},
						},
					}, nil).
					AnyTimes()
			},
			expect: "",
		},
		{
			name:           "ignore variable set if include-variable-set not specified",
			workspaceId:    "w-test-ignore-variable-set-workspace",
			variableSetIds: []string{"varset1"},
			diffOpt:        &DiffOption{varFile: "testdata/terraform.tfvars"},
			setClient: func(mc *mocks.MockVariables, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-ignore-variable-set-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								Key:         "environment",
								Value:       "development",
								Description: "env",
							},
							{
								Key:      "ENV",
								Value:    "TEST",
								Category: tfe.CategoryEnv,
							},
						},
					}, nil).
					AnyTimes()
				mvsv.EXPECT().
					List(context.TODO(), "varset1", nil).
					Return(&tfe.VariableSetVariableList{
						Items: []*tfe.VariableSetVariable{
							{
								Key:   "param1",
								Value: "value1",
							},
							{
								Key:      "SET_ENV",
								Value:    "TEST",
								Category: tfe.CategoryEnv,
							},
						},
					}, nil).
					AnyTimes()
			},
			expect: "",
		},
		{
			name:           "show diff invariable set if include-variable-set specified",
			workspaceId:    "w-test-variable-set-workspace",
			variableSetIds: []string{"varset2"},
			diffOpt:        &DiffOption{varFile: "testdata/terraform.tfvars", includeVariableSet: true},
			setClient: func(mc *mocks.MockVariables, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-variable-set-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								Key:         "environment",
								Value:       "development",
								Description: "env",
							},
							{
								Key:      "ENV",
								Value:    "TEST",
								Category: tfe.CategoryEnv,
							},
						},
					}, nil).
					AnyTimes()
				mvsv.EXPECT().
					List(context.TODO(), "varset2", nil).
					Return(&tfe.VariableSetVariableList{
						Items: []*tfe.VariableSetVariable{
							{
								Key:   "param1",
								Value: "value1",
							},
							{
								Key:      "SET_ENV",
								Value:    "TEST",
								Category: tfe.CategoryEnv,
							},
						},
					}, nil).
					AnyTimes()
			},
			expect: `  environment = "development"
- param1      = "value1"
`,
		},
		{
			name:           "return error if not able to list variable list",
			workspaceId:    "w-test-access-error",
			variableSetIds: []string{},
			diffOpt:        &DiffOption{},
			setClient: func(mc *mocks.MockVariables, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-access-error", nil).
					Return(nil, tfe.ErrInvalidWorkspaceID)
			},
			expect:    "",
			wantErr:   true,
			expectErr: "invalid value for workspace ID",
		},
		{
			name:           "return error if failed to readvars file",
			workspaceId:    "w-test-no-vars-workspace",
			variableSetIds: []string{},
			diffOpt:        &DiffOption{varFile: "testdata/invalid.tfvars"},
			setClient: func(mc *mocks.MockVariables, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-no-vars-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{},
					}, nil).
					AnyTimes()
			},
			expect:    "",
			wantErr:   true,
			expectErr: "Argument or block definition required",
		},
		{
			name:           "return error if undefined variable set id specified",
			workspaceId:    "w-test-undefined-variable-set-workspace",
			variableSetIds: []string{"undefined-variable-set"},
			diffOpt:        &DiffOption{varFile: "testdata/terraform.tfvars", includeVariableSet: true},
			setClient: func(mc *mocks.MockVariables, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-undefined-variable-set-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{},
					}, nil).
					AnyTimes()
				mvsv.EXPECT().
					List(context.TODO(), "undefined-variable-set", nil).
					Return(nil, errors.New("resource not found")).
					AnyTimes()
			},
			expect:    "",
			wantErr:   true,
			expectErr: "resource not found",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			var buf bytes.Buffer
			tt.setClient(mockVariables, mockVariableSetVariables)

			err := diff(ctx, tt.workspaceId, mockVariables, tt.variableSetIds, mockVariableSetVariables, tt.diffOpt, &buf)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expect '%s' error, got no error", tt.expectErr)
				} else if !strings.Contains(err.Error(), tt.expectErr) {
					t.Errorf("expect '%s' error, got %s", tt.expectErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("expect no error, got error: %v", err)
			}
			if bufString := replaceNBSPWithSpace(buf.String()); !strings.Contains(bufString, tt.expect) {
				t.Errorf("expect: %s, got: %s", tt.expect, bufString)
			}
		})
	}
}

func TestNewDiffOption(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		expect *DiffOption
	}{
		{
			name: "default value",
			args: []string{},
			expect: &DiffOption{
				varFile: "terraform.tfvars",
			},
		},
		{
			name: "default value",
			args: []string{"--var-file", "testdata/terraform.tfvars"},
			expect: &DiffOption{
				varFile: "testdata/terraform.tfvars",
			},
		},
		{
			name: "enable include env option",
			args: []string{"--include-env"},
			expect: &DiffOption{
				varFile:    "terraform.tfvars",
				includeEnv: true,
			},
		},
		{
			name: "enable include variable set option",
			args: []string{"--include-variable-set"},
			expect: &DiffOption{
				varFile:            "terraform.tfvars",
				includeVariableSet: true,
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			app := cli.NewApp()
			set := flagSet(diffFlags())
			set.Parse(tt.args)
			ctx := cli.NewContext(app, set, nil)

			sut := NewDiffOption(ctx)

			if !reflect.DeepEqual(tt.expect, sut) {
				t.Errorf("expect '%v', got '%v'", tt.expect, sut)
			}
		})
	}
}

func replaceNBSPWithSpace(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\u00A0' {
			return ' '
		}
		return r
	}, s)
}
