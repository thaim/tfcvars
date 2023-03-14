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
						Key: tfe.String("environment"),
						Value: tfe.String("test"),
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
