// Copyright (c) HashiCorp, Inc.

package cronitor

type Request struct {
	URL             string            `json:"url"`
	Headers         map[string]string `json:"headers,omitempty"`
	Cookies         map[string]string `json:"cookies,omitempty"`
	Body            string            `json:"body,omitempty"`
	Method          string            `json:"method"`
	TimeoutSeconds  int               `json:"timeout_seconds"`
	Regions         []string          `json:"regions,omitempty"`
	FollowRedirects bool              `json:"follow_redirects"`
	VerifySsl       bool              `json:"verify_ssl"`
}

type Monitor struct {
	Name              string   `json:"name"`
	Assertions        []string `json:"assertions"`
	Disabled          bool     `json:"disabled"`
	FailureTolerance  *int     `json:"failure_tolerance,omitempty"`
	GraceSeconds      *int     `json:"grace_seconds,omitempty"`
	Group             *string  `json:"group,omitempty"`
	Key               *string  `json:"key,omitempty"`
	Notify            []string `json:"notify"`
	Paused            bool     `json:"paused"`
	Platform          string   `json:"platform"`
	RealertInterval   string   `json:"realert_interval"`
	Request           *Request `json:"request,omitempty"`
	Running           bool     `json:"running"`
	Schedule          string   `json:"schedule"`
	ScheduleTolerance *int     `json:"schedule_tolerance,omitempty"`
	Tags              []string `json:"tags"`
	Timezone          *string  `json:"timezone,omitempty"`
	Type              string   `json:"type"`
	Environments      []string `json:"environments"`
}

type Notifications struct {
	Emails    []string `json:"emails,omitempty"`
	Slack     []string `json:"slack,omitempty"`
	Pagerduty []string `json:"pagerduty,omitempty"`
	Phones    []string `json:"phones,omitempty"`
	Webhooks  []string `json:"webhook,omitempty"`
}

type NotificationList struct {
	Name          string        `json:"name"`
	Key           string        `json:"key"`
	Notifications Notifications `json:"notifications,omitempty"`
}
