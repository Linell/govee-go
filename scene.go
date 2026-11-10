package govee

import "context"

// Scene is a named preset; apply it with SetScene.
type Scene struct {
	Name string `json:"name"`
	Command
}

func (c *Client) Scenes(ctx context.Context, d Device) ([]Scene, error) {
	return c.scenes(ctx, "/device/scenes", d)
}

func (c *Client) DIYScenes(ctx context.Context, d Device) ([]Scene, error) {
	return c.scenes(ctx, "/device/diy-scenes", d)
}

func (c *Client) scenes(ctx context.Context, path string, d Device) ([]Scene, error) {
	var res struct {
		Payload struct{ Capabilities []Capability }
	}
	if err := c.post(ctx, path, d, nil, &res); err != nil {
		return nil, err
	}
	var scenes []Scene
	for _, capability := range res.Payload.Capabilities {
		for _, o := range capability.Parameters.Options {
			scenes = append(scenes, Scene{o.Name, Command{capability.Type, capability.Instance, o.Value}})
		}
	}
	return scenes, nil
}
