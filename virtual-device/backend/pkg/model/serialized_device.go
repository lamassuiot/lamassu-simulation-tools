package model

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strconv"
)

type SerializedTelemetryData struct {
	BatteryLevel int `json:"battery_level"`
	Temperature  int `json:"temperature"`
	Humidity     int `json:"humidity"`
}

func (t *TelemetryData) Serialize() SerializedTelemetryData {
	return SerializedTelemetryData{
		BatteryLevel: t.BatteryLevel,
		Temperature:  t.Temperature,
		Humidity:     t.Humidity,
	}
}

type SerializedSlot struct {
	ID             string     `json:"id"`
	Certificate    string     `json:"certificate"`
	PrivateKey     string     `json:"private_key"`
	SerialNumber   string     `json:"serial_number"`
	Status         SlotStatus `json:"status"`
	IssuingCA      string     `json:"issuing_ca"`
	ExpirationDate string     `json:"expiration_date"`
}

func (s Slot) Serialize() SerializedSlot {
	b64PemKeyString := ""
	if s.PrivateKey != nil {
		pemKeyString := pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(s.PrivateKey),
			},
		)
		b64PemKeyString = base64.StdEncoding.EncodeToString(pemKeyString)
	}

	return SerializedSlot{
		ID:             s.ID,
		Certificate:    "certi",
		PrivateKey:     b64PemKeyString,
		Status:         s.Status,
		SerialNumber:   s.SerialNumber,
		IssuingCA:      s.IssuingCA,
		ExpirationDate: strconv.Itoa(int(s.ExpirationDate.Unix())),
	}
}

type SerializedDeviceState struct {
	Status                   DeviceStatus            `json:"status"`
	SerialNumber             string                  `json:"serial_number"`
	Model                    string                  `json:"model"`
	TelemetryData            SerializedTelemetryData `json:"telemetry_data"`
	TelemetryDataRateSeconds int                     `json:"telemetry_data_rate"`
	Slots                    []SerializedSlot        `json:"slots"`
	MqttClient               string                  `json:"mqtt_provider"`
	MqttConnected            bool                    `json:"mqtt_connected"`
}

func (d DeviceState) Serialize() SerializedDeviceState {
	serializedSlots := make([]SerializedSlot, len(d.Slots))
	for i, slot := range d.Slots {
		serializedSlots[i] = slot.Serialize()
	}

	return SerializedDeviceState{
		Status:                   d.Status,
		SerialNumber:             d.SerialNumber,
		Model:                    d.Model,
		TelemetryData:            d.TelemetryData.Serialize(),
		TelemetryDataRateSeconds: d.TelemetryDataRateSeconds,
		Slots:                    serializedSlots,
		MqttClient:               "mqtt",
		MqttConnected:            d.MqttConnected,
	}
}
