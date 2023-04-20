package iothub

import (
	"testing"
)

var device = Device{
	HubName:     "myhub",
	DeviceID:    "foo",
	PrivKeyPath: "key.pem",
}

func TestClientID(t *testing.T) {
	want := device.DeviceID
	got := device.ClientID()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUsername(t *testing.T) {
	want := "myhub.azure-devices.net/foo"
	got := device.Username()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCommandTopic(t *testing.T) {
	want := "/devices/foo/messages/devicebound/#"
	got := device.CommandTopic()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTelemetryTopic(t *testing.T) {
	want := "/devices/foo/messages/events"
	got := device.TelemetryTopic()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
