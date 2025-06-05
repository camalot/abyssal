package providers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	// "github.com/camalot/abyssal/models/abyssal"
	"github.com/camalot/abyssal/libs/templates"
	"github.com/camalot/abyssal/models/helm"
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"gopkg.in/op/go-logging.v1"
	"gopkg.in/yaml.v3"

	// "gopkg.in/yaml.v3"
	yq "github.com/mikefarah/yq/v4/pkg/yqlib"
)

type HelmProvider struct {
	Directory string                 `yaml:"directory"`
	Selector  string                 `yaml:"selector"`
	Targets   []helm.HelmChartTarget `yaml:"targets"`

	HelmSelector  string `yaml:"helmSelector"`
}

func NewHelmProvider() *HelmProvider {
	return &HelmProvider{}
}

func (p *HelmProvider) getFilesRecursive(dir string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			subFiles, err := p.getFilesRecursive(dir + "/" + entry.Name())
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
		} else if entry.Type().IsRegular() && (strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".yml")) {
			files = append(files, path.Join(dir, entry.Name()))
		}
	}
	return files, nil
}

func (p *HelmProvider) GetFiles() []string {
	files, err := p.getFilesRecursive(p.Directory)
	if err != nil {
		return nil
	}
	return files
}

func (p *HelmProvider) Load() error {

	// get all yaml files in the directory
	// parse each file and load the targets
	targets := []helm.HelmChartTarget{}
	files := p.GetFiles()
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		// parse the content and load the targets
		var chart []helm.HelmChartTarget

		ymlPrefs := &yq.YamlPreferences{
			Indent:             2,
			EvaluateTogether:   false,
			UnwrapScalar:       false,
			ColorsEnabled:      !true, // Disable colors for output
			PrintDocSeparators: false,
		}

		logging.SetLevel(logging.CRITICAL, "") // Set logging level to critical to suppress yq logs
		encoder := yq.NewYamlEncoder(*ymlPrefs)
		decoder := yq.NewYamlDecoder(*ymlPrefs)
		query := p.Selector

		// logrus.Debugf("Evaluating query '%s' on index.yaml", query)
		evaluator := yq.NewStringEvaluator()
		result, err := evaluator.Evaluate(query, string(content), encoder, decoder)
		if err != nil {
			return fmt.Errorf("failed to evaluate query '%s' on index.yaml: %w", query, err)
		}
		// trim the result to get the version string
		if result == "" {
			return fmt.Errorf("no result found for query '%s' on index.yaml", query)
		}

		// result is a YAML String of []HelmChartTarget
		err = yaml.Unmarshal([]byte(result), &chart)
		if err != nil {
			return fmt.Errorf("failed to unmarshal YAML for query '%s': %w", query, err)
		}

		targets = append(targets, chart...)
	}
	p.Targets = targets
	return nil
}

func (p *HelmProvider) CheckVersionOutOfDate(target helm.HelmChartTarget) (bool, string, string, error) {
	// This function should implement the logic to check if the version is out of date
	// For now, we will just return false, empty strings and nil error
	// You can implement the actual logic based on your requirements
	// pull repo data
	entriesUrl, err := url.JoinPath(strings.TrimSpace(target.RepoURL), "index.yaml")
	if err != nil {
		return false, "", "", fmt.Errorf("failed to join path: %w", err)
	}

	// Fetch the index.yaml file from the Helm repository
	// this shoudl cache the index.yaml file in the future
	entriesYaml, err := getURLContent(entriesUrl)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to fetch index.yaml from %s: %w", entriesUrl, err)
	}
	queryTemplate := templates.NewTemplate("query", p.HelmSelector, target)
	query, err := queryTemplate.Render()
	if err != nil {
		return false, "", "", fmt.Errorf("failed to render query template: %w", err)
	}

	ymlPrefs := &yq.YamlPreferences{
		Indent: 2,
		EvaluateTogether: false,
		UnwrapScalar: false,
		ColorsEnabled: !true, // Disable colors for output
		PrintDocSeparators: false,
		LeadingContentPreProcessing: false,
	}
	encoder := yq.NewYamlEncoder(*ymlPrefs)
	decoder := yq.NewYamlDecoder(*ymlPrefs)

	// logrus.Debugf("Evaluating query '%s' on index.yaml", query)
	evaluator := yq.NewStringEvaluator()
	result, err := evaluator.Evaluate(query, string(entriesYaml), encoder, decoder)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to evaluate query '%s' on index.yaml: %w", query, err)
	}
	// trim the result to get the version string
	if result == "" {
		fmt.Println(string(entriesYaml))
		return false, "", "", fmt.Errorf("no result found for query '%s' on index.yaml", query)
	}
	
	result = cleanValueForVersion(strings.Trim(strings.TrimSpace(result), "\n"))
	
	expectedVersion, err := version.NewVersion(result)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to parse version '%s': %w", result, err)
	}
	currentVersion, err := version.NewVersion(target.TargetRevision)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to parse current version '%s': %w", target.TargetRevision, err)
	}
	if currentVersion.LessThan(expectedVersion) {
		logrus.Debugf("%s is out of date: current version %s, expected version %s", target.ChartName, currentVersion.String(), expectedVersion.String())
		return true, currentVersion.String(), expectedVersion.String(), nil
	} else if currentVersion.Equal(expectedVersion) {
		logrus.Debugf("%s is up to date: current version %s, expected version %s", target.ChartName, currentVersion.String(), expectedVersion.String())
		return false, currentVersion.String(), expectedVersion.String(), nil
	} else {
		logrus.Debugf("%s has a newer version: current version %s, expected version %s", target.ChartName, currentVersion.String(), expectedVersion.String())
		return false, currentVersion.String(), expectedVersion.String(), nil
	}
}



func getURLContent(url string) ([]byte, error) {

	client := &http.Client{
		Timeout: time.Duration(time.Duration(10).Seconds()), // Set a timeout for the HTTP request
	}

	resp, err := client.Get(url)
	if err != nil {
			return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func cleanValueForVersion(value string) string {
	// Remove any leading or trailing whitespace
	value = strings.TrimSpace(value)

	// Remove any leading 'v' character
	value = strings.TrimPrefix(value, "v")

	// remove any quotes
	value = strings.Trim(value, "\"'")

	// Remove any trailing characters that are not digits, dots, or hyphens
	// for i := len(value) - 1; i >= 0; i-- {
	// 	if !(value[i] >= '0' && value[i] <= '9') && value[i] != '.' && value[i] != '-' {
	// 		value = value[:i]
	// 	} else {
	// 		break
	// 	}
	// }

	return value
}
