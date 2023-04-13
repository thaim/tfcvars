package main

import (
	"github.com/urfave/cli/v2"
)

type DiffOption struct {
}

func NewDiffOption(c *cli.Context) *DiffOption {
	return &DiffOption{}
}

func Diff(c *cli.Context) error {
	return nil
}
