package notifiers

import (
	"github.com/camalot/abyssal/config"
	"github.com/camalot/abyssal/libs/providers"
)

type Notifier interface {
	// Notify sends a notification with the given payload.
	Notify(payload interface{}) error
	// this method checks if the notifier needs to send a notification
	// this can be used to avoid duplicate notifications
	NeedsNotification(payload interface{}) bool

	CreatePayload(config config.NotifierElement, result *providers.ProviderCheckResult) (interface{}, error)
	GetNotifierConfig() config.NotifierElement

	IsEnabled() bool
	GetName() string
	GetType() string
}
