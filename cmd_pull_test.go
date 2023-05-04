package main

import (
	"bytes"
	"context"
	"flag"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/mocks"
	"github.com/urfave/cli/v2"
)

func TestCmdPull(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockVariables := mocks.NewMockVariables(ctrl)

	cases := []struct {
		name        string
		workspaceId string
		pullOpt     *PullOption
		setClient   func(*mocks.MockVariables)
		expect      string
		wantErr     bool
		expectErr   string
	}{
		{
			name:        "pull empty variable",
			workspaceId: "w-test-no-vars-workspace",
			pullOpt:     &PullOption{},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-no-vars-workspace", nil).
					Return(&tfe.VariableList{
						Items: nil,
					}, nil).
					AnyTimes()
			},
			expect:    "",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "pull single variable",
			workspaceId: "w-test-single-variable-workspace",
			pullOpt:     &PullOption{},
			setClient: func(mc *mocks.MockVariables) {
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
			expect:    "var1 = \"value1\"\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "pull multiple variables",
			workspaceId: "w-test-multiple-variables-workspace",
			pullOpt:     &PullOption{},
			setClient: func(mc *mocks.MockVariables) {
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
			expect:    "var1 = \"value1\"\nvar2 = \"value2\"\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "pull sensitive variable",
			workspaceId: "w-test-sensitive-variable-workspace",
			pullOpt:     &PullOption{},
			setClient: func(mc *mocks.MockVariables) {
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
			expect:    "// var1 = \"***\"\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "treat multiple variable types",
			workspaceId: "w-test-linclude-multiple-variable-types-workspace",
			pullOpt:     &PullOption{},
			setClient: func(mc *mocks.MockVariables) {
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
			expect:    "var1 = \"value1\"\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "return error if failed to access terraform cloud",
			workspaceId: "w-test-access-error",
			pullOpt:     &PullOption{},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-access-error", nil).
					Return(nil, tfe.ErrInvalidWorkspaceID)
			},
			expect:    "",
			wantErr:   true,
			expectErr: "invalid value for workspace ID",
		},
		{
			name:        "pull tuple string",
			workspaceId: "w-test-variable-tuple",
			pullOpt:     &PullOption{},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-variable-tuple", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								Key:   "var1",
								Value: `["ap-northeast-1a", "ap-northeast-1c", "ap-northeast-1d"]`,
								HCL:   true,
							},
						},
					}, nil).
					AnyTimes()
			},
			expect:    "var1 = [\"ap-northeast-1a\", \"ap-northeast-1c\", \"ap-northeast-1d\"]\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "pull variables include env with include-env option enabled",
			workspaceId: "w-test-variables-include-env-enabled-workspace",
			pullOpt:     &PullOption{includeEnv: true},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-variables-include-env-enabled-workspace", nil).
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
							{
								Key:      "var3",
								Value:    "value3",
								Category: tfe.CategoryEnv,
							},
						},
					}, nil).
					AnyTimes()
			},
			expect:    "var1 = \"value1\"\nvar2 = \"value2\"\nvar3 = \"value3\"\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "pull variables include env with include-env option disabled",
			workspaceId: "w-test-variables-include-env-disabled-workspace",
			pullOpt:     &PullOption{includeEnv: false},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-variables-include-env-disabled-workspace", nil).
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
							{
								Key:      "var3",
								Value:    "value3",
								Category: tfe.CategoryEnv,
							},
						},
					}, nil).
					AnyTimes()
			},
			expect:    "var1 = \"value1\"\nvar2 = \"value2\"\n",
			wantErr:   false,
			expectErr: "",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			tt.setClient(mockVariables)
			var buf bytes.Buffer

			err := pull(ctx, tt.workspaceId, mockVariables, tt.pullOpt, &buf)

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
			if bufString := buf.String(); bufString != tt.expect {
				t.Errorf("expect '%s', got '%s'", tt.expect, buf.String())
			}
		})
	}
}

func TestNewPullOption(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		expect *PullOption
	}{
		{
			name: "default value",
			args: []string{},
			expect: &PullOption{
				varFile:     "terraform.tfvars",
				overwrite:   false,
				prevVarfile: nil,
			},
		},
		{
			name: "custom var file",
			args: []string{"--var-file", "custom.tfvars"},
			expect: &PullOption{
				varFile:     "custom.tfvars",
				overwrite:   false,
				prevVarfile: nil,
			},
		},
		{
			name: "enable overwite option",
			args: []string{"--overwrite"},
			expect: &PullOption{
				varFile:     "terraform.tfvars",
				overwrite:   true,
				prevVarfile: nil,
			},
		},
		{
			name: "ignore overwite option if merge option specified",
			args: []string{"--overwrite", "--merge"},
			expect: &PullOption{
				varFile:     "terraform.tfvars",
				overwrite:   false,
				prevVarfile: nil,
			},
		},
		{
			name: "enable include env option",
			args: []string{"--include-env"},
			expect: &PullOption{
				varFile:    "terraform.tfvars",
				overwrite:  false,
				includeEnv: true,
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			app := cli.NewApp()
			set := flagSet(pullFlags())
			set.Parse(tt.args)
			ctx := cli.NewContext(app, set, nil)

			sut := NewPullOption(ctx)

			if !reflect.DeepEqual(tt.expect, sut) {
				t.Errorf("expect '%v', got '%v'", tt.expect, sut)
			}
		})
	}
}

func flagSet(flags []cli.Flag) *flag.FlagSet {
	set := flag.NewFlagSet("", flag.ContinueOnError)
	for _, f := range flags {
		f.Apply(set)
	}

	return set
}
