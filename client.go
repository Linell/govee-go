// Package govee is a small client for the Govee Developer API.
//
//	client := govee.New(apiKey)
//	devices, _ := client.Devices(ctx)
//	client.Control(ctx, devices[0], govee.SetBrightness(50))
package govee

import (
	"bytes"
	"cmp"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const DefaultBaseURL = "https://openapi.api.govee.com/router/api/v1"

// Client talks to the Govee API. Its fields may be changed after New.
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func New(apiKey string) *Client {
	return &Client{APIKey: apiKey, BaseURL: DefaultBaseURL, HTTPClient: http.DefaultClient}
}

// Error is returned when Govee rejects a request (e.g. 401 bad key, 429 rate limited).
type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string { return fmt.Sprintf("govee: %d %s", e.Code, e.Message) }

type request struct {
	RequestID string  `json:"requestId"`
	Payload   payload `json:"payload"`
}

type payload struct {
	SKU        string   `json:"sku"`
	Device     string   `json:"device"`
	Capability *Command `json:"capability,omitempty"`
}

func (c *Client) post(ctx context.Context, path string, d Device, cmd *Command, out any) error {
	body, err := json.Marshal(request{rand.Text(), payload{d.SKU, d.ID, cmd}})
	if err != nil {
		return err
	}
	return c.send(ctx, http.MethodPost, path, bytes.NewReader(body), out)
}

func (c *Client) send(ctx context.Context, method, path string, body io.Reader, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Govee-API-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var status struct {
		Code         int
		Msg, Message string
	}
	if err := json.Unmarshal(raw, &status); err != nil || res.StatusCode != http.StatusOK || status.Code != http.StatusOK {
		return &Error{Code: cmp.Or(status.Code, res.StatusCode), Message: cmp.Or(status.Msg, status.Message, res.Status)}
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(raw, out)
}
