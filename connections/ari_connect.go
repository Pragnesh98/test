package connections

import (
	"context"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
	"github.com/CyCoreSystems/ari/client/native"
)

var (
	ARIApplication  = "hello-world"
	ARIUsername     = "asterisk"
	ARIPassword     = "asterisk"
	ARIURL          = "http://localhost:8088/ari"
	ARIWebsocketURL = "ws://localhost:8088/ari/events"
)

// ConnectARI connects to Asterisk ARI
func ConnectARI(
	ctx context.Context,
) (ari.Client, error) {
	ymlogger.LogInfo("ARIConnect", "Connecting to ARI")
	ariClient, err := native.Connect(&native.Options{
		Application:  configmanager.ConfStore.ARIApplication,
		Username:     configmanager.ConfStore.ARIUsername,
		Password:     configmanager.ConfStore.ARIPassword,
		URL:          configmanager.ConfStore.ARIURL,
		WebsocketURL: configmanager.ConfStore.ARIWebsocketURL,
	})
	if err != nil {
		ymlogger.LogErrorf("ARIConnect", "Failed to build ARI Client. Error:[%#v]", err)
		return nil, err
	}
	ymlogger.LogInfo("ARIConnect", "Successfully connected to ARI")
	if err := ariClient.Application().Subscribe(&ari.Key{ID: ariClient.ApplicationName(), App: ariClient.ApplicationName()}, "channel:"); err != nil {
		ymlogger.LogErrorf("InitChanOpHandler", "Error while subscribing for all the events. Error: [%#v] [%s]", err, ariClient.ApplicationName())
	}
	if err := ariClient.Application().Subscribe(&ari.Key{ID: ariClient.ApplicationName(), App: ariClient.ApplicationName()}, "bridge:"); err != nil {
		ymlogger.LogErrorf("InitChanOpHandler", "Error while subscribing for all the events. Error: [%#v] [%s]", err, ariClient.ApplicationName())
	}
	// Uncomment if the following events are needed
	// if err := ariClient.Application().Subscribe(&ari.Key{ID: ariClient.ApplicationName(), App: ariClient.ApplicationName()}, "endpoint:"); err != nil {
	// 	ymlogger.LogErrorf("InitChanOpHandler", "Error while subscribing for all the events. Error: [%#v] [%s]", err, ariClient.ApplicationName())
	// }
	// if err := ariClient.Application().Subscribe(&ari.Key{ID: ariClient.ApplicationName(), App: ariClient.ApplicationName()}, "deviceState:"); err != nil {
	// 	ymlogger.LogErrorf("InitChanOpHandler", "Error while subscribing for all the events. Error: [%#v] [%s]", err, ariClient.ApplicationName())
	// }
	ymlogger.LogInfo("ARIConnect", "Subscribed to all the channel events")
	return ariClient, nil
}
