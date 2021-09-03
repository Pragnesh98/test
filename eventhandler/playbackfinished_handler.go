package eventhandler

import (
	"context"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/asterisk"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
)

func (cH *CallHandlers) PlaybackFinishedHandler(
	ctx context.Context,
	ariClient ari.Client,
) {
	callSID := call.GetSID(cH.ChannelHandler.ID())
	ymlogger.LogInfo(callSID, "Playback Finished")
	ymlogger.LogInfof(callSID, "Transcript from the STT: [%s]", call.GetTranscript(cH.ChannelHandler.ID()))
	end := cH.ChannelHandler.Subscribe(ari.Events.StasisEnd)
	defer end.Cancel()
	ymlogger.LogDebug(callSID, "PlaybackFinishedHandler")

	call.SetPlaybackFinished(cH.ChannelHandler.ID(), true)
	var err error
	hangupStr := call.GetHangupString(cH.ChannelHandler.ID())
	if len(hangupStr) <= 0 {
		hangupStr = "normal"
	}
	if call.GetShouldDisconnect(cH.ChannelHandler.ID()) {
		asterisk.HangupChannel(ctx, cH.ChannelHandler.ID(), callSID, hangupStr)
		call.SetDisconnectedBy(cH.ChannelHandler.ID(), "bot")
		return
	}

	if call.GetInterSnoopHandler(cH.ChannelHandler.ID()) != nil {
		ymlogger.LogInfo(callSID, "Hanging up snoop channel handler")
		call.GetInterSnoopHandler(cH.ChannelHandler.ID()).Hangup()
	}

	if strings.ToLower(configmanager.ConfStore.STTInterjectStreamingType) == "streaming" {
		sttHandler := call.GetSTTHandler(cH.ChannelHandler.ID())
		if sttHandler != nil {
			ymlogger.LogInfo(callSID, "Closing the Streaming STT Handler")
			sttHandler.Close()
		}
	}

	//handles googles streaming case
	cancel := call.GetStreamSTTCancel(cH.ChannelHandler.ID())
	if cancel != nil {
		ymlogger.LogDebug(callSID, "Cancelling the Streaming STT Context")
		cancel()
	}

	if call.GetInterRecordHandler(cH.ChannelHandler.ID()) != nil {
		call.GetInterRecordHandler(cH.ChannelHandler.ID()).Stop()
	}
	if len(call.GetTranscript(cH.ChannelHandler.ID())) > 0 {
		call.AddTranscript(cH.ChannelHandler.ID(), call.GetTranscript(cH.ChannelHandler.ID()))
		cH.processUserText(ctx, cH.ChannelHandler.ID(), call.GetTranscript(cH.ChannelHandler.ID()), true)
		return
	}
	// Uncomment this if you don't want to capture Voice and DTMF parallely
	// if botResp.Data.CaptureDTMF {
	// 	return recordHandle.Subscribe(ari.Events.RecordingStarted), recordHandle.Subscribe(ari.Events.RecordingFinished), recordHandle, err
	// }
	if call.GetDTMFCaptured(cH.ChannelHandler.ID()) {
		ymlogger.LogInfo(callSID, "Setting DTMF Captured to False")
		call.SetDTMFCaptured(cH.ChannelHandler.ID(), false)
		return
	}

	// Check if we should redirect to live agent
	// TODO: Comment this code
	// if call.GetShouldForward(cH.ChannelHandler.ID()) && len(call.GetForwardingNumber(cH.ChannelHandler.ID()).E164Format) > 0 {
	// 	err := asterisk.RedirectChannel(
	// 		ctx,
	// 		cH.ChannelHandler.ID(),
	//		callSID,
	// 		call.GetForwardingNumber(cH.ChannelHandler.ID()),
	// 		call.GetPipeType(cH.ChannelHandler.ID()),
	// 	)
	// 	if err != nil {
	// 		ymlogger.LogErrorf(callSID, "Failed to redirect the call. Error: [%#v]", err)
	// 	}
	// 	return
	// }

	// Check if we should dial another leg
	if call.GetShouldForward(cH.ChannelHandler.ID()) && len(call.GetForwardingNumber(cH.ChannelHandler.ID()).E164Format) > 0 {
		// Create the bridge
		cH.BridgeHandler, err = asterisk.CreateBridge(ctx, ariClient, cH.ChannelHandler.Key())
		if err != nil {
			ymlogger.LogErrorf(callSID, "Failed to create the bridge. Error: [%#v]", err)
			return
		}
		if cH.BridgeHandler != nil {
			call.SetBridgeHandler(cH.ChannelHandler.ID(), cH.BridgeHandler)
			cH.BridgeDes = cH.BridgeHandler.Subscribe(ari.Events.BridgeDestroyed)
			cH.ChanEnterBridge = cH.BridgeHandler.Subscribe(ari.Events.ChannelEnteredBridge)
			cH.ChanLeftBridge = cH.BridgeHandler.Subscribe(ari.Events.ChannelLeftBridge)
			cH.ChannelHold = cH.BridgeHandler.Subscribe(ari.Events.ChannelHold)
			cH.ChannelUnhold = cH.BridgeHandler.Subscribe(ari.Events.ChannelUnhold)
		}

		// bridgeHandler.MOH("default")
		// Manager the Bridge
		//go HandleBridge(ctx, bridgeHandler, channelHandler.ID(), channelHandler)
		// Add the current active channel to the bridge
		if !asterisk.ChannelExists(ctx, cH.ChannelHandler.ID()) {
			ymlogger.LogInfof(callSID, "Parent channel not added to the bridge. Channel [%s] does not exist", cH.ChannelHandler.ID())
			cH.BridgeHandler.Delete()
			return
		}
		ymlogger.LogInfo(callSID, "Adding parent channel to the bridge")
		cH.BridgeHandler.AddChannel(cH.ChannelHandler.ID())

		//Set calledID
		callerID := call.GetCallerID(cH.ChannelHandler.ID())
		// callerID = call.GetDialedNumber(cH.ChannelHandler.ID())

		if call.GetBotOptions(cH.ChannelHandler.ID()).ForwardingCallerID != "" {
			callerID, err = ParseNumber(ctx, call.GetBotOptions(cH.ChannelHandler.ID()).ForwardingCallerID)
			if err != nil {
				ymlogger.LogErrorf(callSID, "Failed to parse ForwardingCallerID number from botOptions. Error: [%#v]", err)
			}
		}

		// Create the channel to dial to another user
		ymlogger.LogInfo(callSID, "Going to create the new call")
		channelData, err := asterisk.CreateCall(
			ctx,
			call.GetForwardingNumber(cH.ChannelHandler.ID()),
			callerID,
			"incoming",
			callerID,
			call.GetPipeType(cH.ChannelHandler.ID()),
		)
		if err != nil || channelData.ID == "" {
			ymlogger.LogErrorf(callSID, "Failed to create call to forwarding number. Error: [%#v]", err)
			return
		}
		ymlogger.LogInfof(callSID, "Created the call with ChannelID: [%#v]", channelData.ID)
		call.SetSID(channelData.ID, callSID)
		call.SetCreatedTime(channelData.ID, time.Now())
		call.SetDirection(channelData.ID, call.DirectionOutbound.String())
		call.SetDialedNumber(channelData.ID, call.GetForwardingNumber(cH.ChannelHandler.ID()))
		call.SetCallerID(channelData.ID, call.GetCallerID(cH.ChannelHandler.ID()))
		call.SetPipeType(channelData.ID, call.GetPipeType(cH.ChannelHandler.ID()))
		call.SetIsChild(channelData.ID, true)
		call.SetParentUniqueID(channelData.ID, cH.ChannelHandler.ID())
		call.SetShouldForward(cH.ChannelHandler.ID(), false)

		// Set Children UniqueID on Parent Channel
		call.SetChildUniqueID(cH.ChannelHandler.ID(), channelData.ID)

		// Snoop the newly created channel
		snoopChannelData, err := asterisk.SnoopChannel(ctx, channelData.ID, callSID, "both", "optone_"+callSID)
		if err != nil {
			ymlogger.LogErrorf(callSID, "Failed to snoop the channel for operator tone. Error: [%#v]", err)
		}
		call.SetSID(snoopChannelData.ID, "optone_"+callSID)
		call.SetParentUniqueID(snoopChannelData.ID, cH.ChannelHandler.ID())
		// Add the new channel to the bridge
		// ymlogger.LogInfo(callSID, "Adding the child channel to the bridge")
		// bridgeHandler.AddChannel(channelData.ID)
		return
	}

	if call.GetCaptureDTMF(cH.ChannelHandler.ID()) && !call.GetCaptureVoice(cH.ChannelHandler.ID()) {
		ymlogger.LogDebugf(callSID, "Going to process the DTMF. ChannelID: [%s]", cH.ChannelHandler.ID())
		go cH.processDTMF(ctx, callSID)
		return
	}
	cH.RecordHandler, err = asterisk.Record(ctx, callSID, cH.ChannelHandler, call.GetAuthenticateUser(cH.ChannelHandler.ID()), call.GetRecordingBeep(cH.ChannelHandler.ID()))
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to start recording. Error: [%#v]", err)
	}
	return
}

func (cH *CallHandlers) processDTMF(
	ctx context.Context,
	callSID string,
) {
	sTime := time.Now()
	for {
		time.Sleep(500 * time.Millisecond)
		if call.GetDTMFCaptured(cH.ChannelHandler.ID()) {
			break
		}
		if time.Since(sTime) > time.Duration(5*time.Second) {
			ymlogger.LogDebugf(callSID, "Processed the DTMF. ChannelID: [%s]", cH.ChannelHandler.ID())
			cH.processUserText(ctx, cH.ChannelHandler.ID(), "", false)
			break
		}
	}
	return
	// Sleep for 5 second to see if DTMF has been captured
	// time.Sleep(5 * time.Second)
	// if len(call.GetDigits(cH.ChannelHandler.ID())) == 0 {
	// 	ymlogger.LogDebugf(callSID, "Processed the DTMF. ChannelID: [%s]", cH.ChannelHandler.ID())
	// 	cH.processUserText(ctx, cH.ChannelHandler.ID(), "")
	// 	return
	// }
	// return
}
