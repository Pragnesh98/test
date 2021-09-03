package eventhandler

import (
	"github.com/CyCoreSystems/ari"
)

// CallHandlers contains all the handlers and listeners needed for the call.
type CallHandlers struct {
	CallSID string
	// Handlers
	ChannelHandler     *ari.ChannelHandle
	PlaybackHandler    *ari.PlaybackHandle
	RecordHandler      *ari.LiveRecordingHandle
	BridgeHandler      *ari.BridgeHandle
	SnoopHandler       *ari.ChannelHandle
	InterSnoopHandler  *ari.ChannelHandle
	OpToneSnoopHandler *ari.ChannelHandle
	// Listeners
	ChanDest          ari.Subscription
	ChanHangup        ari.Subscription
	PlaybackStarted   ari.Subscription
	PlaybackFinished  ari.Subscription
	RecordingStarted  ari.Subscription
	RecordingFinished ari.Subscription
	DtmfReceived      ari.Subscription
	BridgeDes         ari.Subscription
	ChanEnterBridge   ari.Subscription
	ChanLeftBridge    ari.Subscription
	ChannelHold       ari.Subscription
	ChannelUnhold     ari.Subscription
	End               ari.Subscription
}

// NewCallHandlers returns the CallHandlers struct initialized with CallSID
func NewCallHandlers(callSID string) *CallHandlers {
	return &CallHandlers{
		CallSID: callSID,
	}
}
