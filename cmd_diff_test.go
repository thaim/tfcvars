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
	mockVariableSets := mocks.NewMockVariableSets(ctrl)
	mockVariableSetVariables := mocks.NewMockVariableSetVariables(ctrl)

	cases := []struct {
		name        string
		workspaceId string
		diffOpt     *DiffOption
		setClient   func(*mocks.MockVariables, *mocks.MockVariableSets, *mocks.MockVariableSetVariables)
		expect      string
		wantErr     bool
		expectErr   string
	}{
		{
			name:        "show no diff with empty local variable and empty variable list",
			workspaceId: "w-test-no-vars-workspace",
			diffOpt:     &DiffOption{},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
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
			name:        "show no diff with same variables",
			workspaceId: "w-test-single-variable-workspace",
			diffOpt:     &DiffOption{varFile: "testdata/terraform.tfvars"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
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
			name:        "show no diff compare with tfvars include comments",
			workspaceId: "w-test-vars-with-comment-workspace",
			diffOpt:     &DiffOption{varFile: "testdata/withcomment.tfvars"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-vars-with-comment-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								Key:   "environment",
								Value: "test",
							},
							{
								Key:   "port",
								Value: "3000",
								HCL:   true,
							},
							{
								Key:   "terraform",
								Value: "true",
								HCL:   true,
							},
						},
					}, nil).
					AnyTimes()
			},
			expect: "",
		},
		{
			name:        "show no diff with mutiple variables",
			workspaceId: "w-test-multiple-variables-workspace",
			diffOpt:     &DiffOption{varFile: "testdata/mixedtypes.tfvars"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
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
								HCL:   true,
							},
							{
								Key:   "terraform",
								Value: "true",
								HCL:   true,
							},
							{
								Key:   "availability_zones",
								Value: `["ap-northeast-1a", "ap-northeast-1c", "ap-northeast-1d"]`,
								HCL:   true,
							},
							{
								Key:   "tags",
								Value: `{repo = "github.com/thaim/tfcvars"}`,
								HCL:   true,
							},
						},
					}, nil).
					AnyTimes()
			},
			expect: "",
		},
		{
			name:        "show diff with different key",
			workspaceId: "w-test-single-variable-different-key-workspace",
			diffOpt:     &DiffOption{varFile: "testdata/terraform.tfvars"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
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
			name:        "show no diff include env category with include-env disabled",
			workspaceId: "w-test-variable-include-env-not-show-diff-workspace",
			diffOpt:     &DiffOption{varFile: "testdata/terraform.tfvars", includeEnv: false},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-variable-include-env-not-show-diff-workspace", nil).
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
			name:        "show diff include consecutive multiple insert lines and delete lines",
			workspaceId: "w-test-variable-consecutive-multiple-lines-workspace",
			diffOpt:     &DiffOption{varFile: "testdata/mixedtypes.tfvars"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-variable-consecutive-multiple-lines-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								Key:   "delete1",
								Value: "value1",
							},
							{
								Key:   "delete2",
								Value: "value2",
							},
							{
								Key:   "environment",
								Value: "test",
							},
							{
								Key:   "port",
								Value: "3000",
								HCL:   true,
							},
							{
								Key:   "terraform",
								Value: "true",
								HCL:   true,
							},
							{
								Key:   "availability_zones",
								Value: "[\"ap-northeast-1a\", \"ap-northeast-1c\", \"ap-northeast-1d\"]",
								HCL:   true,
							},
						},
					}, nil).
					AnyTimes()
			},
			expect: `  environment        = "test"
  port               = "3000"
  terraform          = "true"
  availability_zones = ["ap-northeast-1a", "ap-northeast-1c", "ap-northeast-1d"]
- delete1            = "value1"
- delete2            = "value2"
+ tags = {
+   repo = "github.com/thaim/tfcvars"
+ }
`,
		},
		{
			name:        "show diff include env category with include-env enabled",
			workspaceId: "w-test-variable-include-env-show-diff-workspace",
			diffOpt:     &DiffOption{varFile: "testdata/terraform.tfvars", includeEnv: true},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
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
			expect: `  environment = "development"
- ENV         = "TEST"
`,
		},
		{
			name:        "ignore variable set if include-variable-set not specified",
			workspaceId: "w-test-ignore-variable-set-workspace",
			diffOpt:     &DiffOption{varFile: "testdata/terraform.tfvars"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
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
			name:        "show diff in variable set if include-variable-set specified",
			workspaceId: "w-test-variable-set-workspace",
			diffOpt:     &DiffOption{varFile: "testdata/terraform.tfvars", includeVariableSet: true},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
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
				mvs.EXPECT().
					ListForWorkspace(context.TODO(), "w-test-variable-set-workspace", nil).
					Return(&tfe.VariableSetList{
						Items: []*tfe.VariableSet{
							{
								ID: "varset2",
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
			name:        "return error if not able to list variable list",
			workspaceId: "w-test-access-error",
			diffOpt:     &DiffOption{},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-access-error", nil).
					Return(nil, tfe.ErrInvalidWorkspaceID)
			},
			expect:    "",
			wantErr:   true,
			expectErr: "invalid value for workspace ID",
		},
		{
			name:        "return error if failed to readvars file",
			workspaceId: "w-test-no-vars-workspace",
			diffOpt:     &DiffOption{varFile: "testdata/invalid.tfvars"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
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
			name:        "return error if not allowed to list variable set",
			workspaceId: "w-test-not-allowed-to-list-variable-set-variables",
			diffOpt:     &DiffOption{varFile: "testdata/terraform.tfvars", includeVariableSet: true},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-not-allowed-to-list-variable-set-variables", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{},
					}, nil).
					AnyTimes()
				mvs.EXPECT().
					ListForWorkspace(context.TODO(), "w-test-not-allowed-to-list-variable-set-variables", nil).
					Return(nil, errors.New("failed to list variable set in workspace")).
					AnyTimes()
			},
			expect:    "",
			wantErr:   true,
			expectErr: "failed to list variable set in workspace",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			var buf bytes.Buffer
			tt.setClient(mockVariables, mockVariableSets, mockVariableSetVariables)

			err := diff(ctx, tt.workspaceId, mockVariables, mockVariableSets, mockVariableSetVariables, tt.diffOpt, &buf)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expect '%s' error, got no error", tt.expectErr)
				} else if !strings.Contains(err.Error(), tt.expectErr) {
					t.Errorf("expect '%s' error, got '%s'", tt.expectErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("expect no error, got error: '%v'", err)
			}
			if bufString := replaceNBSPWithSpace(buf.String()); bufString != tt.expect {
				t.Errorf("expect: '%s', got: '%s'", tt.expect, bufString)
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
