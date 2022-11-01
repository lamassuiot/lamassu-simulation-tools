package mqtt

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type awsIotCoreMQTT struct {
	awsIotCoreEndpoint string
	awsIotCoreCA       string
	logsChannel        chan MQTTLog
	mqttClient         *MQTT.Client
}

func NewAWSIoTCoreMQTTClient(endpoint, awsIotCoreCA string, logsChannel chan MQTTLog) MqttDeviceService {
	return &awsIotCoreMQTT{
		awsIotCoreEndpoint: endpoint,
		awsIotCoreCA:       awsIotCoreCA,
		logsChannel:        logsChannel,
	}
}

func (c *awsIotCoreMQTT) Connect(certificate *x509.Certificate, key *rsa.PrivateKey, deviceID string) error {
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile(c.awsIotCoreCA)
	if err != nil {
		return err
	}

	certpool.AppendCertsFromPEM(pemCerts)

	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw})
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	// pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.certificate.Raw})
	// pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(c.key)})
	tlsCert, err := tls.X509KeyPair(pemCert, pemKey)
	if err != nil {
		return err
	}

	tlsconfig := &tls.Config{
		RootCAs:            certpool,
		Certificates:       []tls.Certificate{tlsCert},
		InsecureSkipVerify: true,
	}

	opts := MQTT.NewClientOptions()
	opts.AddBroker("tls://" + c.awsIotCoreEndpoint + ":8883")
	opts.SetClientID(deviceID)
	opts.SetTLSConfig(tlsconfig)
	opts.SetDefaultPublishHandler(c.DefaultMessageHandler)
	opts.SetConnectionLostHandler(c.onConnectionLostHandler)

	mqttClient := MQTT.NewClient(opts)
	c.logsChannel <- MQTTLog{Type: MQTTTLogTypeInfo, Title: "Connecting to AWS IoT Core ..."}

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		sleepTime := time.Duration(time.Second * 5)
		c.logsChannel <- MQTTLog{Type: MQTTTLogTypeError, Title: "Error connecting to AWS IoT Core: " + token.Error().Error()}
		c.logsChannel <- MQTTLog{Type: MQTTTLogTypeInfo, Title: fmt.Sprintf("Retrying in %d seconds", sleepTime/time.Second)}
		time.Sleep(sleepTime)

		//Not sure if this is needed
		mqttClient.Disconnect(1000)
		mqttClient = MQTT.NewClient(opts)

		c.logsChannel <- MQTTLog{Type: MQTTTLogTypeInfo, Title: "Retrying ..."}
		if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
			c.logsChannel <- MQTTLog{Type: MQTTTLogTypeError, Title: "Error connecting to AWS IoT Core: " + token.Error().Error()}
			c.logsChannel <- MQTTLog{Type: MQTTTLogTypeError, Title: "Desisting"}
			return token.Error()
		}
	}

	c.mqttClient = &mqttClient
	return nil
}

func (c *awsIotCoreMQTT) onConnectionLostHandler(cl MQTT.Client, reason error) {
	c.logsChannel <- MQTTLog{Type: MQTTTLogTypeError, Title: "Connection Lost: " + reason.Error()}
}

func (c *awsIotCoreMQTT) DefaultMessageHandler(client MQTT.Client, msg MQTT.Message) {
	c.logsChannel <- MQTTLog{Type: MQTTTLogTypeInfo, Title: ">> Incoming message - using default message handler for topic: " + msg.Topic(), Message: string(msg.Payload())}
}

func (c *awsIotCoreMQTT) Publish(topic string, payload []byte) error {
	if condition := (*c.mqttClient).IsConnected(); !condition {
		fmt.Println("Not connected. Not publishing message: " + string(payload))
		return nil
	}

	resp := (*c.mqttClient).Publish(topic, 0, false, payload)
	if resp.WaitTimeout(time.Second*5) && resp.Error() != nil {
		c.logsChannel <- MQTTLog{Type: MQTTTLogTypeError, Title: "<< " + topic, Message: string(payload)}
		return resp.Error()
	}
	c.logsChannel <- MQTTLog{Type: MQTTTLogTypeInfo, Title: "<< " + topic, Message: string(payload)}
	return nil
}

func (c *awsIotCoreMQTT) Subscribe(topic string, callback func(topic string, payload []byte)) error {
	loggerFunc := func(client MQTT.Client, msg MQTT.Message) {
		payload := msg.Payload()
		topic := msg.Topic()
		fmt.Println(">> Incoming message - using custom message handler for topic: " + topic)
		c.logsChannel <- MQTTLog{Type: MQTTTLogTypeInfo, Title: ">> " + topic, Message: string(payload)}
		callback(topic, payload)
	}
	resp := (*c.mqttClient).Subscribe(topic, 0, loggerFunc)
	if resp.WaitTimeout(time.Second*5) && resp.Error() != nil {
		c.logsChannel <- MQTTLog{Type: MQTTTLogTypeError, Title: "[Subscribe] " + topic}
		return resp.Error()
	}
	c.logsChannel <- MQTTLog{Type: MQTTTLogTypeInfo, Title: "[Subscribe] " + topic}
	return nil
}

func (c *awsIotCoreMQTT) Disconnect() error {
	fmt.Println("disconnecting")
	(*c.mqttClient).Disconnect(uint(time.Second))
	return nil
}

func (c *awsIotCoreMQTT) IsConnected() bool {
	return (*c.mqttClient).IsConnected()
}
