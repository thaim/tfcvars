package main

import (
	"os"
	"path/filepath"

	"github.com/antonholmquist/jason"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)


func NewTfeClient(c *cli.Context) (*tfe.Client, error) {
	token := c.String("tfetoken")
	if token == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Error().Err(err).Msg("failed to get user home directory")
			return nil, err
		}
		fp, err := os.Open(filepath.Join(home, ".terraform.d/credentials.tfrc.json"))
		if err != nil {
			log.Error().Err(err).Msg("cannot find terraform cloud credential file")
			return nil, err
		}
		defer fp.Close()

		v, err := jason.NewObjectFromReader(fp)
		if err != nil {
			log.Error().Err(err).Msg("cannot read from terraform cloud credential file")
			return nil, err
		}

		token, err = v.GetString("credentials", "app.terraform.io", "token")
		if err != nil {
			log.Error().Err(err).Msg("cannot retrieve credentails from terraform cloud credential file")
			return nil, err
		}
	}

	config := &tfe.Config{
		Token: token,
	}
	tfeClient, err := tfe.NewClient(config)
	if err != nil {
		log.Error().Err(err).Msg("cannot create tfe client")
		return nil, err
	}

	return tfeClient, nil
}
