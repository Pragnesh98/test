package eventhandler

import (
	"context"
	"runtime"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
)

func ChannelUnHoldHandler(
	ctx context.Context,
	channelID string,
	v *ari.ChannelUnhold,
) {
	ymlogger.LogDebugf("GoRoutines", "GoRoutine Started. [%#v]", runtime.NumGoroutine())
	callSID := call.GetSID(channelID)
	// Set Hold Time
	ymlogger.LogInfof(callSID, "Adding Hold Duration: [%v]", v.Timestamp)
	call.SetHoldDuration(channelID, GetTimeDifference(ctx, time.Time(v.Timestamp), call.GetHoldTime(channelID)))
	ymlogger.LogDebugf("GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())
	return
}
