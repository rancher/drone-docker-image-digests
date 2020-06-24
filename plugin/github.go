package plugin

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v31/github"
	"golang.org/x/oauth2"
)

func NewGhClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func getGhReleaseAndImageList(settings Settings) (*github.RepositoryRelease, []string, error) {
	var list []string
	release, _, err := ghClient.Repositories.GetReleaseByTag(ctx, settings.GithubOwner, settings.GithubRepo, settings.GithubTag)
	if err != nil {
		return release, list, err
	}

	for _, asset := range release.Assets {
		if asset.GetName() == settings.InputFile {
			if list, err = getLinesFromAsset(asset); err != nil {
				return release, list, err
			}
			break
		}
	}

	if len(list) == 0 {
		return release, list, fmt.Errorf("no outputFile %s found or contents were empty, can not proceed", settings.InputFile)
	}

	for k, im := range list {
		list[k] = cleanImage(im, settings.Registry)
	}

	return release, list, nil
}

func deleteGhExistingOutputAsset(assets []*github.ReleaseAsset, settings Settings) error {
	for _, asset := range assets {
		if asset.GetName() == settings.OutputFile {
			// delete output if it exists
			_, err := ghClient.Repositories.DeleteReleaseAsset(
				ctx, settings.GithubOwner, settings.GithubRepo, asset.GetID())
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func createGhReleaseAsset(release *github.RepositoryRelease, settings Settings, contents fmt.Stringer) (err error) {
	var tempFile *os.File
	if tempFile, err = ioutil.TempFile("", "*.txt"); err != nil {
		return err
	}

	defer func() {
		err = os.Remove(tempFile.Name())
	}()

	if _, err = tempFile.WriteString(contents.String()); err != nil {
		return err
	}

	if _, err = tempFile.Seek(0, io.SeekStart); err != nil {
		return err
	}

	_, _, err = ghClient.Repositories.UploadReleaseAsset(
		ctx,
		settings.GithubOwner,
		settings.GithubRepo,
		release.GetID(),
		&github.UploadOptions{
			Name: settings.OutputFile,
		},
		tempFile)
	return
}

func getLinesFromAsset(asset *github.ReleaseAsset) ([]string, error) {
	res, err := http.Get(asset.GetBrowserDownloadURL())
	if err != nil {
		return []string{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []string{}, err
	}

	if len(body) == 0 {
		return []string{}, fmt.Errorf("contents of %s was empty", asset.GetName())
	}

	list := strings.Split(string(body), "\n")

	return list, nil
}
