package main

import (
	"log"
	"os"

	"github.com/drone-plugins/drone-plugin-lib/urfave"
	_ "github.com/joho/godotenv/autoload"
	"github.com/luthermonson/drone-docker-image-shas/plugin"
	"github.com/urfave/cli/v2"
)

var version string // build number set at compile-time

func main() {
	app := &cli.App{
		Name:    "my plugin",
		Usage:   "my plugin usage",
		Action:  run,
		Version: version,
		Flags:   append(urfave.Flags(), plugin.Flags...),
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	return plugin.Exec(c, urfave.PipelineFromContext(c))
}
