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
	"github.com/jakehl/goid"
)

type azureIotHubMQTT struct {
	azureIotHubEndpoint string
	azureIotHubCA       string
	azureDpsEndpoint    string
	azureScopeID        string
	logsChannel         chan MQTTLog
	mqttClient          *MQTT.Client
}

func NewAzureIotHubMQTTClient(azureIotHubEndpoint, azureIotHubCA string, azureDpsEndpoint string, azureScopeID string, logsChannel chan MQTTLog) MqttDeviceService {
	return &azureIotHubMQTT{
		azureIotHubEndpoint: azureIotHubEndpoint,
		azureIotHubCA:       azureIotHubCA,
		azureDpsEndpoint:    azureDpsEndpoint,
		azureScopeID:        azureScopeID,
		logsChannel:         logsChannel,
	}
}

func (c *azureIotHubMQTT) Connect(certificate *x509.Certificate, key *rsa.PrivateKey, deviceID string) error {
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile(c.azureIotHubCA)
	if err != nil {
		return err
	}

	certpool.AppendCertsFromPEM(pemCerts)

	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw})
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	fmt.Println(string(pemCert))
	fmt.Println(string(pemKey))

	tlsCert, err := tls.X509KeyPair(pemCert, pemKey)
	if err != nil {
		return err
	}

	tlsconfig := &tls.Config{
		RootCAs:            certpool,
		Certificates:       []tls.Certificate{tlsCert},
		InsecureSkipVerify: true,
	}

	c.logsChannel <- MQTTLog{Type: MQTTTLogTypeInfo, Title: "Connecting to DPS"}
	dpsOpts := MQTT.NewClientOptions()
	dpsOpts.AddBroker("tls://" + c.azureDpsEndpoint + ":8883")
	dpsOpts.SetClientID(deviceID)
	dpsOpts.SetTLSConfig(tlsconfig)
	dpsOpts.SetDefaultPublishHandler(c.DefaultMessageHandler)

	username := c.azureScopeID + "/registrations/" + deviceID + "/api-version=2019-03-31"
	fmt.Println(username)
	dpsOpts.SetUsername(username)

	mqttClient := MQTT.NewClient(dpsOpts)
	resp := mqttClient.Connect()
	if resp.WaitTimeout(time.Second*5) && resp.Error() != nil {
		c.logsChannel <- MQTTLog{Type: MQTTTLogTypeError, Title: "Connect to DPS error", Message: resp.Error().Error()}
		return resp.Error()
	}

	c.logsChannel <- MQTTLog{Type: MQTTTLogTypeSuccess, Title: "Connected to DPS"}

	c.mqttClient = &mqttClient

	err = c.Subscribe("$dps/registration/res/#", func(topic string, payload []byte) {
		fmt.Println(string(topic))
		fmt.Println(string(payload))
	})
	if err != nil {
		return err
	}

	reqID := goid.NewV4UUID().String()

	err = c.Publish(fmt.Sprintf("$dps/registrations/PUT/iotdps-register/?$rid=%s", reqID), []byte(`{"registrationId":"`+deviceID+`"}`))
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

	c.Disconnect()

	// time.Sleep(2 * time.Second)

	hubOpts := MQTT.NewClientOptions()
	hubOpts.AddBroker("tls://" + c.azureIotHubEndpoint + ":8883")
	hubOpts.SetClientID(deviceID)
	hubOpts.SetTLSConfig(tlsconfig)
	hubOpts.SetDefaultPublishHandler(c.DefaultMessageHandler)

	hubUsername := fmt.Sprintf("%s/%s/api-version=2016-11-14", c.azureIotHubEndpoint, deviceID)
	hubOpts.SetUsername(hubUsername)

	mqttClient = MQTT.NewClient(hubOpts)
	c.mqttClient = &mqttClient

	c.logsChannel <- MQTTLog{Type: MQTTTLogTypeInfo, Title: "Connecting to IotHub"}
	resp = mqttClient.Connect()
	if resp.WaitTimeout(time.Second*5) && resp.Error() != nil {
		c.logsChannel <- MQTTLog{Type: MQTTTLogTypeError, Title: "Connect to IotHub error", Message: resp.Error().Error()}
		return resp.Error()
	}
	c.logsChannel <- MQTTLog{Type: MQTTTLogTypeSuccess, Title: "Connected to IotHub"}

	if err != nil {
		return err
	}
	return nil
}

func (c *azureIotHubMQTT) DefaultMessageHandler(client MQTT.Client, msg MQTT.Message) {
	c.logsChannel <- MQTTLog{Type: MQTTTLogTypeInfo, Title: ">> Incoming messgae - using default message handler for topic: " + msg.Topic(), Message: string(msg.Payload())}
}

func (c *azureIotHubMQTT) Publish(topic string, payload []byte) error {
	resp := (*c.mqttClient).Publish(topic, 0, false, payload)
	if resp.WaitTimeout(time.Second*5) && resp.Error() != nil {
		c.logsChannel <- MQTTLog{Type: MQTTTLogTypeError, Title: "<< " + topic, Message: string(payload)}
		return resp.Error()
	}
	c.logsChannel <- MQTTLog{Type: MQTTTLogTypeInfo, Title: "<< " + topic, Message: string(payload)}
	return nil
}

func (c *azureIotHubMQTT) Subscribe(topic string, callback func(topic string, payload []byte)) error {
	loggerFunc := func(client MQTT.Client, msg MQTT.Message) {
		payload := msg.Payload()
		topic := msg.Topic()
		c.logsChannel <- MQTTLog{Type: MQTTTLogTypeInfo, Title: "aaa >> " + topic, Message: string(payload)}
		callback(topic, payload)
	}
	resp := (*c.mqttClient).Subscribe(topic, 0, loggerFunc)
	if resp.WaitTimeout(time.Second*5) && resp.Error() != nil {
		c.logsChannel <- MQTTLog{Type: MQTTTLogTypeError, Title: ">> " + topic}
		return resp.Error()
	}
	c.logsChannel <- MQTTLog{Type: MQTTTLogTypeInfo, Title: ">> " + topic}
	return nil
}

func (c *azureIotHubMQTT) Disconnect() error {
	(*c.mqttClient).Disconnect(uint(time.Second))
	return nil
}

func (c *azureIotHubMQTT) IsConnected() bool {
	return (*c.mqttClient).IsConnected()
}
