package notifiers

import "github.com/camalot/abyssal/config"

type NotifierFactory struct {

}

func NewNotifier(notifierElement config.NotifierElement, config *config.AppConfiguration) Notifier {
	switch notifierElement.Type {
	case "github":
		return NewGithubNotifier(notifierElement, config)

	// case "slack":
	// 	return NewSlackNotifier()
	// case "discord":
	// 	return NewDiscordNotifier()
	// case "email":
	// 	return NewEmailNotifier()
	// case "jira":
	// 	return NewJiraNotifier()
	default:
		return nil
	}
}
