package main

import (
	"strconv"

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
