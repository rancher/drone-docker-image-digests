package plugin

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/drone-plugins/drone-plugin-lib/drone"
	"github.com/google/go-github/v31/github"
	dockerclient "github.com/moby/moby/client"
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
		Name:    "github-tag",
		EnvVars: []string{"PLUGIN_GITHUB_TAG"},
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
	&cli.StringFlag{
		Name:    "registry",
		EnvVars: []string{"PLUGIN_REGISTRY"},
		Value:   "docker.io",
	},
	&cli.IntFlag{
		Name:    "threads",
		EnvVars: []string{"PLUGIN_THREADS"},
		Value:   1,
	},
}

type Settings struct {
	GithubRepository string
	GithubOwner      string
	GithubRepo       string
	GithubToken      string
	GithubTag        string
	InputFile        string
	OutputFile       string
	Registry         string
	Threads          int
}

var (
	ctx      context.Context
	ghClient *github.Client
	dClient  *dockerclient.Client
)

func NewSettingsFromContext(c *cli.Context) (Settings, error) {
	settings := Settings{
		GithubRepository: c.String("github-repository"),
		GithubToken:      c.String("github-token"),
		GithubTag:        c.String("github-tag"),
		InputFile:        c.String("input-file"),
		OutputFile:       c.String("output-file"),
		Registry:         c.String("registry"),
		Threads:          c.Int("threads"),
	}

	if settings.GithubToken == "" {
		return settings, errors.New("github token required")
	}

	splitRepository := strings.Split(settings.GithubRepository, "/")
	settings.GithubOwner = splitRepository[0]
	settings.GithubRepo = splitRepository[1]

	// package variables
	ctx = c.Context
	ghClient = NewGhClient(c.Context, settings.GithubToken)
	dockerclient, err := dockerclient.NewEnvClient()
	if err != nil {
		return settings, err
	}
	dClient = dockerclient

	return settings, nil
}

func Exec(c *cli.Context, pipeline drone.Pipeline) error {
	settings, err := NewSettingsFromContext(c)
	if err != nil {
		return err
	}

	fmt.Printf("pulling release from github: %s@%s\n", settings.GithubRepository, settings.GithubTag)
	release, list, err := getGhReleaseAndImageList(settings)
	if err != nil {
		return err
	}

	fmt.Printf("pulling image digests for %d images pulled from %s using %d threads\n",
		len(list), settings.InputFile, settings.Threads)
	digests := getDigests(list, settings.Threads)

	fmt.Printf("deleting digests asset %s from the release if it already exists\n",
		settings.OutputFile)
	if err := deleteGhExistingOutputAsset(release.Assets, settings); err != nil {
		return err
	}

	fmt.Printf("writing release asset file %s with %d digests\n", settings.OutputFile, len(digests))
	return createGhReleaseAsset(release, settings, digests)
}
