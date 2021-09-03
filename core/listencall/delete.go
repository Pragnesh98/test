package listencall

import (
	"context"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/contracts"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/asterisk"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

// Delete hangs up the listening call
func Delete(
	ctx context.Context,
	req contracts.DeleteListenCallRequest,
) (
	*contracts.DeleteListenCallResponse,
	error,
) {
	if err := asterisk.HangupChannel(ctx, *req.ChannelID, *req.ChannelID, "normal"); err != nil {
		ymlogger.LogErrorf(*req.ChannelID, "Failed to destroy listen channel. Error: [%#v]", err.Error())
		return &contracts.DeleteListenCallResponse{}, err
	}

	if err := asterisk.HangupChannel(ctx, *req.ChannelID+"snoop", *req.ChannelID, "normal"); err != nil {
		ymlogger.LogErrorf(*req.ChannelID, "Failed to destroy listen snoop channel. Error: [%#v]", err.Error())
		return &contracts.DeleteListenCallResponse{}, err
	}
	bridgHandler := call.GetBridgeHandler(*req.ChannelID + "snoop")
	if bridgHandler == nil {
		ymlogger.LogInfo(*req.ChannelID, "Bridge handler not found for the listen channel")
	} else {
		bridgHandler.Delete()
	}
	response := new(contracts.DeleteListenCallResponse)
	responseData := new(contracts.SingleDeleteListenCallResponse)
	responseData.Msg = "Listen Channel Hungup Successfully"
	responseData.Status = "success"
	response.ResponseData = *responseData
	return response, nil
}
