package providers

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/camalot/abyssal/config"
	"github.com/camalot/abyssal/libs/templates"
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"gopkg.in/op/go-logging.v1"
	"gopkg.in/yaml.v3"

	yq "github.com/mikefarah/yq/v4/pkg/yqlib"
)

type ArgoAppOfAppsProvider struct {
	Directory string `yaml:"directory"`
	Selector  string `yaml:"selector"`

	Targets         []ProviderTarget `yaml:"targets"`
	Evaluator       string           `yaml:"evaluator"`
	EntriesSelector string           `yaml:"entries"`

	IncludePreRelease bool `yaml:"includePreRelease,omitempty"` // used to include pre-release versions in the evaluation

	TargetNameFrom string `yaml:"nameFrom,omitempty"` // used to set the name of the target from a field in the target map

	config   *config.AppConfiguration `yaml:"-"`
	UseCache bool                     `yaml:"-"`
}

func NewArgoAppOfAppsProvider(providerElement config.ProviderElement, config *config.AppConfiguration) *ArgoAppOfAppsProvider {
	p := &ArgoAppOfAppsProvider{
		Selector:        config.Settings.Providers.ArgoAppOfApps.BaseSelector,
		EntriesSelector: config.Settings.Providers.ArgoAppOfApps.EntriesSelector,
		Evaluator:       config.Settings.Providers.ArgoAppOfApps.EvaluatorSelector,

		TargetNameFrom:    "chartName",
		IncludePreRelease: config.Settings.Providers.ArgoAppOfApps.IncludePreRelease,

		config:   config,
		UseCache: true,
	}
	if providerElement.Extra["useCache"] != nil {
		if useCache, ok := providerElement.Extra["useCache"].(bool); ok {
			p.UseCache = useCache
		} else {
			p.UseCache = true
		}
	} else {
		p.UseCache = true // default to true if not set
	}

	if providerElement.Extra["directory"] == nil {
		p.Directory = "./"
	} else if dir, ok := providerElement.Extra["directory"].(string); ok {
		p.Directory = dir
	} else {
		p.Directory = "./" // default to current directory if not set or invalid
	}
	if providerElement.Extra["selector"] != nil {
		p.Selector = providerElement.Extra["selector"].(string)
	} else {
		p.Selector = config.Settings.Providers.ArgoAppOfApps.BaseSelector
	}
	return p
}

func (p *ArgoAppOfAppsProvider) getFilesRecursive(dir string) ([]string, error) {
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

func (p *ArgoAppOfAppsProvider) GetFiles() []string {
	files, err := p.getFilesRecursive(p.Directory)
	if err != nil {
		return nil
	}
	return files
}

func (p *ArgoAppOfAppsProvider) Load() error {
	// get all yaml files in the directory
	// parse each file and load the targets
	targets := []ProviderTarget{}
	files := p.GetFiles()
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		// parse the content and load the targets
		var target []ProviderTarget

		ymlPrefs := &yq.YamlPreferences{
			Indent:             2,
			EvaluateTogether:   false,
			UnwrapScalar:       false,
			ColorsEnabled:      !true, // Disable colors for output
			PrintDocSeparators: false,
		}

		evaluatorTemplate := templates.NewTemplate("evaluator", p.Evaluator, p)

		logging.SetLevel(logging.CRITICAL, "") // Set logging level to critical to suppress yq logs
		encoder := yq.NewYamlEncoder(*ymlPrefs)
		decoder := yq.NewYamlDecoder(*ymlPrefs)
		query, _ := evaluatorTemplate.Render()

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
		err = yaml.Unmarshal([]byte(result), &target)
		if err != nil {
			return fmt.Errorf("failed to unmarshal YAML for query '%s': %w", query, err)
		}

		// set the name of the target from the targetNameFrom field
		for i := range target {
			if p.TargetNameFrom != "" {
				if name, ok := target[i].Map[p.TargetNameFrom].(string); ok && name != "" {
					target[i].Name = name   // set the name field to the value of targetNameFrom
					target[i].Source = file // set the source field to the file path
				}
			}
		}
		targets = append(targets, target...)
	}
	p.Targets = targets
	return nil
}

func (p *ArgoAppOfAppsProvider) GetTargets() ([]ProviderTarget, error) {
	if len(p.Targets) == 0 {
		if err := p.Load(); err != nil {
			return nil, fmt.Errorf("failed to load targets: %w", err)
		}
	}

	// Convert the targets to a slice of interfaces
	targets := make([]ProviderTarget, len(p.Targets))
	for i, target := range p.Targets {
		targets[i] = target
	}
	return targets, nil
}

func (p *ArgoAppOfAppsProvider) CheckVersionOutOfDate(target ProviderTarget) (ProviderCheckResult, error) {
	repoUrl, ok := target.Map["repoURL"].(string)
	if !ok || repoUrl == "" {
		err := fmt.Errorf("failed to get repoURL from target: missing or not a string")
		return ProviderCheckResult{
			Outdated:        false,
			Target:          target,
			CurrentVersion:  "",
			ExpectedVersion: "",
			Error:           err.Error(),
			State:           ProviderCheckStateError,
		}, err
	}
	targetRevision, ok := target.Map["targetRevision"].(string)
	if !ok || targetRevision == "" {
		err := fmt.Errorf("failed to get targetRevision from target: missing or not a string")
		return ProviderCheckResult{
			Outdated:        false,
			Target:          target,
			CurrentVersion:  "",
			ExpectedVersion: "",
			Error:           err.Error(),
			State:           ProviderCheckStateError,
		}, err
	}

	// pull repo data
	entriesUrl, err := url.JoinPath(strings.TrimSpace(repoUrl), "index.yaml")
	if err != nil {
		return ProviderCheckResult{
			Outdated:        false,
			Target:          target,
			CurrentVersion:  "",
			ExpectedVersion: "",
			Error:           fmt.Sprintf("failed to join path: %v", err),
			State:           ProviderCheckStateError,
		}, fmt.Errorf("failed to join path: %w", err)
	}

	// Fetch the index.yaml file from the Helm repository
	// this should cache the index.yaml file in the future
	entriesYaml, err := p.getURLContent(strings.TrimSpace(repoUrl), "index.yaml")
	if err != nil {
		return ProviderCheckResult{
			Outdated:        false,
			Target:          target,
			CurrentVersion:  "",
			ExpectedVersion: "",
			Error:           fmt.Sprintf("failed to fetch index.yaml from %s: %v", entriesUrl, err),
			State:           ProviderCheckStateError,
		}, fmt.Errorf("failed to fetch index.yaml from %s: %w", entriesUrl, err)
	}
	queryTemplate := templates.NewTemplate("query", p.EntriesSelector, target)
	query, err := queryTemplate.Render()
	if err != nil {
		return ProviderCheckResult{
			Outdated:        false,
			Target:          target,
			CurrentVersion:  "",
			ExpectedVersion: "",
			Error:           fmt.Errorf("failed to render query template: %w", err).Error(),
			State:           ProviderCheckStateError,
		}, fmt.Errorf("failed to render query template: %w", err)
	}

	ymlPrefs := &yq.YamlPreferences{
		Indent:                      2,
		EvaluateTogether:            false,
		UnwrapScalar:                false,
		ColorsEnabled:               !true, // Disable colors for output
		PrintDocSeparators:          false,
		LeadingContentPreProcessing: false,
	}
	encoder := yq.NewYamlEncoder(*ymlPrefs)
	decoder := yq.NewYamlDecoder(*ymlPrefs)

	// logrus.Debugf("Evaluating query '%s' on index.yaml", query)
	evaluator := yq.NewStringEvaluator()
	os.Setenv("ABYSSAL_INCLUDE_PRERELEASE", fmt.Sprintf("%t", p.IncludePreRelease)) // Disable yq debug output
	result, err := evaluator.Evaluate(query, string(entriesYaml), encoder, decoder)
	os.Unsetenv("ABYSSAL_INCLUDE_PRERELEASE") // Unset the environment variable

	if err != nil {
		return ProviderCheckResult{
			Outdated:        false,
			Target:          target,
			CurrentVersion:  "",
			ExpectedVersion: "",
			Error:           fmt.Sprintf("failed to evaluate query '%s' on index.yaml: %v", query, err),
			State:           ProviderCheckStateSkipped,
		}, fmt.Errorf("failed to evaluate query '%s' on index.yaml: %w", query, err)
	}
	// trim the result to get the version string
	if result == "" {
		return ProviderCheckResult{
			Outdated:        false,
			Target:          target,
			CurrentVersion:  "",
			ExpectedVersion: "",
			Error:           fmt.Sprintf("no result found for query '%s' on index.yaml", query),
			State:           ProviderCheckStateSkipped,
		}, fmt.Errorf("no result found for query '%s' on index.yaml", query)
	}

	logrus.Debugf("Result of query '%s': %s\n", query, result)

	result = cleanValueForVersion(strings.Trim(strings.TrimSpace(result), "\n"))

	expectedVersion, err := version.NewVersion(result)
	if err != nil {
		logrus.Debugln(string(entriesYaml))
		return ProviderCheckResult{
			Outdated:        false,
			Target:          target,
			CurrentVersion:  "",
			ExpectedVersion: "",
			Error:           fmt.Sprintf("failed to parse expectedVersion '%s': %v", result, err),
			State:           ProviderCheckStateError,
		}, fmt.Errorf("failed to parse expectedVersion '%s': %w", result, err)
	}

	currentVersion, err := version.NewVersion(targetRevision)
	if err != nil {
		return ProviderCheckResult{
			Outdated:        false,
			Target:          target,
			CurrentVersion:  "",
			ExpectedVersion: "",
			Error:           fmt.Sprintf("failed to parse current version '%s': %v", targetRevision, err),
			State:           ProviderCheckStateError,
		}, fmt.Errorf("failed to parse current version '%s': %w", targetRevision, err)
	}
	if currentVersion.LessThan(expectedVersion) {
		logrus.Debugf("%s is out of date: current version %s, expected version %s", target.Name, currentVersion.String(), expectedVersion.String())
		return ProviderCheckResult{
			Outdated:        true,
			Target:          target,
			CurrentVersion:  currentVersion.String(),
			ExpectedVersion: expectedVersion.String(),
			State:           ProviderCheckStateSuccess,
		}, nil
	} else if currentVersion.Equal(expectedVersion) {
		logrus.Debugf("%s is up to date: current version %s, expected version %s", target.Name, currentVersion.String(), expectedVersion.String())
		return ProviderCheckResult{
			Outdated:        false,
			Target:          target,
			CurrentVersion:  currentVersion.String(),
			ExpectedVersion: expectedVersion.String(),
			State:           ProviderCheckStateSuccess,
		}, nil
	} else if currentVersion.GreaterThan(expectedVersion) {
		return ProviderCheckResult{
			Outdated:        false,
			Target:          target,
			CurrentVersion:  currentVersion.String(),
			ExpectedVersion: expectedVersion.String(),
			State:           ProviderCheckStateError,
		}, fmt.Errorf("%s has a newer version: current version %s, expected version %s", target.Name, currentVersion.String(), expectedVersion.String())
	} else {
		return ProviderCheckResult{
			Outdated:        true,
			Target:          target,
			CurrentVersion:  currentVersion.String(),
			ExpectedVersion: expectedVersion.String(),
			Error:           fmt.Sprintf("unexpected version comparison: current version %s, expected version %s", currentVersion.String(), expectedVersion.String()),
			State:           ProviderCheckStateError,
		}, fmt.Errorf("unexpected version comparison: current version %s, expected version %s", currentVersion.String(), expectedVersion.String())
	}
}

func (p *ArgoAppOfAppsProvider) GetName() string {
	return "Argo - App Of Apps"
}

func (p *ArgoAppOfAppsProvider) GetMarkdownTableHeader() string {
	// return "| Package | Source | Repository | Current Version | Expected Version | Status |\n" +
	// 	"|---------|--------|------------|-----------------|------------------|--------|\n"
	return `<table>
	<thead>
		<tr>
			<th>Chart Name</th>
			<th>Source</th>
			<th>Repository</th>
			<th>Current Version</th>
			<th>Expected Version</th>
			<th>Status</th>
		</tr>
	</thead>
	<tbody>`
}

func (p *ArgoAppOfAppsProvider) GetMarkdownTableFooter() string {
	return `</tbody>
</table>`
}

func (p *ArgoAppOfAppsProvider) GetMarkdownLegend() string {
	return fmt.Sprintf(`---

### Status Legend

- %s Up to date
- %s Outdated
- %s Error
- %s Skipped`, ProviderCheckStateEmojiSuccess, ProviderCheckStateEmojiFailure, ProviderCheckStateEmojiError, ProviderCheckStateEmojiSkipped)
}

func (p *ArgoAppOfAppsProvider) getStatusIcon(result ProviderCheckResult) string {
	if result.State == ProviderCheckStateError {
		return string(ProviderCheckStateEmojiError) // Error state
	} else if result.State == ProviderCheckStateSkipped {
		return string(ProviderCheckStateEmojiSkipped) // Skipped state
	} else if result.State == ProviderCheckStateSuccess && !result.Outdated {
		return string(ProviderCheckStateEmojiSuccess) // Up to date state
	} else if result.State == ProviderCheckStateSuccess && result.Outdated {
		return string(ProviderCheckStateEmojiFailure) // Outdated state
	} else {
		return string(ProviderCheckStateEmojiUnknown) // Unknown state
	}
}

func (p *ArgoAppOfAppsProvider) GetMarkdownTableRow(result ProviderCheckResult) string {
	status := p.getStatusIcon(result)
	rows := strings.Builder{}
	row := fmt.Sprintf(`
		<tr>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
		</tr>`,
		result.Target.Name,
		result.Target.Source,
		result.Target.Map["repoURL"],
		result.CurrentVersion,
		result.ExpectedVersion,
		status,
	)

	rows.WriteString(row)
	if result.Error != "" {
		errorRow := fmt.Sprintf(`
		<tr>
			<td colspan="6" style="color: red;">%s %s</td>
		</tr>`, status, result.Error)
		rows.WriteString(errorRow)
	}
	return rows.String()
}

func (p *ArgoAppOfAppsProvider) writeCachedContent(cacheKey []byte, content []byte) error {
	if !p.UseCache {
		logrus.Debugf("Cache is disabled, not writing content for %x", cacheKey)
		return nil // return nil to indicate that we are not caching
	}
	cacheFilePath := path.Join(os.TempDir(), fmt.Sprintf("%x.yaml", cacheKey))
	// write the content to the cache file
	if err := os.WriteFile(cacheFilePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write cache file %s: %w", cacheFilePath, err)
	}
	logrus.Debugf("Cached content for %x at %s", cacheKey, cacheFilePath)
	return nil
}

func (p *ArgoAppOfAppsProvider) getCachedContent(cacheKey []byte, contentURL string) ([]byte, error) {
	if !p.UseCache {
		logrus.Debugf("Cache is disabled, fetching content from %s", contentURL)
		return nil, nil // return nil to indicate that we need to fetch the content
	}

	cacheFilePath := path.Join(os.TempDir(), fmt.Sprintf("%x.yaml", cacheKey))
	// check if the cache file exists
	if fileInfo, err := os.Stat(cacheFilePath); err == nil {
		// check if the cache file is older than 24 hours
		if time.Since(fileInfo.ModTime()) < 24*time.Hour {
			// read the cache file
			content, err := os.ReadFile(cacheFilePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read cache file %s: %w", cacheFilePath, err)
			}
			logrus.Debugf("Cache hit for %s, returning cached content", contentURL)
			return content, nil
		} else {
			logrus.Debugf("Cache file %s is older than 24 hours, fetching new content", cacheFilePath)
			// if the cache file is older than 24 hours, delete it
			if err := os.Remove(cacheFilePath); err != nil {
				return nil, fmt.Errorf("failed to remove old cache file %s: %w", cacheFilePath, err)
			}
		}
	}

	logrus.Debugf("Cache miss for %s, fetching from URL", contentURL)
	return nil, nil // return nil to indicate that we need to fetch the content
}

func (p *ArgoAppOfAppsProvider) getURLContent(baseUrl string, pathSegments ...string) ([]byte, error) {
	// Normalize the URL to ensure it matches the keys in the authentication map
	normalizedBaseUrl, _ := url.Parse(baseUrl)
	contentURL := normalizedBaseUrl.JoinPath(pathSegments...) // Join the base URL with the path segments

	// create MD5 hash of the URL to use as a cache key
	cacheKey := md5.Sum([]byte(contentURL.String()))

	content, err := p.getCachedContent(cacheKey[:], contentURL.String())
	if content != nil {
		return content, err
	}

	client := &http.Client{
		Timeout: 10 * time.Second, // Set a timeout for the HTTP request
	}

	// do we have auth info for this URL?
	// check if p.config.Settings.Authentication is not nil and has a key for the URL
	if p.config.Settings.Authentication != nil {
		authType := ""
		// check if Authentication map has an entry for this URL
		auth, ok := p.config.Settings.Authentication[normalizedBaseUrl.String()]
		if ok {
			authType = strings.ToLower(auth.Type)
			switch authType {
			case "basic":
				client = &http.Client{
					Transport: &basicAuthTransport{
						base:     http.DefaultTransport,
						username: p.envTemplateValue(auth.Username),
						password: p.envTemplateValue(auth.Password),
					},
					Timeout: 10 * time.Second,
				}

			case "token":
				client = &http.Client{
					Transport: &authTransport{
						base:       http.DefaultTransport,
						authHeader: fmt.Sprintf("Token %s", p.envTemplateValue(auth.Token)),
					},
					Timeout: 10 * time.Second,
				}

			case "bearer":
				token := p.envTemplateValue(auth.Token)
				if token == "" {
					return nil, fmt.Errorf("bearer token is empty for %s", normalizedBaseUrl.String())
				}
				logrus.Debugf("Using Bearer token for %s", normalizedBaseUrl.String())
				client = &http.Client{
					Transport: &authTransport{
						base:       http.DefaultTransport,
						authHeader: fmt.Sprintf("Bearer %s", token),
					},
					Timeout: 10 * time.Second,
				}
			default:
			}
		}
	}

	resp, err := client.Get(contentURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	finalContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", contentURL.String(), err)
	}
	err = p.writeCachedContent(cacheKey[:], finalContent)
	if err != nil {
		logrus.Errorf("Failed to write cache content for %s: %v", contentURL.String(), err)
	}
	return finalContent, nil
}

type ArgoAppOfAppsProviderAuthenticationTemplateData struct {
	EnvironmentVariables map[string]string
}

func (p *ArgoAppOfAppsProvider) envTemplateValue(template string) string {
	// This function should implement the logic to render a template with environment variables
	// create a map of environment variables from os.Environ
	envVars := make(map[string]string)

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			// fmt.Printf("Adding env var: %s=%s\n", parts[0], parts[1])
			envVars[parts[0]] = parts[1]
		}
	}
	result := templates.NewTemplate("templated-value", template, &ArgoAppOfAppsProviderAuthenticationTemplateData{
		EnvironmentVariables: envVars,
	})
	rendered, err := result.Render()
	if err != nil {
		return template
	}
	logrus.Debugf("Rendered template '%s' to '%s'", template, rendered[:5])
	return rendered
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

type authTransport struct {
	base       http.RoundTripper
	authHeader string
}

func (a *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", a.authHeader)
	return a.base.RoundTrip(req)
}

// For Basic Auth
type basicAuthTransport struct {
	base     http.RoundTripper
	username string
	password string
}

func (b *basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(b.username, b.password)
	return b.base.RoundTrip(req)
}
