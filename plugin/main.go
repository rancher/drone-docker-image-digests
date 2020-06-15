package plugin

import (
	"fmt"
	"strings"

	"github.com/drone-plugins/drone-plugin-lib/drone"
	"github.com/google/go-github/v31/github"
	"github.com/urfave/cli/v2"
)

var Flags = []cli.Flag{
	&cli.StringFlag{
		Name:    "github-repository",
		EnvVars: []string{"PLUGIN_GITHUB_REPOSITORY"},
	},
	&cli.StringFlag{
		Name:    "github-token",
		EnvVars: []string{"PLUGIN_GITHUB_TOKEN"},
	},
	&cli.StringFlag{
		Name:    "release-tag",
		EnvVars: []string{"PLUGIN_RELEASE_TAG"},
	},
	&cli.StringFlag{
		Name:    "input-file",
		EnvVars: []string{"PLUGIN_INPUT_FILE"},
		Value:   "images.txt",
	},
	&cli.StringFlag{
		Name:    "output-file",
		EnvVars: []string{"PLUGIN_OUTPUT_FILE"},
		Value:   "images-digests.txt",
	},
}

type Settings struct {
	GithubRepository string
	ReleaseTag       string
	InputFile        string
	OutputFile       string
}

func NewSettingsFromContext(c *cli.Context) Settings {
	return Settings{
		GithubRepository: c.String("github-repository"),
		ReleaseTag:       c.String("release-tag"),
		InputFile:        c.String("input-file"),
		OutputFile:       c.String("output-file"),
	}
}

func Exec(c *cli.Context, pipeline drone.Pipeline) error {
	settings := NewSettingsFromContext(c)
	client := github.NewClient(nil)

	splitRepository := strings.Split(settings.GithubRepository, "/")
	release, _, err := client.Repositories.GetReleaseByTag(c.Context, splitRepository[0], splitRepository[1], settings.ReleaseTag)
	if err != nil {
		return err
	}

	fmt.Printf("%+v", release.Assets)
	return nil
}
