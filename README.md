# Azure IoT Hub over MQTT in Go

[![GoDoc](https://godoc.org/github.com/mtraver/iothub?status.svg)](https://godoc.org/github.com/mtraver/iothub)
[![Go Report Card](https://goreportcard.com/badge/github.com/mtraver/iothub)](https://goreportcard.com/report/github.com/mtraver/iothub)

Package iothub eases interaction with Azure IoT Hub over MQTT. It handles TLS configuration and authentication. It also makes it easy to construct the fully-qualified MQTT topics that IoT Hub uses for telemetry and cloud-to-device communication.

Currently only X.509 self-signed authentication is supported. See https://learn.microsoft.com/en-us/azure/iot-edge/how-to-authenticate-downstream-device?view=iotedge-1.4#x509-self-signed-authentication.

# Requirements

## Azure CA certs

You'll need Azure's trusted root certs. You can find them here: https://learn.microsoft.com/en-us/azure/security/fundamentals/azure-ca-details. Most importantly, as of Feb 2023, it should contain the DigiCert Global Root G2 cert.

Save them in a single .pem file and use its path when constructing a `Device`.
