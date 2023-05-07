package main

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func TestNewTfvarsVariable(t *testing.T) {
	cases := []struct {
		name   string
		vars   []*tfe.Variable
		expect *Tfvars
	}{
		{
			name: "empty variable list",
			vars: []*tfe.Variable{},
			expect: &Tfvars{
				filename: "generated.tfvars",
				vardata: []byte(""),
				vars: []*tfe.Variable{},
			},
		},
		{
			name: "single variable list",
			vars: []*tfe.Variable{
				{
					Key: "env",
					Value: "test",
				},
			},
			expect: &Tfvars{
				filename: "generated.tfvars",
				vardata: []byte("env = \"test\"\n"),
				vars: []*tfe.Variable{
					{
						Key: "env",
						Value: "test",
					},
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual := NewTfvarsVariable(tt.vars)

			if (tt.expect.filename != actual.filename ||
				!bytes.Equal(tt.expect.vardata, actual.vardata) ||
				!reflect.DeepEqual(tt.expect.vars, actual.vars)) {
				t.Errorf("expect '%v', got '%v' (%s)", tt.expect, actual, actual.vardata)
			}
		})
	}
}

func TestTfvars_BuildHCLFile(t *testing.T) {
	cases := []struct {
		name   string
		tfvars *Tfvars
		expect *hclwrite.File
		wantErr bool
		expectErr string
	}{
		{
			name: "empty hcl file",
			tfvars: NewTfvarsVariable([]*tfe.Variable{}),
			expect: func() *hclwrite.File {
				f := hclwrite.NewEmptyFile()
				return f
			}(),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			file, err := tt.tfvars.BuildHCLFile()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expect '%s' error, got no error", tt.expectErr)
				} else if !strings.Contains(err.Error(), tt.expectErr) {
					t.Errorf("expect %s error, got %s", tt.expectErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("expect no error, got error: '%v'", err)
			}

			if !bytes.Equal(file.Bytes(), tt.expect.Bytes()) {
				t.Errorf("expect '%s', got '%s'", tt.expect.Bytes(), file.Bytes())
			}
		})
	}
}
