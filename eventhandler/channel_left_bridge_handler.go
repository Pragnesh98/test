package eventhandler

import (
	"context"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/asterisk"
	"github.com/CyCoreSystems/ari"
)

func (cH *CallHandlers) ChannelLeftBridgeHandler(
	ctx context.Context,
	parentChannelID string,
	event *ari.ChannelLeftBridge,
) {
	callSID := call.GetSID(parentChannelID)
	// Check if parent channel has left the bridge
	if event.Channel.ID == parentChannelID {
		for _, childUniqueID := range call.GetChildrenUniqueIDs(parentChannelID) {
			cH.BridgeHandler.RemoveChannel(childUniqueID)
			if call.GetPickupTime(childUniqueID).IsZero() {
				asterisk.HangupChannel(ctx, childUniqueID, callSID, "rejected")
			} else {
				asterisk.HangupChannel(ctx, childUniqueID, callSID, "normal")
			}
		}
		return
	}

	// Check if child channel has hungup and there is only one child channel
	if call.GetIsChild(event.Channel.ID) && len(call.GetChildrenUniqueIDs(parentChannelID)) == 1 {
		cH.BridgeHandler.RemoveChannel(parentChannelID)
		return
	}
	return
}
