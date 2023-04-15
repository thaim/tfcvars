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
			name:        "",
			workspaceId: "w-test-no-vars-workspace",
			diffOpt:     &DiffOption{},
			setClient: func(mc *mocks.MockVariables) {
				mc.EXPECT().
					List(context.TODO(), "w-test-no-vars-workspace", nil).
					Return(&tfe.VariableList{
						Items: nil,
					}, nil).
					AnyTimes()
			},
			expect: "",
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
			if bufString := buf.String(); bufString != tt.expect {
				t.Errorf("expect %s, got id: %s", tt.expect, buf.String())
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
			name:   "default value",
			args:   []string{},
			expect: &DiffOption{},
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