package eventhandler

import (
	"context"
	"runtime"
	"strconv"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/callback"
	"bitbucket.org/yellowmessenger/asterisk-ari/newrelic"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/analytics"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"

	"github.com/CyCoreSystems/ari"
)

func ChannelDestroyedHandler(
	ctx context.Context,
	channelID string,
	v *ari.ChannelDestroyed,
) {
	ymlogger.LogDebugf("GoRoutines", "GoRoutine Started. [%#v]", runtime.NumGoroutine())
	callSID := call.GetSID(channelID)
	// Set Dialing to Created Time if it is not set
	if call.GetDialingTime(channelID).IsZero() {
		call.SetDialingTime(channelID, call.GetCreatedTime(channelID))
	}
	// Set Ring Duration if call was not picked up or failed
	if call.GetPickupTime(channelID).IsZero() {
		call.SetRingDuration(channelID, GetTimeDifference(ctx, time.Time(v.Timestamp), call.GetRingingTime(channelID)))
	}
	// Set all the required fields in call data
	call.SetCallFinished(channelID, true)
	call.SetEndTime(channelID, time.Time(v.Timestamp))
	causeInfo := calculateTelcoMessage(ctx, channelID, v.Cause, v.CauseTxt)
	call.SetCause(channelID, causeInfo)
	call.SetDuration(channelID, GetTimeDifference(ctx, time.Time(v.Timestamp), call.GetDialingTime(channelID)))
	call.SetBillDuration(channelID, GetTimeDifference(ctx, time.Time(v.Timestamp), call.GetPickupTime(channelID)))
	call.SetStatus(channelID, calculateStatus(ctx, channelID).String())

	if call.GetIsChild(channelID) {
		// Remove Parent from the bridge
		if call.GetCompleteCall(call.GetParentUniqueID(channelID)).BridgeHandler != nil {
			call.GetCompleteCall(call.GetParentUniqueID(channelID)).BridgeHandler.RemoveChannel(call.GetParentUniqueID(channelID))
		}
		cH := NewCallHandlers(call.GetSID(call.GetParentUniqueID(channelID)))
		cH.ChannelHandler = call.GetChannelHandler(call.GetParentUniqueID(channelID))
		call.SetChildLegStatus(call.GetParentUniqueID(channelID), call.GetStatus(channelID))
		// cH.processUserText(ctx, call.GetParentUniqueID(channelID), call.GetStatus(channelID), false)
		ymlogger.LogDebugf(callSID, "Complete call details: [%#v]", call.GetCompleteCall(channelID))
		ymlogger.LogDebugf("GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())
		return
	}

	if call.GetBillDuration(channelID) > 0 {
		fileURL, err := uploadCallRecording(ctx, channelID, callSID, call.GetBotID(channelID))
		if err != nil {
			ymlogger.LogErrorf(callSID, "Failed uploading Recording. Error: [%#v]", err)
		}
		call.SetRecordingURL(channelID, fileURL)
	}

	if len(call.GetCallbackURL(channelID)) > 0 {
		callerID := call.GetCallerID(channelID)
		if call.GetDirection(channelID) == "inbound" && (callerID.E164Format == "+918068402303" || callerID.E164Format == "+918068402309" || callerID.E164Format == "+918068402395") {
			ymlogger.LogErrorf(callSID, "Not hitting the call back URL. ChannelID: [%#v]", channelID)
			newrelic.SendCustomEvent("callbacks_metrics", map[string]interface{}{
				"status": "success",
				"value":  1,
			})
		} else {
			err := callback.StoreCallbackRequest(ctx, channelID, callSID)
			if err != nil {
				ymlogger.LogErrorf(callSID, "Failed to hit the call back URL. Error: [%#v]", err)
				newrelic.SendCustomEvent("callbacks_metrics", map[string]interface{}{
					"status": "failure",
					"value":  1,
				})
			}
		}
	}

	// make update request with only callLaatencyINfoList

	event, err := analytics.PrepareAnalyticsEvent(
		analytics.CallEnd,
		call.GetBotID(channelID),
		call.GetCallerID(channelID).E164Format,
		call.GetDialedNumber(channelID).E164Format,
		call.GetDirection(channelID),
		analytics.AdditionalParams{
			Value:     "1",
			TelcoCode: strconv.Itoa(causeInfo.Code),
			TelcoText: causeInfo.Text,
		},
	)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while preparing the analytics event. Error: [%#v]", err)
	} else {
		go event.Push(ctx, callSID)
	}
	ymlogger.LogInfof(
		callSID,
		"Complete call details: [%#v]",
		call.GetCompleteCall(channelID),
	)

	// Delete Children UniqueId as well
	for _, uniqueID := range call.GetChildrenUniqueIDs(channelID) {
		ymlogger.LogDebugf(callSID, "Going to delete child channel data from the map. ChannelID: [%#v]", uniqueID)
		call.DeleteCall(uniqueID)
	}
	ymlogger.LogDebugf(callSID, "Going to delete the channel data from the map. ChannelID: [%#v]", channelID)

	call.DeleteCall(channelID)
	// call.DeleteCall(callSID)
	ymlogger.LogDebugf("GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())
	return
}
