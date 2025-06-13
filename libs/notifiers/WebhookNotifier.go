package notifiers

import (
	"github.com/camalot/abyssal/config"
	"github.com/camalot/abyssal/libs/providers"
	"github.com/camalot/abyssal/libs/templates"
)

type WebhookNotifier struct {
	Enabled bool   `yaml:"enabled"`
	Url     string `yaml:"url"`
	Name    string `yaml:"name"`

	// Authentication
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	// Headers to be sent with the request
	Headers map[string]string `yaml:"headers,omitempty"`
	// Custom HTTP method to use for the request
	Method string `yaml:"method,omitempty"` // e.g., POST, GET, PUT, DELETE
	// Custom payload to send with the request
	Payload interface{} `yaml:"payload,omitempty"`

	NotifierConfig config.NotifierElement `yaml:"-"`
}

type WebhookNotifierPayload struct {
	Result *providers.ProviderCheckResult `json:"result,omitempty"`
}

func NewWebhookNotifier(config config.NotifierElement) *WebhookNotifier {
	if config.Type != "webhook" {
		return nil
	}

	url := ""
	if val, ok := config.Extra["url"]; ok {
		url = templates.EnvironmentVariableTemplate(val.(string))
	}

	username := ""
	if val, ok := config.Extra["username"]; ok {
		username = templates.EnvironmentVariableTemplate(val.(string))
	}
	password := ""
	if val, ok := config.Extra["password"]; ok {
		password = templates.EnvironmentVariableTemplate(val.(string))
	}
	headers := make(map[string]string)
	if val, ok := config.Extra["headers"]; ok {
		for key, value := range val.(map[string]interface{}) {
			headers[key] = templates.EnvironmentVariableTemplate(value.(string))
		}
	}
	method := "POST" // Default method
	if val, ok := config.Extra["method"]; ok {
		method := templates.EnvironmentVariableTemplate(val.(string))
		if method == "" {
			method = "POST"
		}
	}
	var payload interface{}
	payload = nil
	if val, ok := config.Extra["payload"]; ok {
		payload = val
	}

	name := "Webhook Notifier"
	if val, ok := config.Extra["name"]; ok {
		name = templates.EnvironmentVariableTemplate(val.(string))
	}
	notifier := &WebhookNotifier{
		Enabled: config.Enabled,
		Url:     url,
		Name:    name,
		Username: username,
		Password: password,
		Headers:  headers,
		Method:   method,
		Payload:  payload,

		NotifierConfig: config,
	}
	return notifier
}
// Notify sends a notification with the given payload.
func (n *WebhookNotifier) Notify(payload interface{}) error {
	return nil
}

// this method checks if the notifier needs to send a notification
// this can be used to avoid duplicate notifications
func (n *WebhookNotifier) NeedsNotification(payload interface{}) bool {
	return true
}

func (n *WebhookNotifier) HasNotification(payload interface{}) (bool, []interface{}) {
	return false, nil
}

func (n *WebhookNotifier) CloseNotification(payload interface{}) error {
	return nil
}

func (n *WebhookNotifier) CreatePayload(config config.NotifierElement, result *providers.ProviderCheckResult) (interface{}, error) {

	return WebhookNotifierPayload{
		Result: result,
	}, nil
}

func (n *WebhookNotifier) GetNotifierConfig() config.NotifierElement {
	return config.NotifierElement{}
}

func (n *WebhookNotifier) ProcessResult(result *providers.ProviderCheckResult) error {
	return nil
}

func (n *WebhookNotifier) IsEnabled() bool {
	return n.Enabled
}

func (n *WebhookNotifier) GetName() string {
	return n.Name
}

func (n *WebhookNotifier) GetType() string {
	return "webhook"
}
