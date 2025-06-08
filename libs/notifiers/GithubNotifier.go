package notifiers

import (
	"context"
	// "fmt"
	"os"
	"strings"

	"github.com/camalot/abyssal/config"
	"github.com/camalot/abyssal/libs/providers"
	"github.com/camalot/abyssal/libs/templates"
	"github.com/google/go-github/v72/github"
)

type GithubNotifier struct {
	Enabled bool   `yaml:"enabled"`
	Title   string `yaml:"title"`
	Body    string `yaml:"body"`

	RepositoryName string   `yaml:"repository,omitempty"`
	Organization   string   `yaml:"organization,omitempty"`
	AccessToken    string   `yaml:"token,omitempty"`
	IssueLabels    []string `yaml:"issueLabels,omitempty"`

	NotifierConfig config.NotifierElement `yaml:"-"`
}

func NewGithubNotifier(notifierElement config.NotifierElement, config *config.AppConfiguration) *GithubNotifier {
	repoName := "default-repo"
	if val, ok := notifierElement.Extra["repository"]; ok {
		repoName = envTemplateValue(val.(string))
	}

	org := "default-org"
	if val, ok := notifierElement.Extra["organization"]; ok {
		org = envTemplateValue(val.(string))
	}

	token := ""
	if val, ok := notifierElement.Extra["token"]; ok {
		token = envTemplateValue(val.(string))
	}

	labels := []string{}
	rawList, ok := notifierElement.Extra["labels"].([]interface{})
	if ok {
		labels = interfaceSliceToStringSlice(rawList)
	}

	return &GithubNotifier{
		Enabled:        notifierElement.Enabled,
		Title:          notifierElement.Extra["title"].(string),
		Body:           notifierElement.Extra["body"].(string),
		RepositoryName: repoName,
		Organization:   org,
		AccessToken:    token,
		IssueLabels:    labels,
		NotifierConfig: notifierElement,
	}
}

type GithubNotificationPayload struct {
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	IssueLabels []string `json:"issueLabels,omitempty"`
}

func (g *GithubNotifier) createIssue(title, body string, labels []string) error {
	// fmt.Printf("Creating issue with title: %s\n", title)
	// fmt.Printf("Creating issue with body: %s\n", body)
	// fmt.Printf("Creating issue with labels: %v\n", labels)

	// return nil

	client := github.NewClient(nil).WithAuthToken(g.AccessToken)
	issue := &github.IssueRequest{
		Title:  github.Ptr(title),
		Body:   github.Ptr(body),
		Labels: &labels,
		Type:   github.Ptr("issue"),
		State:  github.Ptr("open"),
	}
	_, _, err := client.Issues.Create(context.Background(), g.Organization, g.RepositoryName, issue)
	return err
}

func (g *GithubNotifier) findIssue(title string) (*[]github.Issue, error) {
	client := github.NewClient(nil).WithAuthToken(g.AccessToken)



	issues, _, err := client.Issues.ListByRepo(context.Background(), g.Organization, g.RepositoryName, &github.IssueListByRepoOptions{
		Labels: []string{"abyssal"},
		State:  "open",
	})
	if err != nil {
		return nil, err
	}
	var foundIssues []github.Issue
	for _, issue := range issues {
		if issue.GetTitle() == title {
			foundIssues = append(foundIssues, *issue)
		}
	}
	if len(foundIssues) == 0 {
		return nil, nil // No issue found
	}
	return &foundIssues, nil
}

func (g *GithubNotifier) GetNotifierConfig() config.NotifierElement {
	return g.NotifierConfig
}

func (g *GithubNotifier) CreatePayload(config config.NotifierElement, result *providers.ProviderCheckResult) (interface{}, error) {
	template := templates.NewTemplate("github-notification", g.Title, result)
	renderedTitle, err := template.Render()
	if err != nil {
		return nil, err // Error occurred while rendering the title
	}

	template = templates.NewTemplate("github-notification-body", g.Body, result)
	renderedBody, err := template.Render()
	if err != nil {
		return nil, err // Error occurred while rendering the body
	}
	return GithubNotificationPayload{
		Title: renderedTitle,
		Body:  renderedBody,
	}, nil
}

func (g *GithubNotifier) IsEnabled() bool {
	return g.Enabled
}

func (g *GithubNotifier) GetType() string {
	return "github"
}
func (g *GithubNotifier) GetName() string {
	return "GitHub Notifier"
}

// Notify sends a notification with the given payload.
func (g *GithubNotifier) Notify(payload interface{}) error {
	if !g.Enabled {
		return nil // No notification sent if not enabled
	}
	if !g.NeedsNotification(payload) {
		return nil // No notification needed if already exists
	}

	ghPayload, ok := payload.(GithubNotificationPayload)
	if !ok {
		return nil // Invalid payload type
	}

	err := g.createIssue(ghPayload.Title, ghPayload.Body, g.IssueLabels)
	if err != nil {
		return err
	}
	// Implement the logic to send a notification to GitHub
	return nil
}

// NeedsNotification checks if the notifier needs to send a notification
// this can be used to avoid duplicate notifications
func (g *GithubNotifier) NeedsNotification(payload interface{}) bool {
	if !g.Enabled {
		return false // No notification needed if not enabled
	}
	ghPayload, ok := payload.(GithubNotificationPayload)
	if !ok {
		// fmt.Println("Invalid payload type for GitHub notifier")
		return false // Invalid payload type
	}

	issues, err := g.findIssue(ghPayload.Title)
	if err != nil {
		// fmt.Printf("Error checking for existing issue: %v\n", err)
		return false // Error occurred while checking for existing issue
	}

	if issues != nil && len(*issues) > 0 {
		// fmt.Printf("Issue with title '%s' already exists.\n", ghPayload.Title)
		return false // Notification already exists
	}
	// fmt.Printf("No existing issue found with title '%s'. Proceeding to create a new issue.\n", ghPayload.Title)
	return true
}

func interfaceSliceToStringSlice(slice interface{}) []string {
	s := []string{}
	if slice == nil {
			return s
	}
	for _, v := range slice.([]interface{}) {
			if str, ok := v.(string); ok {
					s = append(s, str)
			}
	}
	return s
}

type GithubNotifierTemplateData struct {
	EnvironmentVariables map[string]string
}
func envTemplateValue(template string) string {
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
	result := templates.NewTemplate("templated-value", template, &GithubNotifierTemplateData{
		EnvironmentVariables: envVars,
	})
	rendered, err := result.Render()
	if err != nil {
		return template
	}
	return rendered
}
