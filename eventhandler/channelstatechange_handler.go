package eventhandler

import (
	"context"
	"runtime"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/analytics"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"

	"github.com/CyCoreSystems/ari"
)

func ChannelStateChangeHandler(
	ctx context.Context,
	channelID string,
	v *ari.ChannelStateChange,
) {
	callSID := call.GetSID(channelID)
	ymlogger.LogDebugf("GoRoutines", "GoRoutine Started. [%#v]", runtime.NumGoroutine())
	if v.Channel.GetState() == call.StateDialing.String() {
		ymlogger.LogDebugf(callSID, "Call Dialing Time: [%#v]", time.Time(v.Timestamp).String())
		call.SetDialingTime(channelID, time.Time(v.Timestamp))
	} else if v.Channel.GetState() == call.StateRinging.String() {
		ymlogger.LogDebugf(callSID, "Call Ringing Time: [%#v]", time.Time(v.Timestamp).String())
		call.SetRingingTime(channelID, time.Time(v.Timestamp))
	} else if v.Channel.GetState() == call.StateUp.String() {
		ymlogger.LogDebugf(callSID, "Call Pickup Time: [%#v]. Ring Duration: [%#v]", time.Time(v.Timestamp).String(), int(time.Time(v.Timestamp).Sub(call.GetRingingTime(channelID)).Seconds()))
		call.SetRingDuration(channelID, GetTimeDifference(ctx, time.Time(v.Timestamp), call.GetRingingTime(channelID)))
		call.SetPickupTime(channelID, time.Time(v.Timestamp))
	}
	if len(callSID) > 0 {
		go sendEvent(ctx, callSID, call.GetBotID(channelID), v.Channel.GetState(), call.GetCallerID(channelID).E164Format, call.GetDialedNumber(channelID).E164Format, call.DirectionOutbound.String())
	}
	ymlogger.LogDebugf("GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())
	return
}

func sendEvent(
	ctx context.Context,
	callSID string,
	botID string,
	state string,
	callerID string,
	dialedNumber string,
	direction string,
) {
	var eventType analytics.EventType
	switch state {
	case call.StateDialing.String():
		eventType = analytics.CallDial
	case call.StateRinging.String():
		eventType = analytics.CallRing
	case call.StateUp.String():
		eventType = analytics.CallPick
	default:
		eventType = "Unknown"
	}
	event, err := analytics.PrepareAnalyticsEvent(eventType, botID, callerID, dialedNumber, direction, analytics.AdditionalParams{})
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while preparing the analytics event. Error: [%#v]", err)
	} else {
		event.Push(ctx, callSID)
	}
	return
}
