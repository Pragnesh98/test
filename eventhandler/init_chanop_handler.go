package eventhandler

import (
	"context"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
)

func InitChannelOpHandler(
	ctx context.Context,
	ariClient ari.Client,
) {
	chanStateChange := ariClient.Bus().Subscribe(nil, ari.Events.ChannelStateChange)
	defer chanStateChange.Cancel()
	chanDest := ariClient.Bus().Subscribe(nil, ari.Events.ChannelDestroyed)
	defer chanDest.Cancel()
	chanUnhold := ariClient.Bus().Subscribe(nil, ari.Events.ChannelUnhold)
	defer chanUnhold.Cancel()
	chanHold := ariClient.Bus().Subscribe(nil, ari.Events.ChannelHold)
	defer chanHold.Cancel()
	for {
		select {
		case e := <-chanHold.Events():
			v := e.(*ari.ChannelHold)
			ymlogger.LogDebugf("InitChanOpHandler", "Got Channel Hold Event: [%#v]", v)
			go ChannelHoldHandler(ctx, v.Channel.ID, v)
		case e := <-chanUnhold.Events():
			v := e.(*ari.ChannelUnhold)
			ymlogger.LogDebugf("InitChanOpHandler", "Got Channel Unhold Event: [%#v]", v)
			go ChannelUnHoldHandler(ctx, v.Channel.ID, v)
		case e := <-chanStateChange.Events():
			v := e.(*ari.ChannelStateChange)
			ymlogger.LogInfof("InitChanOpHandler", "Got Channel State Change Event. Event: [%#v]", v)
			go ChannelStateChangeHandler(ctx, v.Channel.ID, v)
		case e := <-chanDest.Events():
			v := e.(*ari.ChannelDestroyed)
			ymlogger.LogInfof("InitChanOpHandler", "Got Channel Destroyed Event. [%#v] TimeStamp: [%v]", v, time.Time(v.Timestamp).String())
			if isSnoopChannel(v.Channel.Name) {
				continue
			}
			go ChannelDestroyedHandler(ctx, v.Channel.ID, v)
		case <-ctx.Done():
			return
		}
	}
}
