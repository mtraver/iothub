package iothub_test

import (
	"log"
	"time"

	"github.com/mtraver/iothub"
)

func Example() {
	d := iothub.Device{
		HubName:  "my-hub",
		DeviceID: "my-device",
		// roots.pem must contain Azure's trusted root certs. See the README for more info.
		CACerts:     "roots.pem",
		CertPath:    "my-device.x509",
		PrivKeyPath: "my-device.pem",
	}

	client, err := d.NewClient()
	if err != nil {
		log.Fatalf("Failed to make MQTT client: %v", err)
	}

	if token := client.Connect(); !token.Wait() || token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}

	if token := client.Publish(d.TelemetryTopic(), 1, false, []byte("{\"temp\": 18.0}")); !token.Wait() || token.Error() != nil {
		log.Printf("Failed to publish: %v", token.Error())
	}

	client.Disconnect(250)
	time.Sleep(500 * time.Millisecond)
}
