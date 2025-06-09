package notifiers

import (
	"github.com/camalot/abyssal/config"
	"github.com/camalot/abyssal/libs/providers"
)

/*
- type: jira
	enabled: false
	title: "[Abyssal] Bump {{ .Target.Name }} from v{{ .CurrentVersion }} to v{{ .ExpectedVersion }}"
	body: |
		Bump {{ .Target.Name }} from v{{ .CurrentVersion }} to v{{ .ExpectedVersion }}.
	user: "{{ .EnvironmentVariables.JIRA_USER }}"
	token: "{{ .EnvironmentVariables.JIRA_TOKEN }}"
	url: "{{ .EnvironmentVariables.JIRA_URL }}"
	project: "{{ .EnvironmentVariables.JIRA_PROJECT }}"
	issueType: "{{ .EnvironmentVariables.JIRA_ISSUE_TYPE }}"
	labels: ["abyssal"]
*/

type JiraNotifier struct {
	Enabled bool   `yaml:"enabled"`
	Title   string `yaml:"title"`
	Body    string `yaml:"body"`

	AccessToken string   `yaml:"token,omitempty"`
	IssueLabels []string `yaml:"issueLabels,omitempty"`

	Url       string `yaml:"url,omitempty"`
	Project   string `yaml:"project,omitempty"`
	IssueType string `yaml:"issueType,omitempty"`

	Authentication struct {
		User  string `yaml:"user,omitempty"`
		Token string `yaml:"token,omitempty"`
	}

	NotifierConfig config.NotifierElement `yaml:"-"`
}

func NewJiraNotifier(notifierElement config.NotifierElement, config *config.AppConfiguration) *JiraNotifier {
	token := ""
	if val, ok := notifierElement.Extra["token"]; ok {
		token = envTemplateValue(val.(string))
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

	return &JiraNotifier{
		Enabled:        notifierElement.Enabled,
		Title:          title,
		Body:           body,
		AccessToken:    token,
		IssueLabels:    labels,
		NotifierConfig: notifierElement,
	}
}

// Notify sends a notification with the given payload.
func (j *JiraNotifier) Notify(payload interface{}) error {
	// Implementation for sending a notification to Jira
	// This would typically involve using the Jira API to create an issue or comment
	return nil
}

// this method checks if the notifier needs to send a notification
// this can be used to avoid duplicate notifications
func (j *JiraNotifier) NeedsNotification(payload interface{}) bool {
	// Implementation for sending a notification to Jira
	// This would typically involve using the Jira API to create an issue or comment
	return false
}
func (j *JiraNotifier) HasNotification(payload interface{}) (bool, []interface{}) {
	// Implementation for sending a notification to Jira
	// This would typically involve using the Jira API to create an issue or comment
	return false, nil
}

func (j *JiraNotifier) CloseNotification(payload interface{}) error {
	// Implementation for sending a notification to Jira
	// This would typically involve using the Jira API to create an issue or comment
	return nil
}

func (j *JiraNotifier) CreatePayload(config config.NotifierElement, result *providers.ProviderCheckResult) (interface{}, error) {
	// Implementation for sending a notification to Jira
	// This would typically involve using the Jira API to create an issue or comment
	return nil, nil
}
func (j *JiraNotifier) GetNotifierConfig() config.NotifierElement {
	// Implementation for sending a notification to Jira
	// This would typically involve using the Jira API to create an issue or comment
	return config.NotifierElement{}
}

func (j *JiraNotifier) ProcessResult(result *providers.ProviderCheckResult) error {
	// Implementation for sending a notification to Jira
	// This would typically involve using the Jira API to create an issue or comment
	return nil
}

func (j *JiraNotifier) IsEnabled() bool {
	// Implementation for sending a notification to Jira
	// This would typically involve using the Jira API to create an issue or comment
	return false
}
func (j *JiraNotifier) GetName() string {
	return "Jira Notifier"
}
func (j *JiraNotifier) GetType() string {
	return "jira"
}
