package main

import (
	"context"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/mocks"
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

	mockVariables.EXPECT().
		Update(context.TODO(), "w-test-no-vars-workspace", "", nil).
		Return(&tfe.Variable{}, nil).
		Times(0)
	mockVariables.EXPECT().
		Create(context.TODO(), "w-test-no-vars-workspace", nil).
		Return(&tfe.Variable{}, nil).
		Times(0)

	cases := []struct {
		name        string
		workspaceId string
		vars        *tfe.VariableList
		expect      string
		wantErr     bool
		expectErr   string
	}{
		{
			name:        "push no variable",
			workspaceId: "w-test-no-vars-workspace",
			vars:        &tfe.VariableList{},
			expect:      "",
			wantErr:     false,
			expectErr:   "",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			pushOpt := &PushOption{}

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
