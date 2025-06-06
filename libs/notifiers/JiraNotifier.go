package notifiers

type JiraNotifier struct {
	Enabled bool `yaml:"enabled"`
}

// Notify sends a notification with the given payload.
func (j *JiraNotifier) Notify(payload interface{}) error {
	if !j.Enabled {
		return nil // No notification sent if not enabled
	}
	// Implement the logic to send a notification to Jira
	return nil
}

// NeedsNotification checks if the notifier needs to send a notification
// this can be used to avoid duplicate notifications
func (j *JiraNotifier) NeedsNotification(payload interface{}) bool {
	if !j.Enabled {
		return false // No notification needed if not enabled
	}
	// Implement the logic to determine if a notification is needed
	return false
}
