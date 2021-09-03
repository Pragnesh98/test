package enqueuecallworker

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/callstore"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/contracts"
	"bitbucket.org/yellowmessenger/asterisk-ari/eventhandler"
	"bitbucket.org/yellowmessenger/asterisk-ari/newrelic"
	"bitbucket.org/yellowmessenger/asterisk-ari/phonenumber"
	"bitbucket.org/yellowmessenger/asterisk-ari/queuemanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/analytics"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/asterisk"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

var EnqueueCallChan = make(chan EnqueueCallParams)

// EnqueueCallParams holds the data for scheduling a call
type EnqueueCallParams struct {
	contracts.EnqueueCallRequest
	CallSID string `json:"call_sid"`
}

// EnqueueCallWorker is the worker for the queue messages
type EnqueueCallWorker struct {
	EnqueueCallParams
}

func (ecW *EnqueueCallWorker) Process(jobMsg []byte, botRateLimits *queuemanager.BotRateLimits) queuemanager.QueueJobResult {
	var ecP EnqueueCallParams
	if err := json.Unmarshal(jobMsg, &ecP); err != nil {
		ymlogger.LogErrorf("EnqueueCallWorker", "Error while unmarshalling the JSON. JobMsg: [%s] Error: [%#v]", string(jobMsg), err)
		return queuemanager.QueueJobResult{Status: queuemanager.Failure}
	}

	var toPhoneNumber, fromPhoneNumber phonenumber.PhoneNumber
	toPhoneNumber, err := eventhandler.ParseNumber(nil, *ecP.To)
	if err != nil {
		ymlogger.LogErrorf(ecP.CallSID, "Failed to parse To number from the request. Error: [%#v]", err)
		// Set the default Raw number for each format
		toPhoneNumber = phonenumber.PhoneNumber{
			RawNumber:      *ecP.To,
			E164Format:     *ecP.To,
			LocalFormat:    *ecP.To,
			NationalFormat: *ecP.To,
		}
	}
	fromPhoneNumber, err = eventhandler.ParseNumber(nil, *ecP.From)
	if err != nil {
		ymlogger.LogErrorf(ecP.CallSID, "Failed to parse From number from the request. Error: [%#v]", err)
		// Set the default Raw number for each format
		fromPhoneNumber = phonenumber.PhoneNumber{
			RawNumber:      *ecP.From,
			E164Format:     *ecP.From,
			LocalFormat:    *ecP.From,
			NationalFormat: *ecP.From,
		}
	}
	ymlogger.LogDebugf(ecP.CallSID, "Parsed FromNumber: [%#v], Parsed ToNumber: [%#v]", fromPhoneNumber, toPhoneNumber)

	ymlogger.LogDebugf(ecP.CallSID, "Waiting for ratelimiter for bot: [%#v]", fromPhoneNumber)

	// Ideally this wait and the work after wait should be done in a separate goroutine. Current implementation adds
	// dependency between the bots. If one bot is being throttled, all other campaigns are slowed down since wait for
	// the throttled bot happens on shared thread.
	//
	// If there's an issue with the ratelimiter (like any deadlocks or memory leaks etc), comment out the Wait call.
	// botRateLimits.Wait(context.Background(), fromPhoneNumber.E164Format)

	botRateLimitConf := botRateLimits.GetBotRateLimitConf(fromPhoneNumber.E164Format)
	if botRateLimitConf != nil {
		loc, _ := time.LoadLocation("Asia/Kolkata")
		if botRateLimitConf.MinHour != 0 && botRateLimitConf.MaxHour != 0 && (time.Now().In(loc).Hour() > botRateLimitConf.MaxHour || time.Now().In(loc).Hour() < botRateLimitConf.MinHour) {
			ymlogger.LogDebugf(ecP.CallSID, "Call not allowed at this time for this bot: [%s]", fromPhoneNumber.E164Format)
			return queuemanager.QueueJobResult{
				Status:   queuemanager.TempFailure,
				Priority: 9,
				Delay:    100000, // in MS
			}
		}
	}

	channelRes, err := asterisk.CreateCall(nil, toPhoneNumber, fromPhoneNumber, "incoming", fromPhoneNumber, *ecP.PipeType)
	if err != nil {
		ymlogger.LogErrorf("CreateCall", "Failed to create call. Error: [%#v]", err.Error())
		return queuemanager.QueueJobResult{
			Status:   queuemanager.TempFailure,
			Priority: 9,
			Delay:    10000, // in MS
		}
	}

	// Set the call data
	call.SetBotRateLimiter(channelRes.ID, botRateLimits.GetBotRateLimiter(fromPhoneNumber.E164Format))
	call.SetSID(channelRes.ID, ecP.CallSID)
	call.SetCreatedTime(channelRes.ID, time.Now())
	call.SetDirection(channelRes.ID, call.DirectionOutbound.String())
	call.SetDialedNumber(channelRes.ID, toPhoneNumber)
	call.SetCallerID(channelRes.ID, fromPhoneNumber)
	call.SetCallLatencyStore(channelRes.ID, callstore.LatencyStore{})
	call.SetCallMessageStore(channelRes.ID, callstore.MessageStore{})

	if ecP.CallbackURL != nil && len(*ecP.CallbackURL) > 0 {
		call.SetCallbackURL(channelRes.ID, configmanager.ConfStore.InboundCallbackURL)
	}
	call.SetRecordingFilename(channelRes.ID, ecP.CallSID)
	if ecP.RecordingFileName != nil && len(*ecP.RecordingFileName) > 0 {
		call.SetRecordingFilename(channelRes.ID, ecP.CallSID+"_"+*ecP.RecordingFileName)
	}
	if strings.ToLower(*ecP.PipeType) == strings.ToLower(call.PipeTypeSIP.String()) {
		call.SetPipeType(channelRes.ID, strings.ToLower(*ecP.PipeType))
	}
	if ecP.MaxBotFailureCount != nil && *ecP.MaxBotFailureCount > 0 {
		call.SetMaxBotFailureCount(channelRes.ID, *ecP.MaxBotFailureCount)
	} else {
		call.SetMaxBotFailureCount(channelRes.ID, 7)
	}

	var eP = new(contracts.ExtraParams)
	if ecP.ExtraParams != nil {
		call.SetExtraParams(channelRes.ID, ecP.ExtraParams)
		eP.ExtractExtraParams(ecP.ExtraParams)
		if err == nil && len(eP.BotID) > 0 {
			call.SetBotID(channelRes.ID, eP.BotID)
		}
		if err == nil && len(eP.CampaignID) > 0 {
			call.SetCampaignID(channelRes.ID, eP.CampaignID)
		}
		ecP.ExtraParams = nil
	}
	newrelic.SendCustomEvent("callbacks_metrics", map[string]interface{}{
		"status": "scheduled",
		"value":  1,
	})
	if ecP.TTSOptions != nil && len(ecP.TTSOptions.Message) > 0 {
		call.SetWelcomeMsgAvailable(channelRes.ID, true)
		call.SetTTSOptions(channelRes.ID, ecP.TTSOptions)
	}
	event, err := analytics.PrepareAnalyticsEvent(
		analytics.CallStart,
		eP.BotID,
		fromPhoneNumber.E164Format,
		toPhoneNumber.E164Format,
		call.DirectionOutbound.String(),
		analytics.AdditionalParams{},
	)
	if err != nil {
		ymlogger.LogErrorf(ecP.CallSID, "Error while preparing the analytics event. Error: [%#v]", err)
		return queuemanager.QueueJobResult{Status: queuemanager.Success}
	}
	go event.Push(context.Background(), ecP.CallSID)
	return queuemanager.QueueJobResult{Status: queuemanager.Success}
}

// InitEnqueueCallDequeue start listening to channel to initiate the call
// func InitEnqueueCallDequeue(ctx context.Context) {
// 	for {
// 		select {
// 		case callParam := <-EnqueueCallChan:
// 			ymlogger.LogInfof(callParam.CallSID, "Dequeued Call: [%#v]", callParam)
// 			handleEnqueuCall(callParam)
// 			time.Sleep(time.Duration(configmanager.ConfStore.CampaignDelayPerCall))
// 		case <-ctx.Done():
// 			return
// 		default:
// 		}
// 	}
// }

// func handleEnqueuCall(callParam EnqueueCallParams) {
// 	var toPhoneNumber, fromPhoneNumber phonenumber.PhoneNumber
// 	toPhoneNumber, err := eventhandler.ParseNumber(nil, *callParam.To)
// 	if err != nil {
// 		ymlogger.LogErrorf(callParam.CallSID, "Failed to parse To number from the request. Error: [%#v]", err)
// 		// Set the default Raw number for each format
// 		toPhoneNumber = phonenumber.PhoneNumber{
// 			RawNumber:      *callParam.To,
// 			E164Format:     *callParam.To,
// 			LocalFormat:    *callParam.To,
// 			NationalFormat: *callParam.To,
// 		}
// 	}
// 	fromPhoneNumber, err = eventhandler.ParseNumber(nil, *callParam.From)
// 	if err != nil {
// 		ymlogger.LogErrorf(callParam.CallSID, "Failed to parse From number from the request. Error: [%#v]", err)
// 		// Set the default Raw number for each format
// 		fromPhoneNumber = phonenumber.PhoneNumber{
// 			RawNumber:      *callParam.From,
// 			E164Format:     *callParam.From,
// 			LocalFormat:    *callParam.From,
// 			NationalFormat: *callParam.From,
// 		}
// 	}
// 	ymlogger.LogDebugf(callParam.CallSID, "Parsed FromNumber: [%#v], Parsed ToNumber: [%#v]", fromPhoneNumber, toPhoneNumber)
// 	ymlogger.LogInfof(callParam.CallSID, "Details: %#v", callParam.From)
// 	channelRes, err := asterisk.CreateCall(nil, toPhoneNumber, fromPhoneNumber, "incoming", fromPhoneNumber, *callParam.PipeType)
// 	if err != nil {
// 		ymlogger.LogErrorf("CreateCall", "Failed to create call. Error: [%#v]", err.Error())
// 		return
// 	}

// 	// Set the call data
// 	call.SetSID(channelRes.ID, callParam.CallSID)
// 	call.SetCreatedTime(channelRes.ID, time.Now())
// 	call.SetDirection(channelRes.ID, call.DirectionOutbound.String())
// 	call.SetDialedNumber(channelRes.ID, toPhoneNumber)
// 	call.SetCallerID(channelRes.ID, fromPhoneNumber)
// 	if callParam.CallbackURL != nil && len(*callParam.CallbackURL) > 0 {
// 		call.SetCallbackURL(channelRes.ID, configmanager.ConfStore.InboundCallbackURL)
// 	}
// 	if strings.ToLower(*callParam.PipeType) == strings.ToLower(call.PipeTypeSIP.String()) {
// 		call.SetPipeType(channelRes.ID, strings.ToLower(*callParam.PipeType))
// 	}
// 	return
// }
