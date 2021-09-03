package eventhandler

import (
	"context"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/globals"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
)

func (cH *CallHandlers) StatisEndHandler(
	ctx context.Context,
	ariClient ari.Client,
	handler *ari.ChannelHandle,
) {
	ymlogger.LogInfo(call.GetSID(handler.ID()), "Statis End")
	globals.DecrementNoOfCalls()

	ymlogger.LogInfof(call.GetSID(handler.ID()), "Number of calls [%d]. Number of call objects [%d]", globals.GetNoOfCalls(), globals.GetNoOfCallObject())
	cH.ChanDest.Cancel()
	cH.ChanDest = nil
	cH.ChanHangup.Cancel()
	cH.ChanHangup = nil
	cH.PlaybackStarted.Cancel()
	cH.PlaybackStarted = nil
	cH.PlaybackFinished.Cancel()
	cH.PlaybackFinished = nil
	cH.RecordingStarted.Cancel()
	cH.RecordingStarted = nil
	cH.RecordingFinished.Cancel()
	cH.RecordingFinished = nil
	cH.DtmfReceived.Cancel()
	cH.DtmfReceived = nil
	cH.BridgeDes.Cancel()
	cH.BridgeDes = nil
	cH.ChanLeftBridge.Cancel()
	cH.ChanEnterBridge = nil
	cH.ChanLeftBridge.Cancel()
	cH.ChanLeftBridge = nil
	cH.End.Cancel()
	cH.End = nil
	return
}
