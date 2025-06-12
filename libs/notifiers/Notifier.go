package notifiers

import (
	"regexp"

	"github.com/camalot/abyssal/config"
	"github.com/camalot/abyssal/libs/providers"
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


