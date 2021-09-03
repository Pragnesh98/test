package analytics

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

// EventType describes the type of the event
type EventType string

const (
	CallStart   EventType = "call-start"
	CallDial              = "call-dial"
	CallRing              = "call-ring"
	CallPick              = "call-pick"
	CallEnd               = "call-end"
	TTSChar               = "tts-char"
	STTDur                = "stt-dur"
	CallEnqueue           = "call-enqueue"
)

// AnalyticsQueue is the queue name for analytics
const AnalyticsQueue = "druid-voice-queue"

// Event is structure of event which needs to be sent
type Event struct {
	Queue string `json:"queue"`
	Data  string `json:"data"`
}

type eventData struct {
	EventName EventType `json:"event"`
	CallSID   string    `json:"sid,omitempty"`
	BotID     string    `json:"botId"`
	CallerID  string    `json:"callerId"`
	UserID    string    `json:"uid"`
	Direction string    `json:"direction,omitempty"`
	TimeStamp string    `json:"timestamp"`
	AdditionalParams
}

type AdditionalParams struct {
	CallSID     string `json:"sid,omitempty"`
	UTMCampaign string `json:"utm_campaign,omitempty"`
	Value       string `json:"value"`
	TelcoCode   string `json:"telcoCode,omitempty"`
	TelcoText   string `json:"telco_text,omitempty"`
}

func (e Event) Push(
	ctx context.Context,
	callSID string,
) {
	eventJSON, err := json.Marshal(e)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while marshalling the JSON. Error: [%#v]", err)
		return
	}
	ymlogger.LogDebugf(callSID, "Hitting the Analytics API with the request body: [%s]", string(eventJSON))
	req, err := http.NewRequest(
		http.MethodPost,
		configmanager.ConfStore.AnalyticsEndpoint,
		bytes.NewBuffer(eventJSON),
	)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while forming the Analytics HTTP request. Error: [%#v]", err)
		return
	}
	req.Host = "app.yellowmessenger.com"
	req.Header.Set("Authorization", configmanager.ConfStore.GoogleAccessToken)
	req.Header.Set("Content-Type", "application/json")

	// Initlialize HTTP client
	client := &http.Client{
		Transport: &http.Transport{
			Dial:                (&net.Dialer{Timeout: 3 * time.Second}).Dial,
			TLSHandshakeTimeout: 3 * time.Second,
		},
		Timeout: time.Duration(5 * time.Second),
	}
	defer client.CloseIdleConnections()

	// Make the http request
	response, err := client.Do(req)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while getting the response from Analytics API. Error: [%#v]", err)
		return
	}
	defer response.Body.Close()

	if response != nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
		ymlogger.LogErrorf(callSID, "Got non 2xx response from Analytics API . Response Code: [%d]", response.StatusCode)
		return
	}
	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while reading the response of Analytics API. Error: [%#v]", err)
		return
	}
	ymlogger.LogInfof(callSID, "Got the response from Analytics API. [%#v]", string(respBody))
	return
}

func PrepareAnalyticsEvent(
	eventType EventType,
	botID string,
	callerID string,
	phoneNo string,
	direction string,
	addParams AdditionalParams,
) (*Event, error) {
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	eD := eventData{
		EventName: eventType,
		BotID:     botID,
		CallerID:  callerID,
		UserID:    phoneNo,
		Direction: direction,
		TimeStamp: strconv.FormatInt(currentTime, 10),
	}
	switch eventType {
	case CallEnd:
		eD.Value = "1"
		eD.TelcoCode = addParams.TelcoCode
		eD.TelcoText = addParams.TelcoText
	case CallStart:
		fallthrough
	case CallDial:
		fallthrough
	case CallRing:
		fallthrough
	case CallPick:
		fallthrough
	case CallEnqueue:
		fallthrough
	default:
		eD.Value = "1"
	}
	if len(addParams.CallSID) > 0 {
		eD.CallSID = addParams.CallSID
	}
	if len(addParams.UTMCampaign) > 0 {
		eD.UTMCampaign = addParams.UTMCampaign
	}
	if len(addParams.Value) > 0 {
		eD.Value = addParams.Value
	}
	eventData, err := json.Marshal(eD)
	if err != nil {
		log.Printf("Error while marshalling the event data. [%#v]", err)
		return nil, err
	}
	event := &Event{
		Queue: AnalyticsQueue,
		Data:  string(eventData),
	}
	return event, nil
}
