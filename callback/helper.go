package callback

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/callstore"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/metrics"
	"bitbucket.org/yellowmessenger/asterisk-ari/models/mysql"
	"bitbucket.org/yellowmessenger/asterisk-ari/newrelic"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"

	guuid "github.com/google/uuid"
)

var timeLayout = "2006-01-02T15:04:05.000"

// CallbackRequest holds the parameters to be sent in the call back
type CallbackRequest struct {
	TraceID          string            `json:"traceId"`
	SID              string            `json:"sid"`
	Direction        string            `json:"direction"`
	Status           string            `json:"status"`
	PhoneNumber      string            `json:"phone_no"`
	StartTime        string            `json:"start_time"`
	RingingTime      int               `json:"ringing_time"`
	Duration         int               `json:"duration"`
	CallerID         string            `json:"caller_id"`
	ForwardingNumber string            `json:"forwarding_number"`
	DialTime         string            `json:"dial_time"`
	PickTime         string            `json:"pick_time"`
	HoldDuration     int               `json:"hold_duration"`
	EndTime          string            `json:"end_time"`
	TelcoCode        int               `json:"telco_code"`
	TelcoText        string            `json:"telco_text"`
	DisconnectedBy   string            `json:"disconnected_by"`
	RecordingURL     string            `json:"recording_url"`
	BotFailed        bool              `json:"bot_failed"`
	ChildLegs        []CallbackRequest `json:"child_legs,omitempty"`
	ExtraParams      interface{}       `json:"extra_params"`
}

// StoreCallbackRequest makes the callback to callback url
func StoreCallbackRequest(
	ctx context.Context,
	channelID string,
	callSID string,
) error {
	callbackReqBody := prepareCallbackRequestBody(ctx, channelID, callSID)
	callbackReqBodyJSON, err := json.Marshal(callbackReqBody)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to prepare the call back request body. Error: [%#v]", err)
		return err
	}
	ymlogger.LogDebugf(callSID, "Saving Callback record with request Body: [%v], ChannelID: [%s]", string(callbackReqBodyJSON), channelID)
	if err = mysql.InsertCallbackRecord(callSID, call.GetCallbackURL(channelID), string(callbackReqBodyJSON)); err != nil {
		ymlogger.LogErrorf(callSID, "Failed to save the callback record in DB. Error: [%#v]", err)
		return err
	}
	ymlogger.LogInfo(callSID, "Saving CallStore record")
	callStore, err := prepareUpdateCallStore(channelID, callbackReqBody)
	ymlogger.LogInfof(callSID, "Saving CallStore record with request Body: [%v]",  string(callbackReqBodyJSON))

	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while preparing update call store request body. Error: [%#v]", err)
	} else {
		go callStore.Update(callSID)
	}
	sendMetric(callbackReqBody, call.GetCampaignID(channelID), call.GetDirection(channelID))
	return nil
}

func prepareCallbackRequestBody(
	ctx context.Context,
	channelID string,
	callSID string,
) CallbackRequest {
	var req CallbackRequest
	req = fillParams(req, callSID, channelID)
	if len(call.GetChildrenUniqueIDs(channelID)) > 0 {
		for _, child := range call.GetChildrenUniqueIDs(channelID) {
			var cReq CallbackRequest
			cReq = fillParams(cReq, callSID, child)
			req.ChildLegs = append(req.ChildLegs, cReq)
		}
	}
	return req
}

func fillParams(req CallbackRequest, callSID, channelID string) CallbackRequest {
	req.SID = callSID
	req.TraceID = guuid.New().String()
	req.Direction = call.GetDirection(channelID)
	req.Status = call.GetStatus(channelID)
	req.PhoneNumber = call.GetDialedNumber(channelID).E164Format
	if !call.GetCreatedTime(channelID).IsZero() {
		req.StartTime = call.GetCreatedTime(channelID).Format(timeLayout)
	}
	req.RingingTime = call.GetRingDuration(channelID)
	req.Duration = call.GetDuration(channelID)
	req.CallerID = call.GetCallerID(channelID).E164Format
	req.ForwardingNumber = call.GetForwardingNumber(channelID).E164Format
	if !call.GetDialingTime(channelID).IsZero() {
		req.DialTime = call.GetDialingTime(channelID).Format(timeLayout)
	}
	if !call.GetPickupTime(channelID).IsZero() {
		req.PickTime = call.GetPickupTime(channelID).Format(timeLayout)
	}
	req.HoldDuration = call.GetHoldDuration(channelID)
	if !call.GetEndTime(channelID).IsZero() {
		req.EndTime = call.GetEndTime(channelID).Format(timeLayout)
	}
	if len(call.GetCause(channelID)) > 0 {
		req.TelcoCode = call.GetCause(channelID)[0].Code
		req.TelcoText = call.GetCause(channelID)[0].Text
	}
	disconnectedBy := call.GetDisconnectedBy(channelID)
	if len(disconnectedBy) > 0 {
		req.DisconnectedBy = disconnectedBy
	} else {
		req.DisconnectedBy = "user"
	}
	req.RecordingURL = call.GetRecordingURL(channelID)
	req.BotFailed = call.GetBotFailed(channelID)
	if call.GetExtraParams(channelID) != nil {
		req.ExtraParams = call.GetExtraParams(channelID)
	}

	return req
}

func sendMetric(req CallbackRequest, campaignID, direction string) {
	eventData := map[string]interface{}{
		"caller_id":   req.CallerID,
		"campaign_id": campaignID,
		"status":      req.Status,
		"telco_code":  strconv.Itoa(req.TelcoCode),
		"direction":   direction,
		"bot_failed":  strconv.FormatBool(req.BotFailed),
		"count":       1,
	}
	if err := newrelic.SendCustomEvent("call_stats", eventData); err != nil {
		ymlogger.LogErrorf("NewRelicMetric", "Failed to send call_stats metric to new relic. Error: [%#v]", err)
	}
	filters := make(map[string]string)
	fields := make(map[string]interface{})
	filters["from"] = req.CallerID
	filters["status"] = req.Status
	filters["direction"] = "outgoing"
	filters["bot_failed"] = strconv.FormatBool(req.BotFailed)
	fields["count"] = 1
	metric, err := metrics.NewMetric("call.stats", filters, fields)
	if err != nil {
		ymlogger.LogErrorf("SendMetric", "Failed to create metric. Error: [%#v]", err)
		return
	}
	if err := metrics.SendMetric(metric); err != nil {
		ymlogger.LogErrorf("SendMetric", "Failed to send metrics. Error: [%#v]", err)
		return
	}
	ymlogger.LogInfof("SendMetric", "Successfully sent the call stats metric. Filters: [%#v] Fields: [%#v]", filters, fields)
	return
}

// MockMakeCallbackRequest makes the callback to callback url
func MockMakeCallbackRequest(
	ctx context.Context,
	toNumber string,
	callSID string,
	extraParams interface{},
) error {
	time.Sleep(time.Duration(rand.Intn(30-1)+1) * time.Second)
	var req CallbackRequest
	req.SID = callSID
	req.Direction = "outbound"
	req.Status = "failed"
	req.PhoneNumber = toNumber
	req.StartTime = time.Now().Format(timeLayout)
	req.RingingTime = 0
	req.Duration = 0
	req.CallerID = "+918068402307"
	req.DialTime = time.Now().Format(timeLayout)
	req.PickTime = time.Now().Format(timeLayout)
	req.EndTime = time.Now().Format(timeLayout)
	req.TelcoCode = 0
	req.TelcoText = "Unknown"
	req.DisconnectedBy = "bot"
	req.RecordingURL = ""
	req.BotFailed = false
	req.ExtraParams = extraParams

	callbackReqBodyJSON, err := json.Marshal(req)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to prepare the call back request body. Error: [%#v]", err)
		return err
	}
	ymlogger.LogDebugf(callSID, "Hitting Callback URL with request Body: [%v]", string(callbackReqBodyJSON))
	callbackReq, err := http.NewRequest(http.MethodPost, "https://app.yellowmessenger.com/integrations/voice/callback", bytes.NewBuffer(callbackReqBodyJSON))
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to prepare the call back request. Error: [%#v]", err)
		return err
	}
	callbackReq.Header.Set("Content-Type", "application/json")
	callbackReq.Header.Set("Connection", "close")

	client := &http.Client{
		Transport: &http.Transport{
			Dial:                (&net.Dialer{Timeout: 3 * time.Second}).Dial,
			TLSHandshakeTimeout: 3 * time.Second,
		},
		Timeout: time.Duration(5 * time.Second),
	}
	defer client.CloseIdleConnections()
	for i := 0; i < configmanager.ConfStore.CallbackMaxTries; i++ {
		response, err := client.Do(callbackReq)
		if response == nil || response.StatusCode < 200 || response.StatusCode >= 300 || err != nil {
			ymlogger.LogErrorf(callSID, "Retry: [%d]. Failed hitting the callback URL. Response: [%#v]. Error: [%#v]. Retrying", (i + 1), response, err)
			continue
		}
		defer response.Body.Close()
		ymlogger.LogInfof(callSID, "Successful response from the callback. StatusCode: [%#v]", response.StatusCode)
		respBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			ymlogger.LogErrorf(callSID, "Failed to get body from the response. Error: [%#v]", err)
		}
		ymlogger.LogInfof(callSID, "Successful response from the callback. Body: [%#v]", string(respBody))
		break
	}
	return nil
}

func prepareUpdateCallStore(
	channelID string,
	callBackReq CallbackRequest,
) (*callstore.CallStore, error) {

	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return nil, err
	}
	var sTime time.Time
	if len(callBackReq.StartTime) > 0 {
		sTime, err = time.ParseInLocation(timeLayout, callBackReq.StartTime, loc)
		if err != nil {
			return nil, err
		}
	}
	var dTime time.Time
	if len(callBackReq.DialTime) > 0 {
		dTime, err = time.ParseInLocation(timeLayout, callBackReq.DialTime, loc)
		if err != nil {
			return nil, err
		}
	}
	var pTime time.Time
	if len(callBackReq.PickTime) > 0 {
		pTime, err = time.ParseInLocation(timeLayout, callBackReq.PickTime, loc)
		if err != nil {
			return nil, err
		}
	}
	var eTime time.Time
	if len(callBackReq.EndTime) > 0 {
		eTime, err = time.ParseInLocation(timeLayout, callBackReq.EndTime, loc)
		if err != nil {
			return nil, err
		}
	}

	callStore := &callstore.CallStore{
		BotID:            call.GetBotID(channelID),
		CampaignID:       call.GetCampaignID(channelID),
		Status:           callBackReq.Status,
		StartTime:        sTime.UTC().Format(timeLayout),
		DialTime:         dTime.UTC().Format(timeLayout),
		PickupTime:       pTime.UTC().Format(timeLayout),
		EndTime:          eTime.UTC().Format(timeLayout),
		RingingDuration:  callBackReq.RingingTime,
		ForwardingNumber: callBackReq.ForwardingNumber,
		Duration:         callBackReq.Duration,
		TelcoCode:        callBackReq.TelcoCode,
		TelcoText:        callBackReq.TelcoText,
		DisconnectedBy:   callBackReq.DisconnectedBy,
		RecordingURL:     callBackReq.RecordingURL,
		Transcripts:      call.GetTranscripts(channelID),
		BotFailed:        callBackReq.BotFailed,
		STTDuration:      call.GetSTTDuration(channelID),
		TTSCharacters:    call.GetTTSCharacters(channelID),
		ExtraParams:      callBackReq.ExtraParams,
	}
	latencyStore := call.GetCallLatencyStore(channelID)
	if latencyStore != nil {
		callStore.LatencyInfo = latencyStore.GetLatencies(channelID)
	}
	messageStore := call.GetCallMessageStore(channelID)
	if messageStore != nil {
		callStore.Messages = messageStore.GetMessages(channelID)
	}
	return callStore, nil
}
