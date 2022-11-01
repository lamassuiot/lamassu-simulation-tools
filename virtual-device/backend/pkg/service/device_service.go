package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	mathRand "math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jakehl/goid"
	"github.com/lamassuiot/lamassu-vdevice/pkg/model"
	"github.com/lamassuiot/lamassu-vdevice/pkg/mqtt"
	"github.com/lamassuiot/lamassu-vdevice/pkg/service/store"
	estClient "github.com/lamassuiot/lamassuiot/pkg/est/client"
	"github.com/lamassuiot/lamassuiot/pkg/utils"
	"github.com/robfig/cron/v3"
	"golang.org/x/exp/slices"
)

type DeviceServiceImpl struct {
	deviceStore *store.DeviceStateStore

	cronInstance        *cron.Cron
	telemetryDataCronID *cron.EntryID

	dmsUrl            string
	lamassuGatewayURL string

	mqttProviderInstances map[model.CloudProviderType]mqtt.MqttDeviceService
}

type DeviceService interface {
	ResetDeviceState()

	GenerateNewSlot()

	Enroll(slotID string) error
	Reenroll(slotID string) error

	ConnectCloudProvider(cloudProvider model.CloudProviderType, slotID string) error
	DisconnectCloudProvider()

	GetSensorData()
	UpdateGetSensorDataInterval(interval int) // in seconds
}

// returns a new instance of DeviceServiceImpl and a channel to receive updates of the device state
func New(dmsUrl, lamassuGatewayURL string, mqttProviderInstances map[model.CloudProviderType]mqtt.MqttDeviceService) (DeviceService, chan model.DeviceState) {
	c := cron.New(cron.WithSeconds())
	c.Start()

	deviceStateStore, updateDeviceStateChannel := store.New()

	svc := DeviceServiceImpl{
		deviceStore:           deviceStateStore,
		cronInstance:          c,
		dmsUrl:                dmsUrl,
		lamassuGatewayURL:     lamassuGatewayURL,
		mqttProviderInstances: mqttProviderInstances,
	}

	fmt.Println("Initializing device state")
	svc.ResetDeviceState()

	return &svc, updateDeviceStateChannel
}

func (d *DeviceServiceImpl) ResetDeviceState() {
	defaultSlots := []model.Slot{
		{
			ID:           "default",
			Status:       model.SlotStatusNeedsProvisioning,
			Certificate:  nil,
			PrivateKey:   nil,
			SerialNumber: "",
		},
	}

	newDeviceState := model.DeviceState{
		Status:                   model.DeviceStatusEmpty,
		SerialNumber:             goid.NewV4UUID().String(),
		Model:                    "Raspberry Pi 4",
		TelemetryDataRateSeconds: 5,
		TelemetryData:            model.TelemetryData{},
		Slots:                    defaultSlots,
		MqttConnected:            false,
	}

	d.deviceStore.SetDeviceState(&newDeviceState)

	d.UpdateGetSensorDataInterval(newDeviceState.TelemetryDataRateSeconds)
}

func (d *DeviceServiceImpl) GenerateNewSlot() {
	device := d.deviceStore.GetDeviceState()

	device.Slots = append(device.Slots, model.Slot{
		ID:           strconv.Itoa(len(device.Slots)),
		Status:       model.SlotStatusNeedsProvisioning,
		Certificate:  nil,
		PrivateKey:   nil,
		SerialNumber: "",
		IssuingCA:    "",
	})

	d.deviceStore.SetDeviceState(device)
}

func (d *DeviceServiceImpl) Enroll(slotID string) error {
	idx := slices.IndexFunc(d.deviceStore.GetDeviceState().Slots, func(s model.Slot) bool { return s.ID == slotID })
	if idx == -1 {
		fmt.Println("device not found")
		return fmt.Errorf("slot with id %s not found", slotID)
	}

	fmt.Println(idx)
	keyBytes, _ := rsa.GenerateKey(rand.Reader, 2048)
	device := d.deviceStore.GetDeviceState()
	slot := device.Slots[idx]

	slot.PrivateKey = keyBytes
	slot.Certificate = nil
	slot.Status = model.SlotStatusPendingProvisioning

	device.Slots[idx] = slot
	d.deviceStore.SetDeviceState(device)

	commonName := d.deviceStore.GetDeviceState().SerialNumber
	if slot.ID != "default" {
		commonName = slot.ID + ":" + commonName
	}

	subj := pkix.Name{
		CommonName:         commonName,
		Country:            []string{"ES"},
		Province:           []string{"Gipuzkoa"},
		Locality:           []string{"Donostia"},
		Organization:       []string{"Lamassu"},
		OrganizationalUnit: []string{"IT"},
	}

	template := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrBytes, _ := x509.CreateCertificateRequest(rand.Reader, &template, keyBytes)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
	pemString := base64.StdEncoding.EncodeToString(pemBytes)

	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("error parsing certificate request: %v", err)
	}

	slot.CertificateRequest = csr

	values := map[string]string{
		"serial_number":       d.deviceStore.GetDeviceState().SerialNumber,
		"model":               d.deviceStore.GetDeviceState().Model,
		"slot":                slot.ID,
		"certificate_request": pemString,
	}
	json_data, _ := json.Marshal(values)
	fmt.Println(string(json_data))

	resp, err := http.Post(d.dmsUrl+"/enroll", "application/json", bytes.NewReader(json_data))
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("error sending enrollment request: %v", err)
	}

	fmt.Println(resp.StatusCode)

	type EnrollMessageOut struct {
		IssuingCA   string `json:"issuing_ca"`
		Certificate string `json:"certificate"`
	}

	enrollRespBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading enrollment response: %v", err)
	}

	var enrollResp EnrollMessageOut
	json.Unmarshal(enrollRespBytes, &enrollResp)

	decodedCert, err := base64.StdEncoding.DecodeString(string(enrollResp.Certificate))
	if err != nil {
		return fmt.Errorf("error decoding certificate: %v", err)
	}

	certBlock, _ := pem.Decode(decodedCert)
	certificate, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("error parsing certificate: %v", err)
	}

	slot.Certificate = certificate
	slot.SerialNumber = utils.InsertNth(utils.ToHexInt(certificate.SerialNumber), 2)
	slot.IssuingCA = enrollResp.IssuingCA
	slot.ExpirationDate = certificate.NotAfter
	slot.Status = model.SlotStatusProvisioned

	device.Slots[idx] = slot
	d.deviceStore.SetDeviceState(device)

	return nil
}

func (d *DeviceServiceImpl) Reenroll(slotID string) error {
	device := d.deviceStore.GetDeviceState()
	idx := slices.IndexFunc(device.Slots, func(s model.Slot) bool { return s.ID == slotID })
	if idx == -1 {
		return fmt.Errorf("slot with id %s not found", slotID)
	}

	slot := device.Slots[idx]
	slot.Status = model.SlotStatusReenrollmentUnderway
	device.Slots[idx] = slot
	d.deviceStore.SetDeviceState(device)

	devManagerUrl, err := url.Parse(d.lamassuGatewayURL)
	if err != nil {
		return fmt.Errorf("error parsing lamassu gateway url: %v", err)
	}

	devManagerUrl.Path = "api/devmanager"
	client, err := estClient.NewESTClient(nil, devManagerUrl, slot.Certificate, slot.PrivateKey, nil, true)
	if err != nil {
		return fmt.Errorf("error creating EST client: %v", err)
	}

	ctx := context.Background()
	crt, err := client.Reenroll(ctx, slot.CertificateRequest)
	if err != nil {
		return fmt.Errorf("error reenrolling: %v", err)
	}

	slot.Certificate = crt
	slot.SerialNumber = utils.InsertNth(utils.ToHexInt(crt.SerialNumber), 2)
	slot.ExpirationDate = crt.NotAfter
	slot.Status = model.SlotStatusProvisioned
	device.Slots[idx] = slot
	d.deviceStore.SetDeviceState(device)

	return nil
}

func (d *DeviceServiceImpl) ConnectCloudProvider(cloudProvider model.CloudProviderType, slotID string) error {
	device := d.deviceStore.GetDeviceState()
	idx := slices.IndexFunc(device.Slots, func(s model.Slot) bool { return s.ID == slotID })
	if idx == -1 {
		return fmt.Errorf("slot with id %s not found", slotID)
	}

	slot := device.Slots[idx]

	mqttClient := d.mqttProviderInstances[model.CouldProviderTypeAzure]
	err := mqttClient.Connect(slot.Certificate, slot.PrivateKey, device.SerialNumber)
	if err != nil {
		return fmt.Errorf("error connecting to cloud provider: %v", err)
	}

	switch cloudProvider {
	case model.CouldProviderTypeAzure:
		err := mqttClient.Subscribe("$iothub/twin/PATCH/properties/desired/#", func(topic string, payload []byte) {
			var req struct {
				Reenroll bool   `json:"require_reenrollment"`
				Version  string `json:"version"`
			}
			json.Unmarshal(payload, &req)
			if req.Reenroll {
				d.Reenroll(slot.ID)
			}
		})

		fmt.Println("could not subscribe to azure twin topic", err)

	case model.CloudProviderTypeAWS:
		mqttClient.Subscribe(fmt.Sprintf("$aws/things/%s", device.SerialNumber), func(topic string, payload []byte) {
			fmt.Printf("====================================ACCEPTED==========================================\n")
			fmt.Printf("TOPIC: %s\n", topic)
			fmt.Printf("MSG: %s\n", payload)
			fmt.Printf("==============================================================================\n\n")
		})

		// significa que no hay shadow y hay que crearlo
		mqttClient.Subscribe(fmt.Sprintf("$aws/things/%s/shadow/get/rejected", device.SerialNumber), func(topic string, payload []byte) {
			fmt.Printf("====================================REJECTED==========================================\n")
			fmt.Printf("TOPIC: %s\n", topic)
			fmt.Printf("MSG: %s\n", payload)
			fmt.Printf("Publishing on update topic to create device shadow\n")
			fmt.Printf("==============================================================================\n\n")

			deviceState := `{
				"state": {
					"reported" : {
						"need_rotation" : false
					}
				}
			}`
			mqttClient.Publish(fmt.Sprintf("$aws/things/%s/shadow/update", device.SerialNumber), []byte(deviceState))
			fmt.Printf("==============================================================================\n")
		})

		mqttClient.Subscribe(fmt.Sprintf("$aws/things/%s/shadow/update/delta", device.SerialNumber), func(topic string, payload []byte) {
			fmt.Printf("====================================DELTA==========================================\n")
			fmt.Printf("TOPIC: %s\n", topic)
			fmt.Printf("MSG: %s\n", payload)
			fmt.Printf("Here goes the reenrollment process.")
			fmt.Printf("==============================================================================\n\n")
		})
		time.Sleep(time.Second * 2)

		mqttClient.Publish(fmt.Sprintf("$aws/things/%s/shadow/get", device.SerialNumber), []byte{})
		time.Sleep(time.Second * 2)
	}

	return nil
}

func (d *DeviceServiceImpl) DisconnectCloudProvider() {

}

func (d *DeviceServiceImpl) GetSensorData() {
	device := d.deviceStore.GetDeviceState()

	device.TelemetryData = model.TelemetryData{
		Temperature:  mathRand.Intn(25),
		Humidity:     mathRand.Intn(50),
		BatteryLevel: mathRand.Intn(101),
	}

	d.deviceStore.SetDeviceState(device)

}

func (d *DeviceServiceImpl) UpdateGetSensorDataInterval(interval int) {
	fmt.Println("UpdateGetSensorDataInterval")

	if d.telemetryDataCronID != nil {
		d.cronInstance.Remove(*d.telemetryDataCronID)
	}

	newTelemetryDataCronID, err := d.cronInstance.AddFunc(fmt.Sprintf("0/%d * * * * *", interval), d.GetSensorData)
	if err != nil {
		fmt.Println("error adding cron job for telemetry data", err)
		return
	}

	d.telemetryDataCronID = &newTelemetryDataCronID
}
