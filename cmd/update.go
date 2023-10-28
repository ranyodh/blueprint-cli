package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var cmdUpdate = &cli.Command{
	Name:  "update",
	Usage: "update a cluster",
	Flags: []cli.Flag{
		configFlag,
		debugFlag,
	},
	Before: actions(initLogging),
	Action: func(c *cli.Context) error {
		// read the cluster config
		cfg, err := initBlueprint(c)
		if err != nil {
			return err
		}

		log.Info("Updating Components")
		err = installComponents(cfg)
		if err != nil {
			return fmt.Errorf("failed to update components: %w", err)
		}
		return nil
	},
}
