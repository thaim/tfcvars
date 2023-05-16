package main

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func TestNewTfvarsVariable(t *testing.T) {
	cases := []struct {
		name    string
		vars    []*tfe.Variable
		expect  *Tfvars
		wantErr bool
	}{
		{
			name: "empty variable list",
			vars: []*tfe.Variable{},
			expect: &Tfvars{
				filename: "generated.tfvars",
				vardata:  []byte(""),
				vars:     []*tfe.Variable{},
			},
		},
		{
			name: "single variable list",
			vars: []*tfe.Variable{
				{
					Key:   "env",
					Value: "test",
				},
			},
			expect: &Tfvars{
				filename: "generated.tfvars",
				vardata:  []byte("env = \"test\"\n"),
				vars: []*tfe.Variable{
					{
						Key:   "env",
						Value: "test",
					},
				},
			},
		},
		{
			name: "return nil if variable is broken",
			vars: []*tfe.Variable{
				{
					Value: "valwithoutkey",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual := NewTfvarsVariable(tt.vars)

			if tt.wantErr {
				if actual != nil {
					t.Errorf("expect to return nil, got valid value '%s'", actual.vardata)
				}
				return
			}

			if tt.expect.filename != actual.filename ||
				!bytes.Equal(tt.expect.vardata, actual.vardata) ||
				!reflect.DeepEqual(tt.expect.vars, actual.vars) {
				t.Errorf("expect '%v', got '%v' (%s)", tt.expect, actual, actual.vardata)
			}
		})
	}
}

func TestTfvars_BuildHCLFile(t *testing.T) {
	cases := []struct {
		name      string
		tfvars    *Tfvars
		expect    *hclwrite.File
		wantErr   bool
		expectErr string
	}{
		{
			name:   "empty tfvars file",
			tfvars: NewTfvarsVariable([]*tfe.Variable{}),
			expect: func() *hclwrite.File {
				f := hclwrite.NewEmptyFile()
				return f
			}(),
		},
		{
			name: "single element file",
			tfvars: NewTfvarsVariable([]*tfe.Variable{
				{
					Key:   "env",
					Value: "test",
				},
			}),
			expect: func() *hclwrite.File {
				f := hclwrite.NewEmptyFile()
				body := f.Body()
				body.SetAttributeValue("env", cty.StringVal("test"))
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

func TestTfvars_BuildHCLFileString(t *testing.T) {
	cases := []struct {
		name   string
		tfvars *Tfvars
		expect string
	}{
		{
			name:   "empty tfvars file",
			tfvars: NewTfvarsVariable([]*tfe.Variable{}),
			expect: "",
		},
		{
			name: "single element file",
			tfvars: NewTfvarsVariable([]*tfe.Variable{
				{
					Key:   "env",
					Value: "test",
				},
			}),
			expect: "env = \"test\"\n",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.tfvars.BuildHCLFileString()

			if actual != tt.expect {
				t.Errorf("expect '%s', got '%s'", tt.expect, actual)
			}
		})
	}
}
