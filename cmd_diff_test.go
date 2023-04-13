package main

import (
	"reflect"
	"testing"

	"github.com/urfave/cli/v2"
)

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

