package main

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestString(t *testing.T) {
	cases := []struct {
		name string
		ctyValue cty.Value
		expect string
	}{
		{
			name: "primitive string",
			ctyValue: cty.StringVal("value"),
			expect: "value",
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
