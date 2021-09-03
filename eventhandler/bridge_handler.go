package eventhandler

import (
	"context"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
)

func (cH *CallHandlers) HandleBridge(
	ctx context.Context,
	bridgeHandler *ari.BridgeHandle,
	parentChannelID string,
	parentChannelHandler *ari.ChannelHandle,
) {
	// Delete the bridge when we exit
	defer bridgeHandler.Delete()

	callSID := call.GetSID(parentChannelID)
	bridgeDes := bridgeHandler.Subscribe(ari.Events.BridgeDestroyed)
	defer bridgeDes.Cancel()

	chanEnterBridge := bridgeHandler.Subscribe(ari.Events.ChannelEnteredBridge)
	defer chanEnterBridge.Cancel()

	chanLeftBridge := bridgeHandler.Subscribe(ari.Events.ChannelLeftBridge)
	defer chanLeftBridge.Cancel()

	for {
		select {
		case <-ctx.Done():
			return
		case e := <-bridgeDes.Events():
			v := e.(*ari.BridgeDestroyed)
			ymlogger.LogDebugf(callSID, "Got Bridge Destroyed Event: [%#v]", v)
			return
		case e := <-chanEnterBridge.Events():
			v := e.(*ari.ChannelEnteredBridge)
			ymlogger.LogDebugf(callSID, "Got Channel Entered Bridge Event: [%#v]", v)
		case e := <-chanLeftBridge.Events():
			v := e.(*ari.ChannelLeftBridge)
			ymlogger.LogDebugf(callSID, "Got Channel Left Bridge Event: [%#v]", v)
			go cH.ChannelLeftBridgeHandler(ctx, parentChannelID, v)
		}
	}
}
