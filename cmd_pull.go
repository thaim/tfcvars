package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/urfave/cli/v2"
)

type PullOption struct {
}

func NewPullOption(c *cli.Context) *PullOption {
	var opt = &PullOption{}

	return opt
}

func Pull(c *cli.Context) error {
	ctx := context.Background()
	log.Println("pull command")

	tfeClient, err := buildClient(c)
	if err != nil {
		log.Fatal(err)
		return err
	}
	showOpt := NewPullOption(c)

	return pull(ctx, tfeClient, showOpt)
}

func pull(ctx context.Context, tfeClient *tfe.Client, pullOpt *PullOption) error {
	w, err := tfeClient.Workspaces.Read(ctx, organization, workspaceName)
	if err != nil {
		log.Fatal(err)
		return err
	}

	vars, err := tfeClient.Variables.List(ctx, w.ID, tfe.VariableListOptions{})
	if err != nil {
		log.Fatal(err)
		return err
	}
	for _, v := range(vars.Items) {
		fmt.Println("Key: " + v.Key)
		fmt.Println("Value: " + v.Value)
		fmt.Println("Description: " + v.Description)
		fmt.Println("Sensitive: " + strconv.FormatBool(v.Sensitive))
		fmt.Println()
	}

	return nil
}
