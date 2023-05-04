package main

import (
	"errors"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type Tfvars struct {
	filename string
	vardata []byte
	vars  []*tfe.Variable
}

func NewTfvars() *Tfvars {
	vf := &Tfvars{}

	return vf
}

func (vf *Tfvars) BuildHCLFile() (*hclwrite.File, error) {
	f := hclwrite.NewEmptyFile()
	var diags hcl.Diagnostics
	if vf.vardata != nil {
		f, diags = hclwrite.ParseConfig(vf.vardata, vf.filename, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return nil, errors.New(diags.Error())
		}
	}

	rootBody := f.Body()
	for _, v := range vf.vars {
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

func (vf *Tfvars) BuildHCLFileString() string {
	file, err := vf.BuildHCLFile()
	if err != nil {
		return ""
	}

	return string(file.Bytes())
}
