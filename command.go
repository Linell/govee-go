package govee

import "context"

const (
	onOff        = "devices.capabilities.on_off"
	ranged       = "devices.capabilities.range"
	colorSetting = "devices.capabilities.color_setting"
	segmentColor = "devices.capabilities.segment_color_setting"
)

// Command sets one capability. Build one with the helpers below, or by hand
// from a device's Capabilities for anything they don't cover.
type Command struct {
	Type     string `json:"type"`
	Instance string `json:"instance"`
	Value    any    `json:"value"`
}

type Color struct{ R, G, B uint8 }

func (c Color) rgb() int { return int(c.R)<<16 | int(c.G)<<8 | int(c.B) }

func (c *Client) Control(ctx context.Context, d Device, cmd Command) error {
	return c.post(ctx, "/device/control", d, &cmd, nil)
}

func TurnOn() Command  { return Command{onOff, "powerSwitch", 1} }
func TurnOff() Command { return Command{onOff, "powerSwitch", 0} }

func SetBrightness(percent int) Command { return Command{ranged, "brightness", percent} }

func SetColor(c Color) Command { return Command{colorSetting, "colorRgb", c.rgb()} }

func SetColorTemperature(kelvin int) Command {
	return Command{colorSetting, "colorTemperatureK", kelvin}
}

func SetSegmentColor(c Color, segments ...int) Command {
	return Command{segmentColor, "segmentedColorRgb", map[string]any{"segment": segments, "rgb": c.rgb()}}
}

func SetScene(s Scene) Command { return s.Command }
