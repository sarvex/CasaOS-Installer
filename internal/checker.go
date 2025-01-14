package internal

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/go-resty/resty/v2"
	"gopkg.in/yaml.v3"
)

var client = resty.New()

func init() {
	client.
		SetRetryCount(3).
		SetRetryWaitTime(5 * time.Second).
		SetRetryMaxWaitTime(20 * time.Second)
}

func WriteReleaseToLocal(release *codegen.Release, releasePath string) error {
	// create the yaml file

	os.MkdirAll(filepath.Dir(releasePath), 0755)
	f, err := os.Create(releasePath)
	defer f.Close()

	if err != nil {
		return err
	}

	releaseContent, err := yaml.Marshal(release)
	if err != nil {
		return err
	}

	f.Write([]byte(releaseContent))
	return nil
}
func GetReleaseFromLocal(releasePath string) (*codegen.Release, error) {
	// open the yaml file
	f, err := os.Open(releasePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// decode the yaml file
	var release codegen.Release
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func GetReleaseFromContent(content []byte) (*codegen.Release, error) {
	// decode the yaml file
	var release codegen.Release
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	if err := decoder.Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func GetReleaseFrom(ctx context.Context, releaseURL string) (*codegen.Release, error) {
	// download content from releaseURL
	response, err := client.R().SetContext(ctx).Get(releaseURL)
	if err != nil {
		return nil, err
	}

	if response.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get release from %s - %s", releaseURL, response.Status())
	}

	// parse release
	var release codegen.Release

	if err := yaml.Unmarshal(response.Body(), &release); err != nil {
		return nil, err
	}

	return &release, nil
}

func GetChecksums(filepath string) (map[string]string, error) {
	println("sum filepath", filepath)
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)

	checksums := map[string]string{}

	// get checksums
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)

		if len(fields) != 2 {
			continue
		}

		checksums[fields[1]] = fields[0]
	}

	return checksums, nil
}

func GetChecksumsURL(release codegen.Release, mirror string) string {
	return strings.TrimSuffix(mirror, "/") + release.Checksums
}
