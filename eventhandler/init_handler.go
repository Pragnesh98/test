package eventhandler

import (
	"context"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/asterisk"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
)

func InitHandler(
	ctx context.Context,
	ariClient ari.Client,
	handler func(
		ctx context.Context,
		ariClient ari.Client,
		h *ari.ChannelHandle,
		channelID string,
	),
) {
	// var parentChannelID string
	sub := ariClient.Bus().Subscribe(nil, "StasisStart")
	// chanStateChange := ariClient.Bus().Subscribe(nil, ari.Events.ChannelStateChange)
	// defer chanStateChange.Cancel()
	// chanDest := ariClient.Bus().Subscribe(nil, ari.Events.ChannelDestroyed)
	// defer chanDest.Cancel()
	for {
		select {
		case e := <-sub.Events():
			v := e.(*ari.StasisStart)
			ymlogger.LogInfo("InitHandler", "Got Statis Start Event ", v.Channel.ID, v.Channel.Caller.Number, v.Channel.ChannelVars, v.Channel.Dialplan, v.Channel.Name)

			if isSnoopChannel(v.Channel.Name) {
				go SnoopChannelHandler(ctx, ariClient.Channel().Get(v.Key(ari.ChannelKey, v.Channel.ID)), call.GetParentUniqueID(v.Channel.ID))
				continue
			}
			if isListenChannel(v.Channel.ID) {
				go ListenCallHandler(ctx, ariClient, ariClient.Channel().Get(v.Key(ari.ChannelKey, v.Channel.ID)))
				continue
			}
			if isBargeINChannel(v.Channel.ID) {
				go BargeINCallHandler(ctx, ariClient, ariClient.Channel().Get(v.Key(ari.ChannelKey, v.Channel.ID)))
				continue
			}
			// If channel is child channel, don't process it.
			if call.GetIsChild(v.Channel.ID) {
				if !asterisk.ChannelExists(ctx, call.GetParentUniqueID(v.Channel.ID)) {
					ymlogger.LogInfof(call.GetSID(v.Channel.ID), "Parent channel [%s] does not exist", call.GetParentUniqueID(v.Channel.ID))
					if call.GetBridgeHandler(call.GetParentUniqueID(v.Channel.ID)) != nil {
						call.GetBridgeHandler(call.GetParentUniqueID(v.Channel.ID)).Delete()
					}
					ymlogger.LogInfof(call.GetSID(v.Channel.ID), "Hanging up child channel [%s]", v.Channel.ID)
					asterisk.HangupChannel(ctx, v.Channel.ID, call.GetSID(v.Channel.ID), "normal")
					continue
				}
				ymlogger.LogInfo(call.GetSID(v.Channel.ID), "Adding the child channel to the bridge")
				call.GetBridgeHandler(call.GetParentUniqueID(v.Channel.ID)).AddChannel(v.Channel.ID)
				opToneSnoopHandler := call.GetOpToneSnoopHandler(call.GetParentUniqueID(v.Channel.ID))
				if opToneSnoopHandler != nil {
					ymlogger.LogInfo(call.GetSID(v.Channel.ID), "Hanging up optone snoop channel")
					opToneSnoopHandler.Hangup()
				}
				// Forward user's CLI to Exotel as DTMF
				if isDTMFForwardingNumber(call.GetForwardingNumber(call.GetParentUniqueID(v.Channel.ID)).E164Format, configmanager.ConfStore.DTMFForwardingNumbers) {
					number := call.GetDialedNumber(call.GetParentUniqueID(v.Channel.ID)).WithZeroNationalFormat
					if err := ariClient.Channel().Get(v.Key(ari.ChannelKey, v.Channel.ID)).SendDTMF(number, &ari.DTMFOptions{Before: 1 * time.Second, Between: 300 * time.Millisecond}); err != nil {
						ymlogger.LogError(call.GetSID(v.Channel.ID), "Error while sending the DTMF")
					}
				}
				continue
			}
			// parentChannelID = v.Channel.ID
			go handler(ctx, ariClient, ariClient.Channel().Get(v.Key(ari.ChannelKey, v.Channel.ID)), v.Channel.ID)
		case <-ctx.Done():
			return
		}
	}
}
