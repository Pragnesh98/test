package eventhandler

import (
	"context"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/asterisk"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
)

func BargeINCallHandler(
	ctx context.Context,
	ariClient ari.Client,
	handler *ari.ChannelHandle,
) {
	callSID := call.GetSID(handler.ID())
	bargeINChannelID := call.GetBargeINChannelID(handler.ID())
	var bridgeHandler *ari.BridgeHandle
	var err error
	bridgeHandler = call.GetBridgeHandler(bargeINChannelID)
	if bridgeHandler == nil {
		bridgeHandler, err = asterisk.CreateBridge(ctx, ariClient, handler.Key())
		if err != nil {
			ymlogger.LogErrorf(callSID, "Error while creating the bridge. Error: [%#v]", err)
			handler.Hangup()
			return
		}
	}
	if bridgeHandler != nil {
		call.SetBridgeHandler(bargeINChannelID, bridgeHandler)
	}
	if err := bridgeHandler.AddChannel(handler.ID()); err != nil {
		ymlogger.LogErrorf(callSID, "Failed adding the channel [%s] to bridge. Error: [%#v]", handler.ID(), err)
		handler.Hangup()
		return
	}
	if err := bridgeHandler.AddChannel(bargeINChannelID); err != nil {
		ymlogger.LogErrorf(callSID, "Failed adding the channel [%s] to bridge. Error: [%#v]", bargeINChannelID, err)
		handler.Hangup()
		return
	}
	return
}
