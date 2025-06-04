package notifiers

type GithubNotifier struct {
	Enabled bool	 `yaml:"enabled"`
}

// Notify sends a notification with the given payload.
func (g *GithubNotifier) Notify(payload interface{}) error {
	if !g.Enabled {
		return nil // No notification sent if not enabled
	}
	// Implement the logic to send a notification to GitHub
	return nil
}

// NeedsNotification checks if the notifier needs to send a notification
// this can be used to avoid duplicate notifications
func (g *GithubNotifier) NeedsNotification(payload interface{}) bool {
	if !g.Enabled {
		return false // No notification needed if not enabled
	}
	// Implement the logic to determine if a notification is needed
	return false
}
