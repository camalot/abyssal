package notifiers

import (
	"github.com/camalot/abyssal/config"
	"github.com/camalot/abyssal/libs/providers"
)

type DiscordNotifier struct {
	Enabled bool   `yaml:"enabled"`
	Title   string `yaml:"title"`
	Body    string `yaml:"body"`

	WebhookUrl string `yaml:"webhook-url"`

	NotifierConfig config.NotifierElement `yaml:"-"`
}

type DiscordNotificationPayload struct {
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	IssueLabels []string `json:"issueLabels,omitempty"`

	Result *providers.ProviderCheckResult `json:"result,omitempty"` // This can be used to store the result of the check that triggered the notification
}

func NewDiscordNotifier(notifierElement config.NotifierElement, config *config.AppConfiguration) (*DiscordNotifier, error) {
	notifier := &DiscordNotifier{
		Enabled:        notifierElement.Enabled,
		Title:          notifierElement.Extra["title"].(string),
		Body:           notifierElement.Extra["body"].(string),
		WebhookUrl:     notifierElement.Extra["webhook-url"].(string),
		NotifierConfig: notifierElement,
	}
	return notifier, nil
}

// Notify sends a notification with the given payload.
func (n *DiscordNotifier) Notify(payload interface{}) error {
	return nil
}

// this method checks if the notifier needs to send a notification
// this can be used to avoid duplicate notifications
func (n *DiscordNotifier) NeedsNotification(payload interface{}) bool {
	return false
}

func (n *DiscordNotifier) HasNotification(payload interface{}) (bool, []interface{}) {
	return false, nil
}

func (n *DiscordNotifier) CloseNotification(payload interface{}) error {
	return nil
}

func (n *DiscordNotifier) CreatePayload(config config.NotifierElement, result *providers.ProviderCheckResult) (interface{}, error) {
	return nil, nil
}

func (n *DiscordNotifier) GetNotifierConfig() config.NotifierElement {
	return n.NotifierConfig
}

func (n *DiscordNotifier) ProcessResult(result *providers.ProviderCheckResult) error {
	return nil
}

func (n *DiscordNotifier) IsEnabled() bool {
	return n.Enabled
}

func (n *DiscordNotifier) GetName() string {
	return "Discord Notifier"
}

func (n *DiscordNotifier) GetType() string {
	return "discord"
}
