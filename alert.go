package nightfall

// AlertConfig allows clients to specify where alerts should be delivered when findings are discovered as
// part of a scan. These alerts are delivered asynchronously to all destinations specified in the object instance.
type AlertConfig struct {
	Slack   *SlackAlert   `json:"slack"`
	Email   *EmailAlert   `json:"email"`
	Webhook *WebhookAlert `json:"url"`
}

// SlackAlert contains the configuration required to allow clients to send asynchronous alerts to a Slack
// workspace when findings are detected.
//
// Note that in order for Slack alerts to be delivered to your workspace, you must use authenticate Nightfall
// to your Slack workspace under the Settings menu on the Nightfall Dashboard.
//
// Currently, Nightfall supports delivering alerts to public channels, formatted like "#general".
// Alerts are only sent if findings are detected.
type SlackAlert struct {
	Target string `json:"target"`
}

// EmailAlert contains the configuration required to allow clients to send an asynchronous email message
// when findings are detected. The findings themselves will be delivered as a file attachment on the email.
// Alerts are only sent if findings are detected.
type EmailAlert struct {
	Address string `json:"address"`
}

// WebhookAlert contains the configuration required to allow clients to send a webhook event to an external
// URL when findings are detected. The URL provided must have a route defined on the HTTP POST method,
// and should return a 200 status code upon receipt of the event.
//
// In contrast to other platforms, when using the file scanning APIs, an alert is also sent to this webhook
// *even when there are no findings*.
type WebhookAlert struct {
	Address string `json:"address"`
}
