package asterisk

import (
	"context"

	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"

	"github.com/CyCoreSystems/ari"
)

func Answer(
	ctx context.Context,
	callSID string,
	h *ari.ChannelHandle,
) error {
	ymlogger.LogInfof(callSID, "Going to answer the call. Channel ID: [%#v]", h.ID())
	if err := h.Answer(); err != nil {
		return err
	}
	return nil
}
