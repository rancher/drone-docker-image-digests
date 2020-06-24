package plugin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
)

type Status struct {
	Status string `json:"status,omitempty"`
}

type Digests map[string]string

func (d Digests) String() string {
	var o strings.Builder
	keys := make([]string, 0, len(d))
	for k := range d {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&o, "%s %s\n", k, d[k])
	}
	return o.String()
}

func cleanImage(image, registry string) string {
	if image == "" {
		return ""
	}

	switch registry {
	case "docker.io":
		if len(strings.Split(image, "/")) == 1 {
			image = path.Join("library", image)
		}
	}

	return path.Join(registry, image)
}

func getDigests(list []string, threads int) Digests {
	var wg sync.WaitGroup
	sem := make(chan struct{}, threads)
	wg.Add(len(list))

	var mutex sync.Mutex
	var digests Digests = make(map[string]string)

	for _, image := range list {
		im := image
		go func(im string) {
			image := im
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if image == "" {
				return
			}

			fmt.Printf("pulling image digest for %s\n", image)
			digest, err := getImageDigest(image)
			if err != nil {
				log.Println(err)
			}

			mutex.Lock()
			digests[im] = digest
			mutex.Unlock()
		}(im)
	}
	wg.Wait()

	return digests
}

// pullImage pulls a container image and outputs progress if --verbose flag is set
func getImageDigest(image string) (string, error) {
	var digest string
	resp, err := dClient.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return digest, err
	}
	defer resp.Close()

	scanner := bufio.NewScanner(resp)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var status *Status
		if err := json.Unmarshal(scanner.Bytes(), &status); err != nil {
			return "", err
		}
		if strings.Contains(status.Status, "Digest:") {
			sp := strings.Split(status.Status, "Digest:")
			digest = strings.TrimSpace(sp[1])
		}
	}
	if err := scanner.Err(); err != nil {
		return digest, err
	}

	return digest, nil
}
