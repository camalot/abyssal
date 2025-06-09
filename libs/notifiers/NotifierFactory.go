package notifiers

import (
	"strings"

	"github.com/camalot/abyssal/config"
)

type NotifierFactory struct {
}

func NewNotifier(notifierElement config.NotifierElement, config *config.AppConfiguration) Notifier {
	switch strings.TrimSpace(strings.ToLower(notifierElement.Type)) {
	case "github":
		return NewGithubNotifier(notifierElement, config)
	case "jira":
		return NewJiraNotifier(notifierElement, config)

	// case "slack":
	// 	return NewSlackNotifier()
	// case "discord":
	// 	return NewDiscordNotifier()
	// case "email":
	// 	return NewEmailNotifier()
	default:
		return nil
	}
}
