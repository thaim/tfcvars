package main

import (
	"context"
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/urfave/cli/v2"
)

func pull(c *cli.Context) error {
	ctx := context.Background()
	log.Println("pull command")

	config := &tfe.Config{
		Token: c.String("tfetoken"),
	}
	tfeClient, err := tfe.NewClient(config)
	if err != nil {
		log.Fatal(err)
		return err
	}

	orgs, err := tfeClient.Organizations.List(ctx, tfe.OrganizationListOptions{})
	if err != nil {
		log.Fatal(err)
		return err
	}

	for _, o := range orgs.Items {
		fmt.Println("Name: " + o.Name)
		fmt.Println("Email: " + o.Email)
	}

	return nil
}
