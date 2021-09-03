package eventhandler

import (
	"context"
	"runtime"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
)

func ChannelHoldHandler(
	ctx context.Context,
	channelID string,
	v *ari.ChannelHold,
) {
	ymlogger.LogDebugf("GoRoutines", "GoRoutine Started. [%#v]", runtime.NumGoroutine())
	callSID := call.GetSID(channelID)
	// Set Hold Time
	ymlogger.LogInfof(callSID, "Setting Hold Time: [%v]", v.Timestamp)
	call.SetHoldTime(channelID, time.Time(v.Timestamp))
	return
}
