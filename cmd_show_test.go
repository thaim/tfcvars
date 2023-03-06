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

func TestCmdShow(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockVariables := mocks.NewMockVariables(ctrl)

	mockVariables.EXPECT().
		List(context.TODO(), "w-test-no-vars-workspace", nil).
		Return(&tfe.VariableList{
			Items: nil,
		}, nil).
		AnyTimes()

	mockVariables.EXPECT().
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

	mockVariables.EXPECT().
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

	cases := []struct {
		name        string
		workspaceId string
		expect      string
		wantErr     bool
		expectErr   string
	}{
		{
			name:        "show empty variable",
			workspaceId: "w-test-no-vars-workspace",
			expect:      "",
			wantErr:     false,
			expectErr:   "",
		},
		{
			name:        "show single variable",
			workspaceId: "w-test-single-variable-workspace",
			expect:      "Key: var1\nValue: value1\nDescription: description1\nSensitive: false\n\n",
			wantErr:     false,
			expectErr:   "",
		},
		{
			name:        "show multiple variables",
			workspaceId: "w-test-multiple-variables-workspace",
			expect:      "Key: var1\nValue: value1\nDescription: \nSensitive: false\n\nKey: var2\nValue: value2\nDescription: \nSensitive: false\n\n",
			wantErr:     false,
			expectErr:   "",
		},
		{
			name:        "show sensitive variable",
			workspaceId: "w-test-sensitive-variable-workspace",
			expect:      "Key: var1\nValue: \nDescription: sensitive\nSensitive: true\n\n",
			wantErr:     false,
			expectErr:   "",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			showOpt := &ShowOption{}
			var buf bytes.Buffer

			err := show(ctx, tt.workspaceId, mockVariables, showOpt, &buf)

			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.expectErr) {
					t.Errorf("expect %s error, got %T", tt.expectErr, err)
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
