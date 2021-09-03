package pipehealth

import (
	"context"

	"bitbucket.org/yellowmessenger/asterisk-ari/contracts"
	"bitbucket.org/yellowmessenger/asterisk-ari/spanhealth"
)

func Get(
	ctx context.Context,
) (
	*contracts.GetPipeHealthResponse,
	error,
) {
	messagesCount := 0
	response := new(contracts.GetPipeHealthResponse)
	responseData := new(contracts.SingleGetPipeHealthResponse)
	responseData.ResourceData = new(contracts.HealthResponse)
	responseData.ResourceData.QueueLength = messagesCount
	responseData.ResourceData.PipeHealth = spanhealth.Span
	responseData.Msg = "Successful Request"
	responseData.Status = "success"
	response.ResponseData = *responseData
	return response, nil
}
