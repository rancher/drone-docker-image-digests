package plugin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/drone-plugins/drone-plugin-lib/drone"
	dockerclient "github.com/moby/moby/client"
	"github.com/urfave/cli/v2"
)

var Flags = []cli.Flag{
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
	&cli.StringFlag{
		Name:    "artifacts-base-url",
		EnvVars: []string{"PLUGIN_ARTIFACTS_BASE_URL"},
		Value:   "https://prime.ribs.rancher.io",
	},
	&cli.IntFlag{
		Name:    "threads",
		EnvVars: []string{"PLUGIN_THREADS"},
		Value:   1,
	},
}

type Settings struct {
	GithubTag        string
	InputFile        string
	OutputFile       string
	Registry         string
	ArtifactsBaseURL string
	Threads          int
}

var (
	ctx     context.Context
	dClient *dockerclient.Client
)

func NewSettingsFromContext(c *cli.Context) (Settings, error) {
	settings := Settings{
		GithubTag:        c.String("github-tag"),
		InputFile:        c.String("input-file"),
		OutputFile:       c.String("output-file"),
		Registry:         c.String("registry"),
		ArtifactsBaseURL: c.String("artifacts-base-url"),
		Threads:          c.Int("threads"),
	}

	// package variables
	ctx = c.Context
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

	fmt.Println("getting image list from input")
	list, err := getImageList(settings)
	if err != nil {
		return err
	}

	fmt.Printf("pulling image digests for %d images pulled from %s using %d threads\n",
		len(list), settings.InputFile, settings.Threads)
	digests := getDigests(list, settings.Threads)

	fmt.Printf("writing release asset file %s with %d digests\n", settings.OutputFile, len(digests))

	return createAssetFile(settings, digests)
}

func getImageList(settings Settings) ([]string, error) {
	client := http.Client{Timeout: time.Second * 15}
	res, err := client.Get(settings.ArtifactsBaseURL + "/rancher/" + settings.GithubTag + "/" + settings.InputFile)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	list, err := getLinesFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return list, fmt.Errorf("no outputFile %s found or contents were empty, can not proceed", settings.InputFile)
	}

	for k, im := range list {
		list[k] = cleanImage(im, settings.Registry)
	}

	return list, nil
}

func createAssetFile(settings Settings, contents fmt.Stringer) error {
	return os.WriteFile(settings.OutputFile, []byte(contents.String()), 0644)
}

func getLinesFromReader(body io.Reader) ([]string, error) {
	lines, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return []string{}, errors.New("file was empty")
	}

	return strings.Split(string(lines), "\n"), nil
}
