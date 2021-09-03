package callstore

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

// CallStore contains information about the call
type CallStore struct {
	CallSID          string             `json:"sid,omitempty"`
	Direction        string             `json:"direction,omitempty"`
	BotID            string             `json:"botId,omitempty"`
	CampaignID       string             `json:"campaignId,omitempty"`
	From             string             `json:"from,omitempty"`
	To               string             `json:"to,omitempty"`
	ForwardingNumber string             `json:"forwarding_number,omitempty"`
	Status           string             `json:"status,omitempty"`
	StartTime        string             `json:"start_time,omitempty"`
	DialTime         string             `json:"dial_time,omitempty"`
	PickupTime       string             `json:"pickup_time,omitempty"`
	EndTime          string             `json:"end_time,omitempty"`
	Duration         int                `json:"duration,omitempty"`
	BillDuration     int                `json:"bill_duration,omitempty"`
	RingingDuration  int                `json:"ringing_duration,omitempty"`
	TelcoCode        int                `json:"telco_code,omitempty"`
	TelcoText        string             `json:"telco_text,omitempty"`
	DisconnectedBy   string             `json:"disconnected_by,omitempty"`
	BotFailed        bool               `json:"bot_failed,omitempty"`
	CallbackURL      string             `json:"callback_url,omitempty"`
	RecordingURL     string             `json:"recording_url,omitempty"`
	Transcript       string             `json:"transcript"`
	STTDuration      int64              `json:"stt_duration,omitempty"`
	TTSCharacters    int64              `json:"tts_characters,omitempty"`
	Transcripts      []string           `json:"transcripts,omitempty"`
	ExtraParams      interface{}        `json:"extra_params,omitempty"`
	Host             string             `json:"host,omitempty"`
	LatencyInfo      []LatencyParameter `json:"latency_info"`
	Messages         []Message          `json:"messages"`
}

var callStoreClient *http.Client

// InitCallStoreClient initializes the http client
func InitCallStoreClient() {
	callStoreClient = &http.Client{
		Transport: &http.Transport{
			Dial:                (&net.Dialer{Timeout: 3 * time.Second}).Dial,
			TLSHandshakeTimeout: 3 * time.Second,
		},
		Timeout: time.Duration(5 * time.Second),
	}
}

// GetCallStore gives CallStore struct
func GetCallStore(
	callSID string,
	from string,
	to string,
	direction string,
	callbackURL string,
	host string,
) *CallStore {
	return &CallStore{
		CallSID:     callSID,
		From:        from,
		To:          to,
		Direction:   direction,
		CallbackURL: callbackURL,
		Host:        host,
	}
}

// Create creates the new record for a call in mongoDB.
func (c *CallStore) Create() {
	jsonData, err := json.Marshal(c)
	if err != nil {
		ymlogger.LogErrorf(c.CallSID, "Failed to marshal the data into JSON. Error: [%#v]", err)
		return
	}
	ymlogger.LogDebugf(c.CallSID, "Hitting the Logs Store API with the request body: [%s]", string(jsonData))

	req, err := http.NewRequest(
		http.MethodPost,
		configmanager.ConfStore.LogStoreEndpoint,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		ymlogger.LogErrorf(c.CallSID, "Error while forming the Analytics HTTP request. Error: [%#v]", err)
		return
	}

	req.Host = "app.yellowmessenger.com"
	req.Header.Set("Authorization", configmanager.ConfStore.GoogleAccessToken)
	req.Header.Set("Content-Type", "application/json")

	// Make the http request
	response, err := callStoreClient.Do(req)
	if err != nil {
		ymlogger.LogErrorf(c.CallSID, "Error while getting the response from Log Store API. Error: [%#v]", err)
		return
	}
	defer response.Body.Close()

	if response != nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
		ymlogger.LogErrorf(c.CallSID, "Got non 2xx response from Log Store API . Response Code: [%d]", response.StatusCode)
		return
	}
	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ymlogger.LogErrorf(c.CallSID, "Error while reading the response of Log Store API. Error: [%#v]", err)
		return
	}
	ymlogger.LogInfof(c.CallSID, "Got the response from Log Store API. [%#v]", string(respBody))
	return
}

// Update updates the call in mongoDB
func (c *CallStore) Update(
	callSID string,
) {
	jsonData, err := json.Marshal(c)
	if err != nil {
		ymlogger.LogErrorf(c.CallSID, "Failed to marshal the data into JSON. Error: [%#v]", err)
		return
	}
	ymlogger.LogDebugf(callSID, "Hitting the Logs Store API with the request body: [%s]", string(jsonData))
	req, err := http.NewRequest(
		http.MethodPut,
		configmanager.ConfStore.LogStoreEndpoint+"/"+callSID,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while forming the Analytics HTTP request. Error: [%#v]", err)
		return
	}

	req.Host = "app.yellowmessenger.com"
	req.Header.Set("Authorization", configmanager.ConfStore.GoogleAccessToken)
	req.Header.Set("Content-Type", "application/json")

	// Make the http request
	response, err := callStoreClient.Do(req)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while getting the response from Log Store API. Error: [%#v]", err)
		return
	}
	defer response.Body.Close()

	if response != nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
		ymlogger.LogErrorf(callSID, "Got non 2xx response from Log Store API . Response Code: [%d]", response.StatusCode)
		return
	}
	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while reading the response of Log Store API. Error: [%#v]", err)
		return
	}
	ymlogger.LogInfof(callSID, "Got the response from Log Store API. [%#v]", string(respBody))
	return
}
