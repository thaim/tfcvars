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
	mockVariables := mocks.NewMockVariables(ctrl)

	cases := []struct {
		name        string
		workspaceId string
		showOpt     *ShowOption
		setClient   func(*mocks.MockVariables)
		expect      string
		wantErr     bool
		expectErr   string
	}{
		{
			name:        "show empty variable",
			workspaceId: "w-test-no-vars-workspace",
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-no-vars-workspace", nil).
					Return(&tfe.VariableList{
						Items: nil,
					}, nil).
					AnyTimes()
			},
			showOpt:   &ShowOption{},
			expect:    "",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "show single variable",
			workspaceId: "w-test-single-variable-workspace",
			showOpt:     &ShowOption{},
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
			expect:    "Key: var1\nValue: value1\nDescription: description1\nSensitive: false\n\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "show multiple variables",
			workspaceId: "w-test-multiple-variables-workspace",
			showOpt:     &ShowOption{},
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
			expect:    "Key: var1\nValue: value1\nDescription: \nSensitive: false\n\nKey: var2\nValue: value2\nDescription: \nSensitive: false\n\n",
			wantErr:   false,
			expectErr: "",
		},
		{
			name:        "show sensitive variable",
			workspaceId: "w-test-sensitive-variable-workspace",
			showOpt:     &ShowOption{},
			setClient: func(mc *mocks.MockVariables) {
				mockVariables.EXPECT().
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
			showOpt:     &ShowOption{variableKey: "var2"},
			setClient: func(mc *mocks.MockVariables) {
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
			name:        "show local variable",
			workspaceId: "",
			showOpt:     &ShowOption{local: true, varFile: "testdata/terraform.tfvars"},
			setClient:   func(mc *mocks.MockVariables) {}, // do nothing
			expect:      "Key: environment\nValue: development\nDescription: \nSensitive: false\n\n",
			wantErr:     false,
			expectErr:   "",
		},
		{
			name:        "return HCL parse error",
			workspaceId: "not-used",
			showOpt:     &ShowOption{local: true, varFile: "testdata/invalid.tfvars"},
			setClient:   func(mc *mocks.MockVariables) {}, // do nothing
			wantErr:     true,
			expectErr:   "Argument or block definition required",
		},
		{
			name:        "not allowed to list variables",
			workspaceId: "w-test-not-allowed-to-list-variables",
			showOpt:     &ShowOption{},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-not-allowed-to-list-variables", gomock.Any()).
					Return(nil, errors.New("failed to list variables")).
					AnyTimes()
			}, // do nothing
			wantErr:   true,
			expectErr: "failed to list variables",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			showOpt := tt.showOpt
			var buf bytes.Buffer
			tt.setClient(mockVariables)

			err := show(ctx, tt.workspaceId, mockVariables, showOpt, &buf)

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
				t.Errorf("expect %s, got id: %s", tt.expect, buf.String())
			}
		})
	}
}

func TestNewShowOption(t *testing.T) {
	cases := []struct {
		name   string
		flags  []cli.Flag
		args   []string
		expect *ShowOption
	}{
		{
			name:  "default value",
			flags: showFlags(),
			args:  []string{},
			expect: &ShowOption{
				varFile: "terraform.tfvars",
				local:   false,
			},
		},
		{
			name:  "custom var file",
			flags: showFlags(),
			args:  []string{"--var-file", "custom.tfvars"},
			expect: &ShowOption{
				varFile: "custom.tfvars",
				local:   false,
			},
		},
		{
			name:  "enable local option",
			flags: showFlags(),
			args:  []string{"--local"},
			expect: &ShowOption{
				varFile: "terraform.tfvars",
				local:   true,
			},
		},
		{
			name:  "specify variable",
			flags: showFlags(),
			args:  []string{"--variable", "environment"},
			expect: &ShowOption{
				varFile:     "terraform.tfvars",
				variableKey: "environment",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			app := cli.NewApp()
			set := flagSet(tt.flags)
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
