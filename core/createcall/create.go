package createcall

import (
	"context"
	"os"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/callstore"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/contracts"
	"bitbucket.org/yellowmessenger/asterisk-ari/eventhandler"
	"bitbucket.org/yellowmessenger/asterisk-ari/globals"
	"bitbucket.org/yellowmessenger/asterisk-ari/newrelic"
	"bitbucket.org/yellowmessenger/asterisk-ari/phonenumber"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/analytics"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/asterisk"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

var timeLayout = "2006-01-02T15:04:05"

func Create(
	ctx context.Context,
	req contracts.CreateCallRequest,
) (
	*contracts.CreateCallResponse,
	error,
) {
	callSID := call.GenerateCallSID()
	currentTime := time.Now()
	var hostName string
	hostName, err := os.Hostname()
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while getting host name of the server. Error: [%#v]", err)
	}
	ymlogger.LogInfof(callSID, "Create Call Request: [%#v]", req)
	var toPhoneNumber, fromPhoneNumber phonenumber.PhoneNumber
	fromPhoneNumber, err = eventhandler.ParseNumber(nil, *req.From)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to parse From number from the request. Error: [%#v]", err)
		// Set the default Raw number for each format
		fromPhoneNumber = phonenumber.PhoneNumber{
			RawNumber:              *req.From,
			E164Format:             *req.From,
			LocalFormat:            *req.From,
			NationalFormat:         *req.From,
			WithZeroNationalFormat: *req.From,
		}
	}
	toPhoneNumber, err = eventhandler.ParseNumber(nil, *req.To)
	if err != nil || strings.HasPrefix(strings.ToLower(*req.To), "sip") {
		ymlogger.LogErrorf(callSID, "Failed to parse To number from the request. Error: [%#v]", err)
		// Set the default Raw number for each format
		toPhoneNumber = phonenumber.PhoneNumber{
			RawNumber:              *req.To,
			E164Format:             *req.To,
			LocalFormat:            *req.To,
			NationalFormat:         *req.To,
			WithZeroNationalFormat: *req.To,
		}
	}
	ymlogger.LogDebugf(callSID, "Parsed FromNumber: [%#v], Parsed ToNumber: [%#v]", fromPhoneNumber, toPhoneNumber)
	channelRes, err := asterisk.CreateCall(nil, toPhoneNumber, fromPhoneNumber, "incoming", fromPhoneNumber, *req.PipeType)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to create call. Error: [%#v]", err.Error())
		return &contracts.CreateCallResponse{}, err
	}
	//Increment call count.
	globals.IncrementNoOfCalls()
	ymlogger.LogInfof(callSID, "Number of calls [%d]. Number of call objects [%d]", globals.GetNoOfCalls(), globals.GetNoOfCallObject())
	// Set the call data
	call.SetSID(channelRes.ID, callSID)
	call.SetCreatedTime(channelRes.ID, currentTime)
	call.SetDirection(channelRes.ID, call.DirectionOutbound.String())
	call.SetDialedNumber(channelRes.ID, toPhoneNumber)
	call.SetCallerID(channelRes.ID, fromPhoneNumber)
	call.SetRecordingFilename(channelRes.ID, callSID)

	call.SetCallLatencyStore(channelRes.ID, callstore.LatencyStore{})
	call.SetCallMessageStore(channelRes.ID, callstore.MessageStore{})
	
	if req.RecordingFileName != nil && len(*req.RecordingFileName) > 0 {
		call.SetRecordingFilename(channelRes.ID, callSID+"_"+*req.RecordingFileName)
	}
	if req.CallbackURL != nil && len(*req.CallbackURL) > 0 {
		call.SetCallbackURL(channelRes.ID, configmanager.ConfStore.InboundCallbackURL)
	}
	if strings.ToLower(*req.PipeType) == strings.ToLower(call.PipeTypeSIP.String()) {
		call.SetPipeType(channelRes.ID, strings.ToLower(*req.PipeType))
	}
	newrelic.SendCustomEvent("callbacks_metrics", map[string]interface{}{
		"status": "scheduled",
		"value":  1,
	})
	event, err := analytics.PrepareAnalyticsEvent(
		analytics.CallStart,
		call.GetBotID(channelRes.ID),
		fromPhoneNumber.E164Format,
		toPhoneNumber.E164Format,
		call.DirectionOutbound.String(),
		analytics.AdditionalParams{},
	)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while preparing the analytics event. Error: [%#v]", err)
	} else {
		go event.Push(ctx, callSID)
	}
	response := new(contracts.CreateCallResponse)
	responseData := new(contracts.SingleCreateCallResponse)
	resourceData := contracts.CreateCall{
		SID:         callSID,
		CreatedTime: currentTime.Format(timeLayout),
		From:        fromPhoneNumber.E164Format,
		To:          toPhoneNumber.E164Format,
		Status:      call.StatusInitiated.String(),
		ChannelID:   channelRes.ID,
		ServerHost:  hostName,
	}
	if req.CallbackURL != nil && len(*req.CallbackURL) > 0 {
		resourceData.CallbackURL = *req.CallbackURL
	}
	responseData.ResourceData = &resourceData
	responseData.Msg = "Call Initiated Successfully"
	responseData.Status = "success"
	response.ResponseData = *responseData
	return response, nil
}
