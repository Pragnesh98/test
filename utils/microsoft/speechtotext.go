package microsoft

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/helper"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/speechlogging"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

var microsoftHTTPClient *http.Client

type AudioSource struct {
	Content string `json:"content"`
}
type Audio struct {
	AudioSource `json:"audio_source"`
}
type Config struct {
	SampleRateHertz int    `json:"sample_rate_hertz"`
	LanguageCode    string `json:"language_code"`
	Encoding        string `json:"encoding"`
}
type AzureOptions struct {
	EndpointID string `json:"endpoint_id"`
}
type Options struct {
	AzureOptions AzureOptions `json:"azure_options"`
}
type MicrosoftSTTRequest struct {
	Audio   Audio   `json:"audio"`
	Config  Config  `json:"config"`
	Options Options `json:"options"`
}

type ResponseAlternative struct {
	Transcript string  `json:"transcript"`
	Confidence float64 `json:"confidence"`
}
type ResponseResult struct {
	Alternative []ResponseAlternative `json:"alternatives"`
}
type RecognizeResponse struct {
	Results []ResponseResult `json:"results"`
}
type MicrosoftSTTResponse struct {
	RecognizeResponse RecognizeResponse
}

func InitAzureSTTHTTPClient() {
	microsoftHTTPClient = &http.Client{
		Timeout: 20000 * time.Millisecond,
	}
}

func GetSpeechDataFromFile(callSID string, filename string) ([]byte, error) {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	if fileInfo.Size() == 0 {
		return nil, errors.New("File size is zero")
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while reading the file for speech to text. Error: [%#v]", err)
		return nil, err
	}
	return data, err
}

func CreateRequest(channelID string, callSID string, data []byte) (*http.Request, error) {

	dataBase64 := b64.StdEncoding.EncodeToString(data)
	sttEndpoint := configmanager.ConfStore.MicrosoftSDKEndpoint
	ymlogger.LogDebugf(callSID, "STT Language: [%s]", call.GetSTTLanguage(channelID))
	// Set Query Parameters

	requestBody := &MicrosoftSTTRequest{
		Audio: Audio{
			AudioSource{
				Content: dataBase64},
		},
		Config: Config{
			LanguageCode: call.GetSTTLanguage(channelID),
		},
	}
	req, err := json.Marshal(requestBody)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Couldn't marshall STT request body. Error: [%#v]", err)
		return nil, err
	}

	sttReq, err := http.NewRequest(http.MethodPost, sttEndpoint, bytes.NewBuffer(req))
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while preparing STT request. Error: [%#v]", err)
		return sttReq, err
	}
	return sttReq, err
}

func GetSTTHTTPResponse(request *http.Request, campaignID string) (response *http.Response, err error, latency int64) {

	for i := 0; i < configmanager.ConfStore.ARIAPIRetry; i++ {
		// Capture API response time
		sTime := time.Now()
		// http request
		response, err := microsoftHTTPClient.Do(request)
		if response != nil {
			defer response.Body.Close()
		}
		latency = time.Since(sTime).Milliseconds()
		// Send API response time to newrelic
		go helper.SendResponseTimeMetric("voice_microsoft_stt", campaignID, latency)

		if err != nil || response == nil || response.StatusCode < 200 || response.StatusCode >= 300 {
			continue
		}
		break
	}
	return
}

func GetResultTextFromSTTResponse(callSID string, respData []byte) (string, error) {
	var text string
	var sttResponse *MicrosoftSTTResponse
	err := json.Unmarshal(respData, &sttResponse)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while unmarshalling the response. Error: [%#v] Body: [%#v]", err, string(respData))
		return text, err
	}

	if len(sttResponse.RecognizeResponse.Results) == 0 || len(sttResponse.RecognizeResponse.Results[0].Alternative) == 0 {
		ymlogger.LogInfof(callSID, "Empty transcript from results from Azure SDK")
		return text, err
	}

	finalResult := sttResponse.RecognizeResponse.Results[0].Alternative[0]
	results := sttResponse.RecognizeResponse.Results[0].Alternative
	for _, result := range results {
		ymlogger.LogInfof(callSID, "[%v] (confidence=%v)", result.Transcript, result.Confidence)
		if result.Confidence > finalResult.Confidence {
			finalResult = result
		}
	}

	return finalResult.Transcript, nil
}

func GetTextFromSpeech(
	ctx context.Context,
	channelID string,
	callSID string,
	fileName string,
) (string, error) {
	var text string
	//get speech data
	data, err := GetSpeechDataFromFile(callSID, fileName)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Couldn't recieve data from file. Error: [%#v]", err)
		return text, err
	}

	// create the STT request
	sttRequest, err := CreateRequest(channelID, callSID, data)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Request couldn't be prepared. Error: [%#v]", err)
		return text, err
	}

	// get response
	response, err, latency := GetSTTHTTPResponse(sttRequest, call.GetCampaignID(channelID))
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while getting the response for STT request. Error: [%#v]", err.Error())
		return text, err
	}
	if response != nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
		return text, errors.New("Non 2xx response")
	}

	//parse Response
	respData, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to read the response from STT Response body. Error: [%#v]", err)
		return text, err
	}
	resultTranscript, err := GetResultTextFromSTTResponse(callSID, respData)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while procesing the response. Error: [%#v] Body: [%#v]", err, string(respData))
		return text, err
	}

	// Send request and response for logging
	callStartTime := call.GetPickupTime(channelID)
	botID := call.GetCallerID(channelID).E164Format
	userID := call.GetDialedNumber(channelID).E164Format

	logsRequest := &speechlogging.LoggingRequest{}
	logsRequest.
		SetCallSID(callSID).
		SetAudio(call.GetCaptureVoiceOTP(channelID), call.GetSTTLanguage(channelID), data).
		SetLatencyMillis(latency).
		SetSttService("azure SDK").
		SetTranscript(resultTranscript).
		SetBotID(botID).
		SetUserID(userID).
		SetCallStartTime(callStartTime)

	go speechlogging.Send(logsRequest, speechlogging.URL)

	// Send accuracy to new relic
	go helper.SendAccuracyMetric("voice_azure_stt", call.GetCampaignID(channelID), resultTranscript)
	return resultTranscript, nil
}
