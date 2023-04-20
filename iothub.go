// Package iothub eases interaction with Azure IoT Hub over MQTT.
// It handles TLS configuration and authentication. It also makes it easy to construct
// the fully-qualified MQTT topics that IoT Hub uses for telemetry and cloud-to-device
// communication.
package iothub

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const azureDevicesEndpoint = "azure-devices.net"

// DeviceIDFromCert gets the Common Name from an X.509 cert, which for the purposes of this package is considered to be the device ID.
func DeviceIDFromCert(certPath string) (string, error) {
	certBytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("iothub: cert file does not exist: %v", certPath)
		}

		return "", fmt.Errorf("iothub: failed to read cert: %v", err)
	}

	block, _ := pem.Decode(certBytes)
	if block == nil || block.Type != "CERTIFICATE" {
		return "", fmt.Errorf("iothub: failed to decode PEM certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}

	return cert.Subject.CommonName, nil
}

// Device represents an IoT Hub device.
type Device struct {
	HubName     string `json:"hub_name"`
	DeviceID    string `json:"device_id"`
	CertPath    string `json:"cert_path"`
	PrivKeyPath string `json:"priv_key_path"`
}

// NewClient creates a github.com/eclipse/paho.mqtt.golang Client that may be used to connect to the device's Hub's MQTT broker using TLS,
// which Azure IoT Hub requires. By default it sets up a github.com/eclipse/paho.mqtt.golang ClientOptions with the minimal
// options required to establish a connection:
//
//   - Client ID
//   - Username
//   - TLS configuration that supplies root CA certs and the device's cert
//   - Broker
//
// By passing in options you may customize the ClientOptions. Options are functions with this signature:
//
//	func(*Device, *mqtt.ClientOptions) error
//
// They modify the ClientOptions. The option functions are applied to the ClientOptions in the order given before the
// Client is created. For example, if you wish to set the connect timeout, you might write this:
//
//	func ConnectTimeout(t time.Duration) func(*Device, *mqtt.ClientOptions) error {
//		return func(d *Device, opts *mqtt.ClientOptions) error {
//			opts.SetConnectTimeout(t)
//			return nil
//		}
//	}
//
// Using option functions allows for sensible defaults — no options are required to establish a
// connection — without loss of customizability.
//
// For more information about connecting to Azure IoT Hub's MQTT brokers see https://learn.microsoft.com/en-us/azure/iot-hub/iot-hub-mqtt-support#tlsssl-configuration.
func (d *Device) NewClient(caCerts io.Reader, options ...func(*Device, *mqtt.ClientOptions) error) (mqtt.Client, error) {
	// Load CA certs.
	pemCerts, err := ioutil.ReadAll(caCerts)
	if err != nil {
		return nil, fmt.Errorf("iothub: failed to read CA certs: %v", err)
	}
	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(pemCerts) {
		return nil, fmt.Errorf("iothub: no certs were parsed from given CA certs")
	}

	// Import client certificate/key pair
	cert, err := tls.LoadX509KeyPair(d.CertPath, d.PrivKeyPath)
	if err != nil {
		return nil, fmt.Errorf("iothub: failed to load x509 key pair: %w", err)
	}

	tlsConf := &tls.Config{
		RootCAs:      certpool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	broker := d.Broker()

	// See https://learn.microsoft.com/en-us/azure/iot-hub/iot-hub-mqtt-support#tlsssl-configuration
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker.URL())
	opts.SetClientID(d.ClientID())
	opts.SetUsername(d.Username())
	opts.SetTLSConfig(tlsConf)

	for _, option := range options {
		if err := option(d, opts); err != nil {
			return nil, err
		}
	}

	return mqtt.NewClient(opts), nil
}

func (d *Device) Broker() MQTTBroker {
	return MQTTBroker{
		Host: fmt.Sprintf("%s.%s", d.HubName, azureDevicesEndpoint),
		Port: 8883,
	}
}

// ClientID returns the device ID, since that is what IoT Hub requires.
// See https://learn.microsoft.com/en-us/azure/iot-hub/iot-hub-mqtt-support#using-the-mqtt-protocol-directly-as-a-device
func (d *Device) ClientID() string {
	return d.DeviceID
}

// Username returns a username formatted as required by IoT Hub.
func (d *Device) Username() string {
	// The IoT Hub documentation recommends including an API version in the username, like this:
	// "{iotHub-hostname}/{device-id}/?api-version=2021-04-12". However I found that including it (and trying
	// out several versions, because the docs do not say which should be used or which is the latest and the
	// recommended version is different depending on where you look in the docs) results in failure to connect. I get
	// "Connection Refused: Server Unavailable" when it's included. Therefore an API version is not included here.
	// See https://learn.microsoft.com/en-us/azure/iot-hub/iot-hub-mqtt-support#using-the-mqtt-protocol-directly-as-a-device.
	return fmt.Sprintf("%s.%s/%s", d.HubName, azureDevicesEndpoint, d.DeviceID)
}

// CommandTopic returns the MQTT topic to which the device can subscribe to get commands.
// For more information see https://learn.microsoft.com/en-us/azure/iot-hub/iot-hub-mqtt-support#receiving-cloud-to-device-messages.
func (d *Device) CommandTopic() string {
	return fmt.Sprintf("/devices/%v/messages/devicebound/#", d.DeviceID)
}

// TelemetryTopic returns the MQTT topic to which the device should publish telemetry events.
// For more information see https://learn.microsoft.com/en-us/azure/iot-hub/iot-hub-mqtt-support#sending-device-to-cloud-messages.
func (d *Device) TelemetryTopic() string {
	return fmt.Sprintf("/devices/%v/messages/events", d.DeviceID)
}
