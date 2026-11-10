package govee_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	govee "github.com/Linell/govee-go"
)

var lamp = govee.Device{SKU: "H6008", ID: "AA:BB"}

// fakeGovee serves reply for every request and records the last request body.
func fakeGovee(t *testing.T, status int, reply string, got *map[string]any) *govee.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Govee-API-Key") != "key" {
			t.Errorf("missing API key header")
		}
		if got != nil {
			json.NewDecoder(r.Body).Decode(got)
		}
		w.WriteHeader(status)
		w.Write([]byte(reply))
	}))
	t.Cleanup(srv.Close)
	client := govee.New("key")
	client.BaseURL = srv.URL
	return client
}

func TestDevices(t *testing.T) {
	client := fakeGovee(t, 200, `{"code":200,"message":"success","data":[{"sku":"H6008","device":"AA:BB","deviceName":"Lamp",
		"capabilities":[{"type":"devices.capabilities.range","instance":"brightness","parameters":{"dataType":"INTEGER","range":{"min":1,"max":100,"precision":1}}}]}]}`, nil)

	devices, err := client.Devices(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 1 || devices[0].Name != "Lamp" || devices[0].Capabilities[0].Parameters.Range.Max != 100 {
		t.Fatalf("unexpected devices: %+v", devices)
	}
	if !devices[0].Supports("brightness") || devices[0].Supports("colorRgb") {
		t.Error("Supports is wrong")
	}
}

func TestControlSendsCapability(t *testing.T) {
	var got map[string]any
	client := fakeGovee(t, 200, `{"code":200,"msg":"success"}`, &got)

	if err := client.Control(context.Background(), lamp, govee.SetColor(govee.Color{R: 255})); err != nil {
		t.Fatal(err)
	}
	payload := got["payload"].(map[string]any)
	capability := payload["capability"].(map[string]any)
	if got["requestId"] == "" || payload["device"] != "AA:BB" || capability["instance"] != "colorRgb" || capability["value"] != float64(0xFF0000) {
		t.Fatalf("unexpected request: %+v", got)
	}
}

func TestState(t *testing.T) {
	client := fakeGovee(t, 200, `{"code":200,"msg":"success","payload":{"capabilities":[
		{"type":"devices.capabilities.online","instance":"online","state":{"value":true}}]}}`, nil)

	states, err := client.State(context.Background(), lamp)
	if err != nil {
		t.Fatal(err)
	}
	if len(states) != 1 || states[0].Instance != "online" || states[0].Value != true {
		t.Fatalf("unexpected states: %+v", states)
	}
}

func TestScenesCarryTheirCommand(t *testing.T) {
	client := fakeGovee(t, 200, `{"code":200,"msg":"success","payload":{"capabilities":[{"type":"devices.capabilities.dynamic_scene","instance":"lightScene",
		"parameters":{"dataType":"ENUM","options":[{"name":"Sunrise","value":{"paramId":4280,"id":3853}}]}}]}}`, nil)

	scenes, err := client.Scenes(context.Background(), lamp)
	if err != nil {
		t.Fatal(err)
	}
	cmd := govee.SetScene(scenes[0])
	if scenes[0].Name != "Sunrise" || cmd.Instance != "lightScene" || cmd.Value.(map[string]any)["id"] != float64(3853) {
		t.Fatalf("unexpected scene: %+v", scenes)
	}
}

func TestErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		status int
		reply  string
		code   int
	}{
		"rejected in body":  {200, `{"code":400,"msg":"bad params"}`, 400},
		"rejected in HTTP":  {429, `{"code":429,"message":"rate limited"}`, 429},
		"non-JSON response": {502, `<html>bad gateway</html>`, 502},
	} {
		t.Run(name, func(t *testing.T) {
			err := fakeGovee(t, tc.status, tc.reply, nil).Control(context.Background(), lamp, govee.TurnOn())
			var apiErr *govee.Error
			if !errors.As(err, &apiErr) || apiErr.Code != tc.code {
				t.Fatalf("got %v, want code %d", err, tc.code)
			}
		})
	}
}
