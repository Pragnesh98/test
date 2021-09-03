package azure

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/newrelic"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/helper"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

var token string

var azureTTSHTTPClient *http.Client

// InitAzureTTSHTTPClient initializes the bot's HTTP client
func InitAzureTTSHTTPClient() {
	azureTTSHTTPClient = &http.Client{
		Transport: &http.Transport{
			// Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConnsPerHost:   150,
			MaxIdleConns:          150,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   3 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: 20000 * time.Millisecond,
	}
}

func GetSpeechFile(
	ctx context.Context,
	channelID string,
	callSID string,
	botID string,
	callerID string,
	text string,
	textType string,
	language string,
) (string, error) {

	// Check if the file already exists
	fileName := fmt.Sprintf("%x", md5.Sum([]byte(text)))
	filePath := configmanager.ConfStore.TTSFilePath + fileName + "_" + textType + "_azure"
	speechFile := filePath + ".alaw"
	trimmedSLNFile := filePath + "_trimmed.sln"

	if _, err := os.Stat(trimmedSLNFile); !os.IsNotExist(err) {
		ymlogger.LogInfof(callSID, "File already exists: [%#v]", trimmedSLNFile)
		return trimmedSLNFile, nil
	}

	if _, err := os.Stat(speechFile); !os.IsNotExist(err) {
		ymlogger.LogInfof(callSID, "File already exists: [%#v]", speechFile)
		return speechFile, nil
	}

	var finalResponse []byte
	botResponseText := []string{text}

	if strings.ToLower(textType) == "ssml" {
		// Check for multiple ssml texts
		SSMLList := helper.GetSSMLList(text)
		if len(SSMLList) == 0 {
			ymlogger.LogInfof(callSID, "No valid SSML text found for TTS")
			return speechFile, errors.New("Invalid SSML text")
		}
		botResponseText = SSMLList
	}

	var err error
	// Delete the file if there is any error
	defer checkErrAndDelete(callSID, speechFile, err)

	// get Speech for each SSML text
	for _, ssml := range botResponseText {
		var response *http.Response
		response, err = GetSpeech(ctx, channelID, callSID, botID, callerID, ssml, textType, language)
		if err != nil {
			ymlogger.LogErrorf(callSID, "Failed to get the response from TTS. Error: [%#v]", err)
			return speechFile, err
		}

		if response != nil {
			defer response.Body.Close()
		}
		// Read the content from the response
		var respData []byte
		respData, err = ioutil.ReadAll(response.Body)
		if err != nil {
			ymlogger.LogErrorf(callSID, "Failed to read the response from TTS Response body. Error: [%#v]", err)
			return speechFile, err
		}
		// concat speech results
		finalResponse = append(finalResponse, respData...)
	}

	// Save final audio to file
	if err = ioutil.WriteFile(speechFile, finalResponse, 0644); err != nil {
		ymlogger.LogErrorf(callSID, "Failed to write the content to the file. Error: [%#v]", err)
		return speechFile, err
	}

	if call.GetBotOptions(channelID) != nil && strings.ToLower(call.GetBotOptions(channelID).TTSQuality) == "high" {
		return speechFile, nil
	}

	// Convert audio file to sln
	var audioFile string
	audioFile, err = helper.ConvertAudioFile(ctx, speechFile)
	if err != nil {
		ymlogger.LogInfof(callSID, "Error while converting file. speechFile:[%s] audioFile:[%s]", speechFile, audioFile)
		return audioFile, err
	}

	ymlogger.LogInfof(callSID, "Trimming silence from the end of the file [%v]", audioFile)
	trimmedSLNFile, err = helper.TrimSilence(ctx, audioFile)
	if err != nil {
		ymlogger.LogInfof(callSID, "Error while trimming the file. slnFile:[%s] TrimmedFile:[%s], Error: [%#v]", audioFile, trimmedSLNFile, err)
		return audioFile, err
	}
	
	return trimmedSLNFile, nil
}

func GetSpeech(
	ctx context.Context,
	channelID string,
	callSID string,
	botID string,
	callerID string,
	text string,
	textType string,
	language string,
) (*http.Response, error) {

	// azureTTSAPIKey := configmanager.ConfStore.AzureTTSAPIKey
	// if call.GetBotOptions(channelID).UseNewMsSubscritpion {
	// 	azureTTSAPIKey = configmanager.ConfStore.AzureTTSAPIKeyNew
	// }

	// Get Token
	// token, err := GetAuthorizationToken(azureTTSAPIKey)
	// if err != nil {
	// 	ymlogger.LogErrorf(callSID, "Error while generating authorization token. Error: [%#v]", err)
	// 	return nil, err
	// }

	// Prepare TTS request
	postData := []byte(text)
	ttsReq, err := http.NewRequest(http.MethodPost, configmanager.ConfStore.AzureTTSEndpoint, bytes.NewBuffer(postData))
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while preparing TTS request. Error: [%#v]", err)
		return nil, err
	}
	var ttsQuality string
	if call.GetBotOptions(channelID) != nil {
		ttsQuality = strings.ToLower(call.GetBotOptions(channelID).TTSQuality)
	}

	switch ttsQuality {
	case "high":
		ttsReq.Header.Add("X-Microsoft-OutputFormat", "riff-16khz-16bit-mono-pcm")
	default:
		ttsReq.Header.Add("X-Microsoft-OutputFormat", "riff-8khz-8bit-mono-alaw")
	}

	ttsReq.Header.Add("Content-type", "application/ssml+xml")
	ttsReq.Header.Add("Authorization", "Bearer "+token)
	ttsReq.Header.Add("User-Agent", "GoClient")

	var response *http.Response
	var sTime time.Time
	for i := 0; i < configmanager.ConfStore.ARIAPIRetry; i++ {
		sTime = time.Now()
		// Make the http request
		response, err = azureTTSHTTPClient.Do(ttsReq)

		if err != nil || response == nil || response.StatusCode < 200 || response.StatusCode >= 300 {
			continue
		}
		break
	}
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while getting the response for TTS request. Error: [%#v]", err)
		return nil, err
	}
	if response != nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
		return nil, errors.New("Non 2xx response")
	}
	// Capture API response time
	if err := newrelic.SendCustomEvent("voice_azure_tts", map[string]interface{}{
		"campaign_id":   call.GetCampaignID(channelID),
		"response_time": time.Since(sTime).Milliseconds(),
	}); err != nil {
		ymlogger.LogErrorf("NewRelicMetric", "Failed to send voice_azure_stt metric to newrelic. Error: [%#v]", err)
	}
	// Send TTS Character metric to newrelic
	go helper.SendTTSCharactersMetric(callSID, channelID, "voice_azure_tts", text, botID, callerID)
	return response, nil
}

// RenewAzureTTSToken generates Google Token periodically
func RenewAzureTTSToken(ctx context.Context) {
	token = GetAuthorizationToken(configmanager.ConfStore.AzureTTSAPIKey)
	ticker := time.NewTicker(10 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			token = GetAuthorizationToken(configmanager.ConfStore.AzureTTSAPIKey)
		}
	}
	return
}

func GetAuthorizationToken(azureTTSAPIKey string) string {
	ymlogger.LogDebug("GenerateAzureTTSToken", "Generating the Azure TTS Token")
	postData := []byte("")
	tokenReq, err := http.NewRequest(http.MethodPost, configmanager.ConfStore.AzureTokenEndpoint, bytes.NewBuffer(postData))
	if err != nil {
		ymlogger.LogErrorf("GenerateAzureTTSToken", "Error while creating the request for generating Azure TTS Token. Error: [%#v]", err)
		return ""
	}

	tokenReq.Header.Set("Ocp-Apim-Subscription-Key", azureTTSAPIKey)
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
		ymlogger.LogErrorf("GenerateAzureTTSToken", "Error while generating Azure TTS Token. Error: [%#v]", err)
		return ""
	}
	if response != nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
		ymlogger.LogErrorf("GenerateAzureTTSToken", "Error while generating Azure TTS Token. Error: [%#v]", err)
		return ""
	}
	respData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ymlogger.LogErrorf("GenerateAzureTTSToken", "Error while reading the response for generating Azure TTS Token. Error: [%#v]", err)
		return ""
	}
	return string(respData)
}

func checkErrAndDelete(callSID string, filePath string, err error) {
	if err != nil {
		if _, err = os.Stat(filePath); !os.IsNotExist(err) {
			if err = os.Remove(filePath); err != nil {
				ymlogger.LogErrorf(callSID, "Error while removing the file. [%s]. Error: [%#v]", filePath, err)
			}
		}
	}
	return
}
