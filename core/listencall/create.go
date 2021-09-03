package listencall

import (
	"context"
	"os"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/contracts"
	"bitbucket.org/yellowmessenger/asterisk-ari/eventhandler"
	"bitbucket.org/yellowmessenger/asterisk-ari/phonenumber"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/asterisk"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

var timeLayout = "2006-01-02T15:04:05"

func Create(
	ctx context.Context,
	req contracts.CreateListenCallRequest,
) (
	*contracts.CreateListenCallResponse,
	error,
) {
	callSID := "listen" + call.GenerateCallSID()
	currentTime := time.Now()
	var hostName string
	hostName, err := os.Hostname()
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while getting host name of the server. Error: [%#v]", err)
	}
	ymlogger.LogInfof(callSID, "Listen Call Request: [%#v]", req)
	var toPhoneNumber, fromPhoneNumber phonenumber.PhoneNumber
	fromPhoneNumber, err = eventhandler.ParseNumber(nil, *req.From)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to parse From number from the request. Error: [%#v]", err)
		// Set the default Raw number for each format
		fromPhoneNumber = phonenumber.PhoneNumber{
			RawNumber:      *req.From,
			E164Format:     *req.From,
			LocalFormat:    *req.From,
			NationalFormat: *req.From,
		}
	}
	toPhoneNumber, err = eventhandler.ParseNumber(nil, *req.To)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to parse To number from the request. Error: [%#v]", err)
		// Set the default Raw number for each format
		toPhoneNumber = phonenumber.PhoneNumber{
			RawNumber:      *req.To,
			E164Format:     *req.To,
			LocalFormat:    *req.To,
			NationalFormat: *req.To,
		}
	}
	ymlogger.LogDebugf(callSID, "Parsed FromNumber: [%#v], Parsed ToNumber: [%#v]", fromPhoneNumber, toPhoneNumber)
	channelRes, err := asterisk.CreateCallWithID(nil, callSID, toPhoneNumber, fromPhoneNumber, "incoming", fromPhoneNumber, *req.PipeType)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to create listen call. Error: [%#v]", err.Error())
		return &contracts.CreateListenCallResponse{}, err
	}

	// Set the call data
	call.SetSID(channelRes.ID, callSID)
	call.SetCreatedTime(channelRes.ID, currentTime)
	call.SetDirection(channelRes.ID, call.DirectionOutbound.String())
	call.SetDialedNumber(channelRes.ID, toPhoneNumber)
	call.SetCallerID(channelRes.ID, fromPhoneNumber)
	call.SetListenChannelID(channelRes.ID, *req.ChannelID)

	if strings.ToLower(*req.PipeType) == strings.ToLower(call.PipeTypeSIP.String()) {
		call.SetPipeType(channelRes.ID, strings.ToLower(*req.PipeType))
	}

	response := new(contracts.CreateListenCallResponse)
	responseData := new(contracts.SingleCreateListenCallResponse)
	resourceData := contracts.CreateListenCall{
		SID:             callSID,
		CreatedTime:     currentTime.Format(timeLayout),
		From:            fromPhoneNumber.E164Format,
		To:              toPhoneNumber.E164Format,
		Status:          call.StatusInitiated.String(),
		ChannelID:       channelRes.ID,
		ListenChannelID: *req.ChannelID,
		ServerHost:      hostName,
	}

	responseData.ResourceData = &resourceData
	responseData.Msg = "Listen Initiated Successfully"
	responseData.Status = "success"
	response.ResponseData = *responseData
	return response, nil
}
