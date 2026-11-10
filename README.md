# govee-go

A small, typed Go client for the [Govee Developer API](https://developer.govee.com/reference/get-you-devices). No dependencies.

```sh
go get github.com/Linell/govee-go
```

Get an API key from the Govee Home app (Profile → Settings → Apply for API Key).

## Usage

```go
client := govee.New(os.Getenv("GOVEE_API_KEY"))

devices, err := client.Devices(ctx)
lamp := devices[0]

client.Control(ctx, lamp, govee.TurnOn())
client.Control(ctx, lamp, govee.SetBrightness(40))
client.Control(ctx, lamp, govee.SetColor(govee.Color{R: 255, G: 120}))
client.Control(ctx, lamp, govee.SetColorTemperature(2700))
client.Control(ctx, lamp, govee.SetSegmentColor(govee.Color{B: 255}, 0, 1, 2))

scenes, err := client.Scenes(ctx, lamp) // or client.DIYScenes
client.Control(ctx, lamp, govee.SetScene(scenes[0]))

states, err := client.State(ctx, lamp)
```

Anything the helpers don't cover can be sent as a raw `govee.Command` built from `lamp.Capabilities`.

## Good to know

- **Check before you ask.** Not every device supports every call. Use `lamp.Supports("lightScene")` (or `"brightness"`, `"colorRgb"`, …) first.
- **Groups** made in the Govee app show up as devices with SKU `SameModeGroup`. They support on/off only; state and scenes return an error.
- **Errors** from Govee come back as `*govee.Error`, whose `Code` is Govee's status: 400 bad value, 401 bad key, 429 rate limited.
- **Rate limits** are low for state and scenes (30/minute per device), so cache where you can.

## License

MIT
