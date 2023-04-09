package main

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/mocks"
	"github.com/urfave/cli/v2"
)

func TestCmdPush(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockVariables := mocks.NewMockVariables(ctrl)

	mockVariables.EXPECT().
		List(context.TODO(), "w-test-no-vars-workspace", nil).
		Return(&tfe.VariableList{
			Items: nil,
		}, nil).
		AnyTimes()

	cases := []struct {
		name        string
		workspaceId string
		vars        *tfe.VariableList
		setClient   func(*mocks.MockVariables)
		expect      string
		wantErr     bool
		expectErr   string
	}{
		{
			name:        "push no variable",
			workspaceId: "w-test-no-vars-workspace",
			vars:        &tfe.VariableList{},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					Update(context.TODO(), "w-test-no-vars-workspace", gomock.Any(), gomock.Any()).
					Return(&tfe.Variable{}, nil).
					Times(0)
				mc.EXPECT().
					Create(context.TODO(), "w-test-no-vars-workspace", gomock.Any()).
					Return(&tfe.Variable{}, nil).
					Times(0)
			},
			expect:    "",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "create one variable",
			workspaceId: "w-test-no-vars-workspace",
			vars: &tfe.VariableList{
				Items: []*tfe.Variable{
					{
						Key:       "environment",
						Value:     "test",
						Category:  tfe.CategoryTerraform,
						HCL:       false,
						Sensitive: false,
					},
				},
			},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					Update(context.TODO(), "w-test-no-vars-workspace", gomock.Any(), gomock.Any()).
					Return(&tfe.Variable{}, nil).
					Times(0)
				mc.EXPECT().
					Create(context.TODO(), "w-test-no-vars-workspace", tfe.VariableCreateOptions{
						Key:       tfe.String("environment"),
						Value:     tfe.String("test"),
						Category:  tfe.Category(tfe.CategoryTerraform),
						HCL:       tfe.Bool(false),
						Sensitive: tfe.Bool(false),
					}).
					Return(&tfe.Variable{}, nil).
					Times(1)
			},
			expect:    "",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "update one variable",
			workspaceId: "w-test-one-var-workspace",
			vars: &tfe.VariableList{
				Items: []*tfe.Variable{
					{
						ID:        "variable-id-environment",
						Key:       "environment",
						Value:     "test",
						Category:  tfe.CategoryTerraform,
						HCL:       false,
						Sensitive: false,
					},
				},
			},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-one-var-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								ID:        "variable-id-environment",
								Key:       "environment",
								Value:     "test",
								Category:  tfe.CategoryTerraform,
								HCL:       false,
								Sensitive: false,
							},
						},
					}, nil).
					AnyTimes()
				mc.EXPECT().
					Update(context.TODO(), "w-test-one-var-workspace", "variable-id-environment", gomock.Any()).
					Return(&tfe.Variable{
						ID:        "variable-id-environment",
						Key:       "environment",
						Value:     "test2",
						Category:  tfe.CategoryTerraform,
						HCL:       false,
						Sensitive: false,
					}, nil).
					AnyTimes()
				mc.EXPECT().
					Create(context.TODO(), "w-test-one-var-workspace", gomock.Any()).
					Return(&tfe.Variable{}, nil).
					Times(0)
			},
			expect:    "",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "update one variable with exact same value",
			workspaceId: "w-test-one-var-workspace",
			vars: &tfe.VariableList{
				Items: []*tfe.Variable{
					{
						ID:        "variable-id-environment",
						Key:       "environment",
						Value:     "test",
						Category:  tfe.CategoryTerraform,
						HCL:       false,
						Sensitive: false,
					},
				},
			},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-one-var-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								ID:        "variable-id-environment",
								Key:       "environment",
								Value:     "test",
								Category:  tfe.CategoryTerraform,
								HCL:       false,
								Sensitive: false,
							},
						},
					}, nil).
					AnyTimes()
				mc.EXPECT().
					Update(context.TODO(), "w-test-one-var-workspace", "variable-id-environment", gomock.Any()).
					Return(&tfe.Variable{
						ID:        "variable-id-environment",
						Key:       "environment",
						Value:     "test",
						Category:  tfe.CategoryTerraform,
						HCL:       false,
						Sensitive: false,
					}, nil).
					Times(0)
				mc.EXPECT().
					Create(context.TODO(), "w-test-one-var-workspace", gomock.Any()).
					Return(&tfe.Variable{}, nil).
					Times(0)
			},
			expect:    "",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "return error if failed to access terraform cloud",
			workspaceId: "w-test-access-error",
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-access-error", nil).
					Return(nil, tfe.ErrInvalidWorkspaceID)
			},
			expect:    "",
			wantErr:   true,
			expectErr: "invalid value for workspace ID",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			pushOpt := &PushOption{}
			tt.setClient(mockVariables)

			err := push(ctx, tt.workspaceId, mockVariables, pushOpt, tt.vars)

			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.expectErr) {
					t.Errorf("expect %s error, got %T", tt.expectErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("expect no error, got error: %v", err)
			}
		})
	}
}

func TestNewPushOption(t *testing.T) {
	cases := []struct {
		name   string
		flags  []cli.Flag
		args   []string
		expect *PushOption
	}{
		{
			name:  "default value",
			flags: pushFlags(),
			args:  []string{},
			expect: &PushOption{
				varFile: "terraform.tfvars",
			},
		},
		{
			name:  "custom var file",
			flags: pushFlags(),
			args:  []string{"--var-file", "custom.tfvars"},
			expect: &PushOption{
				varFile: "custom.tfvars",
			},
		},
		{
			name:  "variable option",
			flags: pushFlags(),
			args:  []string{"--variable", "key=value"},
			expect: &PushOption{
				varFile:       "terraform.tfvars",
				variableKey:   "key",
				variableValue: "value",
			},
		},
		{
			name:  "variable option with include equal",
			flags: pushFlags(),
			args:  []string{"--variable", "key=value=10"},
			expect: &PushOption{
				varFile:       "terraform.tfvars",
				variableKey:   "key",
				variableValue: "value=10",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			app := cli.NewApp()
			set := flagSet(tt.flags)
			set.Parse(tt.args)
			ctx := cli.NewContext(app, set, nil)

			sut := NewPushOption(ctx)

			if !reflect.DeepEqual(tt.expect, sut) {
				t.Errorf("expect '%v', got '%v'", tt.expect, sut)
			}
		})
	}
}

func TestVariableFile(t *testing.T) {
	cases := []struct {
		name      string
		varfile   string
		expect    *tfe.VariableList
		wantErr   bool
		expectErr string
	}{
		{
			name:    "default value",
			varfile: "testdata/terraform.tfvars",
			expect: &tfe.VariableList{
				Items: []*tfe.Variable{
					{
						Key:   "environment",
						Value: "development",
					},
				},
			},
		},
		{
			name:      "invalid vars file",
			varfile:   "testdata/invalid.tfvars",
			wantErr:   true,
			expectErr: "Argument or block definition required",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := variableFile(tt.varfile)

			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.expectErr) {
					t.Errorf("expect '%s' error, got '%s'", tt.expectErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("expect no error, got error: %v", err)
			}

			if !reflect.DeepEqual(tt.expect, actual) {
				t.Errorf("expect '%v', got '%v'", tt.expect, actual)
			}
		})
	}
}

func TestVariableEqual(t *testing.T) {
	cases := []struct {
		name    string
		options tfe.VariableUpdateOptions
		target  *tfe.Variable
		expect  bool
	}{
		{
			name: "compare same value",
			options: tfe.VariableUpdateOptions{
				Key:         tfe.String("key"),
				Value:       tfe.String("value"),
				Description: tfe.String("description"),
				Category:    tfe.Category(tfe.CategoryTerraform),
				HCL:         tfe.Bool(false),
				Sensitive:   tfe.Bool(false),
			},
			target: &tfe.Variable{
				Key:         "key",
				Value:       "value",
				Description: "description",
				Category:    tfe.CategoryTerraform,
				HCL:         false,
				Sensitive:   false,
			},
			expect: true,
		},
		{
			name: "compare variable with key differ",
			options: tfe.VariableUpdateOptions{
				Key:         tfe.String("key"),
				Value:       tfe.String("value"),
				Description: tfe.String("description"),
				Category:    tfe.Category(tfe.CategoryTerraform),
				HCL:         tfe.Bool(false),
				Sensitive:   tfe.Bool(false),
			},
			target: &tfe.Variable{
				Key:         "changedkey",
				Value:       "value",
				Description: "description",
				Category:    tfe.CategoryTerraform,
				HCL:         false,
				Sensitive:   false,
			},
			expect: false,
		},
		{
			name: "compare variable with value differ",
			options: tfe.VariableUpdateOptions{
				Key:         tfe.String("key"),
				Value:       tfe.String("value"),
				Description: tfe.String("description"),
				Category:    tfe.Category(tfe.CategoryTerraform),
				HCL:         tfe.Bool(false),
				Sensitive:   tfe.Bool(false),
			},
			target: &tfe.Variable{
				Key:         "key",
				Value:       "changed value",
				Description: "description",
				Category:    tfe.CategoryTerraform,
				HCL:         false,
				Sensitive:   false,
			},
			expect: false,
		},
		{
			name: "compare variable with description differ",
			options: tfe.VariableUpdateOptions{
				Key:         tfe.String("key"),
				Value:       tfe.String("value"),
				Description: tfe.String("description"),
				Category:    tfe.Category(tfe.CategoryTerraform),
				HCL:         tfe.Bool(false),
				Sensitive:   tfe.Bool(false),
			},
			target: &tfe.Variable{
				Key:         "key",
				Value:       "value",
				Description: "description has updated",
				Category:    tfe.CategoryTerraform,
				HCL:         false,
				Sensitive:   false,
			},
			expect: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual := variableEqual(tt.options, tt.target)
			if actual != tt.expect {
				t.Errorf("expect '%t', got '%t'", tt.expect, actual)
			}
		})
	}
}
