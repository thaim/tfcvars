package main

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

func String(value cty.Value) string {
	ty := value.Type()
	switch ty {
	case cty.String:
		return value.AsString()
	case cty.Number:
		return value.AsBigFloat().String()
	case cty.Bool:
		return strconv.FormatBool(value.True())
	}

	if ty.IsTupleType() {
		valString := "["
		for idx, elm := range value.AsValueSlice() {
			if idx != 0 {
				valString += ", "
			}
			valString += `"` + String(elm) + `"`
		}
		valString += "]"

		return valString
	}

	return ""
}

func CtyValue(value string) cty.Value {
	p := hclparse.NewParser()
	file, diag := p.ParseHCL([]byte(fmt.Sprintf("key = %s", value)), "file")
	if diag.HasErrors() {
		file, _ = p.ParseHCL([]byte(fmt.Sprintf(`key = "%s"`, value)), "file")
	}
	attr := file.AttributeAtPos(hcl.Pos{Line: 1, Column: 1})
	val, diag := attr.Expr.Value(nil)
	if diag.HasErrors() {
		fmt.Printf("error: %s\n", value)
		return cty.StringVal("")
	}

	return val
}
