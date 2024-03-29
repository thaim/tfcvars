package main

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/exp/slices"
)

func TestString(t *testing.T) {
	cases := []struct {
		name     string
		ctyValue cty.Value
		expect   []string
	}{
		{
			name:     "primitive string",
			ctyValue: cty.StringVal("value"),
			expect:   []string{"value"},
		},
		{
			name:     "primitive number int",
			ctyValue: cty.NumberIntVal(123),
			expect:   []string{"123"},
		},
		{
			name:     "primitive number float",
			ctyValue: cty.NumberFloatVal(123.5),
			expect:   []string{"123.5"},
		},
		{
			name:     "primitive negative number float",
			ctyValue: cty.NumberFloatVal(-543.21),
			expect:   []string{"-543.21"},
		},
		{
			name:     "primitive bool true",
			ctyValue: cty.BoolVal(true),
			expect:   []string{"true"},
		},
		{
			name:     "primitive bool false",
			ctyValue: cty.BoolVal(false),
			expect:   []string{"false"},
		},
		{
			name:     "list string",
			ctyValue: cty.TupleVal([]cty.Value{cty.StringVal("ap-northeast-1a"), cty.StringVal("ap-northeast-1c"), cty.StringVal("ap-northeast-1d")}),
			expect:   []string{`["ap-northeast-1a", "ap-northeast-1c", "ap-northeast-1d"]`},
		},
		{
			name:     "empty map",
			ctyValue: cty.ObjectVal(map[string]cty.Value{}),
			expect:   []string{`{}`},
		},
		{
			name:     "simple map",
			ctyValue: cty.ObjectVal(map[string]cty.Value{"key": cty.StringVal("value"), "key2": cty.StringVal("value2")}),
			expect:   []string{`{key = "value", key2 = "value2"}`, `{key2 = "value2", key = "value"}`},
		},
		{
			name:     "nested map",
			ctyValue: cty.ObjectVal(map[string]cty.Value{"key": cty.StringVal("value"), "key2": cty.ObjectVal(map[string]cty.Value{"key2key": cty.StringVal("nestedValue")})}),
			expect:   []string{`{key = "value", key2 = {key2key = "nestedValue"}}`, `{key2 = {key2key = "nestedValue"}, key = "value"}`},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual := String(tt.ctyValue)

			if !slices.Contains(tt.expect, actual) {
				t.Errorf("expect '%s', got '%s'", tt.expect, actual)
			}
		})
	}
}

func TestIsPrimitive(t *testing.T) {
	cases := []struct {
		name   string
		value  cty.Value
		expect bool
	}{
		{
			name:   "primitive string",
			value:  cty.StringVal("string_value"),
			expect: true,
		},
		{
			name:   "primitive bool",
			value:  cty.BoolVal(true),
			expect: true,
		},
		{
			name:   "primitive int number",
			value:  cty.NumberIntVal(42),
			expect: true,
		},
		{
			name:   "primitive float number",
			value:  cty.NumberFloatVal(123.45),
			expect: true,
		},
		{
			name:   "tuple value",
			value:  cty.TupleVal([]cty.Value{cty.StringVal("ap-northeast-1a"), cty.StringVal("ap-northeast-1c"), cty.StringVal("ap-northeast-1d")}),
			expect: false,
		},
		{
			name:   "object value",
			value:  cty.ObjectVal(map[string]cty.Value{"key": cty.StringVal("value"), "key2": cty.StringVal("value2")}),
			expect: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsPrimitive(tt.value)

			if actual != tt.expect {
				t.Errorf("expect %t, got %t", tt.expect, actual)
			}
		})
	}
}

func TestCtyValue(t *testing.T) {
	cases := []struct {
		name   string
		value  string
		expect cty.Value
	}{
		{
			name:   "primitive string",
			value:  `"value"`,
			expect: cty.StringVal("value"),
		},
		{
			name:   "primitive int",
			value:  `123`,
			expect: cty.NumberFloatVal(123),
		},
		{
			name:   "list string",
			value:  `["ap-northeast-1a", "ap-northeast-1c", "ap-northeast-1d"]`,
			expect: cty.TupleVal([]cty.Value{cty.StringVal("ap-northeast-1a"), cty.StringVal("ap-northeast-1c"), cty.StringVal("ap-northeast-1d")}),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual := CtyValue(tt.value)

			if actual.Equals(tt.expect).False() {
				t.Errorf("expect '%s' (type %s), got '%s' (type %s)",
					String(tt.expect), tt.expect.Type().GoString(), String(actual), actual.Type().GoString())
			}
		})
	}
}

func TestBuildVariableList(t *testing.T) {
	cases := []struct {
		name   string
		key    string
		value  string
		expect *tfe.VariableList
	}{
		{
			name:  "primitive string",
			key:   "environment",
			value: `"test"`,
			expect: &tfe.VariableList{
				Items: []*tfe.Variable{
					{
						Key:   "environment",
						Value: `"test"`,
					},
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual := BuildVariableList(tt.key, tt.value)

			if !reflect.DeepEqual(tt.expect, actual) {
				t.Errorf("expect '%v', got '%v'", tt.expect.Items[0], actual.Items[0])
			}
		})
	}
}

func TestBuildHCLFile(t *testing.T) {
	cases := []struct {
		name      string
		vars      []*tfe.Variable
		filebody  []byte
		filename  string
		expect    *hclwrite.File
		wantErr   bool
		expectErr string
	}{
		{
			name:     "parse valid tfvars file",
			vars:     nil,
			filebody: []byte("environment = \"development\"\n"),
			filename: "testdata/terraform.tfvars",
			expect: func() *hclwrite.File {
				f := hclwrite.NewEmptyFile()
				body := f.Body()
				body.SetAttributeValue("environment", cty.StringVal("development"))
				return f
			}(),
		},
		{
			name: "parse invalid tfvars file",
			vars: nil,
			filebody: func() []byte {
				data, _ := os.ReadFile("testdata/invalid.tfvars")
				return data
			}(),
			filename:  "testdata/invalid.tfvars",
			expect:    nil,
			wantErr:   true,
			expectErr: "test",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := BuildHCLFile(tt.vars, tt.filebody, tt.filename)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expect '%s' error, got no error", tt.expectErr)
				} else if !strings.Contains(err.Error(), tt.expectErr) {
					t.Errorf("expect %s error, got %s", tt.expectErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("expect no error, got error: %v", err)
			}

			if !bytes.Equal(tt.expect.Bytes(), actual.Bytes()) {
				t.Errorf("expect '%s', got '%s'", tt.expect.Bytes(), actual.Bytes())
			}
		})
	}
}

func TestFilterEnv(t *testing.T) {
	cases := []struct {
		name   string
		input  []*tfe.Variable
		expect []*tfe.Variable
	}{
		{
			name: "do not change terraform variables",
			input: []*tfe.Variable{
				{
					Key:      "key1",
					Value:    "value1",
					Category: tfe.CategoryTerraform,
				},
				{
					Key:      "key2",
					Value:    "value2",
					Category: tfe.CategoryTerraform,
				},
				{
					Key:      "key3",
					Value:    "value3",
					Category: tfe.CategoryTerraform,
				},
			},
			expect: []*tfe.Variable{
				{
					Key:      "key1",
					Value:    "value1",
					Category: tfe.CategoryTerraform,
				},
				{
					Key:      "key2",
					Value:    "value2",
					Category: tfe.CategoryTerraform,
				},
				{
					Key:      "key3",
					Value:    "value3",
					Category: tfe.CategoryTerraform,
				},
			},
		},
		{
			name: "remove all env varibles",
			input: []*tfe.Variable{
				{
					Key:      "key1",
					Value:    "value1",
					Category: tfe.CategoryEnv,
				},
				{
					Key:      "key2",
					Value:    "value2",
					Category: tfe.CategoryEnv,
				},
				{
					Key:      "key3",
					Value:    "value3",
					Category: tfe.CategoryEnv,
				},
			},
			expect: []*tfe.Variable{},
		},
		{
			name: "remove env varibles from mixed list",
			input: []*tfe.Variable{
				{
					Key:      "key1",
					Value:    "value1",
					Category: tfe.CategoryTerraform,
				},
				{
					Key:      "key2",
					Value:    "value2",
					Category: tfe.CategoryEnv,
				},
				{
					Key:      "key3",
					Value:    "value3",
					Category: tfe.CategoryTerraform,
				},
				{
					Key:      "key4",
					Value:    "value4",
					Category: tfe.CategoryEnv,
				},
			},
			expect: []*tfe.Variable{
				{
					Key:      "key1",
					Value:    "value1",
					Category: tfe.CategoryTerraform,
				},
				{
					Key:      "key3",
					Value:    "value3",
					Category: tfe.CategoryTerraform,
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual := FilterEnv(tt.input)

			if !reflect.DeepEqual(tt.expect, actual) {
				t.Errorf("expect '%s', got '%s'", toString(tt.expect), toString(actual))
			}
		})
	}
}

func toString(vars []*tfe.Variable) string {
	result := "{"
	for i, v := range vars {
		if i != 0 {
			result += ", "
		}
		result += fmt.Sprintf("%+v", v)
	}
	return result + "}"
}
