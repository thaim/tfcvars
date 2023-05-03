package main

import (
	"errors"
	"fmt"
	"strconv"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rs/zerolog/log"
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

	first := true
	valString := "{"
	for key, val := range value.AsValueMap() {
		if !first {
			valString += ", "
		}
		if val.Type() == cty.String {
			valString += key + " = \"" + String(val) + "\""
		} else {
			valString += key + " = " + String(val)
		}
		first = false
	}
	valString += "}"

	return valString
}

func IsPrimitive(value cty.Value) bool {
	ty := value.Type()
	switch ty {
	case cty.String, cty.Number, cty.Bool:
		return true
	}

	return false
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

func BuildVariableList(key string, value string) *tfe.VariableList {
	vars := &tfe.VariableList{
		Items: []*tfe.Variable{
			{
				Key:   key,
				Value: value,
			},
		},
	}

	return vars
}

func BuildHCLFile(remoteVars []*tfe.Variable, localFile []byte, filename string) (*hclwrite.File, error) {
	var f *hclwrite.File
	var diags hcl.Diagnostics
	f, diags = hclwrite.ParseConfig(localFile, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		log.Error().Msgf("failed to parse existing varfile: %s", diags.Error())
		return nil, errors.New(diags.Error())
	}

	rootBody := f.Body()
	for _, v := range remoteVars {
		if v.Sensitive {
			rootBody.AppendUnstructuredTokens(generateComment(v.Key))
			continue
		}
		if v.HCL {
			rootBody.SetAttributeValue(v.Key, CtyValue(v.Value))
		} else {
			rootBody.SetAttributeValue(v.Key, cty.StringVal(v.Value))
		}
	}

	return f, nil
}
