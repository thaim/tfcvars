package main

import (
	"errors"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type Tfvars struct {
	filename string
	vardata  []byte
	vars     []*tfe.Variable
}

// NewTfvarsVariable create instance from list of tfe.Variable
func NewTfvarsVariable(vars []*tfe.Variable) *Tfvars {
	vf := &Tfvars{}

	vf.vars = vars
	vf.filename = ""
	vf.vardata = nil

	err := vf.convertTfeVariables()
	if err != nil {
		return nil
	}

	return vf
}

// NewTfvarsFile create instance from file
func NewTfvarsFile(filename string) (*Tfvars, error) {
	vf := &Tfvars{}
	var err error

	vf.filename = filename

	_, errExist := os.Stat(filename)
	vf.vardata, err = os.ReadFile(filename)
	if err != nil && errExist == nil {
		// if cannot read file, return nil
		return nil, err
	} else if err != nil && errExist != nil {
		// if file not exist, treat as empty
		vf.vardata = []byte("")
	}

	err = vf.convertVarsfile()
	if err != nil {
		return nil, err
	}

	return vf, nil
}

// convertVarsfile generate list of tfe.Variable from tfvars file
func (vf *Tfvars) convertVarsfile() error {

	if vf.vardata == nil {
		return errors.New("invalid vardata")
	}
	p := hclparse.NewParser()
	f, diags := p.ParseHCL(vf.vardata, vf.filename)
	if diags.HasErrors() {
		return errors.New(diags.Error())
	}

	vf.vars = []*tfe.Variable{}
	attrs, _ := f.Body.JustAttributes()
	for _, attr := range SortAttributes(attrs) {
		val, _ := attr.Expr.Value(nil)
		vf.vars = append(vf.vars, &tfe.Variable{
			Key:   attr.Name,
			Value: String(val),
			HCL:   !IsPrimitive(val),
		})
	}

	return nil
}

// convertTfeVariables generate tfvars file from list of tfe.Varialbe
func (vf *Tfvars) convertTfeVariables() error {
	if vf.vars == nil {
		return errors.New("tfe variable is nil")
	}

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	for _, v := range vf.vars {
		if v.Key == "" {
			return errors.New("invalid key specified")
		}
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

	vf.filename = "generated.tfvars"
	vf.vardata = f.Bytes()

	return nil
}

// BuildHCLFile return hclwrite.File format
func (vf *Tfvars) BuildHCLFile() (*hclwrite.File, error) {
	f, diags := hclwrite.ParseConfig(vf.vardata, vf.filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, errors.New(diags.Error())
	}

	return f, nil
}

// BuildHCLFileString return string of tfvars file contents
func (vf *Tfvars) BuildHCLFileString() string {
	file, err := vf.BuildHCLFile()
	if err != nil {
		return ""
	}

	return string(file.Bytes())
}
