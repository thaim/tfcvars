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

func TestCmdShow(t *testing.T) {
	ctrl := gomock.NewController(t)
	Variables := mocks.NewMockVariables(ctrl)
	VariableSets := mocks.NewMockVariableSets(ctrl)
	VariableSetVariables := mocks.NewMockVariableSetVariables(ctrl)

	cases := []struct {
		name        string
		workspaceId string
		showOpt     *ShowOption
		setClient   func(*mocks.MockVariables, *mocks.MockVariableSets, *mocks.MockVariableSetVariables)
		expect      string
		wantErr     bool
		expectErr   string
	}{
		{
			name:        "show empty variable",
			workspaceId: "w-test-no-vars-workspace",
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-no-vars-workspace", nil).
					Return(&tfe.VariableList{
						Items: nil,
					}, nil).
					AnyTimes()
			},
			showOpt:   &ShowOption{format: "detail"},
			expect:    "",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "show single variable",
			workspaceId: "w-test-single-variable-workspace",
			showOpt:     &ShowOption{format: "detail"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
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
			expect:    "Key: var1\nValue: value1\nDescription: description1\nSensitive: false\n\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "show multiple variables",
			workspaceId: "w-test-multiple-variables-workspace",
			showOpt:     &ShowOption{format: "detail"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
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
			expect:    "Key: var1\nValue: value1\nDescription: \nSensitive: false\n\nKey: var2\nValue: value2\nDescription: \nSensitive: false\n\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "show sensitive variable",
			workspaceId: "w-test-sensitive-variable-workspace",
			showOpt:     &ShowOption{format: "detail"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
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
						},
					}, nil).
					AnyTimes()
			},
			expect:    "Key: var1\nValue: \nDescription: sensitive\nSensitive: true\n\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "show specified variable",
			workspaceId: "w-test-multiple-variables-filter-variable-workspace",
			showOpt:     &ShowOption{variableKey: "var2", format: "detail"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-multiple-variables-filter-variable-workspace", nil).
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
			expect:    "Key: var2\nValue: value2\nDescription: \nSensitive: false\n\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "show variables with tfvars format",
			workspaceId: "w-test-variables-tfvars-workspace",
			showOpt:     &ShowOption{format: "tfvars"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-variables-tfvars-workspace", nil).
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
								Key:   "var3",
								Value: "[\"value3-1\", \"value3-2\"]",
								HCL:   true,
							},
							{
								Key:       "var4",
								Sensitive: true,
							},
							{
								Key:      "var5",
								Value:    "val5",
								Category: tfe.CategoryEnv,
							},
						},
					}, nil).
					AnyTimes()
			},
			expect:    "var1 = \"value1\"\nvar2 = \"value2\"\nvar3 = [\"value3-1\", \"value3-2\"]\n// var4 = \"***\"\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "show variables with table format",
			workspaceId: "w-test-variables-table-workspace",
			showOpt:     &ShowOption{format: "table"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-variables-table-workspace", nil).
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
								Key:   "var3",
								Value: "[\"value3-1\", \"value3-2\"]",
								HCL:   true,
							},
							{
								Key:       "var4",
								Sensitive: true,
							},
							{
								Key:      "var5",
								Value:    "val5",
								Category: tfe.CategoryEnv,
							},
							{
								Key:         "var6",
								Value:       "val6",
								Description: "var6 description",
							},
						},
					}, nil).
					AnyTimes()
			},
			expect: `+------+--------------------------+-----------+------------------+
| KEY  |          VALUE           | SENSITIVE |   DESCRIPTION    |
+------+--------------------------+-----------+------------------+
| var1 | value1                   | false     |                  |
| var2 | value2                   | false     |                  |
| var3 | ["value3-1", "value3-2"] | false     |                  |
| var4 |                          | true      |                  |
| var6 | val6                     | false     | var6 description |
+------+--------------------------+-----------+------------------+
`,
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "show local variable",
			workspaceId: "",
			showOpt:     &ShowOption{local: true, varFile: "testdata/terraform.tfvars", format: "detail"},
			setClient:   func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {}, // do nothing
			expect:      "Key: environment\nValue: development\nDescription: \nSensitive: false\n\n",
			wantErr:     false,
			expectErr:   "",
		},
		{
			name:        "show local variable include HCL format",
			workspaceId: "",
			showOpt:     &ShowOption{local: true, varFile: "testdata/mixedtypes.tfvars", format: "tfvars"},
			setClient:   func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {}, // do nothing
			expect: `environment        = "test"
port               = "3000"
terraform          = "true"
availability_zones = ["ap-northeast-1a", "ap-northeast-1c", "ap-northeast-1d"]
tags = {
  repo = "github.com/thaim/tfcvars"
}
`,
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "show variable include env",
			workspaceId: "w-test-include-env-variable-workspace",
			showOpt:     &ShowOption{varFile: "testdata/terraform.tfvars", includeEnv: true, format: "detail"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-include-env-variable-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								Key:   "var1",
								Value: "value1",
							},
							{
								Key:      "var2",
								Value:    "value2",
								Category: tfe.CategoryEnv,
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
			expect:    "Key: var1\nValue: value1\nDescription: \nSensitive: false\n\nKey: var2\nValue: value2\nDescription: \nSensitive: false\n\nKey: var3\nValue: value3\nDescription: \nSensitive: false\n\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "ignore env variable without include env option",
			workspaceId: "w-test-ignore-env-variable-workspace",
			showOpt:     &ShowOption{varFile: "testdata/terraform.tfvars", includeEnv: false, format: "detail"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-ignore-env-variable-workspace", nil).
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
			expect:    "Key: var1\nValue: value1\nDescription: \nSensitive: false\n\nKey: var2\nValue: value2\nDescription: \nSensitive: false\n\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "show variable include variable set",
			workspaceId: "w-test-include-variable-set-variables-workspace",
			showOpt:     &ShowOption{varFile: "testdata/terraform.tfvars", includeVariableSet: true, format: "detail"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-include-variable-set-variables-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								Key:   "var1",
								Value: "value1",
							},
							{
								Key:      "var2",
								Value:    "value2",
								Category: tfe.CategoryEnv,
							},
						},
					}, nil).
					AnyTimes()
				mvs.EXPECT().
					ListForWorkspace(context.TODO(), "w-test-include-variable-set-variables-workspace", nil).
					Return(&tfe.VariableSetList{
						Items: []*tfe.VariableSet{
							{
								ID: "variable-set-include-variable-set-variables",
							},
						},
					}, nil).
					AnyTimes()
				mvsv.EXPECT().
					List(context.TODO(), "variable-set-include-variable-set-variables", nil).
					Return(&tfe.VariableSetVariableList{
						Items: []*tfe.VariableSetVariable{
							{
								Key:   "var3",
								Value: "value3",
							},
							{
								Key:      "var4",
								Value:    "value4",
								Category: tfe.CategoryEnv,
							},
						},
					}, nil).
					AnyTimes()
			},
			expect:    "Key: var1\nValue: value1\nDescription: \nSensitive: false\n\nKey: var3\nValue: value3\nDescription: \nSensitive: false\n\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "ignore env variable without include env option",
			workspaceId: "w-test-ignore-env-variable-workspace",
			showOpt:     &ShowOption{varFile: "testdata/terraform.tfvars", includeEnv: false, format: "detail"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-ignore-env-variable-workspace", nil).
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
			expect:    "Key: var1\nValue: value1\nDescription: \nSensitive: false\n\nKey: var2\nValue: value2\nDescription: \nSensitive: false\n\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "return HCL parse error",
			workspaceId: "not-used",
			showOpt:     &ShowOption{local: true, varFile: "testdata/invalid.tfvars", format: "detail"},
			setClient:   func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {}, // do nothing
			wantErr:     true,
			expectErr:   "Argument or block definition required",
		},
		{
			name:        "not allowed to list variables",
			workspaceId: "w-test-not-allowed-to-list-variables",
			showOpt:     &ShowOption{format: "detail"},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-not-allowed-to-list-variables", gomock.Any()).
					Return(nil, errors.New("failed to list variables")).
					AnyTimes()
			}, // do nothing
			wantErr:   true,
			expectErr: "failed to list variables",
		},
		{
			name:        "not allowed to list variable set variables",
			workspaceId: "w-test-not-allowed-to-list-variable-set-variables",
			showOpt:     &ShowOption{format: "detail", includeVariableSet: true},
			setClient: func(mc *mocks.MockVariables, mvs *mocks.MockVariableSets, mvsv *mocks.MockVariableSetVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-not-allowed-to-list-variable-set-variables", gomock.Any()).
					Return(&tfe.VariableList{
						Items: nil,
					}, nil).
					AnyTimes()
				mvs.EXPECT().
					ListForWorkspace(context.TODO(), "w-test-not-allowed-to-list-variable-set-variables", nil).
					Return(nil, errors.New("failed to list variable set in workspace")).
					AnyTimes()
			},
			wantErr:   true,
			expectErr: "failed to list variable set in workspace",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			showOpt := tt.showOpt
			var buf bytes.Buffer
			tt.setClient(Variables, VariableSets, VariableSetVariables)

			err := show(ctx, tt.workspaceId, Variables, VariableSets, VariableSetVariables, showOpt, &buf)

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
				t.Errorf("expect %s, got %s", tt.expect, buf.String())
			}
		})
	}
}

func TestNewShowOption(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		expect *ShowOption
	}{
		{
			name: "default value",
			args: []string{},
			expect: &ShowOption{
				varFile: "terraform.tfvars",
				local:   false,
				format:  "detail",
			},
		},
		{
			name: "custom var file",
			args: []string{"--var-file", "custom.tfvars"},
			expect: &ShowOption{
				varFile: "custom.tfvars",
				local:   false,
				format:  "detail",
			},
		},
		{
			name: "enable local option",
			args: []string{"--local"},
			expect: &ShowOption{
				varFile: "terraform.tfvars",
				local:   true,
				format:  "detail",
			},
		},
		{
			name: "specify variable",
			args: []string{"--variable", "environment"},
			expect: &ShowOption{
				varFile:     "terraform.tfvars",
				variableKey: "environment",
				format:      "detail",
			},
		},
		{
			name: "enable include env option",
			args: []string{"--include-env"},
			expect: &ShowOption{
				varFile:    "terraform.tfvars",
				includeEnv: true,
				format:     "detail",
			},
		},
		{
			name: "enable include variable set option",
			args: []string{"--include-variable-set"},
			expect: &ShowOption{
				varFile:            "terraform.tfvars",
				includeVariableSet: true,
				format:             "detail",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			app := cli.NewApp()
			set := flagSet(showFlags())
			set.Parse(tt.args)
			ctx := cli.NewContext(app, set, nil)

			sut := NewShowOption(ctx)

			if !reflect.DeepEqual(tt.expect, sut) {
				t.Errorf("expect '%v', got '%v'", tt.expect, sut)
			}
		})
	}
}

func TestRequireTfcAccess(t *testing.T) {
	cases := []struct {
		name   string
		opts   *ShowOption
		expect bool
	}{
		{
			name:   "default",
			opts:   &ShowOption{local: false},
			expect: true,
		},
		{
			name:   "enable local option",
			opts:   &ShowOption{local: true},
			expect: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual := requireTfcAccess(tt.opts)

			if actual != tt.expect {
				t.Errorf("expect: %t, got: %t", tt.expect, actual)
			}
		})
	}
}
