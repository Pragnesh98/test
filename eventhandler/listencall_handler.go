package eventhandler

import (
	"context"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/asterisk"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"

	"github.com/CyCoreSystems/ari"
)

func ListenCallHandler(
	ctx context.Context,
	ariClient ari.Client,
	handler *ari.ChannelHandle,
) {
	callSID := call.GetSID(handler.ID())
	listenChannelID := call.GetListenChannelID(handler.ID())
	// var bridgeDes, chanEnterBridge, chanLeftBridge ari.Subscription
	bridgeHandler, err := asterisk.CreateBridge(ctx, ariClient, handler.Key())
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while creating the bridge. Error: [%#v]", err)
		handler.Hangup()
		return
	}
	if bridgeHandler != nil {
		// Delete the bridge when we exit
		// defer bridgeHandler.Delete()
		call.SetBridgeHandler(handler.ID(), bridgeHandler)
		// bridgeDes = bridgeHandler.Subscribe(ari.Events.BridgeDestroyed)
		// chanEnterBridge = bridgeHandler.Subscribe(ari.Events.ChannelEnteredBridge)
		// chanLeftBridge = bridgeHandler.Subscribe(ari.Events.ChannelLeftBridge)
	}
	err = bridgeHandler.AddChannel(handler.ID())
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed adding the channel [%s] to bridge. Error: [%#v]", call.GetSID(listenChannelID), err)
		handler.Hangup()
		return
	}
	var snoopChannelData ari.ChannelData
	listenChannelHandler := call.GetChannelHandler(listenChannelID)
	if listenChannelHandler == nil {
		ymlogger.LogInfo(callSID, "The channel to listen is already destroyed. Hanging up the call.")
		handler.Hangup()
		return
	}
	
	snoopChannelData, err = asterisk.SnoopChannel(ctx, listenChannelID, callSID, "both", callSID+"snoop")
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while snooping the call. Error: [%#v]", err)
		handler.Hangup()
		return
	}
	call.SetSID(snoopChannelData.ID, snoopChannelData.ID)
	call.SetBridgeHandler(snoopChannelData.ID, bridgeHandler)
	return
}