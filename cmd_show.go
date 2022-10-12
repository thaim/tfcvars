package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/urfave/cli/v2"
)

// Show display variable list
func Show(c *cli.Context) error {
	ctx := context.Background()
	log.Println("show command")

	tfeClient, err := buildClient(c)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return show(ctx, tfeClient)
}

func show(ctx context.Context, tfeClient *tfe.Client) error {
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
