package notifiers

import (
	"os"
	"regexp"
	"strings"

	"github.com/camalot/abyssal/config"
	"github.com/camalot/abyssal/libs/providers"
	"github.com/camalot/abyssal/libs/templates"
)

type Notifier interface {
	// Notify sends a notification with the given payload.
	Notify(payload interface{}) error
	// this method checks if the notifier needs to send a notification
	// this can be used to avoid duplicate notifications
	NeedsNotification(payload interface{}) bool
	HasNotification(payload interface{}) (bool, []interface{})

	CloseNotification(payload interface{}) error

	CreatePayload(config config.NotifierElement, result *providers.ProviderCheckResult) (interface{}, error)
	GetNotifierConfig() config.NotifierElement

	ProcessResult(result *providers.ProviderCheckResult) error

	IsEnabled() bool
	GetName() string
	GetType() string
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

func regexEscapeString(s string) string {
	// Escape special characters for regex
	re := regexp.MustCompile(`([*+?^$()|[\]])`)
	return re.ReplaceAllString(s, `\$1`)
}

type EnvironmentTemplateData struct {
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
	result := templates.NewTemplate("templated-value", template, &EnvironmentTemplateData{
		EnvironmentVariables: envVars,
	})
	rendered, err := result.Render()
	if err != nil {
		return template
	}
	return rendered
}
