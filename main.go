package main

import (
	"fmt"
	"log"
	"os"

	"github.com/drone-plugins/drone-plugin-lib/urfave"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rancher/drone-docker-image-digests/plugin"
	"github.com/urfave/cli/v2"
)

var (
	Version   = "v0.0.0-dev"
	GitCommit = "HEAD"
)

func main() {
	app := &cli.App{
		Name:    "Docker image digests Drone plugin",
		Usage:   "drone-docker-image-digests",
		Action:  run,
		Version: fmt.Sprintf("%s (%s)", Version, GitCommit),
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
