package eventhandler

import (
	"context"
	"os"
	"runtime"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/callstore"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/globals"
	"bitbucket.org/yellowmessenger/asterisk-ari/newrelic"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/analytics"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"

	"github.com/CyCoreSystems/ari"
)

// Initiliaze Handlers
// var (
// 	channelHandler     *ari.ChannelHandle
// 	playbackHandler    *ari.PlaybackHandle
// 	recordHandler      *ari.LiveRecordingHandle
// 	bridgeHandler      *ari.BridgeHandle
// 	snoopHandler       *ari.ChannelHandle
// 	opToneSnoopHandler *ari.ChannelHandle
// )

// // Initialize Event Listeners
// var (
// 	chanDest          ari.Subscription
// 	chanHangup        ari.Subscription
// 	playbackStarted   ari.Subscription
// 	playbackFinished  ari.Subscription
// 	recordingStarted  ari.Subscription
// 	recordingFinished ari.Subscription
// 	dtmfReceived      ari.Subscription
// 	bridgeDes         ari.Subscription
// 	chanEnterBridge   ari.Subscription
// 	chanLeftBridge    ari.Subscription
// 	end               ari.Subscription
// )

func ChannelHandler(
	ctx context.Context,
	ariClient ari.Client,
	h *ari.ChannelHandle,
	channelID string,
) {
	h.Answer()
	defer h.Hangup()
	ymlogger.LogInfo("ChannelHandler", "Running channel handler")
	ymlogger.LogDebugf("GoRoutines", "GoRoutine Started. [%#v]", runtime.NumGoroutine())
	// Fill the call data
	var callSID, direction string
	var err error
	direction = call.GetDirection(channelID)
	if err != nil {
		ymlogger.LogErrorf("ChannelHandler", "Error while getting getting the direction channel variable. Error: [%#v]", err)
	}
	if direction != call.DirectionOutbound.String() {
		// If Direction is not outbound, generate the CallSID and set the call data
		callSID = call.GenerateCallSID()
		// Set the call data
		call.SetSID(channelID, callSID)
		call.SetCreatedTime(channelID, time.Now())
		call.SetDirection(channelID, call.DirectionInbound.String())
		call.SetCallbackURL(channelID, configmanager.ConfStore.InboundCallbackURL)
		call.SetMaxBotFailureCount(channelID, 7)
		call.SetRecordingFilename(channelID, callSID)
		direction = call.DirectionInbound.String()
	} else { // Otherwise get the CallSID from the call data
		callSID = call.GetSID(channelID)
	}

	call.SetChannelHandler(channelID, h)
	//Increment call count.
	globals.IncrementNoOfCalls()
	ymlogger.LogInfof(callSID, "Number of calls [%d]. Number of call objects [%d]", globals.GetNoOfCalls(), globals.GetNoOfCallObject())

	channelData, err := h.Data()
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while getting the data from channel. Error: [%#v]", err)
	}

	// Start the Magic
	cH := NewCallHandlers(callSID)
	cH.ChannelHandler = h
	defer cH.ChannelHandler.Hangup()

	// Set Parent Unique ID for Snoop Channel
	call.SetParentUniqueID(callSID, channelID)
	call.SetCallLatencyStore(callSID, callstore.LatencyStore{})
	ymlogger.LogInfof(channelID, "SetCallLatencyStore:v[%#v]", call.GetCallLatencyStore(channelID))
	call.SetCallMessageStore(callSID, callstore.MessageStore{})
	
	// Start snooping the channel for recording.
	if h != nil {
		for i := 0; i < configmanager.ConfStore.ARIAPIRetry; i++ {
			cH.SnoopHandler, err = h.Snoop(callSID, &ari.SnoopOptions{
				Spy: "both",
				App: "hello-world",
			})
			if err != nil {
				ymlogger.LogErrorf(callSID, "Error while snooping the call. Error: [%#v]. Retrying.....", err)
				continue
			}
			break
		}
		if err != nil {
			ymlogger.LogErrorf(callSID, "Error while snooping the call. Error: [%#v]", err)
			cH.ChannelHandler.Hangup()
			return
		}
	}
	ParseChannel(ctx, channelData, channelID, callSID, direction)
	if direction == call.DirectionInbound.String() {
		hostName, err := os.Hostname()
		if err != nil {
			ymlogger.LogErrorf(callSID, "Error while getting host name of the server. Error: [%#v]", err)
		}
		callStore := callstore.GetCallStore(
			callSID,
			call.GetDialedNumber(channelID).E164Format,
			call.GetCallerID(channelID).E164Format,
			call.DirectionInbound.String(),
			call.GetCallbackURL(channelID),
			hostName,
		)
		go callStore.Create()
		newrelic.SendCustomEvent("callbacks_metrics", map[string]interface{}{
			"status": "scheduled",
			"value":  1,
		})
		event, err := analytics.PrepareAnalyticsEvent(analytics.CallStart, call.GetBotID(channelID), call.GetCallerID(channelID).E164Format, call.GetDialedNumber(channelID).E164Format, direction, analytics.AdditionalParams{})
		if err != nil {
			ymlogger.LogErrorf(callSID, "Error while preparing the analytics event. Error: [%#v]", err)
		} else {
			go event.Push(ctx, callSID)
		}
	}

	ymlogger.LogInfof(callSID, "From number is [%s] and To number is [%s] ChannelData: [%#v]", call.GetDialedNumber(channelID).E164Format, call.GetCallerID(channelID).E164Format, channelData)
	defer cH.ChannelHandler.Hangup()
	cH.End = cH.ChannelHandler.Subscribe(ari.Events.StasisEnd)
	defer cH.End.Cancel()
	// Initialize all the listeners on channel handler
	cH.ChanDest = cH.ChannelHandler.Subscribe(ari.Events.ChannelDestroyed)
	defer cH.ChanDest.Cancel()
	cH.ChanHangup = cH.ChannelHandler.Subscribe(ari.Events.ChannelHangupRequest)
	defer cH.ChanHangup.Cancel()
	cH.PlaybackStarted = cH.ChannelHandler.Subscribe(ari.Events.PlaybackStarted)
	defer cH.PlaybackStarted.Cancel()
	cH.PlaybackFinished = cH.ChannelHandler.Subscribe(ari.Events.PlaybackFinished)
	defer cH.PlaybackFinished.Cancel()
	cH.RecordingStarted = cH.ChannelHandler.Subscribe(ari.Events.RecordingStarted)
	defer cH.RecordingStarted.Cancel()
	cH.RecordingFinished = cH.ChannelHandler.Subscribe(ari.Events.RecordingFinished)
	defer cH.RecordingFinished.Cancel()
	cH.DtmfReceived = cH.ChannelHandler.Subscribe(ari.Events.ChannelDtmfReceived)
	defer cH.DtmfReceived.Cancel()
	cH.BridgeDes = cH.ChannelHandler.Subscribe(ari.Events.BridgeDestroyed)
	defer cH.BridgeDes.Cancel()
	cH.ChanEnterBridge = cH.ChannelHandler.Subscribe(ari.Events.ChannelEnteredBridge)
	defer cH.ChanEnterBridge.Cancel()
	cH.ChanLeftBridge = cH.ChannelHandler.Subscribe(ari.Events.ChannelLeftBridge)
	defer cH.ChanLeftBridge.Cancel()
	cH.ChannelHold = cH.ChannelHandler.Subscribe(ari.Events.ChannelHold)
	defer cH.ChannelHold.Cancel()
	cH.ChannelUnhold = cH.ChannelHandler.Subscribe(ari.Events.ChannelUnhold)
	defer cH.ChannelUnhold.Cancel()

	ymlogger.LogInfo(callSID, "Going to initialize the call")
	cH.initiliazeCall(ctx)

	if cH.PlaybackHandler != nil {
		cH.PlaybackStarted = cH.PlaybackHandler.Subscribe(ari.Events.PlaybackStarted)
		defer cH.PlaybackStarted.Cancel()
		cH.PlaybackFinished = cH.PlaybackHandler.Subscribe(ari.Events.PlaybackFinished)
		defer cH.PlaybackFinished.Cancel()
	}

	// Channel for DTMF completed events
	dtmfCompleted := make(chan bool)

	for {
		select {
		case e := <-cH.PlaybackStarted.Events():
			ymlogger.LogInfo(callSID, "Got Playback Started Event", channelID, e)
			cH.PlaybackStartedHandler(ctx)
		case e := <-cH.PlaybackFinished.Events():
			ymlogger.LogInfo(callSID, "Got Playback Finished Event", channelID, e)
			cH.PlaybackFinishedHandler(ctx, ariClient)
			if cH.RecordHandler != nil {
				cH.RecordingStarted = cH.RecordHandler.Subscribe(ari.Events.RecordingStarted)
				defer cH.RecordingStarted.Cancel()
				cH.RecordingFinished = cH.RecordHandler.Subscribe(ari.Events.RecordingFinished)
				defer cH.RecordingFinished.Cancel()
			}
			if cH.PlaybackHandler != nil {
				cH.PlaybackStarted = cH.PlaybackHandler.Subscribe(ari.Events.PlaybackStarted)
				defer cH.PlaybackStarted.Cancel()
				cH.PlaybackFinished = cH.PlaybackHandler.Subscribe(ari.Events.PlaybackFinished)
				defer cH.PlaybackFinished.Cancel()
			}
		case e := <-cH.RecordingStarted.Events():
			v := e.(*ari.RecordingStarted)
			ymlogger.LogInfo(callSID, "Got Recording Started Event", channelID, e)
			cH.RecordingStartedHandler(ctx, callSID, v.Recording.Name)
		case e := <-cH.RecordingFinished.Events():
			v := e.(*ari.RecordingFinished)
			ymlogger.LogInfof(callSID, "Got Recording Finished Event now. ChannelID: [%s] RecordingName: [%s]", channelID, v.Recording.Name)
			cH.RecordingFinishedHandler(ctx, v.Recording.Name)
			if cH.PlaybackHandler != nil {
				cH.PlaybackStarted = cH.PlaybackHandler.Subscribe(ari.Events.PlaybackStarted)
				defer cH.PlaybackStarted.Cancel()
				cH.PlaybackFinished = cH.PlaybackHandler.Subscribe(ari.Events.PlaybackFinished)
				defer cH.PlaybackFinished.Cancel()
			}
		case e := <-cH.DtmfReceived.Events():
			v := e.(*ari.ChannelDtmfReceived)
			ymlogger.LogInfof(callSID, "Got DTMF Received Events. Digits: [%#v]", v.Digit)
			if call.GetCaptureDTMF(channelID) {
				ymlogger.LogInfo(callSID, "Setting DTMF Captured to True")
				call.SetDTMFCaptured(channelID, true)
				go cH.ChannelDTMFReceivedHandler(ctx, v.Digit, dtmfCompleted)
			}
		case <-dtmfCompleted:
			digits := call.GetDigits(cH.ChannelHandler.ID())
			if len(digits) > 0 {
				ymlogger.LogDebugf(callSID, "Processing the DTMF")
				cH.processUserText(ctx, cH.ChannelHandler.ID(), strings.Join(digits, ""), false)
				if cH.PlaybackHandler != nil {
					cH.PlaybackStarted = cH.PlaybackHandler.Subscribe(ari.Events.PlaybackStarted)
					defer cH.PlaybackStarted.Cancel()
					cH.PlaybackFinished = cH.PlaybackHandler.Subscribe(ari.Events.PlaybackFinished)
					defer cH.PlaybackFinished.Cancel()
				}
				call.ResetDigits(cH.ChannelHandler.ID())
			}
		case e := <-cH.ChannelHold.Events():
			v := e.(*ari.ChannelHold)
			ymlogger.LogDebugf(callSID, "Got Channel Hold Event: [%#v]", v)
		case e := <-cH.ChannelUnhold.Events():
			v := e.(*ari.ChannelUnhold)
			ymlogger.LogDebugf(callSID, "Got Channel Unhold Event: [%#v]", v)
		case e := <-cH.BridgeDes.Events():
			v := e.(*ari.BridgeDestroyed)
			ymlogger.LogDebugf(callSID, "Got Bridge Destroyed Event: [%#v]", v)
		case e := <-cH.ChanEnterBridge.Events():
			v := e.(*ari.ChannelEnteredBridge)
			ymlogger.LogDebugf(callSID, "Got Channel Entered Bridge Event: [%#v]", v)
		case e := <-cH.ChanLeftBridge.Events():
			v := e.(*ari.ChannelLeftBridge)
			ymlogger.LogDebugf(callSID, "Got Channel Left Bridge Event: [%#v]", v)
			cH.ChannelLeftBridgeHandler(ctx, channelID, v)
			if v.Channel.ID == channelID {
				childLegStatus := call.GetChildLegStatus(channelID)
				if len(childLegStatus) <= 0 {
					time.Sleep(2 * time.Second)
					childLegStatus = call.GetChildLegStatus(channelID)
				}
				cH.processUserText(ctx, channelID, "", false)
				if cH.PlaybackHandler != nil {
					cH.PlaybackStarted = cH.PlaybackHandler.Subscribe(ari.Events.PlaybackStarted)
					defer cH.PlaybackStarted.Cancel()
					cH.PlaybackFinished = cH.PlaybackHandler.Subscribe(ari.Events.PlaybackFinished)
					defer cH.PlaybackFinished.Cancel()
				}
			}
		case e := <-cH.ChanHangup.Events():
			v := e.(*ari.ChannelHangupRequest)
			ymlogger.LogInfo(callSID, "Got Channel hangup request Event", channelID, e, v)
			cH.SnoopHandler.Hangup()
		case e := <-cH.ChanDest.Events():
			v := e.(*ari.ChannelDestroyed)
			ymlogger.LogInfo(callSID, "Got Channel Destroyed Event", channelID, e, v, v.Channel.GetChannelVars())
			// go ChannelDestroyedHandler(ctx, v.Channel.ID, v)
		case e := <-cH.End.Events():
			v := e.(*ari.StasisEnd)
			ymlogger.LogInfof(callSID, "Got Statis End Event: [%s] [%#v] [%#v]", callSID, v, runtime.NumGoroutine())
			cH.StatisEndHandler(ctx, ariClient, h)
			ymlogger.LogDebugf("GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())
			return
		case <-ctx.Done():
			return
		}
	}

}

func (cH *CallHandlers) initiliazeCall(
	ctx context.Context,
) {
	// For the first time, we set the user response as welcome for the bot to start
	cH.processUserText(ctx, cH.ChannelHandler.ID(), "welcome", false)
	return
}
