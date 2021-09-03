package eventhandler

import (
	"context"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

func (cH *CallHandlers) ChannelDTMFReceivedHandler(
	ctx context.Context,
	digit string,
	dtmfCompleted chan bool,
) {
	callSID := call.GetSID(cH.ChannelHandler.ID())
	// If playbackhandler is not nil, stop the playback
	if cH.PlaybackHandler != nil {
		cH.PlaybackHandler.Stop()
	}

	// If Recording is going on stop recording
	if cH.RecordHandler != nil {
		cH.RecordHandler.Stop()
	}

	call.SetDigit(cH.ChannelHandler.ID(), digit)
	ymlogger.LogDebugf(callSID, "ChannelID: [%s] Digits: [%#v]", cH.ChannelHandler.ID(), call.GetDigits(cH.ChannelHandler.ID()))
	digitLen := len(call.GetDigits(cH.ChannelHandler.ID()))
	time.Sleep(time.Duration(configmanager.ConfStore.ContinuousDTMFDelay) * time.Second)
	// If last character is not the same as the digit we got, means there are more digits
	if len(call.GetDigits(cH.ChannelHandler.ID())) > 0 && len(call.GetDigits(cH.ChannelHandler.ID())) > digitLen {
		ymlogger.LogDebugf(callSID, "There are more digits. Digits: [%#v]", call.GetDigits(cH.ChannelHandler.ID()))
		return
	}

	dtmfCompleted <- true
	ymlogger.LogDebugf(callSID, "ChannelID: [%s] Complete Digits: [%#v]", cH.ChannelHandler.ID(), strings.Join(call.GetDigits(cH.ChannelHandler.ID()), ""))
	return
}
