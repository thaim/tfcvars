package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/mocks"
	"github.com/urfave/cli/v2"
)

func TestCmdRemove(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockVariables := mocks.NewMockVariables(ctrl)

	cases := []struct {
		name        string
		workspaceId string
		removeOpt   *RemoveOption
		vars        *tfe.VariableList
		setClient   func(*mocks.MockVariables)
		input       string
		expect      string
		wantErr     bool
		expectErr   string
	}{
		{
			name:        "remove variable",
			workspaceId: "ws-remove-variable",
			removeOpt:   &RemoveOption{variableKey: "environment", autoApprove: true},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().List(gomock.Any(), "ws-remove-variable", nil).Return(&tfe.VariableList{
					Items: []*tfe.Variable{
						{
							ID:    "v-environment",
							Key:   "environment",
							Value: "test",
						},
					},
				}, nil)
				mc.EXPECT().Delete(gomock.Any(), "ws-remove-variable", "v-environment").Return(nil)
			},
		},
		{
			name:        "return error if variable not exist",
			workspaceId: "ws-specified-variable-not-exist",
			removeOpt:   &RemoveOption{variableKey: "environment", autoApprove: true},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().List(gomock.Any(), "ws-specified-variable-not-exist", nil).Return(&tfe.VariableList{
					Items: []*tfe.Variable{
						{
							ID:    "v-terraform",
							Key:   "terraform",
							Value: "true",
						},
						{
							ID:    "v-aws_region",
							Key:   "aws_region",
							Value: "ap-northeast-1",
						},
					},
				}, nil)
			},
			expect:    "",
			wantErr:   true,
			expectErr: "variable 'environment' not found",
		},
		{
			name:        "return error if delete variable in tfc failed",
			workspaceId: "ws-error-raised-in-tfc",
			removeOpt:   &RemoveOption{variableKey: "environment", autoApprove: true},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().List(gomock.Any(), "ws-error-raised-in-tfc", nil).Return(&tfe.VariableList{
					Items: []*tfe.Variable{
						{
							ID:    "v-environment",
							Key:   "environment",
							Value: "test",
						},
						{
							ID:    "v-aws_region",
							Key:   "aws_region",
							Value: "ap-northeast-1",
						},
					},
				}, nil)
				mc.EXPECT().Delete(gomock.Any(), "ws-error-raised-in-tfc", "v-environment").Return(errors.New("failed to delete variable"))
			},
			expect:    "",
			wantErr:   true,
			expectErr: "failed to delete variable",
		},
		{
			name:        "require approve if auto-approve is not specified",
			workspaceId: "w-require-approve",
			removeOpt:   &RemoveOption{variableKey: "environment", autoApprove: false},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().List(gomock.Any(), "w-require-approve", nil).Return(&tfe.VariableList{
					Items: []*tfe.Variable{
						{
							ID:    "v-environment",
							Key:   "environment",
							Value: "test",
						},
						{
							ID:    "v-aws_region",
							Key:   "aws_region",
							Value: "ap-northeast-1",
						},
					},
				}, nil)
				mc.EXPECT().Delete(gomock.Any(), "w-require-approve", "v-environment").Return(nil)
			},
			input:   "yes\n",
			expect:  "delete variable: environment",
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			tt.setClient(mockVariables)
			tt.removeOpt.in = strings.NewReader(tt.input)
			outBuf := new(bytes.Buffer)
			tt.removeOpt.out = outBuf

			err := remove(ctx, tt.workspaceId, mockVariables, tt.removeOpt)

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

func TestNewRemoveOption(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		expect *RemoveOption
	}{
		{
			name:   "default value",
			args:   []string{},
			expect: nil,
		},
		{
			name: "specify variable key",
			args: []string{"--variable", "environment"},
			expect: &RemoveOption{
				variableKey: "environment",
				autoApprove: false,
				in:          os.Stdin,
				out:         os.Stdout,
			},
		},
		{
			name:   "specify only auto-approve",
			args:   []string{"--auto-approve"},
			expect: nil,
		},
		{
			name: "specify all variables",
			args: []string{"--variable", "environment", "--auto-approve"},
			expect: &RemoveOption{
				variableKey: "environment",
				autoApprove: true,
				in:          os.Stdin,
				out:         os.Stdout,
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			app := cli.NewApp()
			set := flagSet(removeFlags())
			set.Parse(tt.args)
			ctx := cli.NewContext(app, set, nil)

			sut := NewRemoveOption(ctx)

			if !reflect.DeepEqual(tt.expect, sut) {
				t.Errorf("expect '%v', got '%v'", tt.expect, sut)
			}
		})
	}
}
