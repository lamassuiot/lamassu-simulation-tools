package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kelseyhightower/envconfig"
	dmsManagerClient "github.com/lamassuiot/lamassuiot/pkg/dms-manager/client"
	dmsApi "github.com/lamassuiot/lamassuiot/pkg/dms-manager/common/api"
	estClient "github.com/lamassuiot/lamassuiot/pkg/est/client"
	"github.com/lamassuiot/lamassuiot/pkg/utils"
	"github.com/lamassuiot/lamassuiot/pkg/utils/client"
	"github.com/robfig/cron/v3"
)

type DMSStatus string

const (
	DMSStatusEmpty        DMSStatus = "EMPTY"
	DMSStatusAwaitingAuth DMSStatus = "AWAITING_AUTH"
	DMSStatusIdle         DMSStatus = "IDLE"
	DMSStatusEnrolling    DMSStatus = "ENROLLING"
)

type DMSState struct {
	Status                       DMSStatus
	Name                         string
	Certificate                  *x509.Certificate
	PrivateKey                   *rsa.PrivateKey
	AuthorizedCAs                []string
	SelectedCAForEnrollment      string
	AutomaticEnrollment          bool
	AutomaticCertificateTransfer bool
}

type DMSStateSerialized struct {
	Status                       DMSStatus `json:"status"`
	Name                         string    `json:"name"`
	AuthorizedCAs                []string  `json:"authorized_cas"`
	SelectedCAForEnrollment      string    `json:"selected_ca_for_enrollment"`
	AutomaticEnrollment          bool      `json:"automatic_enrollment"`
	AutomaticCertificateTransfer bool      `json:"automatic_certificate_transfer"`
}

func (s *DMSState) Serialize() DMSStateSerialized {
	authCAs := []string{}
	if s.AuthorizedCAs != nil {
		authCAs = s.AuthorizedCAs
	}

	return DMSStateSerialized{
		Status:                       s.Status,
		Name:                         s.Name,
		AuthorizedCAs:                authCAs,
		SelectedCAForEnrollment:      s.SelectedCAForEnrollment,
		AutomaticEnrollment:          s.AutomaticEnrollment,
		AutomaticCertificateTransfer: s.AutomaticCertificateTransfer,
	}
}

// -------------------------------------------------------------

type EnrollmentStatus string

const (
	EnrollingStatusStep1 EnrollmentStatus = "STEP_1"
	EnrollingStatusStep2 EnrollmentStatus = "STEP_2"
	EnrollingStatusStep3 EnrollmentStatus = "STEP_3"
	EnrollingStatusStep4 EnrollmentStatus = "STEP_4"
)

type EnrollmentInProcess struct {
	Status                        EnrollmentStatus
	RequestingDate                time.Time
	DeviceModel                   string
	IssuingCA                     string
	DeviceID                      string
	DeviceSlot                    string
	CertificateSigningRequest     *x509.CertificateRequest
	AuthorizedEnrollment          bool
	Certificate                   *x509.Certificate
	SerialNumber                  string
	ExpirationDate                time.Time
	AuthorizedCertificateTransfer bool
}

type EnrollmentInProcessSerialized struct {
	Status                        EnrollmentStatus `json:"status"`
	RequestingDate                time.Time        `json:"requesting_date"`
	DeviceModel                   string           `json:"device_model"`
	IssuingCA                     string           `json:"issuing_ca"`
	DeviceID                      string           `json:"device_id"`
	DeviceSlot                    string           `json:"device_slot"`
	CertificateSigningRequest     string           `json:"certificate_request"`
	AuthorizedEnrollment          bool             `json:"authorized_enrollment"`
	Certificate                   string           `json:"certificate"`
	SerialNumber                  string           `json:"serial_number"`
	ExpirationDate                time.Time        `json:"expiration_date"`
	AuthorizedCertificateTransfer bool             `json:"authorized_certificate_transfer"`
}

func (s *EnrollmentInProcess) Serialize() EnrollmentInProcessSerialized {
	csr := ""
	if s.CertificateSigningRequest != nil {
		if len(s.CertificateSigningRequest.Subject.Country) > 0 {
			csr += "C=" + s.CertificateSigningRequest.Subject.Country[0] + "/"
		}
		if len(s.CertificateSigningRequest.Subject.Province) > 0 {
			csr += "ST=" + s.CertificateSigningRequest.Subject.Province[0] + "/"
		}
		if len(s.CertificateSigningRequest.Subject.Locality) > 0 {
			csr += "L=" + s.CertificateSigningRequest.Subject.Locality[0] + "/"
		}
		if len(s.CertificateSigningRequest.Subject.Organization) > 0 {
			csr += "O=" + s.CertificateSigningRequest.Subject.Organization[0] + "/"
		}
		if len(s.CertificateSigningRequest.Subject.Country) > 0 {
			csr += "OU=" + s.CertificateSigningRequest.Subject.OrganizationalUnit[0] + "/"
		}
		csr += "CN=" + s.CertificateSigningRequest.Subject.CommonName
	}

	crt := ""
	if s.Certificate != nil {
		if len(s.Certificate.Subject.Country) > 0 {
			crt += "C=" + s.Certificate.Subject.Country[0] + "/"
		}
		if len(s.Certificate.Subject.Province) > 0 {
			crt += "ST=" + s.Certificate.Subject.Province[0] + "/"
		}
		if len(s.Certificate.Subject.Locality) > 0 {
			crt += "L=" + s.Certificate.Subject.Locality[0] + "/"
		}
		if len(s.Certificate.Subject.Organization) > 0 {
			crt += "O=" + s.Certificate.Subject.Organization[0] + "/"
		}
		if len(s.Certificate.Subject.Country) > 0 {
			crt += "OU=" + s.Certificate.Subject.OrganizationalUnit[0] + "/"
		}
		crt += "CN=" + s.Certificate.Subject.CommonName
	}

	return EnrollmentInProcessSerialized{
		Status:                        s.Status,
		RequestingDate:                s.RequestingDate,
		DeviceModel:                   s.DeviceModel,
		DeviceID:                      s.DeviceID,
		DeviceSlot:                    s.DeviceSlot,
		CertificateSigningRequest:     csr,
		AuthorizedEnrollment:          s.AuthorizedEnrollment,
		Certificate:                   crt,
		SerialNumber:                  s.SerialNumber,
		ExpirationDate:                s.ExpirationDate,
		AuthorizedCertificateTransfer: s.AuthorizedCertificateTransfer,
		IssuingCA:                     s.IssuingCA,
	}
}

type EnrolledIdentity struct {
	EnrolledTimestamp time.Time
	SerialNumber      string
	DeviceID          string
	DeviceSlot        string
	IssuingCA         string
	IssuingDuration   time.Duration
}

type EnrolledIdentitySerialized struct {
	EnrolledTimestamp int    `json:"enrolled_timestamp"`
	SerialNumber      string `json:"serial_number"`
	DeviceID          string `json:"device_id"`
	DeviceSlot        string `json:"device_slot"`
	IssuingCA         string `json:"issuing_ca"`
	IssuingDuration   int    `json:"issuing_duration"`
}

func (s *EnrolledIdentity) Serialize() EnrolledIdentitySerialized {
	return EnrolledIdentitySerialized{
		EnrolledTimestamp: int(s.EnrolledTimestamp.UnixMilli()),
		SerialNumber:      s.SerialNumber,
		DeviceID:          s.DeviceID,
		DeviceSlot:        s.DeviceSlot,
		IssuingCA:         s.IssuingCA,
		IssuingDuration:   int(s.IssuingDuration.Seconds()),
	}
}

// -------------------------------------------------------------

type Singelton struct {
	ActiveWebSocketConnection *websocket.Conn
	DMS                       DMSState
	EnrollmentInProcess       *EnrollmentInProcess
	DMSManagerClient          dmsManagerClient.LamassuDMSManagerClient
	CronInstance              *cron.Cron
	PeriodicDMSCheckCronID    cron.EntryID
	EnrolledIdentities        []EnrolledIdentity
	LamassuGatewayURL         url.URL
}

var SingeltonInstance *Singelton

type Cfg struct {
	OperatorUsername string `json:"operator_username"`
	OperatorPassword string `json:"operator_password"`
	DMSName          string `json:"dms_name"`
}
type CfgAutoEnrollment struct {
	AutoEnroll bool `json:"auto_enroll"`
}
type CfgSelectedCAForEnrollment struct {
	SelectedCA string `json:"selected_ca"`
}
type CfgAutoTransfer struct {
	AutoTransfer bool `json:"auto_transfer"`
}

type WebSocketMessage struct {
	Type      string      `json:"type"`
	Message   interface{} `json:"message"`
	Timestamp time.Time   `json:"timestamp"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func messageHandler(inMessage WebSocketMessage, connection *websocket.Conn) {
	switch inMessage.Type {
	case "GET_CFG":
		sendWebSocketMessage(
			WebSocketMessage{
				Type:      "DMS_UPDATE",
				Message:   SingeltonInstance.DMS.Serialize(),
				Timestamp: time.Now(),
			},
		)

	case "CFG":
		bytesIn, err := json.Marshal(inMessage.Message)
		if err != nil {
			sendWebSocketMessage(
				WebSocketMessage{
					Type:      "ERROR",
					Message:   "Error parsing command 1",
					Timestamp: time.Now(),
				},
			)
			return
		}
		var cfg Cfg
		json.Unmarshal(bytesIn, &cfg)

		dmsUrl := SingeltonInstance.LamassuGatewayURL
		dmsUrl.Path = "api/dmsmanager"
		authUrl := SingeltonInstance.LamassuGatewayURL
		authUrl.Host = "auth." + authUrl.Host
		dmsCli, err := dmsManagerClient.NewLamassuDMSManagerClientConfig(client.BaseClientConfigurationuration{
			URL:        &dmsUrl,
			AuthMethod: client.AuthMethodJWT,
			AuthMethodConfig: &client.JWTConfig{
				Username:      cfg.OperatorUsername,
				Password:      cfg.OperatorPassword,
				URL:           &authUrl,
				CACertificate: "",
				Insecure:      true,
			},
			Insecure: true,
		})

		if err != nil {
			sendWebSocketMessage(
				WebSocketMessage{
					Type:      "ERROR",
					Message:   "Error creating DMS Client: " + err.Error(),
					Timestamp: time.Now(),
				},
			)
			return
		}

		dms, err := dmsCli.CreateDMS(context.Background(), &dmsApi.CreateDMSInput{
			Subject: dmsApi.Subject{
				CommonName:       cfg.DMSName,
				Organization:     "Lamassu",
				OrganizationUnit: "IT",
				Country:          "ES",
			},
			KeyMetadata: dmsApi.KeyMetadata{
				KeyType: dmsApi.RSA,
				KeyBits: 4096,
			},
		})

		if err != nil {
			sendWebSocketMessage(
				WebSocketMessage{
					Type:      "ERROR",
					Message:   "Error creating DMS Instance: " + err.Error(),
					Timestamp: time.Now(),
				},
			)
			return
		}

		keyPEMString, err := base64.StdEncoding.DecodeString(dms.PrivateKey.(string))
		if err != nil {
			sendWebSocketMessage(
				WebSocketMessage{
					Type:      "ERROR",
					Message:   "Error decoding private key: " + err.Error(),
					Timestamp: time.Now(),
				},
			)
			return
		}

		keyBlock, _ := pem.Decode(keyPEMString)

		key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			sendWebSocketMessage(
				WebSocketMessage{
					Type:      "ERROR",
					Message:   "Error parsing private key: " + err.Error(),
					Timestamp: time.Now(),
				},
			)
			return
		}

		SingeltonInstance.DMS = DMSState{
			Status:     DMSStatusAwaitingAuth,
			Name:       cfg.DMSName,
			PrivateKey: key,
		}
		checkID, err := SingeltonInstance.CronInstance.AddFunc("0/5 * * * * *", func() {
			dms, err := dmsCli.GetDMSByName(context.Background(), &dmsApi.GetDMSByNameInput{
				Name: SingeltonInstance.DMS.Name,
			})
			if err != nil {
				sendWebSocketMessage(
					WebSocketMessage{
						Type:      "ERROR",
						Message:   "Error checking DMS status: " + err.Error(),
						Timestamp: time.Now(),
					},
				)
				return
			}

			if dms.Status != dmsApi.DMSStatusApproved {
				return
			}

			SingeltonInstance.DMS.AuthorizedCAs = dms.AuthorizedCAs
			if len(dms.AuthorizedCAs) > 0 && SingeltonInstance.DMS.SelectedCAForEnrollment == "" {
				SingeltonInstance.DMS.SelectedCAForEnrollment = dms.AuthorizedCAs[0]
			}
			SingeltonInstance.DMS.Status = DMSStatusIdle
			SingeltonInstance.DMS.Certificate = dms.X509Asset.Certificate

			sendWebSocketMessage(
				WebSocketMessage{
					Type:      "DMS_UPDATE",
					Message:   SingeltonInstance.DMS.Serialize(),
					Timestamp: time.Now(),
				},
			)

			// SingeltonInstance.CronInstance.Remove(SingeltonInstance.PeriodicDMSCheckCronID)
		})
		if err != nil {
			sendWebSocketMessage(
				WebSocketMessage{
					Type:      "ERROR",
					Message:   "Error creating DMS Instance: " + err.Error(),
					Timestamp: time.Now(),
				},
			)
			return
		}
		SingeltonInstance.PeriodicDMSCheckCronID = checkID

		sendWebSocketMessage(
			WebSocketMessage{
				Type:      "DMS_UPDATE",
				Message:   SingeltonInstance.DMS.Serialize(),
				Timestamp: time.Now(),
			},
		)

	case "CFG_SELECTED_CA_FOR_ENROLLMENT":
		bytesIn, err := json.Marshal(inMessage.Message)
		if err != nil {
			sendWebSocketMessage(
				WebSocketMessage{
					Type:      "ERROR",
					Message:   "Error parsing command 1",
					Timestamp: time.Now(),
				},
			)
			return
		}

		var cfgSelectedCAForEnrollment CfgSelectedCAForEnrollment
		json.Unmarshal(bytesIn, &cfgSelectedCAForEnrollment)

		if len(cfgSelectedCAForEnrollment.SelectedCA) > 0 {
			SingeltonInstance.DMS.SelectedCAForEnrollment = cfgSelectedCAForEnrollment.SelectedCA
		} else {
			return
		}

		sendWebSocketMessage(
			WebSocketMessage{
				Type:      "DMS_UPDATE",
				Message:   SingeltonInstance.DMS.Serialize(),
				Timestamp: time.Now(),
			},
		)

	case "CFG_AUTO_ENROLLMENT":
		bytesIn, err := json.Marshal(inMessage.Message)
		if err != nil {
			sendWebSocketMessage(
				WebSocketMessage{
					Type:      "ERROR",
					Message:   "Error parsing command 1",
					Timestamp: time.Now(),
				},
			)
			return
		}

		var cfgAutoEnrollment CfgAutoEnrollment
		json.Unmarshal(bytesIn, &cfgAutoEnrollment)

		SingeltonInstance.DMS.AutomaticEnrollment = cfgAutoEnrollment.AutoEnroll
		fmt.Println(SingeltonInstance.DMS.AutomaticEnrollment)
		fmt.Println(SingeltonInstance.DMS.AutomaticCertificateTransfer)

		sendWebSocketMessage(
			WebSocketMessage{
				Type:      "DMS_UPDATE",
				Message:   SingeltonInstance.DMS.Serialize(),
				Timestamp: time.Now(),
			},
		)

	case "CFG_AUTO_TRANSFER":
		bytesIn, err := json.Marshal(inMessage.Message)
		if err != nil {
			sendWebSocketMessage(
				WebSocketMessage{
					Type:      "ERROR",
					Message:   "Error parsing command 1",
					Timestamp: time.Now(),
				},
			)
			return
		}

		var cfgAutoTransfer CfgAutoTransfer
		json.Unmarshal(bytesIn, &cfgAutoTransfer)

		SingeltonInstance.DMS.AutomaticCertificateTransfer = cfgAutoTransfer.AutoTransfer
		fmt.Println(SingeltonInstance.DMS.AutomaticEnrollment)
		fmt.Println(SingeltonInstance.DMS.AutomaticCertificateTransfer)

		sendWebSocketMessage(
			WebSocketMessage{
				Type:      "DMS_UPDATE",
				Message:   SingeltonInstance.DMS.Serialize(),
				Timestamp: time.Now(),
			},
		)

	case "AUTH_ENROLL":
		if SingeltonInstance.EnrollmentInProcess != nil {
			SingeltonInstance.EnrollmentInProcess.AuthorizedEnrollment = true
		}

	case "AUTH_TRANSFER":
		if SingeltonInstance.EnrollmentInProcess != nil {
			SingeltonInstance.EnrollmentInProcess.AuthorizedCertificateTransfer = true
		}
	}
}

func sendWebSocketMessage(message WebSocketMessage) {
	outBytes, err := json.Marshal(&message)
	if err != nil {
		log.Println("error parsing OUT message to JSON string:", err)
		return
	}

	color.Cyan("<< Sending messgae: " + string(outBytes))

	if SingeltonInstance.ActiveWebSocketConnection == nil {
		return
	}

	err = SingeltonInstance.ActiveWebSocketConnection.WriteMessage(1, outBytes)
	if err != nil {
		log.Println("error sending message via WebSocket:", err)
		return
	}
}

func enrollRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// if SingeltonInstance.DMS.Status != DMSStatusIdle {
		// 	http.Error(w, "DMS is not on a valid state",
		// 		http.StatusBadRequest)
		// }

		SingeltonInstance.DMS.Status = DMSStatusEnrolling

		SingeltonInstance.EnrollmentInProcess = &EnrollmentInProcess{
			AuthorizedEnrollment:          false,
			AuthorizedCertificateTransfer: false,
			IssuingCA:                     SingeltonInstance.DMS.SelectedCAForEnrollment,
		}

		if SingeltonInstance.DMS.AutomaticCertificateTransfer {
			SingeltonInstance.EnrollmentInProcess.AuthorizedCertificateTransfer = true
		}

		if SingeltonInstance.DMS.AutomaticEnrollment {
			SingeltonInstance.EnrollmentInProcess.AuthorizedEnrollment = true
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			SingeltonInstance.DMS.Status = DMSStatusIdle
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}

		type EnrollMessage struct {
			SerialNumber       string `json:"serial_number"`
			Model              string `json:"model"`
			Slot               string `json:"slot"`
			CertificateRequest string `json:"certificate_request"`
		}

		var enrollMsg EnrollMessage
		err = json.Unmarshal(body, &enrollMsg)
		if err != nil {
			SingeltonInstance.DMS.Status = DMSStatusIdle
			http.Error(w, "Error parsing request body",
				http.StatusInternalServerError)
		}

		SingeltonInstance.EnrollmentInProcess.DeviceID = enrollMsg.SerialNumber
		SingeltonInstance.EnrollmentInProcess.DeviceModel = enrollMsg.Model
		SingeltonInstance.EnrollmentInProcess.DeviceSlot = enrollMsg.Slot
		SingeltonInstance.EnrollmentInProcess.RequestingDate = time.Now()

		decodedCsr, err := base64.StdEncoding.DecodeString(enrollMsg.CertificateRequest)
		if err != nil {
			SingeltonInstance.DMS.Status = DMSStatusIdle
			http.Error(w, "Error decoding csr body",
				http.StatusInternalServerError)
		}

		parsedCsrPem, _ := pem.Decode(decodedCsr)
		csr, err := x509.ParseCertificateRequest(parsedCsrPem.Bytes)
		if err != nil {
			SingeltonInstance.DMS.Status = DMSStatusIdle
			http.Error(w, "Error inflating csr body",
				http.StatusInternalServerError)
		}

		SingeltonInstance.EnrollmentInProcess.CertificateSigningRequest = csr
		SingeltonInstance.EnrollmentInProcess.Status = EnrollingStatusStep1
		sendWebSocketMessage(
			WebSocketMessage{
				Type:      "ENROLLING_PROCESS_UPDATE",
				Message:   SingeltonInstance.EnrollmentInProcess.Serialize(),
				Timestamp: time.Now(),
			},
		)

		for {
			if SingeltonInstance.EnrollmentInProcess.AuthorizedEnrollment {
				break
			}
			time.Sleep(1 * time.Second)
		}

		devManagerUrl := SingeltonInstance.LamassuGatewayURL
		devManagerUrl.Path = "api/devmanager"

		client, err := estClient.NewESTClient(
			nil,
			&devManagerUrl,
			SingeltonInstance.DMS.Certificate,
			SingeltonInstance.DMS.PrivateKey,
			nil,
			true,
		)
		if err != nil {
			SingeltonInstance.DMS.Status = DMSStatusIdle
			http.Error(w, "Error creating EST client",
				http.StatusInternalServerError)
		}

		ctx := context.Background()
		// ctx = context.WithValue(ctx, estClient.WithXForwardedClientCertHeader, SingeltonInstance.DMS.Certificate)
		crt, err := client.Enroll(ctx, SingeltonInstance.EnrollmentInProcess.IssuingCA, SingeltonInstance.EnrollmentInProcess.CertificateSigningRequest)
		if err != nil {
			SingeltonInstance.DMS.Status = DMSStatusIdle
			http.Error(w, "Error enrolling device",
				http.StatusInternalServerError)
			return
		}

		SingeltonInstance.EnrollmentInProcess.Status = EnrollingStatusStep2

		sendWebSocketMessage(
			WebSocketMessage{
				Type:      "ENROLLING_PROCESS_UPDATE",
				Message:   SingeltonInstance.EnrollmentInProcess.Serialize(),
				Timestamp: time.Now(),
			},
		)

		time.Sleep(time.Second * 2)
		SingeltonInstance.EnrollmentInProcess.Certificate = crt
		SingeltonInstance.EnrollmentInProcess.SerialNumber = utils.InsertNth(utils.ToHexInt(crt.SerialNumber), 2)
		SingeltonInstance.EnrollmentInProcess.ExpirationDate = crt.NotAfter
		SingeltonInstance.EnrollmentInProcess.Status = EnrollingStatusStep3
		sendWebSocketMessage(
			WebSocketMessage{
				Type:      "ENROLLING_PROCESS_UPDATE",
				Message:   SingeltonInstance.EnrollmentInProcess.Serialize(),
				Timestamp: time.Now(),
			},
		)

		for {
			if SingeltonInstance.EnrollmentInProcess.AuthorizedCertificateTransfer {
				break
			}
			time.Sleep(1 * time.Second)
		}

		SingeltonInstance.EnrollmentInProcess.Status = EnrollingStatusStep4
		sendWebSocketMessage(
			WebSocketMessage{
				Type:      "ENROLLING_PROCESS_UPDATE",
				Message:   SingeltonInstance.EnrollmentInProcess.Serialize(),
				Timestamp: time.Now(),
			},
		)

		SingeltonInstance.EnrolledIdentities = append(SingeltonInstance.EnrolledIdentities, EnrolledIdentity{
			EnrolledTimestamp: SingeltonInstance.EnrollmentInProcess.RequestingDate,
			SerialNumber:      SingeltonInstance.EnrollmentInProcess.SerialNumber,
			DeviceID:          SingeltonInstance.EnrollmentInProcess.DeviceID,
			DeviceSlot:        SingeltonInstance.EnrollmentInProcess.DeviceSlot,
			IssuingCA:         SingeltonInstance.EnrollmentInProcess.IssuingCA,
			IssuingDuration:   SingeltonInstance.EnrollmentInProcess.Certificate.NotAfter.Sub(SingeltonInstance.EnrollmentInProcess.Certificate.NotBefore),
		})

		serializedEnrolledIdentites := make([]EnrolledIdentitySerialized, 0)
		for _, v := range SingeltonInstance.EnrolledIdentities {
			serialized := v.Serialize()
			serializedEnrolledIdentites = append(serializedEnrolledIdentites, serialized)
		}

		sendWebSocketMessage(
			WebSocketMessage{
				Type:      "ENROLLED_IDENTITES_UPDATE",
				Message:   serializedEnrolledIdentites,
				Timestamp: time.Now(),
			},
		)
		type EnrollMessageOut struct {
			IssuingCA   string `json:"issuing_ca"`
			Certificate string `json:"certificate"`
		}

		pem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: crt.Raw})
		encodedCert := base64.StdEncoding.EncodeToString(pem)

		enrollMessageOutBytes, _ := json.Marshal(EnrollMessageOut{
			IssuingCA:   crt.Issuer.CommonName,
			Certificate: encodedCert,
		})

		w.Write(enrollMessageOutBytes)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

}

func mainRoute(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	// defer c.Close()
	SingeltonInstance.ActiveWebSocketConnection = c

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read err:", err)
			SingeltonInstance.ActiveWebSocketConnection = nil
			break
		}

		var inMessage WebSocketMessage
		color.HiGreen(">> Incoming message: " + string(message))
		err = json.Unmarshal(message, &inMessage)
		if err == nil {
			messageHandler(inMessage, c)
			continue
		}
		fmt.Println("err parsing message:", err)
	}
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

func main() {
	type Config struct {
		LamassuGateway string `required:"true" split_words:"true"`
	}
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		fmt.Println("error parsing environment variables:", err)
		os.Exit(1)
	}

	c := cron.New(cron.WithSeconds())
	c.Start()

	gatewayUrl, err := url.Parse(config.LamassuGateway)
	if err != nil {
		fmt.Println("error parsing gateway url:", err)
		os.Exit(1)
	}

	SingeltonInstance = &Singelton{
		ActiveWebSocketConnection: nil,
		DMS: DMSState{
			Status: DMSStatusEmpty,
		},
		CronInstance:       c,
		EnrolledIdentities: []EnrolledIdentity{},
		LamassuGatewayURL:  *gatewayUrl,
	}

	spa := spaHandler{staticPath: "build", indexPath: "index.html"}
	router := mux.NewRouter()
	router.PathPrefix("/ws").HandlerFunc(mainRoute)
	router.PathPrefix("/enroll").HandlerFunc(enrollRoute)
	router.PathPrefix("/").Handler(spa)

	srv := &http.Server{
		Handler: router,
		Addr:    ":7002",
	}

	log.Fatal(srv.ListenAndServe())
}
