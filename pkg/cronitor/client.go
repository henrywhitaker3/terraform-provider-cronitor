// Copyright (c) HashiCorp, Inc.

package cronitor

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type Client struct {
	endpoint string
	ApiKey   string
	client   *http.Client

	listKeyRegex *regexp.Regexp
}

type NewClientOpts struct {
	Endpoint string
	ApiKey   string
	Client   *http.Client
}

func NewClient(opts NewClientOpts) *Client {
	if opts.Endpoint == "" {
		opts.Endpoint = "https://cronitor.io"
	}
	if opts.Client == nil {
		opts.Client = http.DefaultClient
	}

	// Ignore the error as it will always compile
	regex, _ := regexp.Compile(`^[0-9a-z0-9-_]+$`)

	return &Client{
		endpoint:     opts.Endpoint,
		ApiKey:       opts.ApiKey,
		client:       opts.Client,
		listKeyRegex: regex,
	}
}

func (c *Client) GetMonitor(ctx context.Context, id string) (*Monitor, error) {
	req, err := c.request(ctx, http.MethodGet, fmt.Sprintf("/api/monitors/%s", id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitor %s: %w", id, err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: url: %s, code %d", ErrFailedGetMonitor, req.URL.String(), resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	mon := &Monitor{}
	if err := json.Unmarshal(body, mon); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return mon, nil
}

func (c *Client) CreateMonitor(ctx context.Context, monitor *Monitor) (*Monitor, error) {
	c.setCreateDefaults(monitor)
	req, err := c.request(ctx, http.MethodPost, "/api/monitors", monitor)
	if err != nil {
		return nil, fmt.Errorf("failed to create monitor request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send create request: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to ready response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("%w: code %d response: %s", ErrFailedCreateMonitor, resp.StatusCode, string(body))
	}

	mon := &Monitor{}
	if err := json.Unmarshal(body, mon); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	return c.GetMonitor(ctx, *mon.Key)
}

func (c *Client) UpdateMonitor(ctx context.Context, monitor *Monitor) (*Monitor, error) {
	if monitor.Key == nil {
		return nil, errors.New("cannot update monitor with empty key")
	}
	req, err := c.request(ctx, http.MethodPut, fmt.Sprintf("/api/monitors/%s", *monitor.Key), monitor)
	if err != nil {
		return nil, fmt.Errorf("failed to build update request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to update monitor: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update monitor, code %d, response %s", resp.StatusCode, string(body))
	}

	return c.GetMonitor(ctx, *monitor.Key)
}

func (c *Client) DeleteMonitor(ctx context.Context, id string) error {
	req, err := c.request(ctx, http.MethodDelete, fmt.Sprintf("/api/monitors/%s", id), nil)
	if err != nil {
		return fmt.Errorf("failed to create request to delete monitor %s: %w", id, err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete monitor: %w", err)
	}

	if resp.StatusCode > 299 {
		return ErrFailedDeleteMonitor
	}

	return nil
}

func (c *Client) GetNotificationList(ctx context.Context, id string) (*NotificationList, error) {
	req, err := c.request(ctx, http.MethodGet, fmt.Sprintf("/v1/templates/%s", id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification list: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get notification list code: %d body: %s", resp.StatusCode, string(body))
	}

	out := &NotificationList{}
	if err := json.Unmarshal(body, out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return out, nil
}

func (c *Client) CreateNotificationList(ctx context.Context, list *NotificationList) (*NotificationList, error) {
	key := make([]byte, 3)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create random bytes: %w", err)
	}

	list.Key = fmt.Sprintf("%s-%s", strings.ToLower(list.Name), hex.EncodeToString(key))
	if !c.listKeyRegex.Match([]byte(list.Key)) {
		return nil, fmt.Errorf("invalid key, only lowercase letters, numbers, dashes and underscores: %s", list.Key)
	}

	req, err := c.request(ctx, http.MethodPost, "/v1/templates", list)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification list: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create notification list code: %d body: %s", resp.StatusCode, string(body))
	}

	return c.GetNotificationList(ctx, list.Key)
}

func (c *Client) UpdateNotificationList(ctx context.Context, list *NotificationList) (*NotificationList, error) {
	req, err := c.request(ctx, http.MethodPut, fmt.Sprintf("/v1/templates/%s", list.Key), list)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to update notification list: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update notification list code: %d body: %s", resp.StatusCode, string(body))
	}

	return c.GetNotificationList(ctx, list.Key)
}

func (c *Client) DeleteNotificationList(ctx context.Context, list *NotificationList) error {
	req, err := c.request(ctx, http.MethodDelete, fmt.Sprintf("/v1/templates/%s", list.Key), list)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete notification list: %w", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to update notification list code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) setCreateDefaults(mon *Monitor) {
	if mon.RealertInterval == "" {
		mon.RealertInterval = "every 8 hours"
	}
	if len(mon.Notify) == 0 {
		mon.Notify = []string{"default"}
	}
	if len(mon.Environments) == 0 {
		mon.Environments = []string{"production"}
	}
	if mon.Request != nil {
		if mon.Request.TimeoutSeconds == 0 {
			mon.Request.TimeoutSeconds = 5
		}
	}
}

func (c *Client) request(ctx context.Context, method, endpoint string, body any) (*http.Request, error) {
	var br io.Reader
	if body != nil {
		by, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		br = bytes.NewReader(by)
	}
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.endpoint, endpoint), br)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	req = req.WithContext(ctx)
	req.SetBasicAuth(c.ApiKey, "")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}
