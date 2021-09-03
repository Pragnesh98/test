package enqueuecall

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/contracts"
	"bitbucket.org/yellowmessenger/asterisk-ari/enqueuecallworker"
	"bitbucket.org/yellowmessenger/asterisk-ari/globals"
	"bitbucket.org/yellowmessenger/asterisk-ari/queuemanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/analytics"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

var timeLayout = "2006-01-02T15:04:05"

func Create(
	req contracts.EnqueueCallRequest,
) (
	*contracts.EnqueueCallResponse,
	error,
) {
	callSID := call.GenerateCallSID()
	ymlogger.LogInfof(callSID, "Enqueue Call Request: [%#v]", req)

	var hostName string
	hostName, err := os.Hostname()
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while getting host name of the server. Error: [%#v]", err)
	}

	// if req.From != nil && req.To != nil && *req.From == "+918068402307" {
	// 	go callbackhelper.MockMakeCallbackRequest(ctx, *req.To, callSID, req.ExtraParams)
	// 	response := new(contracts.EnqueueCallResponse)
	// 	responseData := new(contracts.SingleEnqueueCallResponse)
	// 	resourceData := contracts.EnqueueCall{
	// 		SID:         callSID,
	// 		CreatedTime: time.Now().Format(timeLayout),
	// 		From:        *req.From,
	// 		To:          *req.To,
	// 		Status:      call.StatusInitiated.String(),
	// 	}
	// 	if req.CallbackURL != nil && len(*req.CallbackURL) > 0 {
	// 		resourceData.CallbackURL = *req.CallbackURL
	// 	}
	// 	responseData.ResourceData = &resourceData
	// 	responseData.Msg = "Call Enqueued Successfully"
	// 	responseData.Status = "success"
	// 	response.ResponseData = *responseData
	// 	return response, nil
	// }
	queueMsgParams, err := formQueueMsgParams(callSID, req)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to form Queue Message params. Error: [%#v]", err)
		return &contracts.EnqueueCallResponse{}, errors.New("Failed to form queue params")
	}

	// Enqueuing the msg to Queue
	if err := queueMsgParams.Enqueue(); err != nil {
		ymlogger.LogErrorf(callSID, "Failed to enqueue the job to queue. Error: [%#v]", err)
		return &contracts.EnqueueCallResponse{}, errors.New("Failed to enqueue the call")
	}

	var eP = new(contracts.ExtraParams)
	if req.ExtraParams != nil {
		eP.ExtractExtraParams(req.ExtraParams)
	}
	enqueueCallRes := contracts.EnqueueCall{
		SID:         callSID,
		CreatedTime: time.Now().Format(timeLayout),
		From:        *req.From,
		To:          *req.To,
		Status:      call.StatusInitiated.String(),
		BotID:       eP.BotID,
		CampaignID:  eP.CampaignID,
		Host:        hostName,
	}
	if req.CallbackURL != nil && len(*req.CallbackURL) > 0 {
		enqueueCallRes.CallbackURL = *req.CallbackURL
	}
	globals.IncrementNoOfCalls()
	ymlogger.LogInfof(callSID, "Number of calls [%d]. Number of call objects [%d]", globals.GetNoOfCalls(), globals.GetNoOfCallObject())

	response := new(contracts.EnqueueCallResponse)
	responseData := new(contracts.SingleEnqueueCallResponse)
	resourceData := contracts.EnqueueCall{
		SID:         callSID,
		CreatedTime: time.Now().Format(timeLayout),
		From:        *req.From,
		To:          *req.To,
		Status:      call.StatusInitiated.String(),
		BotID:       eP.BotID,
		CampaignID:  eP.CampaignID,
		Host:        hostName,
	}
	if req.CallbackURL != nil && len(*req.CallbackURL) > 0 {
		resourceData.CallbackURL = *req.CallbackURL
	}
	// analytics for call enqueue
	event, err := analytics.PrepareAnalyticsEvent(
		analytics.CallEnqueue,
		eP.BotID,
		*req.From,
		*req.To,
		call.DirectionOutbound.String(),
		analytics.AdditionalParams{
			UTMCampaign: eP.CampaignID,
			CallSID:     callSID,
		},
	)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while preparing the analytics event. Error: [%#v]", err)
	} else {
		go event.Push(nil, callSID)
	}
	responseData.ResourceData = &resourceData
	responseData.Msg = "Call Enqueued Successfully"
	responseData.Status = "success"
	response.ResponseData = *responseData
	return response, nil
}

func formQueueMsgParams(callSID string, req contracts.EnqueueCallRequest) (queuemanager.QueueMessageParams, error) {
	var callParam enqueuecallworker.EnqueueCallParams
	callParam.From = req.From
	callParam.To = req.To
	callParam.PipeType = req.PipeType
	callParam.MaxBotFailureCount = req.MaxBotFailureCount
	callParam.CallSID = callSID
	if req.CallbackURL != nil && len(*req.CallbackURL) > 0 {
		callParam.CallbackURL = req.CallbackURL
	}
	if req.RecordingFileName != nil && len(*req.RecordingFileName) > 0 {
		callParam.RecordingFileName = req.RecordingFileName
	}
	if req.TTSOptions != nil {
		callParam.TTSOptions = req.TTSOptions
	}
	if req.ExtraParams != nil {
		callParam.ExtraParams = req.ExtraParams
	}
	msg, err := json.Marshal(callParam)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while Marshalling the JSON. Error: [%#v]", err)
		return configmanager.ConfStore.QueueMessageParams, err
	}
	configmanager.ConfStore.QueueMessageParams.Msg = string(msg)
	return configmanager.ConfStore.QueueMessageParams, nil
}
