package main

import (
	"bytes"
	"context"
	"os"
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
		pushOpt     *PushOption
		vars        *tfe.VariableList
		setClient   func(*mocks.MockVariables)
		input       string
		expect      string
		wantErr     bool
		expectErr   string
	}{
		{
			name:        "push no variable",
			workspaceId: "w-test-no-vars-workspace",
			pushOpt:     &PushOption{autoApprove: true}, // TODO do not require confirm if no changes
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
			pushOpt:     &PushOption{autoApprove: true},
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
			pushOpt:     &PushOption{autoApprove: true},
			vars: &tfe.VariableList{
				Items: []*tfe.Variable{
					{
						ID:        "variable-id-environment",
						Key:       "environment",
						Value:     "test2",
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
					Times(1)
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
			pushOpt:     &PushOption{autoApprove: true},
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
			name:        "delete one variable",
			workspaceId: "w-test-delete-one-var-workspace",
			pushOpt:     &PushOption{delete: true, autoApprove: true},
			vars: &tfe.VariableList{
				Items: []*tfe.Variable{
					{
						Key:   "environment",
						Value: "test",
					},
				},
			},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-delete-one-var-workspace", nil).
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
							{
								ID:        "variable-id-port",
								Key:       "port",
								Value:     "3000",
								Category:  tfe.CategoryTerraform,
								HCL:       false,
								Sensitive: false,
							},
						},
					}, nil).
					AnyTimes()
				mc.EXPECT().
					Delete(context.TODO(), "w-test-delete-one-var-workspace", "variable-id-port").
					Return(nil).
					Times(1)
			},
			expect:    "",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "require confirm and update variable after confirmed",
			workspaceId: "w-test-require-confirm-variable",
			pushOpt:     &PushOption{},
			vars: &tfe.VariableList{
				Items: []*tfe.Variable{
					{
						ID:    "variable-id-confirm",
						Key:   "environment",
						Value: "test",
					},
				},
			},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-require-confirm-variable", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								ID:    "variable-id-confirm",
								Key:   "environment",
								Value: "development",
							},
						},
					}, nil).
					AnyTimes()
				mc.EXPECT().
					Update(context.TODO(), "w-test-require-confirm-variable", "variable-id-confirm", gomock.Any()).
					Return(&tfe.Variable{
						ID:        "variable-id-confirm",
						Key:       "environment",
						Value:     "test",
						Category:  tfe.CategoryTerraform,
						HCL:       false,
						Sensitive: false,
					}, nil).
					Times(1)
			},
			input: "yes\n",
		},
		{
			name:        "abort push variables with confirm reject",
			workspaceId: "w-test-abort-confirm",
			pushOpt:     &PushOption{},
			vars: &tfe.VariableList{
				Items: []*tfe.Variable{
					{
						ID:    "variable-id-abort",
						Key:   "environment",
						Value: "test",
					},
				},
			},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-abort-confirm", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								ID:    "variable-id-abort",
								Key:   "environment",
								Value: "development",
							},
						},
					}, nil).
					AnyTimes()
			},
			input: "no\n",
		},
		{
			name:        "abort push variables with no confirm input",
			workspaceId: "w-test-abort-confirm-no-input",
			pushOpt:     &PushOption{},
			vars: &tfe.VariableList{
				Items: []*tfe.Variable{
					{
						ID:    "variable-id-abort-no-input",
						Key:   "environment",
						Value: "test",
					},
				},
			},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-abort-confirm-no-input", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{
							{
								ID:    "variable-id-abort-no-input",
								Key:   "environment",
								Value: "development",
							},
						},
					}, nil).
					AnyTimes()
			},
			input:     "",
			wantErr:   true,
			expectErr: "EOF",
		},
		{
			name:        "return error if failed to access terraform cloud",
			workspaceId: "w-test-access-error",
			pushOpt:     &PushOption{},
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
			tt.setClient(mockVariables)
			tt.pushOpt.in = strings.NewReader(tt.input)
			outBuf := new(bytes.Buffer)
			tt.pushOpt.out = outBuf

			err := push(ctx, tt.workspaceId, mockVariables, tt.pushOpt, tt.vars)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expect '%s' error, got no error", tt.expectErr)
				} else if !strings.Contains(err.Error(), tt.expectErr) {
					t.Errorf("expect %s error, got %T", tt.expectErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("expect no error, got error: %v", err)
			}
			if tt.expect != "" && !bytes.Equal(outBuf.Bytes(), []byte(tt.expect)) {
				t.Errorf("expect '%s', got '%s'", tt.expect, outBuf.Bytes())
			}
		})
	}
}

func TestNewPushOption(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		expect *PushOption
	}{
		{
			name: "default value",
			args: []string{},
			expect: &PushOption{
				varFile: "terraform.tfvars",
				in:      os.Stdin,
				out:     os.Stdout,
			},
		},
		{
			name: "custom var file",
			args: []string{"--var-file", "custom.tfvars"},
			expect: &PushOption{
				varFile: "custom.tfvars",
				in:      os.Stdin,
				out:     os.Stdout,
			},
		},
		{
			name: "variable option",
			args: []string{"--variable", "key=value"},
			expect: &PushOption{
				varFile:       "terraform.tfvars",
				variableKey:   "key",
				variableValue: "value",
				in:            os.Stdin,
				out:           os.Stdout,
			},
		},
		{
			name: "variable option with include equal",
			args: []string{"--variable", "key=value=10"},
			expect: &PushOption{
				varFile:       "terraform.tfvars",
				variableKey:   "key",
				variableValue: "value=10",
				in:            os.Stdin,
				out:           os.Stdout,
			},
		},
		{
			name: "delete option enabled",
			args: []string{"--delete"},
			expect: &PushOption{
				varFile: "terraform.tfvars",
				delete:  true,
				in:      os.Stdin,
				out:     os.Stdout,
			},
		},
		{
			name: "auto-approve option enabled",
			args: []string{"--auto-approve"},
			expect: &PushOption{
				varFile:     "terraform.tfvars",
				autoApprove: true,
				in:          os.Stdin,
				out:         os.Stdout,
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			app := cli.NewApp()
			set := flagSet(pushFlags())
			set.Parse(tt.args)
			ctx := cli.NewContext(app, set, nil)

			sut := NewPushOption(ctx)

			if !reflect.DeepEqual(tt.expect, sut) {
				t.Errorf("expect '%v', got '%v'", tt.expect, sut)
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

func TestConfirm(t *testing.T) {
	cases := []struct {
		name       string
		input      string
		expectBool bool
		expectErr  string
		wantErr    bool
	}{
		{
			name:       "return true from y",
			input:      "y\n",
			expectBool: true,
			expectErr:  "",
			wantErr:    false,
		},
		{
			name:       "return true from yes",
			input:      "yes\n",
			expectBool: true,
			expectErr:  "",
			wantErr:    false,
		},
		{
			name:       "return true from upper yes",
			input:      "YES\n",
			expectBool: true,
			expectErr:  "",
			wantErr:    false,
		},
		{
			name:       "return false from no",
			input:      "no\n",
			expectBool: false,
			expectErr:  "",
			wantErr:    false,
		},
		{
			name:       "return error without newline",
			input:      "yes",
			expectBool: false,
			expectErr:  "EOF",
			wantErr:    true,
		},
		{
			name:       "return error if no input specified",
			expectBool: false,
			expectErr:  "EOF",
			wantErr:    true,
		},
	}

	for _, tt := range cases {
		in := strings.NewReader(tt.input)
		result, err := confirm(in)

		if tt.wantErr {
			if err == nil {
				t.Errorf("expect error '%s', got no error", tt.expectErr)
			} else if !strings.Contains(err.Error(), tt.expectErr) {
				t.Errorf("expect error '%s', got '%v'", tt.expectErr, err)
			}
			return
		}
		if err != nil {
			t.Errorf("expect no error, got error '%v'", err)
		}
		if tt.expectBool != result {
			t.Errorf("expect result '%t', got '%t'", tt.expectBool, result)
		}
	}
}
