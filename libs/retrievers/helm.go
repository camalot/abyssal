package retrievers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/camalot/abyssal/models/abyssal"
	yq "github.com/mikefarah/yq/v4/pkg/yqlib"
	"github.com/sirupsen/logrus"
)


type HelmRetriever struct {

}

// NewHelmRetriever creates a new HelmRetriever instance
func NewHelmRetriever() *HelmRetriever {
	return &HelmRetriever{}
}

func GetURLContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
			return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// OutOfDateVersion checks if the Helm package is out of date
func (h *HelmRetriever) OutOfDateVersion(pkg interface{}) (bool, string, error) {
	// convert pkg to the appropriate type if necessary

	// For example, if pkg is expected to be of type AbyssalPackage:
	abyssalPkg, ok := pkg.(abyssal.AbyssalPackage)
	if !ok {
		return false, "", fmt.Errorf("expected AbyssalPackage, got %T", pkg)
	}

	// pull repo data
	entriesUrl, err := url.JoinPath(abyssalPkg.RepoUrl, "index.yaml")
	if err != nil {
		return false, "", fmt.Errorf("failed to join path: %w", err)
	}

	// Fetch the index.yaml file from the Helm repository
	entriesYaml, err := GetURLContent(entriesUrl)
	if err != nil {
		return false, "", fmt.Errorf("failed to fetch index.yaml from %s: %w", entriesUrl, err)
	}

	// use github.com/mikefarah/yq/v4 to query the index.yaml file

	ymlPrefs := &yq.YamlPreferences{
		Indent: 2,
		EvaluateTogether: false,
		UnwrapScalar: false,
		ColorsEnabled: !true, // Disable colors for output
		PrintDocSeparators: false,
	}
	// Set yq logger to debug level
	eval := yq.NewAllAtOnceEvaluator()
	eval.EvaluateNodes()

	encoder := yq.NewYamlEncoder(*ymlPrefs)
	decoder := yq.NewYamlDecoder(*ymlPrefs)
	query := fmt.Sprintf(".entries[\"%s\"][0] | .version", abyssalPkg.ChartName)
	logrus.Debugf("Evaluating query '%s' on index.yaml: %s", query, string(entriesYaml))
	evaluator := yq.NewStringEvaluator()
	result, err := evaluator.Evaluate(query, string(entriesYaml), encoder, decoder)
	if err != nil {
		return false, "", fmt.Errorf("failed to evaluate query '%s' on index.yaml: %w", query, err)
	}

	logrus.Debugf("Result of query '%s': %v", query, result)
	// Implement the logic to check if the Helm package is out of date
	return false, result, nil
}