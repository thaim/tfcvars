package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/urfave/cli/v2"
)

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
			set := flagSet(pushFlags())
			set.Parse(tt.args)
			ctx := cli.NewContext(app, set, nil)

			sut := NewRemoveOption(ctx)

			if !reflect.DeepEqual(tt.expect, sut) {
				t.Errorf("expect '%v', got '%v'", tt.expect, sut)
			}
		})
	}
}
