package yellowmessenger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
	"mime/multipart"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/helper"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/speechlogging"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

var ymSTTHTTPClient *http.Client

type SpeechToTextAPIResponse struct {
	Model_name          string `json:"model_name"`
	File_size           int64  `json:"file_size"`
	Transcript          string `json:"transcript"`
	Transcript_len      int64  `json:"transcript_len"`
}

// InitYMSTTHTTPClient initializes the bot's HTTP client
func InitYMSTTHTTPClient() {
	ymSTTHTTPClient = &http.Client{
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
	sttURL, err := url.Parse(sttEndpoint)
	if err != nil {
		return
	}
	q, err := url.ParseQuery(sttURL.RawQuery)
	if err != nil {
		return
	}
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
	// open the file
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var text string
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return text, err
	}
	if fileInfo.Size() == 0 {
		return text, errors.New("File size is zero")
	}
	// we can also try bytes
	var fileBody bytes.Buffer
	writer := multipart.NewWriter(&fileBody)

	// https://golang.org/pkg/mime/multipart/#Writer.CreateFormFile
	// must match what is expected by the receiving program
	// so, set field to "fileName" for easier life...
	filePart, err := writer.CreateFormFile("sound_file", fileInfo.Name())
		if err != nil {
			ymlogger.LogErrorf(callSID, "[YMRest] Error when creating form file. Error: [%#v]", err)
			return text, err
	}

	// since we are using mime multipart
	_, err = io.Copy(filePart, file)
	if err != nil {
		ymlogger.LogErrorf(callSID, "[YMRest] io.Copy error. Error: [%#v]", err)
		return text, err
	}

	// populate our header with simple data
	_ = writer.WriteField("title", "Marathi STT.")

	// remember to close writer
	err = writer.Close()
	if err != nil {
		ymlogger.LogErrorf(callSID, "[YMRest] Writer close error. Error: [%#v]", err)
		return text, err
	}

	sttEndpoint := getSTTEndpoint(channelID, configmanager.ConfStore.YMSTTEndpoint)
	// Prepare STT request
	sttReq, err := http.NewRequest(http.MethodPost, sttEndpoint, &fileBody)
	if err != nil {
		ymlogger.LogErrorf(callSID, "[YMRest] Error while sending post request. Error: [%#v]", err)
		return text, err
	}
	ymlogger.LogDebugf(callSID, "[YMRest] STT Language: [%s]", call.GetSTTLanguage(channelID))
	// Set Query Parameters
	q := sttReq.URL.Query()
	q.Add("language", call.GetSTTLanguage(channelID))
	sttReq.URL.RawQuery = q.Encode()
	// Set Request Headers
	sttReq.Header.Set("Content-type", writer.FormDataContentType())

	var response *http.Response
	var latency int64
	for i := 0; i < configmanager.ConfStore.ARIAPIRetry; i++ {
		// Capture API response time
		sTime := time.Now()
		ymlogger.LogDebugf(callSID, "[YMRest] Make the http request [%#v]", sTime)
		response, err = ymSTTHTTPClient.Do(sttReq)
		if response != nil {
			defer response.Body.Close()
		}
		latency = time.Since(sTime).Milliseconds()
		// Send API response time to newrelic
		go helper.SendResponseTimeMetric("voice_ym_stt", call.GetCampaignID(channelID), latency)

		if err != nil || response == nil || response.StatusCode < 200 || response.StatusCode >= 300 {
			continue
		}
		break
	}
	if err != nil {
		ymlogger.LogErrorf(callSID, "[YMRest] Error while getting the response for STT request. Error: [%#v]", err.Error())
		return text, err
	}
	if response != nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
		ymlogger.LogErrorf(callSID, "[YMRest] Non 2xx response. Error: [%#v]", response)
		return text, errors.New("Non 2xx response")
	}
	ymlogger.LogInfof(callSID, "[YMRest] Successfully received response. Response: [%#v]", response)
	
	// Read the content from the response
	respData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ymlogger.LogErrorf(callSID, "[YMRest] Failed to read the response from STT Response body. Error: [%#v]", err)
		return text, err
	}
	var sttResponse SpeechToTextAPIResponse
	err = json.Unmarshal(respData, &sttResponse)
	if err != nil {
		ymlogger.LogErrorf(callSID, "[YMRest] Error while unmarshalling the response. Error: [%#v] Body: [%#v]", err, string(respData))
		return text, err
	}

	ymlogger.LogInfof(callSID, "[YMRest] Successfully formatted response. Response: [%#v]", sttResponse)

	// Send request and response for logging
	callStartTime := call.GetPickupTime(channelID)
	botID := call.GetCallerID(channelID).E164Format
	userID := call.GetDialedNumber(channelID).E164Format

	logsRequest := &speechlogging.LoggingRequest{}
	logsRequest.
		SetCallSID(callSID).
		SetAudio(call.GetCaptureVoiceOTP(channelID), call.GetSTTLanguage(channelID), nil).
		SetLatencyMillis(latency).
		SetSttService("yellowmessenger").
		SetTranscript(sttResponse.Transcript).
		SetBotID(botID).
		SetUserID(userID).
		SetCallStartTime(callStartTime)

	go speechlogging.Send(logsRequest, speechlogging.URL)

	// Send accuracy to new relic
	go helper.SendAccuracyMetric("voice_ym_stt", call.GetCampaignID(channelID), sttResponse.Transcript)
	// Send STT Duration metric
	go helper.SendSTTDurationMetric(callSID, channelID, "voice_ym_stt", fileName, call.GetBotID(channelID), call.GetCallerID(channelID).NationalFormat)
	return sttResponse.Transcript, nil
}