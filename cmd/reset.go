package cmd

import "github.com/urfave/cli/v2"

var cmdReset = &cli.Command{
	Name:  "reset",
	Usage: "reset a cluster",
	Flags: []cli.Flag{
		configFlag,
		debugFlag,
		forceFlag,
	},
	Before: actions(initLogging),
	Action: func(c *cli.Context) error {
		// read the cluster config
		cfg, err := initBlueprint(c)
		if err != nil {
			return err
		}

		switch cfg.Spec.Kubernetes.Provider {
		case "k0s":
			return cmdWrapper(c)
		case "kind":
			return cmdRun("kind", "delete", "clusters", cfg.Metadata.Name)
		}

		return nil
	},
}
