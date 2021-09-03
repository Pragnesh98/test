package bargeincall

import (
	"context"

	"bitbucket.org/yellowmessenger/asterisk-ari/contracts"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/asterisk"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

// Delete hangs up the listening call
func Delete(
	ctx context.Context,
	req contracts.DeleteBargeINCallRequest,
) (
	*contracts.DeleteBargeINCallResponse,
	error,
) {
	if err := asterisk.HangupChannel(ctx, *req.ChannelID, *req.ChannelID, "normal"); err != nil {
		ymlogger.LogErrorf(*req.ChannelID, "Failed to destroy barge in channel. Error: [%#v]", err.Error())
		return &contracts.DeleteBargeINCallResponse{}, err
	}

	response := new(contracts.DeleteBargeINCallResponse)
	responseData := new(contracts.SingleDeleteBargeINCallResponse)
	responseData.Msg = "Barge-IN Channel Hungup Successfully"
	responseData.Status = "success"
	response.ResponseData = *responseData
	return response, nil
}
