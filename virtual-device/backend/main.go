package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/lamassuiot/lamassu-vdevice/pkg/model"
	"github.com/lamassuiot/lamassu-vdevice/pkg/mqtt"
	"github.com/lamassuiot/lamassu-vdevice/pkg/service"
	"github.com/lamassuiot/lamassu-vdevice/pkg/transport"
)

func main() {
	logsChannel := make(chan mqtt.MQTTLog)

	awsMqttClient := mqtt.NewAWSIoTCoreMQTTClient(
		"a3hczhtwc7h4es.iot.eu-west-1.amazonaws.com",
		"/home/ikerlan/lamassu/v2.lamassu-virtual-demos/virtual-device/backend/aws-iotcore-ca.crt",
		logsChannel,
	)

	azureMqttClient := mqtt.NewAzureIotHubMQTTClient(
		"lamassu-hub.azure-devices.net",
		"/home/ikerlan/lamassu/v2.lamassu-virtual-demos/virtual-device/backend/azure-iothub-ca.crt",
		"global.azure-devices-provisioning.net",
		"0ne005927A2",
		logsChannel,
	)

	mqttInstances := map[model.CloudProviderType]mqtt.MqttDeviceService{}
	mqttInstances[model.CloudProviderTypeAWS] = awsMqttClient
	mqttInstances[model.CouldProviderTypeAzure] = azureMqttClient

	deviceState, chanDeviceUpdate := service.New("http://dev-lamassu.zpd.ikerlan.es:7002", "https://dev-lamassu.zpd.ikerlan.es", mqttInstances)

	wsHandler := transport.NewWebsocketHandler(deviceState, chanDeviceUpdate)

	spa := spaHandler{staticPath: "build", indexPath: "index.html"}
	router := mux.NewRouter()

	router.PathPrefix("/ws").HandlerFunc(wsHandler.MainRoute)
	router.PathPrefix("/").Handler(spa)

	srv := &http.Server{
		Handler: router,
		Addr:    ":7001",
	}

	go func() {
		for log := range logsChannel {
			fmt.Println(log)
			wsHandler.SendWebSocketMessage(transport.WebSocketMessage{
				Type:      "MQTT_LOG",
				Message:   log,
				Timestamp: time.Now(),
			})
		}
	}()

	log.Fatal(srv.ListenAndServe())
}

type spaHandler struct {
	staticPath string
	indexPath  string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)

}
