package notifiers

import (
	"context"
	"fmt"
	"os"
	"regexp"
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
		repoName = templates.EnvironmentVariableTemplate(val.(string))
	}

	org := "default-org"
	if val, ok := notifierElement.Extra["organization"]; ok {
		org = templates.EnvironmentVariableTemplate(val.(string))
	}

	token := ""
	if val, ok := notifierElement.Extra["token"]; ok {
		token = templates.EnvironmentVariableTemplate(val.(string))
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

	Result *providers.ProviderCheckResult `json:"result,omitempty"` // This can be used to store the result of the check that triggered the notification
}

func (g *GithubNotifier) createIssue(title, body string, labels []string) error {
	if !g.Enabled {
		fmt.Fprintln(os.Stderr, "GitHub notifier is not enabled, skipping issue creation.")
		return nil // No issue created if not enabled
	}
	fmt.Fprintf(os.Stderr, "Creating issue with title: %s\n", title)
	fmt.Fprintf(os.Stderr, "Creating issue with body: %s\n", body)
	fmt.Fprintf(os.Stderr, "Creating issue with labels: %v\n", labels)

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

func (g *GithubNotifier) commentOnIssue(issue github.Issue, comment string) error {
	if !g.Enabled {
		fmt.Fprintln(os.Stderr, "GitHub notifier is not enabled, skipping commenting on issue.")
		return nil // No comment made if not enabled
	}
	if comment == "" {
		fmt.Fprintln(os.Stderr, "No comment provided, skipping commenting on issue.")
		return nil // No comment made if comment is empty
	}

	fmt.Fprintf(os.Stderr, "Commenting on issue with number: %d\n", issue.GetNumber())
	fmt.Fprintf(os.Stderr, "Comment content: %s\n", comment)

	client := github.NewClient(nil).WithAuthToken(g.AccessToken)
	commentRequest := &github.IssueComment{
		Body: github.Ptr(comment),
	}
	_, _, err := client.Issues.CreateComment(context.Background(), g.Organization, g.RepositoryName, issue.GetNumber(), commentRequest)
	return err
}

func (g *GithubNotifier) closeIssue(issue github.Issue) error {
	fmt.Fprintf(os.Stderr, "Closing issue with number: %d\n", issue.GetNumber())

	client := github.NewClient(nil).WithAuthToken(g.AccessToken)
	issueRequest := &github.IssueRequest{
		State: github.Ptr("closed"),
	}
	_, _, err := client.Issues.Edit(context.Background(), g.Organization, g.RepositoryName, issue.GetNumber(), issueRequest)
	return err
}

func (g *GithubNotifier) findIssue(title, state string, labels []string) (*[]github.Issue, error) {
	if !g.Enabled {
		fmt.Fprintln(os.Stderr, "GitHub notifier is not enabled, skipping issue search.")
		return nil, nil // No issue found if not enabled
	}
	// print to stderror for debugging
	fmt.Fprintf(os.Stderr, "Searching for issue with title: %s\n", title)
	client := github.NewClient(nil).WithAuthToken(g.AccessToken)
	issues, _, err := client.Issues.ListByRepo(context.Background(), g.Organization, g.RepositoryName, &github.IssueListByRepoOptions{
		Labels: labels,
		State:  state,
	})
	if err != nil {
		return nil, err
	}
	var foundIssues []github.Issue
	for _, issue := range issues {
		if issue == nil {
			continue
		}
		// regex match the title
		re, _ := regexp.Compile(`(?i)` + title)
		if re.MatchString(issue.GetTitle()) || title == issue.GetTitle() {
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
		Title:  renderedTitle,
		Body:   renderedBody,
		Result: result,
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

func (g *GithubNotifier) CloseNotification(payload interface{}) error {
	if !g.Enabled {
		return nil // No notification to close if not enabled
	}
	ghIssue, ok := payload.(github.Issue)
	if !ok {
		fmt.Fprintln(os.Stderr, "Invalid payload type for GitHub notifier")
		return fmt.Errorf("invalid payload type for GitHub notifier") // Invalid payload type
	}
	fmt.Fprintf(os.Stderr, "Closing notification for issue with number: %d\n", ghIssue.GetNumber())
	g.commentOnIssue(ghIssue, "Closing this issue as the package is no longer outdated.")
	return g.closeIssue(ghIssue)
}

func (g *GithubNotifier) ProcessResult(result *providers.ProviderCheckResult) error {
	if !g.Enabled {
		fmt.Fprintln(os.Stderr, "GitHub notifier is not enabled, skipping processing.")
		return nil // No processing needed if not enabled
	}
	fmt.Fprintf(os.Stderr, "Processing result for GitHub notifier: %s\n", result.Target.Name)

	payload, err := g.CreatePayload(g.NotifierConfig, result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating payload for GitHub notifier: %v\n", err)
		return err
	}

	if result.Outdated && g.NeedsNotification(payload) {
		if err := g.Notify(payload); err != nil {
			fmt.Fprintf(os.Stderr, "Error sending notification with GitHub notifier: %v\n", err)
			return err // Error occurred while sending notification
		}
		fmt.Fprintln(os.Stderr, "Notification sent successfully.")
	} else if !result.Outdated {
		hasNotification, existingIssues := g.HasNotification(payload)
		if !hasNotification || existingIssues == nil || len(existingIssues) == 0 {
			fmt.Fprintln(os.Stderr, "No existing notification found, nothing to close.")
			return nil // No existing notification found, nothing to close
		}

		// close them all
		for _, issue := range existingIssues {
			ghIssue, ok := issue.(github.Issue)
			if !ok {
				fmt.Fprintln(os.Stderr, "Invalid issue type in existing issues")
				continue // Skip invalid issue type
			}
			fmt.Fprintln(os.Stderr, "Closing notification as the package is no longer outdated.")
			if err := g.CloseNotification(ghIssue); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing issue with GitHub notifier: %v\n", err)
				return err // Error occurred while closing issue
			}
			fmt.Fprintf(os.Stderr, "Closed issue with number: %d\n", ghIssue.GetNumber())
		}
	} else {
		fmt.Fprintln(os.Stderr, "No notification needed for this result.")
	}

	return nil
}

// Notify sends a notification with the given payload.
func (g *GithubNotifier) Notify(payload interface{}) error {
	if !g.Enabled {
		fmt.Fprintln(os.Stderr, "GitHub notifier is not enabled, skipping notification.")
		return nil // No notification sent if not enabled
	}

	if !g.NeedsNotification(payload) {
		return nil // No notification needed if already exists
	}

	// clean up existing issues if NeedsNotification
	hasExistingIssues, existingIssues := g.HasNotification(payload)

	if hasExistingIssues && existingIssues != nil && len(existingIssues) > 0 {
		fmt.Fprintln(os.Stderr, "Found existing issues, closing them before creating a new one.")
		for _, issue := range existingIssues {
			ghIssue, ok := issue.(github.Issue)
			if !ok {
				fmt.Fprintln(os.Stderr, "Invalid issue type in existing issues")
				continue // Skip invalid issue type
			}
			fmt.Fprintf(os.Stderr, "Closing existing issue with number: %d\n", ghIssue.GetNumber())
			if err := g.CloseNotification(ghIssue); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing existing issue with GitHub notifier: %v\n", err)
				return err // Error occurred while closing existing issue
			}
			fmt.Fprintf(os.Stderr, "Closed existing issue with number: %d\n", ghIssue.GetNumber())
		}
	}

	ghPayload, ok := payload.(GithubNotificationPayload)
	if !ok {
		fmt.Fprintln(os.Stderr, "Invalid payload type for GitHub notifier")
		return nil // Invalid payload type
	}

	err := g.createIssue(ghPayload.Title, ghPayload.Body, g.IssueLabels)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating GitHub issue: %v\n", err)
		return err
	}

	return nil
}


func (g *GithubNotifier) HasNotification(payload interface{}) (bool, []interface{}) {
	if !g.Enabled {
		return false, nil // No notification found if not enabled
	}

	var err error
	ghPayload, ok := payload.(GithubNotificationPayload)
	if !ok {
		fmt.Fprintln(os.Stderr, "Invalid payload type for GitHub notifier")
		return false, nil // Invalid payload type
	}

	titleMatch := ghPayload.Title
	template, ok := g.NotifierConfig.Extra["title"].(string)
	if ok {

		regexTitle := templates.NewTemplate("github-notification-search", template, map[string]interface{}{
			"Target": ghPayload.Result.Target,
			"ExpectedVersion": ghPayload.Result.ExpectedVersion,
			"CurrentVersion":  "{{ .CurrentVersion }}",
		})
		tempTemplate, _ := regexTitle.Render()
		fmt.Fprintf(os.Stderr, "Using template for title match: %s\n", regexEscapeString(tempTemplate))
		regexTitle = templates.NewTemplate("github-notification-search", regexEscapeString(tempTemplate), map[string]interface{}{
			"CurrentVersion":  `([^\s]+?)`,
		})
		titleMatch, err = regexTitle.Render()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering title match regex: %v\n", err)
		}
		fmt.Fprintf(os.Stderr, "Using regex title match: %s\n", titleMatch)
	}
	titleMatch = strings.TrimSpace(titleMatch)
	fmt.Fprintf(os.Stderr, "Checking for existing issue with title: %s\n", titleMatch)
	issues, err := g.findIssue(titleMatch, "open", ghPayload.IssueLabels)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking for existing issue: %v\n", err)
		return false, nil // Error occurred while checking for existing issue
	}

	if issues != nil && len(*issues) > 0 {
		fmt.Fprintf(os.Stderr, "Issue with title '%s' already exists.\n", ghPayload.Title)
		// Return true and the existing issues
		existingIssues := []interface{}{}
		for _, issue := range *issues {
			existingIssues = append(existingIssues, issue) // Convert github.Issue to interface{}
		}
		// this should return []github.Issue or nil if no issue exists
		return true, existingIssues
	}
	return false, nil // No existing issue found
}

// NeedsNotification checks if the notifier needs to send a notification
// this can be used to avoid duplicate notifications
func (g *GithubNotifier) NeedsNotification(payload interface{}) bool {
	if !g.Enabled {
		return false // No notification needed if not enabled
	}
	ghPayload, ok := payload.(GithubNotificationPayload)
	if !ok {
		fmt.Fprintln(os.Stderr, "Invalid payload type for GitHub notifier")
		return false // Invalid payload type
	}
	// this finds an existing issue with the exact same title
	issues, err := g.findIssue(ghPayload.Title, "open", g.IssueLabels)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking for existing issue: %v\n", err)
		return false // Error occurred while checking for existing issue
	}

	if issues != nil && len(*issues) > 0 {
		fmt.Fprintf(os.Stderr, "Issue with title '%s' already exists.\n", ghPayload.Title)
		return false // Notification already exists
	}
	fmt.Fprintf(os.Stderr, "No existing issue found with title '%s'. Proceeding to create a new issue.\n", ghPayload.Title)
	return true
}
