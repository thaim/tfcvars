package main

import (
	"bytes"
	"context"
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

	cases := []struct {
		name        string
		workspaceId string
		diffOpt     *DiffOption
		setClient   func(*mocks.MockVariables)
		expect      string
		wantErr     bool
		expectErr   string
	}{
		{
			name:        "show no diff with empty local variable and empty variable list",
			workspaceId: "w-test-no-vars-workspace",
			diffOpt:     &DiffOption{},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-no-vars-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{{}},
					}, nil).
					AnyTimes()
			},
			expect: "",
		},
		{
			name:        "show no diff with same variables",
			workspaceId: "w-test-single-variable-workspace",
			diffOpt:     &DiffOption{varFile: "testdata/terraform.tfvars"},
			setClient: func(mc *mocks.MockVariables) {
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
			name:        "show no diff with mutiple variables",
			workspaceId: "w-test-multiple-variables-workspace",
			diffOpt:     &DiffOption{varFile: "testdata/mixedtypes.tfvars"},
			setClient: func(mc *mocks.MockVariables) {
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
			name:        "show diff with dirrerent key",
			workspaceId: "w-test-single-variable-different-key-workspace",
			diffOpt:     &DiffOption{varFile: "testdata/terraform.tfvars"},
			setClient: func(mc *mocks.MockVariables) {
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
			expect: `
- 		Key:   "environment",
+ 		Key:   "env",
`,
		},
		{
			name:        "return error if not able to list variable list",
			workspaceId: "w-test-access-error",
			diffOpt:     &DiffOption{},
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
			name:        "return error if failed to readvars file",
			workspaceId: "w-test-no-vars-workspace",
			diffOpt:     &DiffOption{varFile: "testdata/invalid.tfvars"},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-no-vars-workspace", nil).
					Return(&tfe.VariableList{
						Items: []*tfe.Variable{{}},
					}, nil).
					AnyTimes()
			},
			expect:    "",
			wantErr:   true,
			expectErr: "Argument or block definition required",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			var buf bytes.Buffer
			tt.setClient(mockVariables)

			err := diff(ctx, tt.workspaceId, mockVariables, tt.diffOpt, &buf)

			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.expectErr) {
					t.Errorf("expect %s error, got %s", tt.expectErr, err.Error())
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
