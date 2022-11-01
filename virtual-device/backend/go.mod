module github.com/lamassuiot/lamassu-vdevice

go 1.18

replace github.com/lamassuiot/lamassuiot => /home/ikerlan/lamassu/lamassuiot

require (
	github.com/eclipse/paho.mqtt.golang v1.4.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.0
	github.com/jakehl/goid v1.1.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lamassuiot/lamassuiot v0.0.5
	github.com/stianeikeland/go-rpio/v4 v4.6.0
	golang.org/x/exp v0.0.0-20220827204233-334a2380cb91
)

require (
	github.com/globalsign/est v1.0.6 // indirect
	github.com/go-chi/chi v4.1.2+incompatible // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/google/go-tpm v0.3.2 // indirect
	go.mozilla.org/pkcs7 v0.0.0-20210826202110-33d05740a352 // indirect
	golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f // indirect
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
)

require (
	github.com/robfig/cron/v3 v3.0.1
	golang.org/x/net v0.0.0-20220520000938-2e3eb7b945c2 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
)
