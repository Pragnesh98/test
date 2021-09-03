package azure

import (
	"strconv"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/helper"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/speechlogging"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

var azureSTTHTTPClient *http.Client

type SpeechToTextAPIResponse struct {
	RecognitionStatus string `json:"RecognitionStatus"`
	DisplayText       string `json:"DisplayText"`
	OffSet            int64  `json:"Offset"`
	Duration          int64  `json:"Duration"`
}

// InitAzureSTTHTTPClient initializes the bot's HTTP client
func InitAzureSTTHTTPClient() {
	azureSTTHTTPClient = &http.Client{
		Transport: &http.Transport{
			// Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConnsPerHost:   100,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   3 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: 20000 * time.Millisecond,
	}
}

func getSTTEndpoint(channelID, sttEndpoint string) (endpoint string) {
	endpoint = sttEndpoint
	botOptions := call.GetBotOptions(channelID)
	if botOptions == nil {
		return
	}
	if botOptions.MicrosoftSTTOptions.EndpointId == "" {
		return
	}
	sttURL, err := url.Parse(sttEndpoint)
	if err != nil {
		return
	}
	q, err := url.ParseQuery(sttURL.RawQuery)
	if err != nil {
		return
	}
	q.Add("cid", botOptions.MicrosoftSTTOptions.EndpointId)
	sttURL.RawQuery = q.Encode()
	endpoint = sttURL.String()
	return
}

func GetTextFromSpeech(
	ctx context.Context,
	channelID string,
	callSID string,
	fileName string,
) (string, error) {
	var text string
	sttOutcome := helper.STTOutcome{
		MetricName  :   "STTOutcome", 
		Stt_engine  :   "Azure",
		Stt_type    :   "REST",
		Reason      :   "FAILED",
	}
	
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return text, err
	}
	if fileInfo.Size() == 0 {
		return text, errors.New("File size is zero")
	}
	// token, err := GetSTTAuthorizationToken()
	if err != nil {
		ymlogger.LogErrorf(callSID, "[AzureRest] Error while generating authorization token. Error: [%#v]", err)
		return text, err
	}

	// Read the audio file
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		ymlogger.LogErrorf(callSID, "[AzureRest] Error while reading the file for speech to text. Error: [%#v]", err)
		return text, err
	}

	sttEndpoint := getSTTEndpoint(channelID, configmanager.ConfStore.AzureSTTEndpoint)
	azureSTTAPIKey := configmanager.ConfStore.AzureSTTAPIKey
	if call.GetBotOptions(channelID).UseNewMsSubscritpion {
		azureSTTAPIKey = configmanager.ConfStore.AzureSTTAPIKeyNew
	}
	// Prepare STT request
	sttReq, err := http.NewRequest(http.MethodPost, sttEndpoint, bytes.NewBuffer(dat))
	if err != nil {
		ymlogger.LogErrorf(callSID, "[AzureRest] Error while preparing TTS request. Error: [%#v]", err)
		return text, err
	}
	ymlogger.LogDebugf(callSID, "[AzureRest] STT Language: [%s]", call.GetSTTLanguage(channelID))
	// Set Query Parameters
	q := sttReq.URL.Query()
	q.Add("language", call.GetSTTLanguage(channelID))
	sttReq.URL.RawQuery = q.Encode()
	// Set Request Headers
	sttReq.Header.Set("Ocp-Apim-Subscription-Key", azureSTTAPIKey)
	sttReq.Header.Set("Content-type", "audio/wav; codecs=audio/pcm; samplerate=16000 and audio/ogg; codecs=opus.")
	// sttReq.Header.Set("Authorization", "Bearer "+token)

	// Initlialize HTTP client
	// client := &http.Client{
	// 	Transport: &http.Transport{
	// 		Dial:                (&net.Dialer{Timeout: 3 * time.Second}).Dial,
	// 		TLSHandshakeTimeout: 3 * time.Second,
	// 		MaxIdleConns:        100,
	// 		MaxIdleConnsPerHost: 20,
	// 	},
	// 	Timeout: time.Duration(10 * time.Second),
	// }
	// defer client.CloseIdleConnections()

	var response *http.Response
	var latency int64
	for i := 0; i < configmanager.ConfStore.ARIAPIRetry; i++ {
		// Capture API response time
		sTime := time.Now()
		ymlogger.LogDebugf(callSID, "[AzureRest] Make the http request [%#v]", sTime)
		response, err = azureSTTHTTPClient.Do(sttReq)
		if response != nil {
			defer response.Body.Close()
		}
		latency = time.Since(sTime).Milliseconds()
		// Send API response time to newrelic
		go helper.SendResponseTimeMetric("voice_azure_stt", call.GetCampaignID(channelID), latency)

		if err != nil || response == nil || response.StatusCode < 200 || response.StatusCode >= 300 {
			continue
		}
		break
	}
	defer func(callSID string) {
		go helper.SendSTTOutcome(sttOutcome.MetricName, channelID, callSID, sttOutcome.Status, sttOutcome.Reason, sttOutcome.Stt_engine, sttOutcome.Stt_type, 0)
	}(callSID)

	ymlogger.LogDebugf(callSID, "[AzureRest] STT Language: [%s]", call.GetSTTLanguage(channelID))
	
	if err != nil {
		ymlogger.LogErrorf(callSID, "[AzureRest] Error while getting the response for STT request. Error: [%#v]", err.Error())
		return text, err
	}
	sttOutcome.Status = strconv.Itoa(response.StatusCode)
	if response != nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
		return text, errors.New("Non 2xx response")
	}

	// Read the content from the response
	respData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ymlogger.LogErrorf(callSID, "[AzureRest] Failed to read the response from STT Response body. Error: [%#v]", err)
		return text, err
	}
	var sttResponse SpeechToTextAPIResponse
	err = json.Unmarshal(respData, &sttResponse)
	if err != nil {
		ymlogger.LogErrorf(callSID, "[AzureRest] Error while unmarshalling the response. Error: [%#v] Body: [%#v]", err, string(respData))
		sttOutcome.Reason = "INV_RESPONSE"
		return text, err
	}
	sttOutcome.Reason = sttResponse.RecognitionStatus
	
	// Send request and response for logging
	callStartTime := call.GetPickupTime(channelID)
	botID := call.GetCallerID(channelID).E164Format
	userID := call.GetDialedNumber(channelID).E164Format

	logsRequest := &speechlogging.LoggingRequest{}
	logsRequest.
		SetCallSID(callSID).
		SetAudio(call.GetCaptureVoiceOTP(channelID), call.GetSTTLanguage(channelID), dat).
		SetLatencyMillis(latency).
		SetSttService("azure").
		SetTranscript(sttResponse.DisplayText).
		SetBotID(botID).
		SetUserID(userID).
		SetCallStartTime(callStartTime)

	go speechlogging.Send(logsRequest, speechlogging.URL)

	// Send accuracy to new relic
	go helper.SendAccuracyMetric("voice_azure_stt", call.GetCampaignID(channelID), sttResponse.DisplayText)
	// Send STT Duration metric
	go helper.SendSTTDurationMetric(callSID, channelID, "voice_azure_stt", fileName, call.GetBotID(channelID), call.GetCallerID(channelID).NationalFormat)
	return sttResponse.DisplayText, nil
}

func GetSTTAuthorizationToken() (string, error) {
	var token string
	postData := []byte("")
	tokenReq, err := http.NewRequest(http.MethodPost, configmanager.ConfStore.AzureTokenEndpoint, bytes.NewBuffer(postData))
	if err != nil {
		return token, err
	}
	tokenReq.Header.Set("Content-type", "application/x-www-form-urlencoded")
	tokenReq.Header.Set("Content-Length", "0")
	tokenReq.Header.Set("Ocp-Apim-Subscription-Key", configmanager.ConfStore.AzureSTTAPIKey)
	// Initlialize HTTP client
	client := &http.Client{
		Transport: &http.Transport{
			Dial:                (&net.Dialer{Timeout: 3 * time.Second}).Dial,
			TLSHandshakeTimeout: 3 * time.Second,
		},
		Timeout: time.Duration(5 * time.Second),
	}
	defer client.CloseIdleConnections()

	var response *http.Response
	for i := 0; i < configmanager.ConfStore.ARIAPIRetry; i++ {
		// Make the http request
		response, err = client.Do(tokenReq)
		if response != nil {
			defer response.Body.Close()
		}
		if err != nil || response == nil || response.StatusCode < 200 || response.StatusCode >= 300 {
			continue
		}
		break
	}
	if err != nil {
		return token, err
	}
	if response != nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
		return token, errors.New("Non 2xx response")
	}
	respData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return token, err
	}
	return string(respData), nil
}
