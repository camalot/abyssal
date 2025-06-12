package notifiers

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/camalot/abyssal/config"
	"github.com/camalot/abyssal/libs/providers"
	"github.com/camalot/abyssal/libs/templates"
	jira2 "github.com/ctreminiom/go-atlassian/v2/jira/v2"
	jira "github.com/ctreminiom/go-atlassian/v2/jira/v3"
	jiramodels "github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"
)

// https://docs.go-atlassian.io/jira-software-cloud/issues/comments

/*
- type: jira
	enabled: false
	title: "[Abyssal] Bump {{ .Target.Name }} from v{{ .CurrentVersion }} to v{{ .ExpectedVersion }}"
	body: |
		Bump {{ .Target.Name }} from v{{ .CurrentVersion }} to v{{ .ExpectedVersion }}.
	user: "{{ .EnvironmentVariables.JIRA_USERNAME }}"
	token: "{{ .EnvironmentVariables.JIRA_TOKEN }}"
	url: "{{ .EnvironmentVariables.JIRA_URL }}"
	project: "{{ .EnvironmentVariables.JIRA_PROJECT_KEY }}"
	board: "{{ .EnvironmentVariables.JIRA_BOARD_ID }}"
	issueType: "{{ .EnvironmentVariables.JIRA_ISSUE_TYPE }}"
	jql: project = {{ .EnvironmentVariables.JIRA_PROJECT_KEY }} AND issuetype = {{ .EnvironmentVariables.JIRA_ISSUE_TYPE }} AND status != Done
	labels: ["abyssal"]
	payload: {}
*/

type JiraNotifier struct {
	Enabled bool   `yaml:"enabled"`
	Title   string `yaml:"title"`
	Body    string `yaml:"body"`

	Url            string                     `yaml:"url,omitempty"`
	BoardId        string                     `yaml:"board,omitempty"`
	Project        string                     `yaml:"project,omitempty"`
	IssueType      string                     `yaml:"issueType,omitempty"`
	IssueLabels    []string                   `yaml:"issueLabels,omitempty"`
	JQL            string                     `yaml:"jql,omitempty"` // JQL query to search for existing issues
	Authentication JiraNotifierAuthentication `yaml:"authentication,omitempty"`

	NotifierConfig config.NotifierElement `yaml:"-"`

	Payload interface{} `yaml:"payload,omitempty"`
}

type JiraNotifierAuthentication struct {
	User  string `yaml:"user,omitempty"`
	Token string `yaml:"token,omitempty"`
}

type JiraNotificationPayload struct {
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	IssueLabels []string `json:"issueLabels,omitempty"`

	Result *providers.ProviderCheckResult `json:"result,omitempty"` // This can be used to store the result of the check that triggered the notification
}

// NewJiraNotifier creates a new JiraNotifier instance based on the provided notifierElement and app configuration.
// It extracts the necessary fields from the notifierElement's Extra map and initializes the JiraNotifier struct.
// The function handles environment variable templates for fields like token, user, url, project, board, issueType, and jql.
// It also converts the labels from a raw list to a string slice.
// This function is used to set up the Jira notifier with the required configuration for sending notifications.
func NewJiraNotifier(notifierElement config.NotifierElement, config *config.AppConfiguration) *JiraNotifier {
	token := ""
	if val, ok := notifierElement.Extra["token"]; ok {
		token = templates.EnvironmentVariableTemplate(val.(string))
	}

	username := ""
	if val, ok := notifierElement.Extra["user"]; ok {
		username = templates.EnvironmentVariableTemplate(val.(string))
	}

	labels := []string{}
	rawList, ok := notifierElement.Extra["labels"].([]interface{})
	if ok {
		labels = interfaceSliceToStringSlice(rawList)
	}

	title := ""
	if val, ok := notifierElement.Extra["title"]; ok {
		title = val.(string)
	}

	body := ""
	if val, ok := notifierElement.Extra["body"]; ok {
		body = val.(string)
	}

	url := ""
	if val, ok := notifierElement.Extra["url"]; ok {
		url = templates.EnvironmentVariableTemplate(val.(string))
	}

	project := ""
	if val, ok := notifierElement.Extra["project"]; ok {
		project = templates.EnvironmentVariableTemplate(val.(string))
	}

	boardId := ""
	if val, ok := notifierElement.Extra["board"]; ok {
		boardId = templates.EnvironmentVariableTemplate(val.(string))
	}

	issueType := ""
	if val, ok := notifierElement.Extra["issueType"]; ok {
		issueType = templates.EnvironmentVariableTemplate(val.(string))
	}
	if issueType == "" {
		issueType = "Task" // Default issue type if not specified
	}

	jql := ""
	if val, ok := notifierElement.Extra["jql"]; ok {
		jql = templates.EnvironmentVariableTemplate(val.(string))
	}

	var payload interface{}
	payload = nil
	if val, ok := notifierElement.Extra["payload"]; ok {
		payload = val
	}

	return &JiraNotifier{
		Enabled:        notifierElement.Enabled,
		Title:          title,
		Body:           body,
		IssueLabels:    labels,
		NotifierConfig: notifierElement,

		Url:       url,
		Project:   project,
		BoardId:   boardId,
		IssueType: issueType,
		JQL:       jql,

		Authentication: JiraNotifierAuthentication{
			User:  templates.EnvironmentVariableTemplate(username),
			Token: templates.EnvironmentVariableTemplate(token),
		},
		Payload: payload,
	}
}

// arrayToJqlList converts a slice of strings to a JQL list format.
// It formats the strings as a comma-separated list enclosed in double quotes.
func arrayToJqlList(array []string) string {
	if len(array) == 0 {
		return ""
	}
	jqlList := ""
	for i, item := range array {
		if i > 0 {
			jqlList += ", "
		}
		jqlList += fmt.Sprintf("\"%s\"", item)
	}
	return jqlList
}

// createJiraV2Client creates a new Jira v2 client using the provided URL and authentication token.
// It returns the client or an error if the notifier is not enabled or if there is an error creating the client.
// This method is used to establish a connection to the Jira API for performing operations like creating issues or comments.
// It uses the Jira v2 client for compatibility with the Jira API.
func (j *JiraNotifier) createJiraV2Client() (*jira2.Client, error) {
	if !j.Enabled {
		return nil, fmt.Errorf("createJiraV2Client called on disabled Jira notifier")
	}
	host, err := j.getHost()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting Jira host: %v\n", err)
		return nil, err
	}

	client, err := jira2.New(nil, host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating Jira client: %v\n", err)
		return nil, err
	}

	// Set the authentication token
	// client.Auth.SetBearerToken(j.Authentication.Token)
	client.Auth.SetBasicAuth(j.Authentication.User, j.Authentication.Token)

	return client, nil
}

// createJiraClient creates a new Jira client using the provided URL and authentication token.
// It returns the client or an error if the notifier is not enabled or if there is an error creating the client.
// This method is used to establish a connection to the Jira API for performing operations like creating issues or comments.
// It uses the Jira v3 client for compatibility with the Jira API.
func (j *JiraNotifier) createJiraClient() (*jira.Client, error) {
	if !j.Enabled {
		return nil, fmt.Errorf("createJiraClient called on disabled Jira notifier")
	}
	host, err := j.getHost()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting Jira host: %v\n", err)
		return nil, err
	}

	client, err := jira.New(nil, host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating Jira client: %v\n", err)
		return nil, err
	}

	// Set the authentication token
	// client.Auth.SetBearerToken(j.Authentication.Token)
	client.Auth.SetBasicAuth(j.Authentication.User, j.Authentication.Token)
	return client, nil
}

// createIssue creates a new Jira issue with the given title, body, and labels.
// It uses the Jira client to perform the issue creation operation.
// If the notifier is not enabled, it returns an error.
// The title and body are used to set the issue's summary and description, respectively.
func (j *JiraNotifier) createIssue(title, body string, labels []string) error {
	if !j.Enabled {
		return fmt.Errorf("createIssue called on disabled Jira notifier")
	}
	client, err := j.createJiraV2Client()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating Jira client: %v\n", err)
		return err
	}

	issue := jiramodels.IssueSchemeV2{
		Fields: &jiramodels.IssueFieldsSchemeV2{
			Summary:     title,
			Description: body,
			Project: &jiramodels.ProjectScheme{
				Key: j.Project,
			},
			IssueType: &jiramodels.IssueTypeScheme{
				Name: j.IssueType,
			},
			Labels: labels,
		},
	}

	_, _, err = client.Issue.Create(context.Background(), &issue, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating Jira issue: %v\n", err)
		return err
	}
	return nil
}

// commentOnIssue adds a comment to the given Jira issue.
// It uses the Jira client to perform the comment operation.
// If the notifier is not enabled, it returns an error.
// The comment is provided as a string and is added to the issue's comments.
func (j *JiraNotifier) commentOnIssue(issue *jiramodels.IssueScheme, comment string) error {
	if !j.Enabled {
		return fmt.Errorf("commentOnIssue called on disabled Jira notifier")
	}
	client, err := j.createJiraV2Client()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating Jira client: %v\n", err)
		return err
	}

	commentScheme := &jiramodels.CommentPayloadSchemeV2{
		Body: comment,
	}

	_, _, err = client.Issue.Comment.Add(context.Background(), issue.ID, commentScheme, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error adding comment to Jira issue: %v\n", err)
		return err
	}
	return nil
}

// closeIssue closes the given Jira issue by updating its status to "Done".
// It uses the Jira client to perform the update operation.
// If the notifier is not enabled, it returns an error.
func (j *JiraNotifier) closeIssue(issue *jiramodels.IssueScheme) error {
	if !j.Enabled {
		return fmt.Errorf("closeIssue called on disabled Jira notifier")
	}
	client, err := j.createJiraClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating Jira client: %v\n", err)
		return err
	}

	var payload = jiramodels.IssueScheme{
		Fields: &jiramodels.IssueFieldsScheme{
			Status: &jiramodels.StatusScheme{
				Name: "Done", // Assuming "Done" is the status to close the issue
			},
		},
	}

	_, err = client.Issue.Update(context.Background(), issue.ID, false, &payload, nil, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error closing Jira issue: %v\n", err)
		return err
	}
	return nil
}

// findIssues searches for existing Jira issues based on the title, states, and labels.
// It constructs a JQL query using the provided parameters and returns a list of issues that match the criteria.
// If the notifier is not enabled, it returns an error.
func (j *JiraNotifier) findIssues(title string, states []string, labels []string) ([]*jiramodels.IssueScheme, error) {
	if !j.Enabled {
		return nil, fmt.Errorf("findIssues called on disabled Jira notifier")
	}

	client, err := j.createJiraClient()

	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating Jira client: %v\n", err)
		return nil, err
	}


	// take the JQL query from the notifier config
	// and append the title to it
	jql := fmt.Sprintf(
		" %s AND summary ~ \"%s\" AND status IN (%s) AND labels IN (%s)",
		j.JQL,
		title,
		arrayToJqlList(states),
		arrayToJqlList(labels),
	)
	fields := []string{"summary", "status", "assignee", "reporter", "created", "updated", "labels"}
	expands := []string{"renderedFields", "changelog"}
	issues, response, err := client.Issue.Search.SearchJQL(context.Background(), jql, fields, expands, 1, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error searching Jira issues: %v\n", err)
		fmt.Fprintf(os.Stderr, "JQL: %s\n", jql)
		fmt.Fprintf(os.Stderr, "Response Status: %s\n", response.Status)
		fmt.Fprintf(os.Stderr, "Response Body: %s\n", response.Bytes.String())
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("error searching Jira issues: %s", response.Status)
	}

	return issues.Issues, nil
}

// getHost extracts the host from the Jira notifier URL.
// It returns an error if the URL is not set or if there is an error parsing the URL.
// This method is used to ensure that the notifier has a valid host to connect to Jira.
func (j *JiraNotifier) getHost() (string, error) {
	if j.Url == "" {
		return "", fmt.Errorf("jira notifier URL is not set. Please check your configuration")
	}
	// Parse the URL to extract the host
	parsedUrl, err := url.Parse(j.Url)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host), nil
}

// Notify sends a notification with the given payload.
// It checks if the notifier is enabled, if a notification is needed, and if there are existing issues to close.
// If the notifier is not enabled, it prints a message to stderr and returns nil.
func (j *JiraNotifier) Notify(payload interface{}) error {
	if !j.Enabled {
		fmt.Fprintln(os.Stderr, "Jira notifier is not enabled, skipping notification.")
		return nil // No notification sent if not enabled
	}

	if !j.NeedsNotification(payload) {
		return nil // No notification needed if already exists
	}

	// clean up existing issues if NeedsNotification
	hasExistingIssues, existingIssues := j.HasNotification(payload)

	if hasExistingIssues && existingIssues != nil && len(existingIssues) > 0 {
		fmt.Fprintln(os.Stderr, "Found existing issues, closing them before creating a new one.")
		for _, issue := range existingIssues {
			jIssue, ok := issue.(jiramodels.IssueScheme)
			if !ok {
				fmt.Fprintln(os.Stderr, "Invalid issue type in existing issues")
				continue // Skip invalid issue type
			}
			fmt.Fprintf(os.Stderr, "Closing existing issue with key: %s\n", jIssue.Key)
			if err := j.CloseNotification(jIssue); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing existing issue with Jira notifier: %v\n", err)
				return err // Error occurred while closing existing issue
			}
			fmt.Fprintf(os.Stderr, "Closed existing issue with key: %s\n", jIssue.Key)
		}
	}

	jPayload, ok := payload.(JiraNotificationPayload)
	if !ok {
		fmt.Fprintln(os.Stderr, "Invalid payload type for Jira notifier")
		return nil // Invalid payload type
	}

	err := j.createIssue(jPayload.Title, jPayload.Body, j.IssueLabels)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Jira issue: %v\n", err)
		return err
	}

	return nil
}

// this method checks if the notifier needs to send a notification
// this can be used to avoid duplicate notifications
func (j *JiraNotifier) NeedsNotification(payload interface{}) bool {
	if !j.Enabled {
		return false
	}
	jPayload, ok := payload.(JiraNotificationPayload)
	if !ok {
		fmt.Fprintln(os.Stderr, "Invalid payload type for Jira notifier")
		return false // Invalid payload type
	}
	// this finds an existing issue with the exact same title
	issues, err := j.findIssues(jPayload.Title, []string{"Open", "In Progress"}, j.IssueLabels)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking for existing issue: %v\n", err)
		return false // Error occurred while checking for existing issue
	}

	if len(issues) > 0 {
		fmt.Fprintf(os.Stderr, "Issue with title '%s' already exists.\n", jPayload.Title)
		return false // Notification already exists
	}
	fmt.Fprintf(os.Stderr, "No existing issue found with title '%s'. Proceeding to create a new issue.\n", jPayload.Title)
	return true
}

// HasNotification checks if there is an existing notification for the given payload.
// It searches for existing Jira issues based on the title and labels.
// If the notifier is not enabled, it returns false and nil.
// If there are existing issues, it returns true and a list of those issues.
func (j *JiraNotifier) HasNotification(payload interface{}) (bool, []interface{}) {
	if !j.Enabled {
		return false, nil
	}

	jPayload, ok := payload.(JiraNotificationPayload)
	if !ok {
		fmt.Fprintln(os.Stderr, "Invalid payload type for Jira notifier")
		return false, nil // Invalid payload type
	}

	issues, err := j.findIssues(jPayload.Title, []string{"Open", "In Progress"}, jPayload.IssueLabels)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error finding Jira issues: %v\n", err)
		return false, nil
	}
	if len(issues) == 0 {
		return false, nil
	}

	hasIssues := len(issues) > 0
	if !hasIssues {
		return false, nil
	}
	issuesList := make([]interface{}, len(issues))
	for i, issue := range issues {
		issuesList[i] = issue
	}
	return true, issuesList
}

// CloseNotification closes the notification for the given payload.
// This method is used to close an existing notification, typically when the package is no longer outdated.
// It takes a payload of type jiramodels.IssueScheme, which represents the Jira issue to be closed.
// It returns an error if the payload is not of the expected type or if there is an error while closing the issue.
// If the notifier is not enabled, it returns nil without performing any action.
// It also comments on the issue to indicate that it is being closed due to the package no longer being outdated.
// If the payload is not of type jiramodels.IssueScheme, it prints an error message to stderr and returns an error.
func (j *JiraNotifier) CloseNotification(payload interface{}) error {
	if !j.Enabled {
		return nil
	}
	jIssue, ok := payload.(jiramodels.IssueScheme)
	if !ok {
		fmt.Fprintln(os.Stderr, "Invalid payload type for Jira notifier")
		return fmt.Errorf("invalid payload type for Jira notifier") // Invalid payload type
	}
	fmt.Fprintf(os.Stderr, "Closing notification for issue: %s (%s)\n", jIssue.Key, jIssue.ID)
	j.commentOnIssue(&jIssue, "Closing this issue as the package is no longer outdated.")
	return j.closeIssue(&jIssue)
}

// CreatePayload creates a payload for the Jira notification.
// This method renders the title and body templates using the provided result.
// It returns a JiraNotificationPayload struct that contains the rendered title, body, and the result.
func (j *JiraNotifier) CreatePayload(config config.NotifierElement, result *providers.ProviderCheckResult) (interface{}, error) {
	if !j.Enabled {
		return nil, nil
	}
	template := templates.NewTemplate("jira-notification", j.Title, result)
	renderedTitle, err := template.Render()
	if err != nil {
		return nil, err // Error occurred while rendering the title
	}

	template = templates.NewTemplate("jira-notification-body", j.Body, result)
	renderedBody, err := template.Render()
	if err != nil {
		return nil, err // Error occurred while rendering the body
	}
	return JiraNotificationPayload{
		Title:  renderedTitle,
		Body:   renderedBody,
		Result: result,
		IssueLabels: j.IssueLabels,
	}, nil
}

// GetNotifierConfig returns the configuration of the notifier.
// This method is used to retrieve the notifier configuration for logging or other purposes.
func (j *JiraNotifier) GetNotifierConfig() config.NotifierElement {
	return j.NotifierConfig
}

// ProcessResult processes the result and sends a notification if needed.
// It checks if the notifier is enabled, creates a payload, and sends a notification if necessary.
// If the result is not outdated, it checks for existing notifications and closes them if found.
func (j *JiraNotifier) ProcessResult(result *providers.ProviderCheckResult) error {
	if !j.Enabled {
		fmt.Fprintln(os.Stderr, "Jira notifier is not enabled, skipping processing.")
		return nil // No processing needed if not enabled
	}
	fmt.Fprintf(os.Stderr, "Processing result for Jira notifier: %s\n", result.Target.Name)

	payload, err := j.CreatePayload(j.NotifierConfig, result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating payload for Jira notifier: %v\n", err)
		return err
	}

	if result.Outdated && j.NeedsNotification(payload) {
		if err := j.Notify(payload); err != nil {
			fmt.Fprintf(os.Stderr, "Error sending notification with Jira notifier: %v\n", err)
			return err // Error occurred while sending notification
		}
		fmt.Fprintln(os.Stderr, "Notification sent successfully.")
	} else if !result.Outdated {
		hasNotification, existingIssues := j.HasNotification(payload)
		if !hasNotification || existingIssues == nil || len(existingIssues) == 0 {
			fmt.Fprintln(os.Stderr, "No existing notification found, nothing to close.")
			return nil // No existing notification found, nothing to close
		}

		// close them all
		for _, issue := range existingIssues {
			jIssue, ok := issue.(jiramodels.IssueScheme)
			if !ok {
				fmt.Fprintln(os.Stderr, "Invalid issue type in existing issues")
				continue // Skip invalid issue type
			}
			fmt.Fprintln(os.Stderr, "Closing notification as the package is no longer outdated.")
			if err := j.CloseNotification(jIssue); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing issue with Jira notifier: %v\n", err)
				return err // Error occurred while closing issue
			}
			fmt.Fprintf(os.Stderr, "Closed issue: %s (%s)\n", jIssue.Key, jIssue.ID)
		}
	} else {
		fmt.Fprintln(os.Stderr, "No notification needed for this result.")
	}

	return nil
}

// IsEnabled checks if the notifier is enabled.
// This method can be used to determine if the notifier should send notifications.
func (j *JiraNotifier) IsEnabled() bool {
	return j.Enabled
}

// GetName returns the name of the notifier.
func (j *JiraNotifier) GetName() string {
	return "Jira Notifier"
}

// GetType returns the type of the notifier.
func (j *JiraNotifier) GetType() string {
	return "jira"
}
