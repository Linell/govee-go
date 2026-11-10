package govee

import (
	"context"
	"net/http"
	"slices"
)

// Device is a light, plug, or appliance on the account.
type Device struct {
	SKU          string       `json:"sku"`
	ID           string       `json:"device"`
	Name         string       `json:"deviceName"`
	Type         string       `json:"type"`
	Capabilities []Capability `json:"capabilities"`
}

// Capability describes one thing a device can do, such as "powerSwitch" or "brightness".
type Capability struct {
	Type       string     `json:"type"`
	Instance   string     `json:"instance"`
	Parameters Parameters `json:"parameters"`
}

type Parameters struct {
	DataType string   `json:"dataType"`
	Unit     string   `json:"unit,omitempty"`
	Range    *Range   `json:"range,omitempty"`
	Options  []Option `json:"options,omitempty"`
	Fields   []Field  `json:"fields,omitempty"`
}

type Range struct {
	Min       int `json:"min"`
	Max       int `json:"max"`
	Precision int `json:"precision"`
}

type Option struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

// Field is one member of a STRUCT-typed capability, such as a segment color's "rgb".
type Field struct {
	Name     string `json:"fieldName"`
	Required bool   `json:"required"`
	Parameters
}

// State is the current value of one capability.
type State struct {
	Type     string `json:"type"`
	Instance string `json:"instance"`
	Value    any    `json:"value"`
}

func (d Device) Supports(instance string) bool {
	return slices.ContainsFunc(d.Capabilities, func(c Capability) bool { return c.Instance == instance })
}

func (c *Client) Devices(ctx context.Context) ([]Device, error) {
	var res struct{ Data []Device }
	err := c.send(ctx, http.MethodGet, "/user/devices", nil, &res)
	return res.Data, err
}

// State reports each capability's current value. If "online" is false, the rest is stale.
func (c *Client) State(ctx context.Context, d Device) ([]State, error) {
	var res struct {
		Payload struct {
			Capabilities []struct {
				Type, Instance string
				State          struct{ Value any }
			}
		}
	}
	if err := c.post(ctx, "/device/state", d, nil, &res); err != nil {
		return nil, err
	}
	states := make([]State, len(res.Payload.Capabilities))
	for i, s := range res.Payload.Capabilities {
		states[i] = State{s.Type, s.Instance, s.State.Value}
	}
	return states, nil
}
