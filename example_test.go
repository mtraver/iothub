package iothub_test

import (
	"log"
	"os"
	"time"

	"github.com/mtraver/iothub"
)

func Example() {
	d := iothub.Device{
		HubName:     "my-hub",
		DeviceID:    "my-device",
		CertPath:    "my-device.x509",
		PrivKeyPath: "my-device.pem",
	}

	// roots.pem should contain the root CA certs listed here:
	// https://learn.microsoft.com/en-us/azure/security/fundamentals/azure-ca-details
	// Most importantly, as of Feb 2023, it should contain the DigiCert Global Root G2 cert.
	certs, err := os.Open("roots.pem")
	if err != nil {
		log.Fatalf("Failed to open certs file: %v", err)
	}
	defer certs.Close()

	client, err := d.NewClient(certs)
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
