package notifiers

type Notifier interface {
	// Notify sends a notification with the given payload.
	Notify(payload interface{}) error
	// this method checks if the notifier needs to send a notification
	// this can be used to avoid duplicate notifications
	NeedsNotification(payload interface{}) bool
}