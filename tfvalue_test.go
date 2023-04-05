package main

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestString(t *testing.T) {
	cases := []struct {
		name     string
		ctyValue cty.Value
		expect   string
	}{
		{
			name:     "primitive string",
			ctyValue: cty.StringVal("value"),
			expect:   "value",
		},
		{
			name:     "primitive number int",
			ctyValue: cty.NumberIntVal(123),
			expect:   "123",
		},
		{
			name:     "primitive number float",
			ctyValue: cty.NumberFloatVal(123.5),
			expect:   "123.5",
		},
		{
			name:     "primitive negative number float",
			ctyValue: cty.NumberFloatVal(-543.21),
			expect:   "-543.21",
		},
		{
			name:     "primitive bool true",
			ctyValue: cty.BoolVal(true),
			expect:   "true",
		},
		{
			name:     "primitive bool false",
			ctyValue: cty.BoolVal(false),
			expect:   "false",
		},
		{
			name:     "list string",
			ctyValue: cty.TupleVal([]cty.Value{cty.StringVal("ap-northeast-1a"), cty.StringVal("ap-northeast-1c"), cty.StringVal("ap-northeast-1d")}),
			expect:   `["ap-northeast-1a", "ap-northeast-1c", "ap-northeast-1d"]`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual := String(tt.ctyValue)

			if actual != tt.expect {
				t.Errorf("expect '%s', got '%s'", tt.expect, actual)
			}
		})
	}
}
