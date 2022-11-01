package transport

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lamassuiot/lamassu-vdevice/pkg/model"
	"github.com/lamassuiot/lamassu-vdevice/pkg/service"
)

type WebsocketHandler struct {
	activeWebSocketConnection *websocket.Conn
	deviceService             service.DeviceService
	webSocketPublisherChan    chan []byte
}

type WebSocketMessage struct {
	Type      string      `json:"type"`
	Message   interface{} `json:"message"`
	Timestamp time.Time   `json:"timestamp"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func NewWebsocketHandler(svc service.DeviceService, updateDeviceStateChan chan model.DeviceState) *WebsocketHandler {
	fmt.Println("NewWebsocketHandler")
	wsSvc := &WebsocketHandler{
		activeWebSocketConnection: nil,
		deviceService:             svc,
		webSocketPublisherChan:    make(chan []byte),
	}

	//send device update messages to the websocket connection
	go func() {
		for updatedDevice := range updateDeviceStateChan {
			wsSvc.SendWebSocketMessage(WebSocketMessage{
				Type:      "DEVICE_STATE_UPDATE",
				Message:   updatedDevice.Serialize(),
				Timestamp: time.Now(),
			})
		}
	}()

	go func() {
		wsSvc.publishToWebsocket(wsSvc.webSocketPublisherChan)
	}()

	return wsSvc
}

func (ws *WebsocketHandler) messageHandler(inMessage WebSocketMessage) {
	bytesIn, err := json.Marshal(inMessage.Message)
	if err != nil {
		log.Println(err)
		return
	}

	switch inMessage.Type {
	case "CHANGE_TELEMETRY_DATA_RATE":
		type SpecificMessage struct {
			NewRate int `json:"new_rate"`
		}

		var newRateMessage SpecificMessage
		json.Unmarshal(bytesIn, &newRateMessage)

		ws.deviceService.UpdateGetSensorDataInterval(newRateMessage.NewRate)

	case "GEN_NEW_ID":
		ws.deviceService.ResetDeviceState()

	case "GEN_NEW_SLOT":
		ws.deviceService.GenerateNewSlot()

	case "ENROLL":
		type SpecificMessage struct {
			SlotID string `json:"slot_id"`
		}

		var msg SpecificMessage
		json.Unmarshal(bytesIn, &msg)

		ws.deviceService.Enroll(msg.SlotID)

	case "REENROLL":
		type SpecificMessage struct {
			SlotID string `json:"slot_id"`
		}

		var msg SpecificMessage
		json.Unmarshal(bytesIn, &msg)

		ws.deviceService.Reenroll(msg.SlotID)

	case "MQTT_CONNECT":
		type SpecificMessage struct {
			SlotID   string `json:"slot_id"`
			Provider string `json:"provider"`
		}

		var msg SpecificMessage
		json.Unmarshal(bytesIn, &msg)

		ws.deviceService.ConnectCloudProvider(model.CouldProviderTypeAzure, msg.SlotID)
	}
}

func (ws *WebsocketHandler) MainRoute(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	ws.activeWebSocketConnection = c

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read err:", err)
			continue
		}

		var inMessage WebSocketMessage
		fmt.Println("<< Incoming message: " + string(message))
		err = json.Unmarshal(message, &inMessage)
		if err == nil {
			ws.messageHandler(inMessage)
			continue
		}
		fmt.Println("err parsing message:", err)
	}
}

func (ws *WebsocketHandler) SendWebSocketMessage(message WebSocketMessage) {
	outBytes, err := json.Marshal(&message)
	if err != nil {
		log.Println("error parsing OUT message to JSON string:", err)
		return
	}

	ws.webSocketPublisherChan <- outBytes
}

func (ws *WebsocketHandler) publishToWebsocket(c chan []byte) {
	for {
		msg := <-c
		// fmt.Println(">> Sending messgae: " + string(msg))

		if ws.activeWebSocketConnection == nil {
			continue
		}

		err := ws.activeWebSocketConnection.WriteMessage(1, msg)
		if err != nil {
			log.Println("error sending message via WebSocket:", err)
			continue
		}

	}
}
