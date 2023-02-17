package main

import (
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-tfe/mocks"
)


type MockClient struct {
	Workspaces *mocks.MockWorkspaces
	Variables *mocks.MockVariables
}

func NewMockTfeClient(ctrl *gomock.Controller) *MockClient {
	return &MockClient{
		Workspaces: mocks.NewMockWorkspaces(ctrl),
		Variables: mocks.NewMockVariables(ctrl),
	}
}
