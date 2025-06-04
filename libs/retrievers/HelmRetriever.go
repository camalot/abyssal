package retrievers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/camalot/abyssal/models/abyssal"
	"github.com/camalot/abyssal/libs/templates"
	"github.com/hashicorp/go-version"
	yq "github.com/mikefarah/yq/v4/pkg/yqlib"
	"github.com/sirupsen/logrus"
	"gopkg.in/op/go-logging.v1"
)


type HelmRetriever struct {

}


// NewHelmRetriever creates a new HelmRetriever instance
func NewHelmRetriever() *HelmRetriever {
	return &HelmRetriever{}
}

func getURLContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
			return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func getFileContent(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return content, nil
}


func cleanValueForVersion(value string) string {
	// Remove any leading or trailing whitespace
	value = strings.TrimSpace(value)

	// Remove any leading 'v' character
	if strings.HasPrefix(value, "v") {
		value = value[1:]
	}

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

func getYAMLFiles(path string) ([]string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var yamlFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml") {
			yamlFiles = append(yamlFiles, file.Name())
		}
	}
	return yamlFiles, nil
}

// CheckVersionOutOfDate checks if the Helm package is out of date
func (h *HelmRetriever) CheckVersionOutOfDate(pkg interface{}) (
	needsUpdate bool, current string, expected string, error error,
) {
	logging.SetLevel(logging.CRITICAL, "")

	// convert pkg to the appropriate type if necessary

	// For example, if pkg is expected to be of type AbyssalPackage:
	abyssalPkg, ok := pkg.(abyssal.AbyssalPackage)
	if !ok {
		return false, "", "", fmt.Errorf("expected AbyssalPackage, got %T", pkg)
	}

	// pull repo data
	entriesUrl, err := url.JoinPath(abyssalPkg.RepoUrl, "index.yaml")
	if err != nil {
		return false, "", "", fmt.Errorf("failed to join path: %w", err)
	}

	// Fetch the index.yaml file from the Helm repository
	entriesYaml, err := getURLContent(entriesUrl)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to fetch index.yaml from %s: %w", entriesUrl, err)
	}

	

	// use github.com/mikefarah/yq/v4 to query the index.yaml file

	ymlPrefs := &yq.YamlPreferences{
		Indent: 2,
		EvaluateTogether: false,
		UnwrapScalar: false,
		ColorsEnabled: !true, // Disable colors for output
		PrintDocSeparators: false,
	}


	encoder := yq.NewYamlEncoder(*ymlPrefs)
	decoder := yq.NewYamlDecoder(*ymlPrefs)
	// query := fmt.Sprintf(".entries[\"%s\"][0] | .version", abyssalPkg.ChartName)
	queryTemplate := templates.NewTemplate("query", abyssalPkg.HelmSelector, abyssalPkg)
	query, err := queryTemplate.Render()
	if err != nil {
		return false, "", "", fmt.Errorf("failed to render query template: %w", err)
	}

	// logrus.Debugf("Evaluating query '%s' on index.yaml", query)
	evaluator := yq.NewStringEvaluator()
	result, err := evaluator.Evaluate(query, string(entriesYaml), encoder, decoder)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to evaluate query '%s' on index.yaml: %w", query, err)
	}
	// trim the result to get the version string
	if result == "" {
		return false, "", "", fmt.Errorf("no result found for query '%s' on index.yaml", query)
	}
	
	result = cleanValueForVersion(strings.Trim(strings.TrimSpace(result), "\n"))
	
	expectedVersion, err := version.NewVersion(result)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to parse version '%s': %w", result, err)
	}

	// get data info
	// read all yaml in path
	logrus.Debugf("Reading all YAML files in path: %s", abyssalPkg.Directory)
	yamlFiles, err := getYAMLFiles(abyssalPkg.Directory)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to read YAML files from %s: %w", abyssalPkg.Directory, err)
	}

	logrus.Debugf("Found YAML files: %v", yamlFiles)

	logrus.Debugf("Rendering selector template for package: %s", abyssalPkg.ValueSelector)
	selectorTemplate := templates.NewTemplate("selector", abyssalPkg.ValueSelector, abyssalPkg)
	selector, err := selectorTemplate.Render()
	if err != nil {
		return false, "", "", fmt.Errorf("failed to render selector template: %w", err)
	}
	logrus.Debugf("Selector template rendered: %s", selector)


	// combine all YAML files into a single string
	var combinedYAML string
	for _, file := range yamlFiles {
		filePath := fmt.Sprintf("%s/%s", abyssalPkg.Directory, file)
		content, err := getFileContent(filePath)
		if err != nil {
			return false, "", "", fmt.Errorf("failed to read file %s: %w", filePath, err)
		}
		// if content ! starts with "---", add it
		if !strings.HasPrefix(string(content), "---") {
			combinedYAML += "---\n"
		}
		// append the content of the file
		combinedYAML += string(content) + "\n"
	}

	result, err = evaluator.Evaluate(selector, combinedYAML, encoder, decoder)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to evaluate selector '%s' on combined YAML files: %w", selector, err)
	}
	result = cleanValueForVersion(strings.Trim(strings.TrimSpace(result), "\n"))
	if result == "" {
		return false, "", "", fmt.Errorf("no result found for selector '%s' on combined YAML files", selector)
	}
	logrus.Debugf("Selector '%s' evaluated on combined YAML files, result: %s", selector, result)

	currentVersion, err := version.NewVersion(result)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to parse current version '%s': %w", result, err)
	}

	logrus.Debugf("Result of query '%s': %v", query, result)
	// Implement the logic to check if the Helm package is out of date

	if currentVersion.LessThan(expectedVersion) {
		logrus.Debugf("%s is out of date: current version %s, expected version %s", abyssalPkg.Name, currentVersion.String(), expectedVersion.String())
		return true, currentVersion.String(), expectedVersion.String(), nil
	} else if currentVersion.Equal(expectedVersion) {
		logrus.Debugf("%s is up to date: current version %s, expected version %s", abyssalPkg.Name, currentVersion.String(), expectedVersion.String())
		return false, currentVersion.String(), expectedVersion.String(), nil
	} else {
		logrus.Debugf("%s has a newer version: current version %s, expected version %s", abyssalPkg.Name, currentVersion.String(), expectedVersion.String())
		return false, currentVersion.String(), expectedVersion.String(), nil
	}
}