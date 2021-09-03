package eventhandler

import (
	"context"
	"runtime"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/asterisk"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"

	"github.com/CyCoreSystems/ari"
)

func SnoopChannelHandler(
	ctx context.Context,
	handler *ari.ChannelHandle,
	parentChannelID string,
) {
	ymlogger.LogDebugf("GoRoutines", "GoRoutine Started. [%#v]", runtime.NumGoroutine())
	if strings.HasPrefix(strings.ToLower(handler.ID()), "listen") && strings.HasSuffix(strings.ToLower(handler.ID()), "snoop") {
		// Wait for 1 second till the time bridge is created
		time.Sleep(1 * time.Second)
		bridgeHandler := call.GetBridgeHandler(handler.ID())
		if bridgeHandler == nil {
			ymlogger.LogInfo(handler.ID(), "The snooping channel has already ended")
			return
		}
		if err := bridgeHandler.AddChannel(handler.ID()); err != nil {
			ymlogger.LogErrorf(handler.ID(), "Failed to add channel to bridge. Error: [%#v]", err)
		}
		return
	}
	if strings.HasPrefix(strings.ToLower(handler.ID()), "optone") {
		call.SetOpToneSnoopHandler(parentChannelID, handler)
		if call.GetBridgeHandler(parentChannelID) != nil {
			call.GetBridgeHandler(parentChannelID).AddChannel(handler.ID())
		}
		ymlogger.LogDebugf("GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())
		return
	}
	if strings.HasPrefix(strings.ToLower(handler.ID()), "inter") {
		recordHandle, err := asterisk.RecordCall(
			ctx,
			call.GetInterRecordingFilename(parentChannelID),
			handler,
		)
		if err != nil {
			ymlogger.LogErrorf(handler.ID(), "Failed to start recording the call. RecordingHandle: [%#v] Error: [%#v]", recordHandle, err)
		}
		call.SetInterRecordHandler(parentChannelID, recordHandle)
		call.SetInterSnoopHandler(parentChannelID, handler)
		return
	}
	recordHandle, err := asterisk.RecordCall(
		ctx,
		call.GetRecordingFilename(parentChannelID),
		handler,
	)
	if err != nil {
		ymlogger.LogErrorf(handler.ID(), "Failed to start recording the call. RecordingHandle: [%#v] Error: [%#v]", recordHandle, err)
	}
	ymlogger.LogDebugf("GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())
	return
}
