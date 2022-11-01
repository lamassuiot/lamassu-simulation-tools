package model

import (
	"crypto/rsa"
	"crypto/x509"
	"time"
)

type SlotStatus string

const (
	SlotStatusNeedsProvisioning    SlotStatus = "NEEDS_PROVISIONING"
	SlotStatusPendingProvisioning  SlotStatus = "PENDING_PROVISIONING"
	SlotStatusProvisioned          SlotStatus = "PROVISIONED"
	SlotStatusNeedsReenrollment    SlotStatus = "NEEDS_REENROLLMENT"
	SlotStatusReenrollmentUnderway SlotStatus = "REENROLLMENT_UNDERWAY"
	SlotStatusExpired              SlotStatus = "EXPIRED"
)

type Slot struct {
	ID                 string
	Certificate        *x509.Certificate
	CertificateRequest *x509.CertificateRequest
	PrivateKey         *rsa.PrivateKey
	SerialNumber       string
	Status             SlotStatus
	IssuingCA          string
	ExpirationDate     time.Time
}

type DeviceStatus string

const (
	DeviceStatusEmpty  DeviceStatus = "EMPTY"
	DeviceStatusWithID DeviceStatus = "WITH_ID"
)

type DeviceState struct {
	Status                   DeviceStatus
	SerialNumber             string
	Model                    string
	TelemetryData            TelemetryData
	TelemetryDataRateSeconds int
	Slots                    []Slot
	MqttConnected            bool
}

type TelemetryData struct {
	Temperature  int
	Humidity     int
	BatteryLevel int
}

type CloudProviderType string

const (
	CloudProviderTypeAWS   CloudProviderType = "AWS"
	CouldProviderTypeAzure CloudProviderType = "AZURE"
)
