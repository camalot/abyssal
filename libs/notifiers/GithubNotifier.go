package notifiers

import (
	"context"

	"github.com/google/go-github/v72/github"
)

type GithubNotifier struct {
	Enabled      bool   `yaml:"enabled"`
	Title        string `yaml:"title"`
	Body         string `yaml:"body"`

	RepositoryName string
	Organization   string
	AccessToken    string
}

type GithubNotificationPayload struct {
	Title string `json:"title"`
}

func (g *GithubNotifier) createIssue(title, body string) error {
	client := github.NewClient(nil).WithAuthToken(g.AccessToken)
	issue := &github.IssueRequest{
		Title:  github.Ptr(title),
		Body:   github.Ptr(body),
		Labels: &[]string{"notification"},
	}
	_, _, err := client.Issues.Create(context.Background(), g.Organization, g.RepositoryName, issue)
	return err
}

func (g *GithubNotifier) findIssue(title string) (*[]github.Issue, error) {
	client := github.NewClient(nil).WithAuthToken(g.AccessToken)
	issues, _, err := client.Issues.ListByRepo(context.Background(), g.Organization, g.RepositoryName, &github.IssueListByRepoOptions{
		Labels: []string{"notification"},
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

// Notify sends a notification with the given payload.
func (g *GithubNotifier) Notify(payload GithubNotificationPayload) error {
	if !g.Enabled {
		return nil // No notification sent if not enabled
	}
	if !g.NeedsNotification(payload) {
		return nil // No notification needed if already exists
	}

	err := g.createIssue(payload.Title, g.Body)
	if err != nil {
		return err
	}
	// Implement the logic to send a notification to GitHub
	return nil
}

// NeedsNotification checks if the notifier needs to send a notification
// this can be used to avoid duplicate notifications
func (g *GithubNotifier) NeedsNotification(payload GithubNotificationPayload) bool {
	if !g.Enabled {
		return false // No notification needed if not enabled
	}

	issues, err := g.findIssue(string(payload.Title))
	if err != nil {
		return false // Error occurred while checking for existing issue
	}

	if issues != nil && len(*issues) > 0 {
		return false // Notification already exists
	}

	return true
}
