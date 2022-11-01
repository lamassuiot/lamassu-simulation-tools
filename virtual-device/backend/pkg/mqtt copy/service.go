package mqtt

import (
	"crypto/rsa"
	"crypto/x509"
)

type MQTTTLogType string

const (
	MQTTTLogTypeInfo    MQTTTLogType = "INFO"
	MQTTTLogTypeError   MQTTTLogType = "ERROR"
	MQTTTLogTypeSuccess MQTTTLogType = "SUCCESS"
)

type MQTTLog struct {
	Type      MQTTTLogType `json:"type"`
	Title     string       `json:"title"`
	Message   string       `json:"message"`
	Timestamp int          `json:"timestamp"`
}

type Service interface {
	Connect(certificate *x509.Certificate, key *rsa.PrivateKey) error
	IsConnected() bool

	Publish(topic string, payload []byte) error
	Subscribe(topic string, callback func(topic string, payload []byte)) error

	Disconnect() error
}
